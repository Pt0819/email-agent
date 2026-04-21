"""
Steam信息提取服务
从Steam邮件中提取游戏促销信息
"""
from datetime import datetime

from app.services.base_service import BaseService
from app.prompts.steam_extraction import (
    get_system_prompt, get_user_prompt, parse_steam_result
)


class SteamExtractService(BaseService):
    """Steam信息提取服务"""

    async def extract(self, email_id: str, subject: str,
                      sender_email: str, content: str,
                      content_html: str = "") -> dict:
        """
        从Steam邮件中提取游戏信息

        Args:
            email_id: 邮件ID
            subject: 邮件主题
            sender_email: 发件人邮箱
            content: 纯文本内容
            content_html: HTML内容

        Returns:
            提取结果dict
        """
        self.log_info(f"开始提取Steam信息: {email_id}")

        try:
            from app.llm import get_llm_manager
            llm_manager = get_llm_manager()

            if not llm_manager.is_available():
                raise Exception("没有可用的LLM Provider")

            # 构建提示词
            system_prompt = get_system_prompt()
            user_prompt = get_user_prompt(
                subject=subject,
                sender_email=sender_email,
                content=content,
                content_html=content_html,
            )

            # 调用LLM
            response = await llm_manager.chat_with_system(
                system_prompt=system_prompt,
                user_content=user_prompt
            )

            # 解析结果
            result_data = parse_steam_result(response.content)

            games = result_data.get("games", [])

            self.log_info(
                f"Steam信息提取完成: {email_id}, "
                f"提取到{len(games)}款游戏"
            )

            return {
                "email_id": email_id,
                "games": games,
                "processed_at": datetime.now().isoformat(),
            }

        except Exception as e:
            self.log_error(f"Steam提取失败: {email_id}", error=str(e))
            raise
