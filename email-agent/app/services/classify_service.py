"""
分类服务
负责邮件分类的核心业务逻辑，支持正则预筛选
"""
from datetime import datetime

from app.services.base_service import BaseService
from app.schemas import (
    ClassifyRequest, ClassifyResponse, ClassificationResult
)
from app.prompts.classification import (
    get_system_prompt, get_user_prompt, parse_classification_result
)
from app.prompts.classify_rules import fast_classify, RuleMatch


class ClassifyService(BaseService):
    """邮件分类服务"""

    # 正则预筛选置信度阈值
    REGEX_CONFIDENCE_THRESHOLD = 0.85

    async def execute(self, request: ClassifyRequest) -> ClassifyResponse:
        """执行分类"""
        return await self.classify(request)

    async def classify(self, request: ClassifyRequest) -> ClassifyResponse:
        """
        分类单封邮件

        流程：
        1. 正则预筛选 - 快速判断
        2. 高置信度直接返回
        3. 低置信度/未匹配调用LLM

        Args:
            request: 分类请求

        Returns:
            分类结果
        """
        self.log_info(f"开始分类邮件: {request.email_id}")

        try:
            # Step 1: 正则预筛选
            rule_match = fast_classify(
                sender_email=request.sender_email,
                subject=request.subject,
                content=request.content
            )

            # Step 2: 高置信度直接返回
            if rule_match.matched and rule_match.confidence >= self.REGEX_CONFIDENCE_THRESHOLD:
                result = ClassificationResult(
                    category=rule_match.category,
                    priority=rule_match.priority or "medium",
                    confidence=rule_match.confidence,
                    reasoning=rule_match.reasoning
                )
                response_obj = ClassifyResponse(
                    email_id=request.email_id,
                    classification=result,
                    processed_at=datetime.now()
                )
                self.log_info(
                    f"邮件分类完成(正则): {request.email_id}, "
                    f"category={result.category}, confidence={result.confidence}"
                )
                return response_obj

            # Step 3: 低置信度或未匹配，调用LLM
            return await self._classify_with_llm(request, rule_match)

        except Exception as e:
            self.log_error(f"分类失败: {request.email_id}", error=str(e))
            raise

    async def _classify_with_llm(
        self,
        request: ClassifyRequest,
        rule_match: RuleMatch = None
    ) -> ClassifyResponse:
        """
        使用LLM进行邮件分类

        Args:
            request: 分类请求
            rule_match: 正则预筛选结果（可选，用于提示LLM）

        Returns:
            分类结果
        """
        from app.llm import get_llm_manager

        if not get_llm_manager().is_available():
            raise Exception("没有可用的LLM Provider")

        # 构建增强的系统提示词
        system_prompt = get_system_prompt()

        # 如果正则有预判结果，在用户提示词中提供参考
        user_prompt = get_user_prompt(
            sender_name=request.sender_name,
            sender_email=request.sender_email,
            subject=request.subject,
            content=request.content,
            received_at=request.received_at
        )

        if rule_match and rule_match.needs_llm and rule_match.category:
            user_prompt += f"\n\n[参考分类]: 正则预筛选建议为 {rule_match.category}，置信度 {rule_match.confidence}，请结合正文内容最终确认。"

        # 调用LLM
        response = await get_llm_manager().chat_with_system(
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
            f"邮件分类完成(LLM): {request.email_id}, "
            f"category={result.category}, confidence={result.confidence}"
        )
        return response_obj

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

    def is_steam_category(self, category: str) -> bool:
        """判断是否为Steam相关分类"""
        return category.startswith("steam_")

    def needs_steam_extract(self, category: str) -> bool:
        """
        判断是否需要触发Steam信息提取

        Args:
            category: 分类结果

        Returns:
            True if should trigger steam extract
        """
        steam_extract_categories = {"steam_promotion", "steam_wishlist", "steam_news"}
        return category in steam_extract_categories