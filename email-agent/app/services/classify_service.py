"""
分类服务
负责邮件分类的核心业务逻辑
"""
from datetime import datetime

from app.services.base_service import BaseService
from app.schemas import (
    ClassifyRequest, ClassifyResponse, ClassificationResult
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
            # TODO: 调用LLM进行分类
            # 目前返回模拟结果
            result = ClassificationResult(
                category="work_normal",
                priority="medium",
                confidence=0.85,
                reasoning=f"基于发件人和内容分析，判断为普通工作邮件"
            )

            response = ClassifyResponse(
                email_id=request.email_id,
                classification=result,
                processed_at=datetime.now()
            )

            self.log_info(f"邮件分类完成: {request.email_id}, category={result.category}")
            return response

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

        self.log_info(f"批量分类完成: 成功{len(results)}封")
        return results