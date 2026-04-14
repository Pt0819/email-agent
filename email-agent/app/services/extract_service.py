"""
提取服务
负责从邮件中提取关键信息
"""
from datetime import datetime

from app.services.base_service import BaseService
from app.schemas import (
    ExtractRequest, ExtractResponse, ExtractionResult,
    ActionItem, MeetingInfo, KeyEntities
)
from app.llm import get_llm_manager
from app.prompts.extraction import (
    get_system_prompt, get_user_prompt, parse_extraction_result
)


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
            llm_manager = get_llm_manager()

            if not llm_manager.is_available():
                raise Exception("没有可用的LLM Provider")

            # 构建提示词
            system_prompt = get_system_prompt()
            user_prompt = get_user_prompt(
                sender_name=request.sender_name or "",
                sender_email=request.sender_email,
                subject=request.subject,
                content=request.content
            )

            # 调用LLM
            response = await llm_manager.chat_with_system(
                system_prompt=system_prompt,
                user_content=user_prompt
            )

            # 解析结果
            result_data = parse_extraction_result(response.content)

            # 构建行动项
            action_items = []
            for item in result_data.get("action_items", []):
                try:
                    action_items.append(ActionItem(
                        task=item.get("task", ""),
                        task_type=item.get("task_type", "review"),
                        deadline=item.get("deadline"),
                        priority=item.get("priority", "medium"),
                        confidence=item.get("confidence", 0.7)
                    ))
                except Exception:
                    continue

            # 构建会议信息
            meetings = []
            for m in result_data.get("meetings", []):
                try:
                    meetings.append(MeetingInfo(
                        title=m.get("title", ""),
                        time=m.get("time", ""),
                        location=m.get("location"),
                        attendees=m.get("attendees", []),
                        meeting_url=m.get("meeting_url")
                    ))
                except Exception:
                    continue

            # 构建关键实体
            entities_data = result_data.get("key_entities", {})
            key_entities = KeyEntities(
                people=entities_data.get("people", []),
                organizations=entities_data.get("organizations", []),
                projects=entities_data.get("projects", []),
                dates=entities_data.get("dates", []),
                amounts=entities_data.get("amounts", [])
            )

            result = ExtractionResult(
                action_items=action_items,
                meetings=meetings,
                key_entities=key_entities,
                summary=result_data.get("summary", ""),
                intent=result_data.get("intent", "information")
            )

            response_obj = ExtractResponse(
                email_id=request.email_id,
                extraction=result,
                processed_at=datetime.now()
            )

            self.log_info(
                f"邮件信息提取完成: {request.email_id}, "
                f"actions={len(action_items)}, meetings={len(meetings)}"
            )
            return response_obj

        except Exception as e:
            self.log_error(f"提取失败: {request.email_id}", error=str(e))
            raise