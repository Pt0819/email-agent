"""
提示词模板模块
"""
from app.prompts.classification import (
    get_system_prompt,
    get_user_prompt,
    parse_classification_result,
    CATEGORIES,
)
from app.prompts.classify_rules import (
    fast_classify,
    is_steam_email,
    RuleMatch,
    STEAM_PATTERNS,
    NORMAL_PATTERNS,
)

__all__ = [
    "get_system_prompt",
    "get_user_prompt",
    "parse_classification_result",
    "CATEGORIES",
    "fast_classify",
    "is_steam_email",
    "RuleMatch",
    "STEAM_PATTERNS",
    "NORMAL_PATTERNS",
]