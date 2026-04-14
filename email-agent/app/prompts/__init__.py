"""
提示词模板模块
"""
from app.prompts.classification import (
    get_system_prompt,
    get_user_prompt,
    parse_classification_result,
    CATEGORIES,
)

__all__ = [
    "get_system_prompt",
    "get_user_prompt",
    "parse_classification_result",
    "CATEGORIES",
]