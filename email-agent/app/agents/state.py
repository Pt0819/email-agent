"""
偏好Agent状态管理
维护用户画像快照、变化历史和洞察记录
"""
from typing import Dict, List, Optional, Any
from datetime import datetime

from app.models.trigger_event import TriggerEvent


class PreferenceState:
    """偏好Agent状态"""

    def __init__(self, user_id: int):
        self.user_id = user_id
        self.current_profile: Dict[str, float] = {}  # tag -> weight
        self.change_history: List[Dict[str, Any]] = []
        self.insights: List[Dict[str, Any]] = []
        self.last_update: Optional[datetime] = None
        self.anomaly_flags: List[str] = []
        self._loaded = False

    def load_profile(self, preferences: List[Dict[str, Any]]):
        """加载用户画像"""
        self.current_profile = {}
        for pref in preferences:
            tag = pref.get("tag", "")
            weight = float(pref.get("weight", 1.0))
            if tag:
                self.current_profile[tag] = weight
        self._loaded = True

    def is_loaded(self) -> bool:
        """是否已加载"""
        return self._loaded

    def get_tag_weight(self, tag: str) -> float:
        """获取标签权重"""
        normalized = tag.lower().strip()
        if normalized in self.current_profile:
            return self.current_profile[normalized]
        for k, v in self.current_profile.items():
            if k.lower() == normalized:
                return v
        return 0.0

    def update_profile(self, tag_changes: List[Dict[str, Any]]):
        """更新画像"""
        for change in tag_changes:
            tag = change.get("tag", "")
            delta = float(change.get("delta", 0))
            if not tag:
                continue

            normalized = tag.lower().strip()
            old_weight = self.get_tag_weight(normalized)
            new_weight = max(0, old_weight + delta)
            self.current_profile[normalized] = new_weight

        self.last_update = datetime.now()
        self._trim_history()

    def record_action(self, action_type: str, result: Dict[str, Any],
                      event: Optional[TriggerEvent] = None):
        """记录行动"""
        record = {
            "type": action_type,
            "result": result,
            "timestamp": datetime.now().isoformat(),
        }
        if event:
            record["event_type"] = event.type.value
            record["event_data"] = event.data
        self.change_history.append(record)

    def add_insight(self, insight: str, decision_type: str,
                    reasoning: str = "", metadata: Optional[Dict[str, Any]] = None):
        """添加洞察记录"""
        record = {
            "insight": insight,
            "decision_type": decision_type,
            "reasoning": reasoning,
            "timestamp": datetime.now().isoformat(),
        }
        if metadata:
            record["metadata"] = metadata
        self.insights.append(record)
        if len(self.insights) > 50:
            self.insights = self.insights[-50:]

    def record_anomaly(self, anomaly_type: str, description: str,
                       game_id: str = "", game_name: str = ""):
        """记录异常"""
        self.anomaly_flags.append(anomaly_type)
        self.add_insight(
            insight=f"异常检测: {description}",
            decision_type="anomaly_detected",
            reasoning=f"检测到{anomaly_type}类型的异常行为",
            metadata={
                "anomaly_type": anomaly_type,
                "game_id": game_id,
                "game_name": game_name,
            }
        )

    def get_profile_summary(self, top_n: int = 10) -> str:
        """获取画像摘要（用于LLM）"""
        if not self.current_profile:
            return "用户暂无偏好数据"
        sorted_tags = sorted(
            self.current_profile.items(),
            key=lambda x: x[1],
            reverse=True
        )[:top_n]
        lines = [f"- {tag}: {weight:.2f}" for tag, weight in sorted_tags]
        return "\n".join(lines) if lines else "用户暂无显著偏好"

    def get_top_tags(self, top_n: int = 15) -> List[Dict[str, Any]]:
        """获取Top标签"""
        if not self.current_profile:
            return []
        sorted_tags = sorted(
            self.current_profile.items(),
            key=lambda x: x[1],
            reverse=True
        )[:top_n]
        return [
            {"tag": tag, "weight": round(weight, 2), "source": "system"}
            for tag, weight in sorted_tags
            if weight > 0.1
        ]

    def has_recent_update(self, hours: int = 2) -> bool:
        """检查是否在指定小时内有更新"""
        if not self.last_update:
            return False
        elapsed = (datetime.now() - self.last_update).total_seconds()
        return elapsed < hours * 3600

    def _trim_history(self):
        """清理历史记录"""
        if len(self.change_history) > 100:
            self.change_history = self.change_history[-100:]

    def to_dict(self) -> Dict[str, Any]:
        """序列化"""
        return {
            "user_id": self.user_id,
            "current_profile": self.current_profile,
            "last_update": self.last_update.isoformat() if self.last_update else None,
            "anomaly_flags": self.anomaly_flags,
            "history_count": len(self.change_history),
            "insights_count": len(self.insights),
        }
