"""健康检查响应模型"""
from pydantic import BaseModel


class HealthResponse(BaseModel):
    """健康检查响应"""
    status: str
    service: str
    version: str = "1.0.0"
    llm_status: str = "ok"