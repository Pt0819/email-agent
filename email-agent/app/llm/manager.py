"""
LLM管理器
负责Provider的初始化、路由和统一调用接口

设计考虑：
- 当前：从config.yaml读取配置，适合个人部署
- 未来：可通过set_user_config扩展为从数据库获取用户级配置
"""
from typing import Dict, List, Optional

from loguru import logger

from app.core.config import get_config
from app.llm.provider import BaseLLMProvider, Message, ChatCompletionResponse
from app.llm.zhipu import ZhipuProvider
from app.llm.deepseek import DeepSeekProvider
from app.llm.mock import MockLLMProvider


class LLMManager:
    """LLM管理器"""

    # Provider注册表
    PROVIDER_CLASSES = {
        "zhipu": ZhipuProvider,
        "deepseek": DeepSeekProvider,
        "mock": MockLLMProvider,  # Mock Provider始终可用
    }

    def __init__(self):
        self._providers: Dict[str, BaseLLMProvider] = {}
        self._default_provider: str = ""
        self._initialized = False

    def initialize(self) -> None:
        """
        从配置文件初始化所有启用的Provider

        当前：从config.yaml读取
        未来：可扩展为从数据库/API获取用户配置
        """
        config = get_config()
        self._default_provider = config.llm.default_provider

        # 始终初始化Mock Provider作为fallback
        self._providers["mock"] = MockLLMProvider()
        logger.info("Provider mock 初始化成功 (fallback)")

        for provider_name, provider_config in config.llm.providers.items():
            if not provider_config.enabled:
                logger.info(f"Provider {provider_name} 已禁用，跳过")
                continue

            if not provider_config.api_key:
                logger.warning(f"Provider {provider_name} 缺少API Key，跳过")
                continue

            provider_cls = self.PROVIDER_CLASSES.get(provider_name)
            if not provider_cls:
                logger.warning(f"未知的Provider: {provider_name}")
                continue

            try:
                provider = provider_cls(
                    api_key=provider_config.api_key,
                    model=provider_config.model,
                    temperature=provider_config.temperature,
                    max_tokens=provider_config.max_tokens,
                    base_url=provider_config.base_url,
                )
                self._providers[provider_name] = provider
                logger.info(f"Provider {provider_name} 初始化成功")
            except Exception as e:
                logger.error(f"Provider {provider_name} 初始化失败: {e}")

        self._initialized = True

        if len(self._providers) == 1 and "mock" in self._providers:
            # 只有mock provider可用
            logger.warning("没有配置真实LLM Provider，将使用Mock Provider进行测试")
            self._default_provider = "mock"
        else:
            logger.info(
                f"LLM Manager初始化完成, "
                f"可用Provider: {list(self._providers.keys())}, "
                f"默认: {self._default_provider}"
            )

    async def chat(
        self,
        messages: List[Message],
        provider: Optional[str] = None,
        **kwargs
    ) -> ChatCompletionResponse:
        """
        统一的LLM调用接口

        Args:
            messages: 消息列表
            provider: 指定Provider（可选，默认使用配置的default_provider）
            **kwargs: 传递给Provider的额外参数

        Returns:
            LLM响应

        Raises:
            Exception: 没有可用的Provider时抛出异常
        """
        self._ensure_initialized()

        provider_name = provider or self._default_provider
        provider_instance = self._providers.get(provider_name)

        if not provider_instance:
            # 尝试使用任意可用的Provider
            if self._providers:
                fallback_name = next(iter(self._providers))
                logger.warning(
                    f"Provider {provider_name} 不可用, "
                    f"回退到 {fallback_name}"
                )
                provider_name = fallback_name
                provider_instance = self._providers[provider_name]
            else:
                raise Exception("没有可用的LLM Provider")

        logger.debug(f"使用Provider: {provider_name}")
        return await provider_instance.chat(messages, **kwargs)

    async def chat_with_system(
        self,
        system_prompt: str,
        user_content: str,
        provider: Optional[str] = None,
        **kwargs
    ) -> ChatCompletionResponse:
        """
        便捷方法：system + user 两轮对话

        Args:
            system_prompt: 系统提示词
            user_content: 用户内容
            provider: 指定Provider
            **kwargs: 额外参数

        Returns:
            LLM响应
        """
        messages = [
            Message(role="system", content=system_prompt),
            Message(role="user", content=user_content),
        ]
        return await self.chat(messages, provider=provider, **kwargs)

    async def validate_provider(self, provider_name: str) -> bool:
        """
        验证指定Provider连接是否有效

        Args:
            provider_name: Provider名称

        Returns:
            是否连接成功
        """
        provider_instance = self._providers.get(provider_name)
        if not provider_instance:
            return False
        return await provider_instance.validate_connection()

    def get_available_providers(self) -> List[str]:
        """获取所有可用的Provider名称"""
        self._ensure_initialized()
        return list(self._providers.keys())

    def is_available(self) -> bool:
        """检查是否有可用的Provider"""
        self._ensure_initialized()
        return len(self._providers) > 0

    def _ensure_initialized(self) -> None:
        """确保Manager已初始化"""
        if not self._initialized:
            self.initialize()


# ---- 全局单例 ----

_manager: Optional[LLMManager] = None


def get_llm_manager() -> LLMManager:
    """获取全局LLM Manager实例"""
    global _manager
    if _manager is None:
        _manager = LLMManager()
        _manager.initialize()
    return _manager