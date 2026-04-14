"""
摘要生成提示词模板
"""
from typing import List


def get_system_prompt() -> str:
    """获取摘要生成系统提示词"""
    return """你是一个专业的邮件摘要生成助手。请根据多封邮件生成每日摘要。

## 任务

根据用户一天的邮件，生成结构化摘要：

### 1. 统计信息
- 总邮件数
- 各分类邮件数量

### 2. 重要邮件
选出当天最重要的3-5封邮件，说明为什么重要。

### 3. 待处理行动项
汇总所有需要执行的行动，按优先级排序。

### 4. 摘要文本
用200-300字总结一天的工作邮件概况。

## 输出格式
请严格按以下JSON格式输出：
{
    "total_emails": 总数,
    "by_category": {"分类名": 数量},
    "important_emails": [
        {"email_id": "ID", "subject": "主题", "importance": "重要性", "reason": "原因"}
    ],
    "action_items_summary": "待处理行动项汇总",
    "priority_actions": [
        {"task": "任务", "priority": "优先级", "from_email": "来源邮件主题"}
    ],
    "summary_text": "一天工作摘要"
}

只输出JSON，不要其他内容。"""


def get_user_prompt(
    emails: List[dict],
    date: str
) -> str:
    """获取用户提示词"""
    # 将邮件列表格式化为文本
    email_lines = []
    for i, email in enumerate(emails):
        email_lines.append(f"""
--- 邮件 {i+1} ---
ID: {email.get('email_id', 'N/A')}
发件人: {email.get('sender_name', '未知')}
主题: {email.get('subject', '(无主题)')}
分类: {email.get('category', 'unknown')}
优先级: {email.get('priority', 'medium')}
摘要: {email.get('summary', '无摘要')}
""")

    emails_text = "\n".join(email_lines)

    return f"""请为 {date} 的邮件生成每日摘要：

{emails_text}

请输出JSON格式的摘要结果。"""


def parse_summary_result(content: str) -> dict:
    """解析LLM返回的摘要结果"""
    import json
    import re

    # 尝试提取JSON
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

    raise ValueError(f"无法解析摘要结果: {content[:200]}")