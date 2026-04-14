"""
分类服务
负责邮件分类的核心业务逻辑
"""
from datetime import datetime

from app.services.base_service import BaseService
from app.schemas import (
    ClassifyRequest, ClassifyResponse, ClassificationResult
)
from app.llm import get_llm_manager
from app.prompts.classification import (
    get_system_prompt, get_user_prompt, parse_classification_result
)


class ClassifyService(BaseService):
    """邮件分类服务"""

    async def execute(self, request: ClassifyRequest) -> ClassifyResponse:
        """执行分类"""
        return await self.classify(request)

    async def classify(self, request: ClassifyRequest) -> ClassifyResponse:
        """
        分类单封邮件

        Args:
            request: 分类请求

        Returns:
            分类结果
        """
        self.log_info(f"开始分类邮件: {request.email_id}")

        try:
            # 获取LLM Manager
            llm_manager = get_llm_manager()

            if not llm_manager.is_available():
                raise Exception("没有可用的LLM Provider")

            # 构建提示词
            system_prompt = get_system_prompt()
            user_prompt = get_user_prompt(
                sender_name=request.sender_name,
                sender_email=request.sender_email,
                subject=request.subject,
                content=request.content,
                received_at=request.received_at
            )

            # 调用LLM
            response = await llm_manager.chat_with_system(
                system_prompt=system_prompt,
                user_content=user_prompt
            )

            # 解析结果
            result_data = parse_classification_result(response.content)

            result = ClassificationResult(
                category=result_data.get("category", "unknown"),
                priority=result_data.get("priority", "medium"),
                confidence=result_data.get("confidence", 0.5),
                reasoning=result_data.get("reasoning", "")
            )

            response_obj = ClassifyResponse(
                email_id=request.email_id,
                classification=result,
                processed_at=datetime.now()
            )

            self.log_info(
                f"邮件分类完成: {request.email_id}, "
                f"category={result.category}, confidence={result.confidence}"
            )
            return response_obj

        except Exception as e:
            self.log_error(f"分类失败: {request.email_id}", error=str(e))
            raise

    async def batch_classify(self, requests: list) -> list[ClassifyResponse]:
        """
        批量分类邮件

        Args:
            requests: 分类请求列表

        Returns:
            分类结果列表
        """
        self.log_info(f"开始批量分类: {len(requests)}封邮件")

        results = []
        for req in requests:
            try:
                result = await self.classify(req)
                results.append(result)
            except Exception as e:
                self.log_error(f"分类邮件失败: {req.email_id}", error=str(e))
                # 返回错误结果
                results.append(ClassifyResponse(
                    email_id=req.email_id,
                    classification=ClassificationResult(
                        category="error",
                        priority="unknown",
                        confidence=0.0,
                        reasoning=f"分类失败: {str(e)}"
                    ),
                    processed_at=datetime.now()
                ))

        self.log_info(f"批量分类完成: 成功{len(results)}封")
        return results