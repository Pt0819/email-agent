"""
服务基类
所有业务服务应继承此类
"""
from typing import Optional
from loguru import logger


class BaseService:
    """服务基类"""

    def __init__(self):
        self.logger = logger

    async def execute(self, *args, **kwargs):
        """执行服务方法，子类实现具体逻辑"""
        raise NotImplementedError("子类必须实现 execute 方法")

    def log_info(self, message: str, **kwargs):
        """记录Info级别日志"""
        self.logger.info(f"{self.__class__.__name__}: {message}", **kwargs)

    def log_error(self, message: str, **kwargs):
        """记录Error级别日志"""
        self.logger.error(f"{self.__class__.__name__}: {message}", **kwargs)

    def log_warning(self, message: str, **kwargs):
        """记录Warning级别日志"""
        self.logger.warning(f"{self.__class__.__name__}: {message}", **kwargs)