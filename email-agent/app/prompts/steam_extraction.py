"""
Steam信息提取提示词模板
"""


def get_system_prompt() -> str:
    """获取Steam提取系统提示词"""
    return """你是一个专业的Steam游戏信息提取助手。你的任务是从Steam相关邮件中提取游戏促销和资讯信息。

## 提取规则

### 需要提取的信息：
1. **游戏名称** - 每款游戏的完整名称
2. **Steam AppID** - 如果能识别到（通常是数字ID或商店链接中的数字）
3. **原价** - 游戏的原始价格（数字，不带货币符号）
4. **折扣价** - 当前的促销价格
5. **折扣率** - 折扣百分比（如50表示半价）
6. **游戏类型/标签** - 如RPG、动作、策略等
7. **促销截止日期** - 如果有提到
8. **商店链接** - Steam商店URL

### 注意事项：
- 一封邮件可能包含多个游戏促销信息
- 价格统一转换为人民币数字（去掉¥符号）
- 如果无法确定某项信息，填null或空字符串
- 游戏标签提取2-5个最相关的

## 输出格式

请严格按以下JSON格式输出：
{
    "games": [
        {
            "app_id": "Steam应用ID或空字符串",
            "name": "游戏名称",
            "genre": "游戏类型",
            "tags": ["标签1", "标签2"],
            "cover_url": "封面图URL或空字符串",
            "store_url": "商店页URL或空字符串",
            "has_deal": true,
            "original_price": 198.00,
            "deal_price": 99.00,
            "discount": 50,
            "deal_end": "2026-05-01 或空字符串"
        }
    ]
}

如果邮件不是Steam游戏相关内容，返回空数组：
{"games": []}
"""


def get_user_prompt(
    subject: str,
    sender_email: str,
    content: str,
    content_html: str = "",
) -> str:
    """获取用户提示词"""
    # 优先使用HTML内容，信息更丰富
    body = content_html if content_html else content
    if len(body) > 5000:
        body = body[:5000] + "..."

    return f"""请从以下Steam邮件中提取游戏信息：

## 邮件信息
- 主题: {subject}
- 发件人: {sender_email}

## 邮件内容:
{body}

请输出JSON格式的提取结果。"""


def parse_steam_result(content: str) -> dict:
    """解析LLM返回的Steam提取结果"""
    import json
    import re

    # 尝试提取JSON（可能包含嵌套数组）
    # 先尝试匹配最外层的 { ... }
    brace_count = 0
    start_idx = -1
    for i, ch in enumerate(content):
        if ch == '{':
            if brace_count == 0:
                start_idx = i
            brace_count += 1
        elif ch == '}':
            brace_count -= 1
            if brace_count == 0 and start_idx >= 0:
                json_str = content[start_idx:i+1]
                try:
                    return json.loads(json_str)
                except json.JSONDecodeError:
                    continue

    # 回退：尝试解析整个响应
    try:
        return json.loads(content)
    except json.JSONDecodeError:
        pass

    raise ValueError(f"无法解析Steam提取结果: {content[:200]}")
