"""
摘要API路由
"""
from typing import List, Dict, Any
from fastapi import APIRouter, HTTPException

from app.schemas import SummaryRequest, DailySummaryResponse
from app.services.summary_service import SummaryService

router = APIRouter(prefix="/summary", tags=["摘要"])


@router.post("/daily", response_model=DailySummaryResponse)
async def generate_daily_summary(request: SummaryRequest):
    """
    生成每日邮件摘要

    Args:
        request: 摘要请求，包含邮件ID列表和日期

    Returns:
        DailySummaryResponse: 每日摘要
    """
    try:
        service = SummaryService()

        # TODO: 从数据库或缓存获取邮件数据
        # 目前使用请求中的数据（由后端传入）
        emails_data = _build_email_data(request)

        result = await service.generate_daily_summary(
            email_ids=request.email_ids,
            emails_data=emails_data,
            date=request.date
        )
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


def _build_email_data(request: SummaryRequest) -> List[Dict[str, Any]]:
    """构建邮件数据列表

    Note: 实际数据应该从数据库或后端获取
    这里只是一个占位实现
    """
    # 返回空列表，由SummaryService处理无邮件的情况
    return []