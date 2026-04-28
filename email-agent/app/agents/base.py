"""
Agent基类
定义Agent组件的通用接口和工具函数
"""
from abc import ABC, abstractmethod
from typing import List, Dict, Any, Optional
from loguru import logger
import json


class BaseAgent(ABC):
    """Agent基类"""

    def __init__(self, name: str):
        self.name = name
        self.logger = logger

    def log(self, level: str, message: str, **kwargs):
        """统一日志方法"""
        getattr(self.logger, level)(f"[{self.name}] {message}", **kwargs)


def parse_tags_json(tags_str: str) -> List[str]:
    """解析标签JSON字符串"""
    if not tags_str:
        return []
    try:
        tags = json.loads(tags_str)
        if isinstance(tags, list):
            return [t for t in tags if isinstance(t, str)]
        return []
    except (json.JSONDecodeError, TypeError):
        return []


def normalize_tag(tag: str) -> str:
    """标准化标签名称"""
    tag = tag.strip()
    if not tag:
        return ""
    return tag.lower().title()


def calculate_playtime_hours(minutes: int) -> float:
    """计算游玩小时数"""
    return round(minutes / 60, 1)


def is_significant_playtime(delta_hours: float, threshold: float = 5.0) -> bool:
    """判断游玩时长变化是否显著"""
    return delta_hours >= threshold


def format_playtime_description(hours: float) -> str:
    """格式化游玩时长描述"""
    if hours < 1:
        return f"{int(hours * 60)}分钟"
    if hours < 24:
        return f"{hours:.1f}小时"
    days = int(hours / 24)
    remaining_hours = hours - days * 24
    if remaining_hours < 1:
        return f"{days}天"
    return f"{days}天{int(remaining_hours)}小时"
