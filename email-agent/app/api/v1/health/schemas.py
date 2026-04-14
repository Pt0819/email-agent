"""健康检查响应模型"""
from typing import List, Optional
from pydantic import BaseModel


class HealthResponse(BaseModel):
    """健康检查响应"""
    status: str
    service: str
    version: str = "1.0.0"
    llm_status: str = "unavailable"  # available, unavailable
    providers: List[str] = []  # 可用的Provider列表
    message: Optional[str] = None