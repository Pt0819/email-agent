"""
推荐理由生成模块
基于用户偏好和游戏信息，使用LLM生成个性化推荐理由
"""
from typing import List, Dict, Any, Optional
from loguru import logger


class RecommendationReasonGenerator:
    """推荐理由生成器"""

    SYSTEM_PROMPT = """你是一个专业的Steam游戏推荐助手。

你的任务是根据用户的历史偏好和游戏信息，生成简短、有说服力的推荐理由。

要求：
1. 推荐理由应该自然流畅，像朋友推荐一样
2. 结合用户的历史游玩偏好和游戏特点
3. 突出游戏与用户偏好的匹配点
4. 理由应该简洁有力，控制在50字以内
5. 不要使用模板化的语言

示例：
- "你最近在玩很多RPG游戏，这款《艾尔登法环》融合了《黑魂》系列的精髓，值得一试。"
- "基于你喜欢的策略游戏，《文明 VI》提供了数百小时的深度体验。"
- "这款游戏和《巫师3》有相似的剧情深度，如果你喜欢开放世界RPG不要错过。"

直接输出推荐理由，不要解释。"""

    def __init__(self, llm_manager):
        self.llm_manager = llm_manager

    async def generate_reason(
        self,
        game_name: str,
        game_genre: str,
        game_tags: List[str],
        user_preferences: List[Dict[str, Any]],
        recent_games: List[Dict[str, Any]] = None,
    ) -> str:
        """
        生成推荐理由

        Args:
            game_name: 游戏名称
            game_genre: 游戏类型
            game_tags: 游戏标签列表
            user_preferences: 用户偏好标签及权重
            recent_games: 用户最近游玩的游戏

        Returns:
            推荐理由字符串
        """
        if not self.llm_manager:
            return self._generate_fallback_reason(game_name, game_genre, user_preferences)

        try:
            # 构建用户偏好描述
            top_prefs = self._format_preferences(user_preferences)
            recent_str = self._format_recent_games(recent_games) if recent_games else ""
            tags_str = ", ".join(game_tags[:5]) if game_tags else game_genre

            user_message = f"""请为以下游戏生成推荐理由：

游戏信息：
- 名称：{game_name}
- 类型：{game_genre}
- 标签：{tags_str}

用户偏好：
{top_prefs}

{f"用户最近在玩：{recent_str}" if recent_str else ""}

直接输出推荐理由，不要解释。"""

            from app.llm.provider import Message

            messages = [
                Message(role="system", content=self.SYSTEM_PROMPT),
                Message(role="user", content=user_message),
            ]

            response = await self.llm_manager.chat(messages)

            if response and response.content:
                reason = response.content.strip()
                # 确保理由不为空
                if reason:
                    return reason

            return self._generate_fallback_reason(game_name, game_genre, user_preferences)

        except Exception as e:
            logger.error(f"生成推荐理由失败: {e}")
            return self._generate_fallback_reason(game_name, game_genre, user_preferences)

    def _format_preferences(self, preferences: List[Dict[str, Any]]) -> str:
        """格式化用户偏好"""
        if not preferences:
            return "暂无明确的偏好标签"

        lines = []
        for i, pref in enumerate(preferences[:5], 1):
            tag = pref.get("tag", "未知")
            weight = pref.get("weight", 1.0)
            source = pref.get("source", "未知")
            weight_bar = "★" * int(weight / 2) if weight else ""
            lines.append(f"{i}. {tag} {weight_bar} (来源: {source})")

        return "\n".join(lines)

    def _format_recent_games(self, games: List[Dict[str, Any]]) -> str:
        """格式化最近游玩"""
        if not games:
            return ""

        names = [g.get("game_name", g.get("name", "未知游戏")) for g in games[:3]]
        return "、".join(names)

    def _generate_fallback_reason(
        self,
        game_name: str,
        game_genre: str,
        preferences: List[Dict[str, Any]]
    ) -> str:
        """生成备用推荐理由（当LLM不可用时）"""
        if not preferences:
            return f"热门{game_genre}游戏，值得一试"

        # 找到最匹配的偏好
        top_pref = preferences[0].get("tag", "") if preferences else ""
        if top_pref and game_genre:
            return f"基于你偏好的{top_pref}游戏，这是款不错的{game_genre}游戏"
        elif game_genre:
            return f"这款{game_genre}游戏与你的偏好匹配"
        else:
            return f"{game_name}是热门游戏推荐"


class RecommendationMatcher:
    """推荐匹配器 - 计算游戏与用户偏好的匹配度"""

    # 标签同义词映射
    TAG_SYNONYMS = {
        "rpg": ["role playing", "jrpg", "arpg", "action rpg"],
        "fps": ["first person", "shooter", "tactical shooter"],
        "act": ["action", "beat 'em up", "hack and slash"],
        "strategy": ["strategic", "turn based", "real-time strategy"],
        "simulation": ["sim", "management", "life sim"],
        "adventure": ["adventure game", "point and click"],
        "puzzle": ["puzzle game", "logic"],
        "horror": ["survival horror", "psychological horror"],
    }

    def __init__(self):
        pass

    def calculate_match_score(
        self,
        game_tags: List[str],
        game_genre: str,
        user_preferences: List[Dict[str, Any]]
    ) -> tuple[float, List[str]]:
        """
        计算游戏与用户偏好的匹配度

        Returns:
            (匹配度分数 0-100, 匹配理由列表)
        """
        if not user_preferences:
            return 50.0, ["根据你的游戏库推荐"]

        game_tags_lower = [t.lower() for t in game_tags]
        game_genre_lower = game_genre.lower() if game_genre else ""

        total_weight = 0.0
        max_possible = 0.0
        matches = []

        for pref in user_preferences:
            pref_tag = pref.get("tag", "").lower()
            weight = pref.get("weight", 1.0)
            max_possible += weight * 2

            # 直接标签匹配
            for game_tag in game_tags_lower:
                if pref_tag == game_tag or self._tag_contains(pref_tag, game_tag):
                    total_weight += weight * 2
                    matches.append(f"包含\"{pref.get('tag')}\"标签")
                    break

            # 类型匹配
            if game_genre_lower and (pref_tag in game_genre_lower or self._tag_contains(pref_tag, game_genre_lower)):
                total_weight += weight * 1.5
                matches.append(f"你常玩{pref.get('tag')}类型")

        # 计算分数
        score = 50.0
        if max_possible > 0:
            score = (total_weight / max_possible) * 50
        score += 50  # 基础分
        score = min(100.0, max(0.0, score))

        # 去重匹配理由
        seen = set()
        unique_matches = []
        for m in matches:
            if m not in seen:
                seen.add(m)
                unique_matches.append(m)

        if not unique_matches:
            unique_matches = ["热门游戏推荐"]

        return score, unique_matches[:3]  # 最多3个匹配理由

    def _tag_contains(self, keyword: str, tag: str) -> bool:
        """检查标签是否包含关键字"""
        # 检查同义词
        for key, synonyms in self.TAG_SYNONYMS.items():
            if keyword == key:
                for syn in synonyms:
                    if syn in tag or key in tag:
                        return True
        return keyword in tag or tag in keyword
