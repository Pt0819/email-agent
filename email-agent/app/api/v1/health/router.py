"""
健康检查API路由
"""
from fastapi import APIRouter

from app.schemas.health import HealthResponse

router = APIRouter(prefix="/health", tags=["系统"])


@router.get("", response_model=HealthResponse)
async def health_check():
    """健康检查接口"""
    return HealthResponse(
        status="ok",
        service="email-agent",
        version="1.0.0",
        llm_status="ok"
    )