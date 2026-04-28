"""
偏好分析Agent
自主感知-决策-行动闭环
"""
from typing import List, Dict, Any, Optional
from datetime import datetime

from app.models.trigger_event import TriggerEvent, TriggerType
from app.agents.state import PreferenceState
from app.agents.perceiver import CompositePerceiver
from app.agents.decider import PreferenceDecider
from app.agents.actor import PreferenceActor
from app.agents.base import parse_tags_json, calculate_playtime_hours


class PreferenceAgent:
    """
    自主偏好分析Agent
    持续观察数据变化，自主决策并执行行动

    架构: Perceive -> Decide -> Act 闭环
    """

    def __init__(self, db_service=None):
        self.db_service = db_service
        self.state: Optional[PreferenceState] = None
        self.perceiver = CompositePerceiver()
        self.decider = PreferenceDecider()
        self.actor = PreferenceActor(db_service=db_service)

    async def initialize(self, user_id: int, current_preferences: List[Dict[str, Any]]):
        """初始化Agent状态"""
        self.state = PreferenceState(user_id)
        if current_preferences:
            self.state.load_profile(current_preferences)

    async def on_event(self, event: TriggerEvent) -> Optional[Dict[str, Any]]:
        """事件入口 - 执行感知-决策-行动闭环"""
        if not self.state:
            return None

        if event.user_id != self.state.user_id:
            return None

        # 1. 感知
        perception = self.perceiver.perceive(event, self.state)
        if not perception:
            return None

        # 2. 决策
        decision = await self.decider.decide(perception, self.state)

        if not decision.action_needed:
            return {
                "perception": perception.summary,
                "decision": decision.type.value,
                "reasoning": decision.reasoning,
                "action_needed": False,
            }

        # 3. 执行
        results = await self.actor.execute(
            decision, self.state, event, self.state.user_id
        )

        # 4. 记录洞察
        if decision.insight:
            self.state.add_insight(
                insight=decision.insight,
                decision_type=decision.type.value,
                reasoning=decision.reasoning,
            )

        return {
            "perception": perception.summary,
            "key_changes": perception.key_changes,
            "decision": decision.type.value,
            "confidence": decision.confidence,
            "reasoning": decision.reasoning,
            "insight": decision.insight,
            "actions_executed": len(results),
            "action_results": results,
            "anomalies": decision.anomalies,
            "tag_changes_count": len(decision.tag_changes),
        }

    async def analyze_full(self, game_library: List[Dict[str, Any]],
                          trigger_type: str = "manual_trigger") -> Dict[str, Any]:
        """全量分析 - 基于游戏库数据"""
        if not self.state:
            return {"success": False, "error": "Agent未初始化"}

        all_insights = []
        all_anomalies = []

        sorted_games = sorted(
            game_library,
            key=lambda g: g.get("playtime_2_weeks", g.get("playtime", 0)),
            reverse=True
        )

        for game in sorted_games[:30]:
            game_id = game.get("game_id", "")
            game_name = game.get("game_name", "未知")
            playtime = int(game.get("playtime", 0))
            playtime_2w = int(game.get("playtime_2_weeks", 0))
            genre = game.get("genre", "")
            tags_str = game.get("tags", "")
            tags = parse_tags_json(tags_str)

            event = TriggerEvent(
                type=TriggerType.LIBRARY_SYNC,
                timestamp=datetime.now(),
                user_id=self.state.user_id,
                data={
                    "game_id": game_id,
                    "game_name": game_name,
                    "playtime_delta": calculate_playtime_hours(playtime_2w),
                    "total_playtime": calculate_playtime_hours(playtime),
                    "genre": genre,
                    "tags": tags,
                }
            )

            result = await self.on_event(event)
            if result and result.get("action_needed") is not False:
                if result.get("insight"):
                    all_insights.append(result["insight"])
                if result.get("anomalies"):
                    all_anomalies.extend(result["anomalies"])

        seen = set()
        unique_insights = []
        for insight in all_insights:
            if insight not in seen:
                seen.add(insight)
                unique_insights.append(insight)

        updated_tags = [
            {"tag": tag, "weight": weight, "source": "system"}
            for tag, weight in self.state.current_profile.items()
            if weight > 0.1
        ]

        new_tags = [t for t in updated_tags if t["weight"] < 0.5]
        stable_tags = [t for t in updated_tags if t["weight"] >= 0.5]

        return {
            "success": True,
            "new_tags": new_tags,
            "updated_tags": stable_tags,
            "insights": unique_insights[:10],
            "anomalies": all_anomalies[:5],
            "recommend_rec": len(all_anomalies) > 0 or len(unique_insights) > 0,
            "profile_summary": self.state.get_profile_summary(top_n=15),
        }

    def get_insights(self, limit: int = 10) -> List[Dict[str, Any]]:
        """获取近期洞察"""
        if not self.state:
            return []
        return self.state.insights[-limit:]

    def get_profile(self) -> Dict[str, Any]:
        """获取当前画像"""
        if not self.state:
            return {}
        return {
            "top_tags": self.state.get_top_tags(15),
            "profile": self.state.current_profile,
            "last_update": self.state.last_update.isoformat() if self.state.last_update else None,
            "anomaly_count": len(self.state.anomaly_flags),
        }
