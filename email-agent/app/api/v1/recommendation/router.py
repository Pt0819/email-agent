"""
推荐API路由
处理推荐理由生成请求
"""
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import List, Dict, Any, Optional
from loguru import logger

from app.llm import get_llm_manager
from app.services.recommendation_service import RecommendationReasonGenerator

router = APIRouter(tags=["recommendation"])


class GenerateReasonRequest(BaseModel):
    """生成推荐理由请求"""
    game_name: str
    game_genre: str = ""
    game_tags: List[str] = []
    user_preferences: List[Dict[str, Any]] = []
    recent_games: List[Dict[str, Any]] = []


class GenerateReasonResponse(BaseModel):
    """生成推荐理由响应"""
    success: bool
    reason: str
    game_name: str


class MatchScoreRequest(BaseModel):
    """匹配度计算请求"""
    game_tags: List[str]
    game_genre: str = ""
    user_preferences: List[Dict[str, Any]] = []


class MatchScoreResponse(BaseModel):
    """匹配度计算响应"""
    score: float
    matches: List[str]


class BatchGenerateRequest(BaseModel):
    """批量生成推荐理由请求"""
    games: List[Dict[str, Any]] = []
    user_preferences: List[Dict[str, Any]] = []
    recent_games: List[Dict[str, Any]] = []


class BatchGenerateResponse(BaseModel):
    """批量生成推荐理由响应"""
    success: bool
    results: List[Dict[str, Any]]


@router.post("/recommendation/reason", response_model=GenerateReasonResponse)
async def generate_recommendation_reason(req: GenerateReasonRequest):
    """生成个性化推荐理由"""
    try:
        llm_manager = get_llm_manager()
        generator = RecommendationReasonGenerator(llm_manager)

        reason = await generator.generate_reason(
            game_name=req.game_name,
            game_genre=req.game_genre,
            game_tags=req.game_tags,
            user_preferences=req.user_preferences,
            recent_games=req.recent_games,
        )

        return GenerateReasonResponse(
            success=True,
            reason=reason,
            game_name=req.game_name,
        )

    except Exception as e:
        logger.error(f"生成推荐理由失败: {e}")
        raise HTTPException(status_code=500, detail=f"生成推荐理由失败: {e}")


@router.post("/recommendation/match", response_model=MatchScoreResponse)
async def calculate_match_score(req: MatchScoreRequest):
    """计算游戏与用户偏好的匹配度"""
    from app.services.recommendation_service import RecommendationMatcher

    matcher = RecommendationMatcher()
    score, matches = matcher.calculate_match_score(
        game_tags=req.game_tags,
        game_genre=req.game_genre,
        user_preferences=req.user_preferences,
    )

    return MatchScoreResponse(score=score, matches=matches)


@router.post("/recommendation/batch-reason", response_model=BatchGenerateResponse)
async def batch_generate_reasons(req: BatchGenerateRequest):
    """批量生成推荐理由"""
    try:
        llm_manager = get_llm_manager()
        generator = RecommendationReasonGenerator(llm_manager)

        results = []
        for game in req.games:
            reason = await generator.generate_reason(
                game_name=game.get("game_name", ""),
                game_genre=game.get("game_genre", ""),
                game_tags=game.get("game_tags", []),
                user_preferences=req.user_preferences,
                recent_games=req.recent_games,
            )

            results.append({
                "game_name": game.get("game_name", ""),
                "game_id": game.get("game_id", ""),
                "reason": reason,
                "success": True,
            })

        return BatchGenerateResponse(success=True, results=results)

    except Exception as e:
        logger.error(f"批量生成推荐理由失败: {e}")
        raise HTTPException(status_code=500, detail=f"批量生成推荐理由失败: {e}")
