"""
邮件分类正则规则预筛选
用于快速识别邮件类型，减少LLM调用
"""
import re
from dataclasses import dataclass
from typing import Optional

# Steam邮件识别规则
STEAM_PATTERNS = {
    "sender_domains": [
        r"steampowered\.com",
        r"store\.steampowered\.com",
        r"news\.steampowered\.com",
        r"steamcommunity\.com",
        r"valvesoftware\.com",
    ],
    "subject_keywords": [
        r"steam",
        r"steam商店",
        r"steam store",
        r"愿望单",
        r"wishlist",
        r"特惠",
        r"特卖",
        r"打折",
        r"促销活动",
        r"\bdeal\b",
        r"\bsale\b",
        r"\bdiscount\b",
    ],
}

# Steam分类细分规则（按优先级排序）
STEAM_SUB_CATEGORIES = [
    # 愿望单 - 优先级最高
    {
        "keywords": [
            r"愿望单",
            r"wishlist",
            r"愿望清单",
            r"正在打折",
            r"降价",
            r"愿望单提醒",
        ],
        "category": "steam_wishlist",
        "priority": "medium",
    },
    # 促销 - 通用Steam促销
    {
        "keywords": [
            r"特惠",
            r"特卖",
            r"促销",
            r"sale",
            r"deal",
            r"打折",
            r"折扣",
            r"优惠",
            r"限时",
            r" bundle",
            r"bundle",
        ],
        "category": "steam_promotion",
        "priority": "medium",
    },
    # 资讯 - 新闻更新
    {
        "keywords": [
            r"新闻",
            r"资讯",
            r"公告",
            r"更新",
            r"社区",
            r"文章",
            r"攻略",
            r"评测",
        ],
        "category": "steam_news",
        "priority": "low",
    },
]

# 普通邮件快速分类规则（按优先级从高到低）
NORMAL_PATTERNS = [
    # 紧急工作 - 最高优先级
    {
        "keywords": [
            r"紧急",
            r"立刻",
            r"马上",
            r"尽快",
            r"尽快处理",
            r"urgent",
            r"asap",
            r"立刻处理",
            r"马上处理",
        ],
        "category": "work_urgent",
        "priority": "critical",
    },
    # 会议相关
    {
        "keywords": [
            r"会议邀请",
            r"参会",
            r"会议通知",
            r"meeting",
            r"邀请您参加",
        ],
        "category": "work_normal",
        "priority": "medium",
    },
    # 任务相关
    {
        "keywords": [
            r"任务指派",
            r"需求变更",
            r"排期",
            r"截止日期",
            r"deadline",
        ],
        "category": "work_normal",
        "priority": "medium",
    },
    # 订阅邮件
    {
        "keywords": [
            r"周刊",
            r"newsletter",
            r"订阅",
            r"邮件订阅",
            r"技术周报",
            r"日报",
        ],
        "category": "subscription",
        "priority": "low",
    },
    # 系统通知
    {
        "keywords": [
            r"验证",
            r"安全提醒",
            r"来自.*noreply",
            r"notification",
            r"系统通知",
            r"账号安全",
            r"密码重置",
        ],
        "category": "notification",
        "priority": "low",
    },
    # 营销推广（非Steam）
    {
        "keywords": [
            r"优惠.*领取",
            r"限时特惠",
            r"全场.*折",
            r"促销",
            r"广告",
            r"推广",
            r"免费领取",
        ],
        "category": "promotion",
        "priority": "low",
    },
    # 垃圾邮件
    {
        "keywords": [
            r"恭喜.*中奖",
            r"免费送",
            r"立刻获奖",
            r"您已中奖",
            r"诈骗",
            r"钓鱼",
        ],
        "category": "spam",
        "priority": "low",
    },
]


@dataclass
class RuleMatch:
    """规则匹配结果"""
    matched: bool
    category: Optional[str] = None
    priority: Optional[str] = None
    confidence: float = 1.0
    reasoning: str = ""
    needs_llm: bool = False  # 是否需要LLM进一步判断


def compile_patterns(patterns: list[str]) -> list[re.Pattern]:
    """编译正则模式"""
    return [re.compile(p, re.IGNORECASE) for p in patterns]


def compile_keywords(keywords: list[str]) -> list[re.Pattern]:
    """编译关键词为正则模式"""
    # 转义特殊字符并构建单词边界匹配
    result = []
    for kw in keywords:
        # 如果关键词已经有正则元字符，直接编译
        if any(c in kw for c in r'\^$.|?*+[]{}()'):
            try:
                result.append(re.compile(kw, re.IGNORECASE))
            except re.error:
                # 如果编译失败，当作普通文本处理
                result.append(re.compile(r'\b' + re.escape(kw) + r'\b', re.IGNORECASE))
        else:
            # 普通文本添加单词边界
            result.append(re.compile(r'\b' + kw + r'\b', re.IGNORECASE))
    return result


# 预编译所有模式
_steam_sender_patterns = compile_patterns(STEAM_PATTERNS["sender_domains"])
_steam_subject_patterns = compile_keywords(STEAM_PATTERNS["subject_keywords"])
_steam_sub_category_patterns = [
    compile_keywords(cfg["keywords"])
    for cfg in STEAM_SUB_CATEGORIES
]
_normal_patterns = [
    compile_keywords(cfg["keywords"])
    for cfg in NORMAL_PATTERNS
]


def is_steam_email(sender_email: str, subject: str, content: str = "") -> bool:
    """
    快速判断是否为Steam邮件

    Args:
        sender_email: 发件人邮箱
        subject: 邮件主题
        content: 邮件内容（可选）

    Returns:
        True if is Steam email
    """
    # 检查发件人域名
    for pattern in _steam_sender_patterns:
        if pattern.search(sender_email):
            return True

    # 检查主题关键词
    for pattern in _steam_subject_patterns:
        if pattern.search(subject):
            return True

    return False


def match_steam_sub_category(subject: str, content: str = "") -> RuleMatch:
    """
    匹配Steam子分类

    Args:
        subject: 邮件主题
        content: 邮件内容（可选）

    Returns:
        匹配结果
    """
    combined = f"{subject} {content[:500] if content else ''}"

    for i, patterns in enumerate(_steam_sub_category_patterns):
        for pattern in patterns:
            if pattern.search(combined):
                cfg = STEAM_SUB_CATEGORIES[i]
                return RuleMatch(
                    matched=True,
                    category=cfg["category"],
                    priority=cfg["priority"],
                    confidence=0.95,
                    reasoning=f"正则匹配Steam{cfg['category'].split('_')[1]}邮件",
                    needs_llm=False
                )

    # Steam邮件但无法细分，默认促销
    return RuleMatch(
        matched=True,
        category="steam_promotion",
        priority="medium",
        confidence=0.8,
        reasoning="正则识别Steam促销邮件",
        needs_llm=False
    )


def match_normal_category(sender_email: str, subject: str, content: str = "") -> RuleMatch:
    """
    匹配普通邮件分类

    Args:
        sender_email: 发件人邮箱
        subject: 邮件主题
        content: 邮件内容（可选）

    Returns:
        匹配结果
    """
    combined = f"{sender_email} {subject} {content[:500] if content else ''}"

    for i, patterns in enumerate(_normal_patterns):
        for pattern in patterns:
            if pattern.search(combined):
                cfg = NORMAL_PATTERNS[i]
                return RuleMatch(
                    matched=True,
                    category=cfg["category"],
                    priority=cfg["priority"],
                    confidence=0.9,
                    reasoning=f"正则匹配{cfg['category']}邮件",
                    needs_llm=False
                )

    # 没有匹配普通规则，需要LLM判断
    return RuleMatch(
        matched=False,
        needs_llm=True
    )


def fast_classify(sender_email: str, subject: str, content: str = "") -> RuleMatch:
    """
    快速分类邮件（正则预筛选）

    使用规则：
    1. 优先检查是否为Steam邮件
    2. Steam邮件使用专门的分类规则
    3. 普通邮件使用快速分类规则
    4. 无法匹配则需要LLM判断

    Args:
        sender_email: 发件人邮箱
        subject: 邮件主题
        content: 邮件内容（可选）

    Returns:
        分类结果
    """
    # Step 1: 快速判断是否为Steam邮件
    if is_steam_email(sender_email, subject, content):
        return match_steam_sub_category(subject, content)

    # Step 2: 检查普通邮件分类规则
    return match_normal_category(sender_email, subject, content)