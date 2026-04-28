"""
决策器模块
实现规则引擎和LLM混合决策
"""
from typing import List, Dict, Any, Optional
from dataclasses import dataclass, field
import json

from app.models.trigger_event import TriggerEvent, DecisionType, TriggerType
from app.agents.state import PreferenceState
from app.agents.perceiver import Perception
from app.llm.manager import get_llm_manager


@dataclass
class Decision:
    """决策结果"""
    type: DecisionType
    confidence: float
    reasoning: str
    insight: str = ""
    actions: List[Dict[str, Any]] = field(default_factory=list)
    anomalies: List[Dict[str, Any]] = field(default_factory=list)
    tag_changes: List[Dict[str, Any]] = field(default_factory=list)
    recommend_rec: bool = False

    @property
    def action_needed(self) -> bool:
        return self.type != DecisionType.NO_ACTION

    def to_dict(self) -> Dict[str, Any]:
        return {
            "type": self.type.value,
            "confidence": self.confidence,
            "reasoning": self.reasoning,
            "insight": self.insight,
            "actions": self.actions,
            "anomalies": self.anomalies,
            "tag_changes": [{"tag": tc.get("tag", ""), "delta": tc.get("delta", 0)} for tc in self.tag_changes],
            "recommend_rec": self.recommend_rec,
        }


class RuleEngine:
    """规则引擎"""

    def __init__(self):
        self.rules = self._build_rules()

    def _build_rules(self) -> Dict[str, List[Dict[str, Any]]]:
        return {
            "profile_update": [
                {
                    "name": "大量游玩",
                    "condition": lambda p: p.metadata.get("playtime_delta", 0) >= 5.0,
                    "decision": DecisionType.PROFILE_UPDATE,
                    "priority": 8,
                },
                {
                    "name": "新增游戏",
                    "condition": lambda p: bool(p.metadata.get("game_name")) and
                                         p.metadata.get("playtime_delta", 0) >= 1.0,
                    "decision": DecisionType.PROFILE_UPDATE,
                    "priority": 7,
                },
                {
                    "name": "购买行为",
                    "condition": lambda p: "purchase" in p.summary.lower() or "购买" in p.summary,
                    "decision": DecisionType.PROFILE_UPDATE,
                    "priority": 9,
                },
                {
                    "name": "高折扣愿望单",
                    "condition": lambda p: p.metadata.get("discount", 0) >= 50,
                    "decision": DecisionType.PROFILE_UPDATE,
                    "priority": 7,
                },
                {
                    "name": "大量邮件同步",
                    "condition": lambda p: p.metadata.get("email_count", 0) > 5,
                    "decision": DecisionType.PROFILE_UPDATE,
                    "priority": 6,
                },
                {
                    "name": "周期性更新",
                    "condition": lambda p: p.metadata.get("days_since_last", 0) > 7,
                    "decision": DecisionType.PROFILE_UPDATE,
                    "priority": 5,
                },
            ],
            "anomaly": [
                {
                    "name": "极端游玩",
                    "condition": lambda p: p.metadata.get("is_anomaly", False),
                    "decision": DecisionType.ANOMALY_DETECTED,
                    "priority": 20,
                    "anomaly_type": "extreme_playtime",
                },
            ],
            "recommendation": [
                {
                    "name": "新促销匹配",
                    "condition": lambda p: p.metadata.get("deal_count", 0) > 0 and
                                         p.metadata.get("playtime_delta", 0) >= 2.0,
                    "decision": DecisionType.GENERATE_RECOMMENDATION,
                    "priority": 10,
                },
            ],
        }

    def evaluate(self, perception: Perception, state: PreferenceState) -> List[Decision]:
        decisions = []
        triggered_rules = []

        for category, rules in self.rules.items():
            for rule in rules:
                try:
                    if rule["condition"](perception):
                        triggered_rules.append((rule, rule["priority"]))
                except Exception:
                    pass

        triggered_rules.sort(key=lambda x: x[1], reverse=True)

        decision_types_seen = set()
        for rule, priority in triggered_rules:
            dt = rule["decision"]
            if dt in decision_types_seen:
                continue
            decision_types_seen.add(dt)

            actions = self._get_actions_for_decision(dt)
            anomalies = []

            if dt == DecisionType.ANOMALY_DETECTED:
                anomalies = [{
                    "type": rule.get("anomaly_type", "unknown"),
                    "game_id": perception.metadata.get("game_id", ""),
                    "game_name": perception.metadata.get("game_name", ""),
                    "description": f"检测到极端游玩行为: {perception.metadata.get('playtime_delta', 0):.1f}小时",
                }]

            decision = Decision(
                type=dt,
                confidence=min(0.9, 0.7 + priority * 0.02),
                reasoning=f"规则触发: {rule['name']}",
                actions=actions,
                anomalies=anomalies,
                tag_changes=self._calculate_tag_changes(perception, state),
                recommend_rec=(dt == DecisionType.GENERATE_RECOMMENDATION),
            )
            decisions.append(decision)

        return decisions

    def _calculate_tag_changes(self, perception: Perception,
                              state: PreferenceState) -> List[Dict[str, float]]:
        """计算标签权重变化"""
        changes = []

        tag_updates = perception.metadata.get("tag_weight_updates", {})
        for tag, new_weight in tag_updates.items():
            old_weight = state.get_tag_weight(tag)
            delta = new_weight - old_weight
            if abs(delta) > 0.05:
                changes.append({"tag": tag, "delta": delta})

        if perception.metadata.get("playtime_delta", 0) > 0:
            delta = perception.metadata["playtime_delta"]
            tags = perception.metadata.get("tags", [])
            genre = perception.metadata.get("genre", "")

            for tag in tags:
                if not tag:
                    continue
                old_weight = state.get_tag_weight(tag)
                new_weight = old_weight + delta * 0.1
                delta_weight = new_weight - old_weight
                if abs(delta_weight) > 0.05:
                    changes.append({"tag": tag, "delta": delta_weight})

            if genre:
                old_weight = state.get_tag_weight(genre)
                new_weight = old_weight + delta * 0.05
                delta_weight = new_weight - old_weight
                if abs(delta_weight) > 0.05:
                    changes.append({"tag": genre, "delta": delta_weight})

        tag_map = {}
        for c in changes:
            tag = c["tag"]
            if tag not in tag_map or abs(c["delta"]) > abs(tag_map[tag]["delta"]):
                tag_map[tag] = c

        return list(tag_map.values())

    def _get_actions_for_decision(self, decision: DecisionType) -> List[Dict[str, Any]]:
        actions = []
        if decision == DecisionType.PROFILE_UPDATE:
            actions.append({"type": "update_tag_weight", "params": {}})
            actions.append({"type": "log_insight", "params": {}})
        if decision == DecisionType.ANOMALY_DETECTED:
            actions.append({"type": "log_insight", "params": {"is_anomaly": True}})
        if decision == DecisionType.GENERATE_RECOMMENDATION:
            actions.append({"type": "update_tag_weight", "params": {}})
            actions.append({"type": "generate_recommendation", "params": {}})
        return actions


class LLMDecider:
    """LLM增强决策器"""

    SYSTEM_PROMPT = """你是一个游戏偏好分析专家。用户正在使用一个Steam游戏推荐平台。

你的任务是分析用户行为数据，判断是否需要更新用户的游戏偏好画像。

分析原则：
1. 关注游玩时长超过5小时的游戏类型
2. 注意用户突然开始玩的新类型（可能表示兴趣转移）
3. 识别用户反复游玩的类型（稳定偏好）
4. 忽略低于1小时的短暂游玩
5. 结合多个信号判断（游玩+购买+愿望单 > 单个信号）

请以JSON格式输出分析结果：
{
    "needs_update": true/false,
    "confidence": 0.0-1.0,
    "reasoning": "分析理由",
    "insight": "关键洞察",
    "tag_changes": [
        {"tag": "开放世界", "delta": 0.15},
        {"tag": "RPG", "delta": 0.10}
    ],
    "anomalies": [],
    "recommend_rec": true/false
}
"""

    def __init__(self):
        self.llm_manager = None

    def _get_llm(self):
        if self.llm_manager is None:
            self.llm_manager = get_llm_manager()
        return self.llm_manager

    async def decide(self, perception: Perception,
                     state: PreferenceState) -> Optional[Decision]:
        """使用LLM进行深度分析"""

        user_content = f"""## 用户行为摘要
{perception.summary}

## 关键变化
{chr(10).join(f"- {c}" for c in perception.key_changes)}

## 当前画像摘要
{state.get_profile_summary(top_n=10)}

## 行为详情
- 游戏: {perception.metadata.get('game_name', 'N/A')}
- 类型: {perception.metadata.get('genre', 'N/A')}
- 标签: {', '.join(perception.metadata.get('tags', []) or ['N/A'])}
- 新增游玩时长: {perception.metadata.get('playtime_delta', 0):.1f}小时
"""

        try:
            llm = self._get_llm()
            response = await llm.chat_with_system(
                system_prompt=self.SYSTEM_PROMPT,
                user_content=user_content,
            )

            content = response.content.strip()
            if "```json" in content:
                start = content.find("```json") + 7
                end = content.find("```", start)
                content = content[start:end].strip()
            elif "```" in content:
                start = content.find("```") + 3
                end = content.find("```", start)
                content = content[start:end].strip()

            result = json.loads(content)

            if not result.get("needs_update", False):
                return None

            tag_changes = [
                {"tag": tc["tag"], "delta": float(tc["delta"])}
                for tc in result.get("tag_changes", [])
                if tc.get("tag") and tc.get("delta", 0) != 0
            ]

            anomalies = [
                {"type": a.get("type", ""), "description": a.get("description", "")}
                for a in result.get("anomalies", [])
            ]

            insight = result.get("insight", result.get("reasoning", ""))
            reasoning = result.get("reasoning", "")

            if not insight and perception.key_changes:
                insight = f"基于分析: {perception.key_changes[0]}"

            return Decision(
                type=DecisionType.PROFILE_UPDATE if tag_changes else DecisionType.NO_ACTION,
                confidence=float(result.get("confidence", 0.7)),
                reasoning=reasoning,
                insight=insight,
                tag_changes=tag_changes,
                anomalies=anomalies,
                recommend_rec=result.get("recommend_rec", False),
                actions=[
                    {"type": "update_tag_weight", "params": {}},
                    {"type": "log_insight", "params": {}},
                ],
            )

        except Exception as e:
            import traceback
            traceback.print_exc()
            return None


class PreferenceDecider:
    """混合决策器: 规则引擎 + LLM"""

    def __init__(self):
        self.rule_engine = RuleEngine()
        self.llm_decider = LLMDecider()

    async def decide(self, perception: Perception,
                     state: PreferenceState) -> Decision:
        """决策入口: 规则优先，LLM增强"""

        rule_decisions = self.rule_engine.evaluate(perception, state)

        if rule_decisions:
            top = rule_decisions[0]
            if top.confidence >= 0.85:
                return top

        if perception.requires_deep_analysis:
            llm_decision = await self.llm_decider.decide(perception, state)
            if llm_decision and llm_decision.action_needed:
                return self._merge_decisions(rule_decisions, llm_decision)

        if rule_decisions:
            return rule_decisions[0]

        return Decision(
            type=DecisionType.NO_ACTION,
            confidence=1.0,
            reasoning="无需行动: 未检测到显著的偏好变化",
            insight="游戏偏好无显著变化",
        )

    def _merge_decisions(self, rule_decisions: List[Decision],
                         llm_decision: Decision) -> Decision:
        """合并规则和LLM决策"""
        merged = llm_decision

        existing_anomaly_types = {a["type"] for a in merged.anomalies}
        for rd in rule_decisions:
            for anomaly in rd.anomalies:
                if anomaly["type"] not in existing_anomaly_types:
                    merged.anomalies.append(anomaly)
                    existing_anomaly_types.add(anomaly["type"])

        tag_map = {tc["tag"]: tc["delta"] for tc in merged.tag_changes}
        for rd in rule_decisions:
            for tc in rd.tag_changes:
                tag = tc["tag"]
                if tag in tag_map:
                    if abs(tc["delta"]) > abs(tag_map[tag]):
                        tag_map[tag] = tc["delta"]
                else:
                    tag_map[tag] = tc["delta"]
        merged.tag_changes = [{"tag": k, "delta": v} for k, v in tag_map.items()]

        return merged
