"""
枚举定义
"""
from enum import Enum


class EmailCategory(str, Enum):
    """邮件分类"""
    WORK_URGENT = "work_urgent"
    WORK_NORMAL = "work_normal"
    PERSONAL = "personal"
    SUBSCRIPTION = "subscription"
    NOTIFICATION = "notification"
    PROMOTION = "promotion"
    SPAM = "spam"
    UNCLASSIFIED = "unclassified"


class EmailPriority(str, Enum):
    """邮件优先级"""
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


class EmailStatus(str, Enum):
    """邮件状态"""
    UNREAD = "unread"
    READ = "read"
    PROCESSED = "processed"
    ARCHIVED = "archived"


class TaskType(str, Enum):
    """任务类型"""
    REPLY = "reply"
    REVIEW = "review"
    SUBMIT = "submit"
    ATTEND = "attend"
    PREPARE = "prepare"
    OTHER = "other"


class EmailIntent(str, Enum):
    """邮件意图"""
    REQUEST = "request"
    INFORMATION = "information"
    INVITATION = "invitation"
    APPROVAL = "approval"
    DISCUSSION = "discussion"
    NOTIFICATION = "notification"