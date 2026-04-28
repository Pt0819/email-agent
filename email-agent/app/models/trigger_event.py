"""
触发事件数据模型
定义Agent感知-决策-行动闭环中的触发事件
"""
from enum import Enum
from dataclasses import dataclass, field, asdict
from typing import Optional, List, Dict, Any
from datetime import datetime


class TriggerType(str, Enum):
    """触发事件类型"""
    # Steam相关
    STEAM_EMAIL_SYNC = "steam_email_sync"    # Steam邮件同步完成
    LIBRARY_SYNC = "library_sync"           # 游戏库同步完成
    PLAYTIME_UPDATE = "playtime_update"     # 游玩时长更新
    NEW_GAME_ADDED = "new_game_added"       # 新增游戏

    # 用户交互
    USER_FEEDBACK = "user_feedback"         # 用户反馈
    GAME_PURCHASED = "game_purchased"       # 用户购买游戏
    GAME_WISHLISTED = "game_wishlisted"     # 添加愿望单

    # 系统任务
    PERIODIC_CHECK = "periodic_check"      # 定时检查
    MANUAL_TRIGGER = "manual_trigger"        # 手动触发

    def __str__(self) -> str:
        return self.value


class DecisionType(str, Enum):
    """决策类型"""
    NO_ACTION = "no_action"                    # 无需行动
    PROFILE_UPDATE = "profile_update"          # 更新画像
    TAG_WEIGHT_ADJUST = "tag_weight_adjust"   # 调整标签权重
    ANOMALY_DETECTED = "anomaly_detected"     # 异常检测
    PREFERENCE_DRIFT = "preference_drift"     # 偏好漂移
    NEW_PATTERN = "new_pattern"               # 新模式识别
    PUSH_NOTIFICATION = "push_notification"   # 推送通知
    GENERATE_RECOMMENDATION = "generate_rec"  # 生成推荐
    REQUEST_CONFIRMATION = "request_confirm"  # 请求确认

    def __str__(self) -> str:
        return self.value


class ActionType(str, Enum):
    """行动类型"""
    UPDATE_PROFILE = "update_profile"           # 更新画像
    UPDATE_TAG_WEIGHT = "update_tag_weight"   # 更新标签权重
    LOG_INSIGHT = "log_insight"              # 记录洞察
    PUSH_NOTIFICATION = "push_notification"  # 推送通知
    GENERATE_RECOMMENDATION = "generate_rec"  # 生成推荐
    TRIGGER_SYNC = "trigger_sync"            # 触发同步

    def __str__(self) -> str:
        return self.value


@dataclass
class TriggerEvent:
    """触发事件"""
    type: TriggerType
    timestamp: datetime
    user_id: int
    data: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """转换为字典"""
        return {
            "type": self.type.value,
            "timestamp": self.timestamp.isoformat(),
            "user_id": self.user_id,
            "data": self.data,
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "TriggerEvent":
        """从字典创建"""
        return cls(
            type=TriggerType(data["type"]),
            timestamp=datetime.fromisoformat(data["timestamp"]),
            user_id=int(data["user_id"]),
            data=data.get("data", {}),
        )

    # ---- 便捷工厂方法 ----

    @classmethod
    def steam_email_sync(cls, user_id: int, email_count: int = 0,
                          game_count: int = 0, deal_count: int = 0) -> "TriggerEvent":
        """Steam邮件同步事件"""
        return cls(
            type=TriggerType.STEAM_EMAIL_SYNC,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "email_count": email_count,
                "game_count": game_count,
                "deal_count": deal_count,
            }
        )

    @classmethod
    def library_sync(cls, user_id: int, total_games: int,
                      new_games: int = 0, total_playtime: int = 0) -> "TriggerEvent":
        """游戏库同步事件"""
        return cls(
            type=TriggerType.LIBRARY_SYNC,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "total_games": total_games,
                "new_games": new_games,
                "total_playtime": total_playtime,
            }
        )

    @classmethod
    def playtime_update(cls, user_id: int, game_id: str,
                          game_name: str = "", playtime_delta: int = 0,
                          total_playtime: int = 0, genre: str = "",
                          tags: Optional[List[str]] = None) -> "TriggerEvent":
        """游玩时长更新事件"""
        return cls(
            type=TriggerType.PLAYTIME_UPDATE,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "game_id": game_id,
                "game_name": game_name,
                "playtime_delta": playtime_delta,
                "total_playtime": total_playtime,
                "genre": genre,
                "tags": tags or [],
            }
        )

    @classmethod
    def new_game_added(cls, user_id: int, game_id: str,
                        game_name: str = "", genre: str = "",
                        tags: Optional[List[str]] = None) -> "TriggerEvent":
        """新增游戏事件"""
        return cls(
            type=TriggerType.NEW_GAME_ADDED,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "game_id": game_id,
                "game_name": game_name,
                "genre": genre,
                "tags": tags or [],
            }
        )

    @classmethod
    def user_feedback(cls, user_id: int, game_id: str,
                        game_name: str = "", action: str = "",
                        deal_id: Optional[int] = None) -> "TriggerEvent":
        """用户反馈事件"""
        return cls(
            type=TriggerType.USER_FEEDBACK,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "game_id": game_id,
                "game_name": game_name,
                "action": action,
                "deal_id": deal_id,
            }
        )

    @classmethod
    def game_purchased(cls, user_id: int, game_id: str,
                         game_name: str = "", genre: str = "",
                         tags: Optional[List[str]] = None,
                         price: float = 0) -> "TriggerEvent":
        """游戏购买事件"""
        return cls(
            type=TriggerType.GAME_PURCHASED,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "game_id": game_id,
                "game_name": game_name,
                "genre": genre,
                "tags": tags or [],
                "price": price,
            }
        )

    @classmethod
    def game_wishlisted(cls, user_id: int, game_id: str,
                          game_name: str = "", genre: str = "",
                          tags: Optional[List[str]] = None,
                          discount: int = 0) -> "TriggerEvent":
        """添加愿望单事件"""
        return cls(
            type=TriggerType.GAME_WISHLISTED,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "game_id": game_id,
                "game_name": game_name,
                "genre": genre,
                "tags": tags or [],
                "discount": discount,
            }
        )

    @classmethod
    def periodic_check(cls, user_id: int, days_since_last: int = 0) -> "TriggerEvent":
        """定时检查事件"""
        return cls(
            type=TriggerType.PERIODIC_CHECK,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "days_since_last": days_since_last,
            }
        )

    @classmethod
    def manual_trigger(cls, user_id: int, reason: str = "") -> "TriggerEvent":
        """手动触发事件"""
        return cls(
            type=TriggerType.MANUAL_TRIGGER,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "reason": reason,
            }
        )
