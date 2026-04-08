"""
分类API路由
"""
from typing import List
from fastapi import APIRouter, HTTPException

from app.schemas import ClassifyRequest, ClassifyResponse, ClassificationResult
from app.services.classify_service import ClassifyService

router = APIRouter(prefix="/classify", tags=["分类"])


@router.post("", response_model=ClassifyResponse)
async def classify_email(request: ClassifyRequest):
    """
    分类单封邮件

    Args:
        request: 分类请求，包含邮件内容

    Returns:
        ClassifyResponse: 分类结果
    """
    try:
        service = ClassifyService()
        result = await service.classify(request)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/batch")
async def batch_classify(requests: List[ClassifyRequest]):
    """
    批量分类邮件

    Args:
        requests: 分类请求列表

    Returns:
        批量分类结果
    """
    try:
        service = ClassifyService()
        results = await service.batch_classify(requests)
        return {"results": results, "total": len(results)}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))