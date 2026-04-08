"""
Pydantic请求模型定义
"""
from typing import List, Optional
from pydantic import BaseModel


class ClassifyRequest(BaseModel):
    """分类请求"""
    email_id: str
    subject: str
    sender_name: Optional[str] = None
    sender_email: str
    content: str
    received_at: Optional[str] = None


class BatchClassifyRequest(BaseModel):
    """批量分类请求"""
    emails: List[ClassifyRequest]


class ExtractRequest(BaseModel):
    """信息提取请求"""
    email_id: str
    subject: str
    sender_name: Optional[str] = None
    sender_email: str
    content: str


class SummaryRequest(BaseModel):
    """摘要生成请求"""
    email_ids: List[str]
    date: str  # YYYY-MM-DD