"""
摘要服务
负责生成每日邮件摘要
"""
from datetime import datetime
from typing import List, Dict, Any

from app.services.base_service import BaseService
from app.schemas import DailySummaryResponse, EmailSummary, ActionItem
from app.prompts.summary import (
    get_system_prompt, get_user_prompt, parse_summary_result
)


class SummaryService(BaseService):
    """邮件摘要生成服务"""

    async def generate_daily_summary(
        self,
        email_ids: List[str],
        emails_data: List[Dict[str, Any]],
        date: str
    ) -> DailySummaryResponse:
        """
        生成每日摘要

        Args:
            email_ids: 邮件ID列表
            emails_data: 邮件数据列表（包含email_id, subject, sender_name, category等）
            date: 日期 (YYYY-MM-DD)

        Returns:
            每日摘要响应
        """
        self.log_info(f"开始生成每日摘要: {date}, 邮件数: {len(emails_data)}")

        try:
            # 获取LLM Manager（延迟导入避免循环依赖）
            from app.llm import get_llm_manager
            llm_manager = get_llm_manager()

            if not llm_manager.is_available():
                raise Exception("没有可用的LLM Provider")

            if not emails_data:
                return DailySummaryResponse(
                    date=date,
                    total_emails=0,
                    by_category={},
                    important_emails=[],
                    action_items=[],
                    summary_text="今日无邮件"
                )

            # 构建提示词
            system_prompt = get_system_prompt()
            user_prompt = get_user_prompt(emails=emails_data, date=date)

            # 调用LLM
            response = await llm_manager.chat_with_system(
                system_prompt=system_prompt,
                user_content=user_prompt
            )

            # 解析结果
            result_data = parse_summary_result(response.content)

            # 构建重要邮件列表
            important_emails = []
            for email_data in result_data.get("important_emails", []):
                important_emails.append(EmailSummary(
                    email_id=email_data.get("email_id", ""),
                    subject=email_data.get("subject", ""),
                    sender=email_data.get("sender", ""),
                    category=email_data.get("category", ""),
                    priority=email_data.get("priority", "medium"),
                    summary=email_data.get("reason", "")
                ))

            # 构建行动项列表
            action_items = []
            for action_data in result_data.get("priority_actions", []):
                action_items.append(ActionItem(
                    task=action_data.get("task", ""),
                    task_type="review",
                    deadline=None,
                    priority=action_data.get("priority", "medium"),
                    confidence=0.8
                ))

            response_obj = DailySummaryResponse(
                date=date,
                total_emails=result_data.get("total_emails", len(emails_data)),
                by_category=result_data.get("by_category", {}),
                important_emails=important_emails,
                action_items=action_items,
                summary_text=result_data.get("summary_text", "")
            )

            self.log_info(
                f"每日摘要生成完成: {date}, "
                f"重要邮件: {len(important_emails)}, "
                f"行动项: {len(action_items)}"
            )
            return response_obj

        except Exception as e:
            self.log_error(f"摘要生成失败: {date}", error=str(e))
            raise