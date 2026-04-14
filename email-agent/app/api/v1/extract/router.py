"""
提取API路由
"""
from fastapi import APIRouter, HTTPException

from app.schemas import ExtractRequest, ExtractResponse
from app.services.extract_service import ExtractService

router = APIRouter(prefix="/extract", tags=["提取"])


@router.post("", response_model=ExtractResponse)
async def extract_email_info(request: ExtractRequest):
    """
    提取邮件关键信息

    Args:
        request: 提取请求，包含邮件内容

    Returns:
        ExtractResponse: 提取结果
    """
    try:
        service = ExtractService()
        result = await service.extract(request)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))