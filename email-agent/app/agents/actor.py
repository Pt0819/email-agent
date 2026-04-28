"""
执行器模块
执行决策并与外部系统交互
"""
from typing import List, Dict, Any, Optional

from app.models.trigger_event import ActionType
from app.agents.state import PreferenceState
from app.agents.decider import Decision
from app.models.trigger_event import TriggerEvent


class PreferenceActor:
    """偏好Agent执行器"""

    ALLOWED_ACTIONS = {
        ActionType.UPDATE_PROFILE,
        ActionType.UPDATE_TAG_WEIGHT,
        ActionType.LOG_INSIGHT,
        ActionType.PUSH_NOTIFICATION,
        ActionType.GENERATE_RECOMMENDATION,
    }

    def __init__(self, db_service=None):
        self.db_service = db_service

    async def execute(self, decision: Decision,
                     state: PreferenceState,
                     event: TriggerEvent,
                     user_id: int) -> List[Dict[str, Any]]:
        """执行决策"""
        results = []

        for action_def in decision.actions:
            action_type_str = action_def.get("type", "")
            try:
                action_type = ActionType(action_type_str)
            except ValueError:
                continue

            if action_type not in self.ALLOWED_ACTIONS:
                continue

            try:
                result = await self._execute_action(
                    action_type, action_def.get("params", {}),
                    decision, state, event, user_id
                )
                results.append({
                    "action": action_type.value,
                    "success": True,
                    "result": result,
                })
                state.record_action(action_type.value, result, event)
            except Exception as e:
                results.append({
                    "action": action_type.value,
                    "success": False,
                    "error": str(e),
                })

        return results

    async def _execute_action(self, action_type: ActionType,
                              params: Dict[str, Any],
                              decision: Decision,
                              state: PreferenceState,
                              event: TriggerEvent,
                              user_id: int) -> Dict[str, Any]:
        """执行单个行动"""

        if action_type == ActionType.UPDATE_TAG_WEIGHT:
            return await self._update_tag_weights(decision.tag_changes, user_id, state)

        elif action_type == ActionType.LOG_INSIGHT:
            return self._log_insight(decision, state, event)

        elif action_type == ActionType.PUSH_NOTIFICATION:
            return self._push_notification(decision, user_id)

        elif action_type == ActionType.GENERATE_RECOMMENDATION:
            return await self._generate_recommendation(decision, user_id)

        elif action_type == ActionType.UPDATE_PROFILE:
            r1 = await self._update_tag_weights(decision.tag_changes, user_id, state)
            r2 = self._log_insight(decision, state, event)
            return {"tag_update": r1, "insight_log": r2}

        return {}

    async def _update_tag_weights(self, tag_changes: List[Dict[str, Any]],
                                  user_id: int,
                                  state: PreferenceState) -> Dict[str, Any]:
        """更新标签权重"""
        if not tag_changes:
            return {"updated": 0}

        state.update_profile(tag_changes)

        if self.db_service:
            try:
                await self.db_service.batch_update_preferences(user_id, tag_changes)
            except Exception as e:
                pass

        return {
            "updated": len(tag_changes),
            "tags": [{"tag": tc.get("tag"), "delta": tc.get("delta")} for tc in tag_changes],
        }

    def _log_insight(self, decision: Decision,
                     state: PreferenceState,
                     event: TriggerEvent) -> Dict[str, Any]:
        """记录洞察"""
        insight_text = decision.insight or decision.reasoning
        if not insight_text and decision.anomalies:
            insight_text = f"检测到异常: {decision.anomalies[0].get('description', '')}"

        if not insight_text:
            return {"logged": False}

        state.add_insight(
            insight=insight_text,
            decision_type=decision.type.value,
            reasoning=decision.reasoning,
            metadata={
                "confidence": decision.confidence,
                "event_type": event.type.value,
                "tag_changes": decision.tag_changes,
                "anomalies": decision.anomalies,
            }
        )

        return {
            "logged": True,
            "insight": insight_text,
            "decision_type": decision.type.value,
        }

    def _push_notification(self, decision: Decision,
                           user_id: int) -> Dict[str, Any]:
        """推送通知"""
        return {"pushed": False, "reason": "推送功能待实现"}

    async def _generate_recommendation(self, decision: Decision,
                                       user_id: int) -> Dict[str, Any]:
        """生成推荐"""
        try:
            from app.llm import get_llm_manager
            from app.services.recommendation_service import RecommendationReasonGenerator

            llm_manager = get_llm_manager()
            generator = RecommendationReasonGenerator(llm_manager)

            # 从决策中获取推荐上下文
            params = decision.reasoning or ""
            game_info = decision.metadata.get("game", {}) if decision.metadata else {}

            if game_info:
                reason = await generator.generate_reason(
                    game_name=game_info.get("game_name", ""),
                    game_genre=game_info.get("game_genre", ""),
                    game_tags=game_info.get("game_tags", []),
                    user_preferences=decision.metadata.get("user_preferences", []),
                )
                return {
                    "generated": True,
                    "game_name": game_info.get("game_name", ""),
                    "reason": reason,
                }

            return {"generated": False, "reason": "缺少游戏信息"}

        except Exception as e:
            return {"generated": False, "reason": f"推荐生成失败: {e}"}
