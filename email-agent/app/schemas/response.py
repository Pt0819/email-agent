"""
Pydantic响应模型定义
"""
from datetime import datetime
from typing import List, Dict, Optional, Any
from pydantic import BaseModel, Field


class ClassificationResult(BaseModel):
    """分类结果"""
    category: str
    priority: str
    confidence: float = Field(ge=0, le=1)
    reasoning: str


class ClassifyResponse(BaseModel):
    """分类响应"""
    email_id: str
    classification: ClassificationResult
    processed_at: datetime = Field(default_factory=datetime.now)


class ActionItem(BaseModel):
    """行动项"""
    task: str
    task_type: str
    deadline: Optional[str] = None
    priority: str = "medium"
    confidence: float = Field(ge=0, le=1, default=0.8)


class MeetingInfo(BaseModel):
    """会议信息"""
    title: str
    time: str
    location: Optional[str] = None
    attendees: List[str] = []
    meeting_url: Optional[str] = None


class KeyEntities(BaseModel):
    """关键实体"""
    people: List[str] = []
    organizations: List[str] = []
    projects: List[str] = []
    dates: List[str] = []
    amounts: List[str] = []


class ExtractionResult(BaseModel):
    """提取结果"""
    action_items: List[ActionItem] = []
    meetings: List[MeetingInfo] = []
    key_entities: KeyEntities = Field(default_factory=KeyEntities)
    summary: str = ""
    intent: str = "information"


class ExtractResponse(BaseModel):
    """提取响应"""
    email_id: str
    extraction: ExtractionResult
    processed_at: datetime = Field(default_factory=datetime.now)


class EmailSummary(BaseModel):
    """邮件摘要"""
    email_id: str
    subject: str
    sender: str
    category: str
    priority: str
    summary: str


class DailySummaryResponse(BaseModel):
    """每日摘要响应"""
    date: str
    total_emails: int
    by_category: Dict[str, int]
    important_emails: List[EmailSummary]
    action_items: List[ActionItem]
    summary_text: str


class HealthResponse(BaseModel):
    """健康检查响应"""
    status: str
    service: str
    version: str = "1.0.0"
    llm_status: str = "ok"


class ErrorResponse(BaseModel):
    """错误响应"""
    code: int
    message: str
    detail: Optional[str] = None