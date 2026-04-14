"""
分类提示词模板
"""
from typing import List


# 分类类别定义
CATEGORIES = [
    {"code": "work_urgent", "name": "紧急工作", "priority": "critical",
     "description": "需要今天或明天完成的紧急任务、领导邮件、重要通知"},
    {"code": "work_normal", "name": "普通工作", "priority": "medium",
     "description": "普通工作沟通、项目进展、会议通知"},
    {"code": "personal", "name": "个人邮件", "priority": "medium",
     "description": "家人朋友、私人事务、个人邀请"},
    {"code": "subscription", "name": "订阅邮件", "priority": "low",
     "description": "新闻简报、技术博客、定期推送"},
    {"code": "notification", "name": "系统通知", "priority": "low",
     "description": "GitHub、Jira、CI/CD、系统告警"},
    {"code": "promotion", "name": "营销推广", "priority": "low",
     "description": "广告、促销、商业推广"},
    {"code": "spam", "name": "垃圾邮件", "priority": "low",
     "description": "诈骗、可疑内容、无价值邮件"},
]

CATEGORY_TABLE = "\n".join(
    f"| {c['code']} | {c['name']} | {c['description']} |"
    for c in CATEGORIES
)

# 优先级定义
PRIORITY_RULES = """
### 优先级判断规则
- critical: 今天截止，或来自直接领导/客户的紧急邮件
- high: 本周截止，有明确行动请求
- medium: 有价值信息，本周内处理即可
- low: 信息性内容，可延后处理
"""


def get_system_prompt() -> str:
    """获取分类系统提示词"""
    return f"""你是一个专业的邮件分类助手。请根据以下规则准确分析邮件并进行分类。

## 分类标准

### 邮件类别
| 类别代码 | 类别名称 | 描述 |
|---------|---------|------|
{CATEGORY_TABLE}

{PRIORITY_RULES}

## 输出格式要求

请严格按以下JSON格式输出，不要包含其他内容：
{{
    "category": "类别代码",
    "priority": "优先级",
    "confidence": 置信度(0-1之间的浮点数),
    "reasoning": "判断理由(用中文简要说明)"
}}

## 注意事项
1. 置信度反映分类的确定程度，高置信度(>0.8)用于典型邮件，低置信度(<0.5)用于模糊邮件
2. 如果邮件涉及多个类别，选择最重要的那个
3. reasoning不要超过50字
"""


def get_user_prompt(
    sender_name: str,
    sender_email: str,
    subject: str,
    content: str,
    received_at: str = None
) -> str:
    """获取用户提示词"""
    time_info = f"\n- 接收时间: {received_at}" if received_at else ""

    return f"""请分析以下邮件并分类：

## 邮件信息
- 发件人: {sender_name or "未知"}
- 发件人邮箱: {sender_email}
- 主题: {subject}
{time_info}

## 正文内容:
{content[:2000]}{"..." if len(content) > 2000 else ""}

请输出JSON格式的分类结果。"""


def parse_classification_result(content: str) -> dict:
    """解析LLM返回的分类结果"""
    import json
    import re

    # 尝试提取JSON
    json_match = re.search(r'\{[^{}]*\}', content, re.DOTALL)
    if json_match:
        try:
            return json.loads(json_match.group())
        except json.JSONDecodeError:
            pass

    # 尝试解析整个响应
    try:
        return json.loads(content)
    except json.JSONDecodeError:
        pass

    raise ValueError(f"无法解析分类结果: {content[:200]}")