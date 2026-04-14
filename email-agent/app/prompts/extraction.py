"""
信息提取提示词模板
"""
from typing import List


def get_system_prompt() -> str:
    """获取信息提取系统提示词"""
    return """你是一个专业的邮件信息提取助手。请从邮件中提取关键信息。

## 提取任务

### 1. 行动项（Action Items）
提取邮件中需要执行的任务：
- task: 任务描述（简洁明确）
- task_type: 任务类型
  - reply: 需要回复
  - review: 需要审阅/查看
  - submit: 需要提交
  - attend: 需要参加
  - prepare: 需要准备
- deadline: 截止时间（ISO格式，如无则为null）
- priority: 优先级 (high/medium/low)

### 2. 会议信息（Meetings）
提取会议安排：
- title: 会议标题
- time: 会议时间
- location: 会议地点（如无则为null）
- attendees: 参会人员列表
- meeting_url: 会议链接（如无则为null）

### 3. 关键实体（Key Entities）
提取重要实体：
- people: 提到的人物名称
- organizations: 提到的组织/公司
- projects: 提到的项目名称
- dates: 重要日期
- amounts: 涉及金额

### 4. 邮件意图（Intent）
判断邮件的主要意图：
- request: 请求行动（需要你做某事）
- information: 提供信息（告知/通知）
- invitation: 邀请参会/参加活动
- approval: 需要审批/确认
- discussion: 需要讨论
- notification: 系统自动通知

### 5. 摘要（Summary）
用1-2句话概括邮件核心内容。

## 输出格式
请严格按以下JSON格式输出：
{
    "action_items": [
        {"task": "任务描述", "task_type": "类型", "deadline": null, "priority": "medium", "confidence": 0.9}
    ],
    "meetings": [
        {"title": "会议标题", "time": "时间", "location": "地点", "attendees": [], "meeting_url": null}
    ],
    "key_entities": {
        "people": [],
        "organizations": [],
        "projects": [],
        "dates": [],
        "amounts": []
    },
    "intent": "意图类型",
    "summary": "邮件摘要"
}

如果没有某个类别的信息，使用空数组或空值。只输出JSON，不要其他内容。"""


def get_user_prompt(
    sender_name: str,
    sender_email: str,
    subject: str,
    content: str
) -> str:
    """获取用户提示词"""
    return f"""请从以下邮件中提取关键信息：

## 邮件信息
- 发件人: {sender_name or "未知"}
- 发件人邮箱: {sender_email}
- 主题: {subject}

## 正文内容:
{content[:3000]}{"..." if len(content) > 3000 else ""}

请输出JSON格式的提取结果。"""


def parse_extraction_result(content: str) -> dict:
    """解析LLM返回的提取结果"""
    import json
    import re

    # 尝试提取JSON（支持嵌套大括号）
    json_match = re.search(r'\{.*\}', content, re.DOTALL)
    if json_match:
        try:
            return json.loads(json_match.group())
        except json.JSONDecodeError:
            pass

    try:
        return json.loads(content)
    except json.JSONDecodeError:
        pass

    raise ValueError(f"无法解析提取结果: {content[:200]}")