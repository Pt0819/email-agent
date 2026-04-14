"""
LLM Provider接口定义
"""
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Any, Dict, List, Optional


@dataclass
class Message:
    """对话消息"""
    role: str  # system, user, assistant
    content: str


@dataclass
class ChatCompletionRequest:
    """聊天完成请求"""
    messages: List[Message]
    model: str = ""
    temperature: float = 0.3
    max_tokens: int = 4096
    stream: bool = False


@dataclass
class ChatCompletionResponse:
    """聊天完成响应"""
    content: str
    model: str
    finish_reason: str = "stop"
    usage: Optional[Dict[str, Any]] = None


class BaseLLMProvider(ABC):
    """LLM Provider基类"""

    @property
    @abstractmethod
    def name(self) -> str:
        """Provider名称"""
        pass

    @property
    @abstractmethod
    def default_model(self) -> str:
        """默认模型"""
        pass

    @abstractmethod
    async def chat(self, messages: List[Message], **kwargs) -> ChatCompletionResponse:
        """
        发送聊天请求

        Args:
            messages: 消息列表
            **kwargs: 其他参数

        Returns:
            响应结果
        """
        pass

    @abstractmethod
    async def validate_connection(self) -> bool:
        """
        验证连接是否有效

        Returns:
            是否连接成功
        """
        pass
