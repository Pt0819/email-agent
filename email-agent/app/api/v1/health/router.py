"""
健康检查API路由
"""
from fastapi import APIRouter

from app.schemas import HealthResponse
from app.llm import get_llm_manager

router = APIRouter(prefix="/health", tags=["系统"])


@router.get("", response_model=HealthResponse)
async def health_check():
    """健康检查接口"""
    # 检查LLM状态
    llm_manager = get_llm_manager()
    if llm_manager.is_available():
        llm_status = "available"
        providers = llm_manager.get_available_providers()
    else:
        llm_status = "unavailable"
        providers = []

    return HealthResponse(
        status="ok",
        service="email-agent",
        version="1.0.0",
        llm_status=llm_status,
        providers=providers
    )