"""
依赖注入
管理服务实例的生命周期
"""
from functools import lru_cache
from typing import Optional

from app.core.config import get_config
from app.services.base_service import BaseService
from app.services.classify_service import ClassifyService
from app.services.extract_service import ExtractService

# 导入类型
config = get_config()


def get_classify_service() -> ClassifyService:
    """获取分类服务实例"""
    return ClassifyService()


def get_extract_service() -> ExtractService:
    """获取提取服务实例"""
    return ExtractService()


# 服务缓存（单例模式）
_service_cache = {}


def get_service(service_class: type) -> BaseService:
    """
    获取服务实例（单例）

    Args:
        service_class: 服务类

    Returns:
        服务实例
    """
    if service_class not in _service_cache:
        _service_cache[service_class] = service_class()
    return _service_cache[service_class]