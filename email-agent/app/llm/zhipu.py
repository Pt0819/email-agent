"""
智谱GLM Provider实现
支持GLM-4-Flash免费模型
"""
import httpx
from loguru import logger
from typing import Any, Dict, List, Optional

from app.llm.provider import BaseLLMProvider, Message, ChatCompletionResponse


class ZhipuProvider(BaseLLMProvider):
    """智谱GLM Provider"""

    # API endpoint
    BASE_URL = "https://open.bigmodel.cn/api/paas/v4"

    # 支持的模型
    MODELS = {
        "glm-4-flash": "GLM-4-Flash (免费)",  # 免费模型
        "glm-4": "GLM-4",
        "glm-4-air": "GLM-4-Air",
    }

    def __init__(
        self,
        api_key: str,
        model: str = "glm-4-flash",
        temperature: float = 0.3,
        max_tokens: int = 4096,
        base_url: Optional[str] = None
    ):
        """
        初始化智谱GLM Provider

        Args:
            api_key: API密钥
            model: 模型名称，默认使用免费模型
            temperature: 温度参数
            max_tokens: 最大token数
            base_url: 自定义API地址
        """
        self._api_key = api_key
        self._model = model
        self._temperature = temperature
        self._max_tokens = max_tokens
        self._base_url = base_url or self.BASE_URL

        if model not in self.MODELS:
            logger.warning(f"未知的模型: {model}, 使用默认模型 glm-4-flash")
            self._model = "glm-4-flash"

    @property
    def name(self) -> str:
        return "zhipu"

    @property
    def default_model(self) -> str:
        return "glm-4-flash"

    async def chat(
        self,
        messages: List[Message],
        model: Optional[str] = None,
        temperature: Optional[float] = None,
        max_tokens: Optional[int] = None,
        **kwargs
    ) -> ChatCompletionResponse:
        """
        发送聊天请求到智谱GLM

        Args:
            messages: 消息列表
            model: 模型名称（可选覆盖）
            temperature: 温度（可选覆盖）
            max_tokens: 最大token（可选覆盖）
            **kwargs: 其他参数

        Returns:
            响应结果
        """
        use_model = model or self._model
        use_temperature = temperature or self._temperature
        use_max_tokens = max_tokens or self._max_tokens

        # 构建请求体
        request_body = {
            "model": use_model,
            "messages": [{"role": m.role, "content": m.content} for m in messages],
            "temperature": use_temperature,
            "max_tokens": use_max_tokens,
        }

        # 添加可选参数
        if "top_p" in kwargs:
            request_body["top_p"] = kwargs["top_p"]

        headers = {
            "Authorization": f"Bearer {self._api_key}",
            "Content-Type": "application/json",
        }

        url = f"{self._base_url}/chat/completions"

        logger.debug(f"智谱GLM请求: model={use_model}, messages_count={len(messages)}")

        try:
            async with httpx.AsyncClient(timeout=60.0) as client:
                response = await client.post(
                    url,
                    json=request_body,
                    headers=headers
                )

                if response.status_code != 200:
                    error_text = response.text
                    logger.error(f"智谱GLM API错误: {response.status_code} - {error_text}")
                    raise Exception(f"智谱GLM API调用失败: {response.status_code}")

                data = response.json()

                # 解析响应
                choices = data.get("choices", [])
                if not choices:
                    raise Exception("智谱GLM返回空响应")

                content = choices[0].get("message", {}).get("content", "")
                finish_reason = choices[0].get("finish_reason", "stop")
                usage = data.get("usage", {})

                logger.debug(f"智谱GLM响应: content_length={len(content)}, usage={usage}")

                return ChatCompletionResponse(
                    content=content,
                    model=use_model,
                    finish_reason=finish_reason,
                    usage=usage
                )

        except httpx.TimeoutException:
            logger.error("智谱GLM请求超时")
            raise Exception("智谱GLM请求超时")
        except Exception as e:
            logger.error(f"智谱GLM请求异常: {e}")
            raise

    async def validate_connection(self) -> bool:
        """
        验证API连接是否有效

        Returns:
            是否连接成功
        """
        try:
            test_message = Message(role="user", content="你好")
            response = await self.chat([test_message], max_tokens=10)
            return len(response.content) > 0
        except Exception as e:
            logger.warning(f"智谱GLM连接验证失败: {e}")
            return False