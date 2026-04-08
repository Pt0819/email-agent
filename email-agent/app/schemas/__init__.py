"""
Pydantic数据验证模型
"""
from app.schemas.enums import (
    EmailCategory,
    EmailPriority,
    EmailStatus,
    TaskType,
    EmailIntent,
)
from app.schemas.request import (
    ClassifyRequest,
    BatchClassifyRequest,
    ExtractRequest,
    SummaryRequest,
)
from app.schemas.response import (
    ClassificationResult,
    ClassifyResponse,
    ActionItem,
    MeetingInfo,
    KeyEntities,
    ExtractionResult,
    ExtractResponse,
    EmailSummary,
    DailySummaryResponse,
    HealthResponse,
    ErrorResponse,
)

__all__ = [
    # 枚举
    "EmailCategory",
    "EmailPriority",
    "EmailStatus",
    "TaskType",
    "EmailIntent",
    # 请求
    "ClassifyRequest",
    "BatchClassifyRequest",
    "ExtractRequest",
    "SummaryRequest",
    # 响应
    "ClassificationResult",
    "ClassifyResponse",
    "ActionItem",
    "MeetingInfo",
    "KeyEntities",
    "ExtractionResult",
    "ExtractResponse",
    "EmailSummary",
    "DailySummaryResponse",
    "HealthResponse",
    "ErrorResponse",
]