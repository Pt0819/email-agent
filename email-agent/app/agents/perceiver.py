"""
感知器模块
实现数据变化的感知和预处理
"""
from abc import ABC, abstractmethod
from typing import Dict, List, Any, Optional

from app.models.trigger_event import TriggerEvent, TriggerType
from app.agents.state import PreferenceState
from app.agents.base import parse_tags_json, calculate_playtime_hours, format_playtime_description


class Perception:
    """感知结果"""

    def __init__(self, summary: str, key_changes: List[str],
                 metadata: Dict[str, Any], requires_deep_analysis: bool = False):
        self.summary = summary
        self.key_changes = key_changes
        self.metadata = metadata
        self.requires_deep_analysis = requires_deep_analysis

    def to_dict(self) -> Dict[str, Any]:
        return {
            "summary": self.summary,
            "key_changes": self.key_changes,
            "metadata": self.metadata,
            "requires_deep_analysis": self.requires_deep_analysis,
        }


class PreferencePerceiver(ABC):
    """感知器基类"""

    @abstractmethod
    def perceive(self, event: TriggerEvent,
                 current_state: PreferenceState) -> Optional[Perception]:
        """处理事件并生成感知"""
        pass

    def _detect_weight_changes(self, old_profile: Dict[str, float],
                                 new_tags: Dict[str, float],
                                 threshold: float = 0.2) -> List[str]:
        """检测权重变化"""
        changes = []
        for tag, new_weight in new_tags.items():
            old_weight = old_profile.get(tag, 0)
            delta = new_weight - old_weight
            if abs(delta) > threshold:
                changes.append(f"标签【{tag}】: {old_weight:.2f} → {new_weight:.2f} ({delta:+.2f})")
        return changes


class PlaytimePerceiver(PreferencePerceiver):
    """游玩时长感知器"""

    SIGNIFICANT_PLAYTIME_HOURS = 5.0
    DEEP_ANALYSIS_HOURS = 10.0
    EXTREME_PLAYTIME_HOURS = 50.0

    def perceive(self, event: TriggerEvent,
                 current_state: PreferenceState) -> Optional[Perception]:
        if event.type != TriggerType.PLAYTIME_UPDATE:
            return None

        data = event.data
        game_id = data.get("game_id", "")
        game_name = data.get("game_name", "未知游戏")
        delta_minutes = int(data.get("playtime_delta", 0))
        total_minutes = int(data.get("total_playtime", 0))
        genre = data.get("genre", "")
        tags = data.get("tags", [])
        if isinstance(tags, str):
            tags = parse_tags_json(tags)

        delta_hours = calculate_playtime_hours(delta_minutes)
        total_hours = calculate_playtime_hours(total_minutes)

        key_changes = []
        requires_analysis = delta_hours >= self.DEEP_ANALYSIS_HOURS
        is_anomaly = delta_hours >= self.EXTREME_PLAYTIME_HOURS

        if delta_hours > 0:
            if delta_hours >= 1:
                desc = f"游玩了《{game_name}》{format_playtime_description(delta_hours)}"
            else:
                desc = f"游玩了《{game_name}》{int(delta_minutes)}分钟"
            key_changes.append(desc)

        if is_anomaly:
            key_changes.append(f"⚠️ 检测到极端游玩行为: {format_playtime_description(delta_hours)}")

        if genre and delta_hours >= self.SIGNIFICANT_PLAYTIME_HOURS:
            current_genre_weight = current_state.get_tag_weight(genre)
            new_genre_weight = current_genre_weight + delta_hours * 0.05
            if new_genre_weight - current_genre_weight > 0.1:
                key_changes.append(f"类型【{genre}】偏好增强")

        tag_weight_updates = {}
        for tag in tags:
            if not tag:
                continue
            old_weight = current_state.get_tag_weight(tag)
            new_weight = old_weight + delta_hours * 0.1
            tag_weight_updates[tag] = new_weight

        if tag_weight_updates and delta_hours >= self.SIGNIFICANT_PLAYTIME_HOURS:
            changes = self._detect_weight_changes(
                current_state.current_profile, tag_weight_updates
            )
            if changes:
                key_changes.extend(changes[:5])

        summary = f"用户游玩了《{game_name}》，新增{format_playtime_description(delta_hours)}"
        if genre:
            summary += f"，类型: {genre}"
        if tags:
            summary += f"，标签: {', '.join(tags[:3])}"

        return Perception(
            summary=summary,
            key_changes=key_changes,
            metadata={
                "game_id": game_id,
                "game_name": game_name,
                "genre": genre,
                "tags": tags,
                "playtime_delta": delta_hours,
                "total_playtime": total_hours,
                "is_anomaly": is_anomaly,
                "tag_weight_updates": tag_weight_updates,
            },
            requires_deep_analysis=requires_analysis
        )


class SteamEmailPerceiver(PreferencePerceiver):
    """Steam邮件变化感知器"""

    def perceive(self, event: TriggerEvent,
                 current_state: PreferenceState) -> Optional[Perception]:
        if event.type not in (TriggerType.STEAM_EMAIL_SYNC, TriggerType.NEW_GAME_ADDED,
                               TriggerType.GAME_PURCHASED, TriggerType.GAME_WISHLISTED,
                               TriggerType.USER_FEEDBACK):
            return None

        data = event.data
        key_changes = []
        requires_analysis = False
        metadata = {}

        if event.type == TriggerType.STEAM_EMAIL_SYNC:
            email_count = data.get("email_count", 0)
            game_count = data.get("game_count", 0)
            deal_count = data.get("deal_count", 0)
            summary = f"Steam邮件同步完成: {email_count}封邮件"
            if game_count > 0:
                summary += f", {game_count}个游戏"
            if deal_count > 0:
                summary += f", {deal_count}个促销"
            if game_count > 5:
                key_changes.append(f"发现{game_count}个新游戏信息")
                requires_analysis = True
            metadata = {"email_count": email_count, "game_count": game_count, "deal_count": deal_count}

        elif event.type == TriggerType.NEW_GAME_ADDED:
            game_name = data.get("game_name", "未知游戏")
            genre = data.get("genre", "")
            tags = data.get("tags", [])
            if isinstance(tags, str):
                tags = parse_tags_json(tags)
            summary = f"新增游戏: 《{game_name}》"
            if genre:
                summary += f" ({genre})"
            key_changes.append(f"用户添加了新游戏: 《{game_name}》")
            if genre:
                key_changes.append(f"新类型: 【{genre}】")
            requires_analysis = True
            metadata = {"game_name": game_name, "genre": genre, "tags": tags}

        elif event.type == TriggerType.GAME_PURCHASED:
            game_name = data.get("game_name", "未知游戏")
            genre = data.get("genre", "")
            price = data.get("price", 0)
            tags = data.get("tags", [])
            if isinstance(tags, str):
                tags = parse_tags_json(tags)
            summary = f"用户购买了《{game_name}》"
            if price > 0:
                summary += f"，价格: ¥{price:.2f}"
            key_changes.append(f"购买行为: 《{game_name}》")
            requires_analysis = True
            metadata = {"game_name": game_name, "genre": genre, "tags": tags, "price": price}

        elif event.type == TriggerType.GAME_WISHLISTED:
            game_name = data.get("game_name", "未知游戏")
            discount = data.get("discount", 0)
            summary = f"用户将《{game_name}》加入愿望单"
            if discount > 0:
                summary += f"（{discount}%折扣）"
            key_changes.append(f"愿望单: 《{game_name}》")
            if discount >= 50:
                key_changes.append(f"高折扣游戏加入愿望单: {discount}%")
                requires_analysis = True
            metadata = {"game_name": game_name, "discount": discount}

        elif event.type == TriggerType.USER_FEEDBACK:
            game_name = data.get("game_name", "未知游戏")
            action = data.get("action", "")
            summary = f"用户对《{game_name}》的反馈: {action}"
            key_changes.append(f"反馈行为: 《{game_name}》- {action}")
            requires_analysis = True
            metadata = {"game_name": game_name, "action": action}

        return Perception(
            summary=summary,
            key_changes=key_changes,
            metadata=metadata,
            requires_deep_analysis=requires_analysis
        )


class PeriodicPerceiver(PreferencePerceiver):
    """定时检查感知器"""

    def perceive(self, event: TriggerEvent,
                 current_state: PreferenceState) -> Optional[Perception]:
        if event.type != TriggerType.PERIODIC_CHECK:
            return None

        data = event.data
        days_since = data.get("days_since_last", 0)

        summary = "定时检查: 偏好画像例行分析"
        key_changes = []
        requires_analysis = False

        if days_since > 7:
            key_changes.append(f"已有{days_since}天未更新画像")
            requires_analysis = True

        if current_state.anomaly_flags:
            key_changes.append(f"存在{len(current_state.anomaly_flags)}个待处理异常")
            requires_analysis = True

        return Perception(
            summary=summary,
            key_changes=key_changes,
            metadata={"days_since_last": days_since},
            requires_deep_analysis=requires_analysis
        )


class CompositePerceiver:
    """组合感知器"""

    def __init__(self):
        self.perceivers: Dict[TriggerType, PreferencePerceiver] = {
            TriggerType.PLAYTIME_UPDATE: PlaytimePerceiver(),
            TriggerType.STEAM_EMAIL_SYNC: SteamEmailPerceiver(),
            TriggerType.LIBRARY_SYNC: SteamEmailPerceiver(),
            TriggerType.NEW_GAME_ADDED: SteamEmailPerceiver(),
            TriggerType.GAME_PURCHASED: SteamEmailPerceiver(),
            TriggerType.GAME_WISHLISTED: SteamEmailPerceiver(),
            TriggerType.USER_FEEDBACK: SteamEmailPerceiver(),
            TriggerType.PERIODIC_CHECK: PeriodicPerceiver(),
        }

    def perceive(self, event: TriggerEvent,
                 current_state: PreferenceState) -> Optional[Perception]:
        perceiver = self.perceivers.get(event.type)
        if not perceiver:
            return None
        return perceiver.perceive(event, current_state)
