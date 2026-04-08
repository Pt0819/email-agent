"""
提取服务
负责从邮件中提取关键信息
"""
from datetime import datetime

from app.services.base_service import BaseService
from app.schemas import ExtractRequest, ExtractResponse, ExtractionResult


class ExtractService(BaseService):
    """邮件信息提取服务"""

    async def execute(self, request: ExtractRequest) -> ExtractResponse:
        """执行信息提取"""
        return await self.extract(request)

    async def extract(self, request: ExtractRequest) -> ExtractResponse:
        """
        提取邮件关键信息

        Args:
            request: 提取请求

        Returns:
            提取结果
        """
        self.log_info(f"开始提取邮件信息: {request.email_id}")

        try:
            # TODO: 调用LLM进行信息提取
            result = ExtractionResult(
                summary=f"这是一封关于{request.subject}的邮件摘要",
                intent="information"
            )

            response = ExtractResponse(
                email_id=request.email_id,
                extraction=result,
                processed_at=datetime.now()
            )

            self.log_info(f"邮件信息提取完成: {request.email_id}")
            return response

        except Exception as e:
            self.log_error(f"提取失败: {request.email_id}", error=str(e))
            raise