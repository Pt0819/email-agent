"""
偏好分析API路由
"""
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import List, Optional
from loguru import logger

from app.agents.preference_agent import PreferenceAgent
from app.agents.base import parse_tags_json


router = APIRouter(prefix="/preference", tags=["偏好分析"])


# ---- Request/Response Schemas ----

class TagPreferenceData(BaseModel):
    tag: str
    weight: float
    source: str = "system"


class LibraryGameData(BaseModel):
    game_id: str
    game_name: str
    playtime: int = 0
    playtime_2_weeks: int = 0
    genre: str = ""
    tags: str = ""
    last_played_at: Optional[str] = None


class PreferenceAnalyzeRequest(BaseModel):
    user_id: int
    game_library: List[LibraryGameData] = []
    current_preferences: List[TagPreferenceData] = []
    trigger_type: str = "manual_trigger"


class TagChangeData(BaseModel):
    tag: str
    delta: float


class AnomalyData(BaseModel):
    type: str
    description: str
    game_id: Optional[str] = None
    game_name: Optional[str] = None


class PreferenceAnalyzeResponse(BaseModel):
    success: bool
    new_tags: List[TagPreferenceData] = []
    updated_tags: List[TagPreferenceData] = []
    insights: List[str] = []
    reasoning: str = ""
    anomalies: List[AnomalyData] = []
    recommend_rec: bool = False


# ---- Agent缓存 ----

_agent_cache = {}


def get_agent(user_id: int) -> PreferenceAgent:
    if user_id not in _agent_cache:
        _agent_cache[user_id] = PreferenceAgent()
    return _agent_cache[user_id]


# ---- API Endpoints ----

@router.post("/analyze", response_model=PreferenceAnalyzeResponse)
async def analyze_preferences(request: PreferenceAnalyzeRequest):
    """
    偏好分析入口
    基于游戏库数据执行完整的感知-决策-行动闭环
    """
    try:
        agent = get_agent(request.user_id)

        current_prefs = [
            {"tag": p.tag, "weight": p.weight}
            for p in request.current_preferences
        ]
        await agent.initialize(request.user_id, current_prefs)

        game_library = []
        for game in request.game_library:
            game_library.append({
                "game_id": game.game_id,
                "game_name": game.game_name,
                "playtime": game.playtime,
                "playtime_2_weeks": game.playtime_2_weeks,
                "genre": game.genre,
                "tags": game.tags,
                "last_played_at": game.last_played_at,
            })

        result = await agent.analyze_full(game_library, request.trigger_type)

        if not result.get("success", False):
            return PreferenceAnalyzeResponse(success=False)

        new_tags = [
            TagPreferenceData(tag=t.get("tag", ""), weight=t.get("weight", 0))
            for t in result.get("new_tags", [])
        ]

        updated_tags = [
            TagPreferenceData(tag=t.get("tag", ""), weight=t.get("weight", 0))
            for t in result.get("updated_tags", [])
        ]

        anomalies = [
            AnomalyData(
                type=a.get("type", ""),
                description=a.get("description", ""),
                game_id=a.get("game_id", ""),
                game_name=a.get("game_name", "")
            )
            for a in result.get("anomalies", [])
        ]

        return PreferenceAnalyzeResponse(
            success=True,
            new_tags=new_tags,
            updated_tags=updated_tags,
            insights=result.get("insights", []),
            reasoning=result.get("profile_summary", ""),
            anomalies=anomalies,
            recommend_rec=result.get("recommend_rec", False),
        )

    except Exception as e:
        logger.error(f"偏好分析失败: {e}")
        import traceback
        traceback.print_exc()
        raise HTTPException(status_code=500, detail=str(e))
