"""
Mock LLM Provider
当没有配置真实API Key时使用，返回预设的分类结果
"""
import json
import re
from loguru import logger
from typing import List, Optional

from app.llm.provider import BaseLLMProvider, Message, ChatCompletionResponse


class MockLLMProvider(BaseLLMProvider):
    """Mock LLM Provider - 用于测试，返回预设结果"""

    # 根据邮件内容的简单规则返回分类结果
    CATEGORY_RULES = [
        (["紧急", "立刻", "尽快", "马上", "urgent"], "work_urgent"),
        (["会议", "会议邀请", "参会", "评审"], "work_normal"),
        (["任务", "需求", "排期", "截止", "完成"], "work_normal"),
        (["聚会", "邀请", "周末", "你好", "嗨"], "personal"),
        (["周刊", "订阅", "newsletter"], "subscription"),
        (["安全提醒", "通知", "验证", "noreply"], "notification"),
        (["特惠", "优惠", "折扣", "促销", "折扣"], "promotion"),
        (["免费", "中奖", "恭喜", "垃圾"], "spam"),
    ]

    PRIORITY_MAP = {
        "work_urgent": "critical",
        "work_normal": "medium",
        "personal": "medium",
        "subscription": "low",
        "notification": "medium",
        "promotion": "low",
        "spam": "low",
    }

    def __init__(self, **kwargs):
        self._model = "mock-llm"

    @property
    def name(self) -> str:
        return "mock"

    @property
    def default_model(self) -> str:
        return "mock-llm"

    def _classify_email(self, subject: str, content: str) -> dict:
        """基于简单规则分类邮件"""
        text = (subject + " " + content).lower()

        category = "work_normal"
        for keywords, cat in self.CATEGORY_RULES:
            for kw in keywords:
                if kw in text:
                    category = cat
                    break
            if category != "work_normal":
                break

        priority = self.PRIORITY_MAP.get(category, "medium")

        # 检测是否包含待办事项
        has_tasks = bool(re.search(r'\d+\.\s|任务|完成|截止', text))
        # 检测是否包含会议信息
        has_meeting = bool(re.search(r'会议|时间.*\d{1,2}:\d{2}|会议室', text))

        action_items = []
        if has_tasks:
            task_matches = re.findall(r'\d+\.\s*(.+?)(?:（|$)', content)
            for task in task_matches:
                action_items.append({
                    "task": task.strip(),
                    "task_type": "follow_up",
                    "priority": "medium"
                })

        meetings = []
        if has_meeting:
            meetings.append({
                "title": subject,
                "time": "",
                "location": "",
                "attendees": []
            })

        return {
            "category": category,
            "priority": priority,
            "confidence": 0.85,
            "reasoning": f"[Mock] 基于关键词匹配分类为 {category}",
            "action_items": action_items,
            "meetings": meetings,
            "summary": f"这是一封关于'{subject}'的{category}类邮件",
            "intent": "inform"
        }

    async def chat(
        self,
        messages: List[Message],
        **kwargs
    ) -> ChatCompletionResponse:
        """模拟LLM响应"""
        # 从用户消息中提取邮件内容
        user_content = ""
        for msg in messages:
            if msg.role == "user":
                user_content = msg.content
                break

        # 判断请求类型（分类 vs 提取 vs 摘要）
        is_system_classify = any(
            "分类" in msg.content for msg in messages if msg.role == "system"
        )
        is_system_extract = any(
            "提取" in msg.content for msg in messages if msg.role == "system"
        )
        is_system_summary = any(
            "摘要" in msg.content for msg in messages if msg.role == "system"
        )

        # 从用户消息中提取subject和content
        subject = ""
        content = ""
        subject_match = re.search(r'主题[：:]\s*(.+)', user_content)
        if subject_match:
            subject = subject_match.group(1).strip()
        content_match = re.search(r'正文[：:]\s*(.+)', user_content, re.DOTALL)
        if content_match:
            content = content_match.group(1).strip()
        else:
            content = user_content

        result = self._classify_email(subject, content)

        # 根据请求类型返回不同格式的结果
        if is_system_classify:
            response_content = json.dumps({
                "category": result["category"],
                "priority": result["priority"],
                "confidence": result["confidence"],
                "reasoning": result["reasoning"]
            }, ensure_ascii=False)
        elif is_system_extract:
            response_content = json.dumps({
                "action_items": result["action_items"],
                "meetings": result["meetings"],
                "summary": result["summary"],
                "intent": result["intent"]
            }, ensure_ascii=False)
        elif is_system_summary:
            response_content = json.dumps({
                "date": "2026-04-17",
                "total_emails": 8,
                "summary": f"今日共收到8封邮件，其中工作邮件4封，个人邮件1封，系统通知1封，其他2封。",
                "key_items": [
                    {"type": "urgent", "description": f"紧急：{subject}" if "紧急" in subject else "无紧急事项"},
                    {"type": "meeting", "description": "1个会议邀请"},
                    {"type": "task", "description": f"{len(result['action_items'])}个待办事项"}
                ],
                "categories": {
                    "work_urgent": 1, "work_normal": 3,
                    "personal": 1, "notification": 1,
                    "subscription": 1, "promotion": 1
                }
            }, ensure_ascii=False)
        else:
            response_content = json.dumps(result, ensure_ascii=False)

        logger.info(f"[Mock LLM] 分类结果: category={result['category']}")

        return ChatCompletionResponse(
            content=response_content,
            model=self._model,
            finish_reason="stop",
            usage={"prompt_tokens": 100, "completion_tokens": 200, "total_tokens": 300}
        )

    async def validate_connection(self) -> bool:
        """Mock始终可用"""
        return True
