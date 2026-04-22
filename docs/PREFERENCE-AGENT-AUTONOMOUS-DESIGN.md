# 偏好分析Agent自主性增强设计

> 日期：2026-04-22
> 版本：v1.0
> 状态：设计文档（优化方向）
> **目标：将偏好分析从被动问答升级为自主决策Agent**

---

## 1. 背景与目标

### 1.1 当前方案（被动式）

当前偏好分析采用"请求-响应"模式：

```
用户触发请求 → Agent分析数据 → 返回画像结果
```

**局限**：
- 需要用户主动触发
- 只能回答预设问题
- 无持续观察能力
- 无法主动决策

### 1.2 优化目标

升级为**自主性Agent**，实现感知-决策-行动闭环：

```
数据变化 → Agent感知 → 自主决策 → 自动行动
```

**目标能力**：
- 持续观察数据变化
- 自主判断何时更新画像
- 自动检测偏好变化
- 主动推送匹配推荐

---

## 2. Agent架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                   PreferenceAgent                      │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌─────────────┐    ┌──────────────┐    ┌───────────┐   │
│  │ 感知层      │───▶│ 决策层        │───▶│ 行动层    │   │
│  │ Perceive   │    │ Decide       │    │ Act       │   │
│  └─────────────┘    └──────────────┘    └───────────┘   │
│         │                  │                  │         │
│         ▼                  ▼                  ▼         │
│  触发事件检测          规则+LLM判断        执行决策      │
│  数据变化感知          自主模式识别        更新画像     │
│                        异常情况发现        推送通知     │
│                                          生成推荐     │
│                                                          │
│  ┌─────────────────────────────────────────────────┐   │
│  │              PreferenceState (状态维护)          │   │
│  │  - 当前画像快照                                  │   │
│  │  - 近期变化历史                                  │   │
│  │  - 决策规则配置                                  │   │
│  │  - 行动日志                                      │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### 2.2 组件职责

| 组件 | 职责 | 输入 | 输出 |
|------|------|------|------|
| **Perceiver** | 感知数据变化 | 触发事件 | Perception |
| **Decider** | 决策判断 | Perception + State | Decision |
| **Actor** | 执行行动 | Decision | ActionResult |
| **State** | 维护状态 | 决策历史 | 当前状态快照 |

---

## 3. 触发事件设计

### 3.1 事件类型

```python
from enum import Enum
from dataclasses import dataclass
from typing import Optional, List
from datetime import datetime

class TriggerType(Enum):
    """触发事件类型"""
    # Steam相关
    STEAM_EMAIL_SYNC = "steam_email_sync"      # Steam邮件同步完成
    LIBRARY_SYNC = "library_sync"             # 游戏库同步完成
    PLAYTIME_UPDATE = "playtime_update"        # 游玩时长更新
    NEW_GAME_ADDED = "new_game_added"          # 新增游戏

    # 用户交互
    USER_FEEDBACK = "user_feedback"            # 用户反馈
    GAME_PURCHASED = "game_purchased"          # 用户购买游戏
    GAME_WISHLISTED = "game_wishlisted"        # 添加愿望单

    # 系统任务
    PERIODIC_CHECK = "periodic_check"          # 定时检查
    MANUAL_TRIGGER = "manual_trigger"          # 手动触发
```

### 3.2 事件数据结构

```python
@dataclass
class TriggerEvent:
    """触发事件"""
    type: TriggerType
    timestamp: datetime
    user_id: int
    data: dict  # 事件相关数据

    @classmethod
    def steam_email_sync(cls, user_id: int, email_data: dict):
        return cls(
            type=TriggerType.STEAM_EMAIL_SYNC,
            timestamp=datetime.now(),
            user_id=user_id,
            data=email_data
        )

    @classmethod
    def playtime_update(cls, user_id: int, game_id: str,
                        playtime_delta: int, total_playtime: int):
        return cls(
            type=TriggerType.PLAYTIME_UPDATE,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "game_id": game_id,
                "playtime_delta": playtime_delta,
                "total_playtime": total_playtime
            }
        )

    @classmethod
    def user_feedback(cls, user_id: int, game_id: str,
                     action: str, deal_id: Optional[int] = None):
        return cls(
            type=TriggerType.USER_FEEDBACK,
            timestamp=datetime.now(),
            user_id=user_id,
            data={
                "game_id": game_id,
                "action": action,  # clicked/purchased/ignored/wishlisted
                "deal_id": deal_id
            }
        )
```

---

## 4. 感知层设计

### 4.1 感知器接口

```python
from abc import ABC, abstractmethod
from typing import List, Dict, Any

@dataclass
class Perception:
    """感知结果"""
    summary: str                    # 感知摘要
    key_changes: List[str]          # 关键变化
    metadata: Dict[str, Any]        # 额外数据
    requires_deep_analysis: bool     # 是否需要深度分析

class PreferencePerceiver(ABC):
    """感知器基类"""

    @abstractmethod
    def perceive(self, event: TriggerEvent,
                 current_state: 'PreferenceState') -> Perception:
        """处理事件并生成感知"""
        pass

    def _detect_changes(self, before: dict, after: dict) -> List[str]:
        """检测变化"""
        changes = []
        for key in set(before.keys()) | set(after.keys()):
            delta = after.get(key, 0) - before.get(key, 0)
            if abs(delta) > 0.1:  # 10%变化
                changes.append(f"{key}: {delta:+.1%}")
        return changes
```

### 4.2 感知器实现

```python
class PlaytimePerceiver(PreferencePerceiver):
    """游玩时长感知器"""

    def perceive(self, event: TriggerEvent,
                 current_state: 'PreferenceState') -> Perception:
        if event.type != TriggerType.PLAYTIME_UPDATE:
            return Perception("", [], {}, False)

        data = event.data
        game_id = data["game_id"]
        delta = data["playtime_delta"]

        # 加载游戏信息
        game_info = self._get_game_info(game_id)

        # 检测变化
        key_changes = []
        requires_analysis = False

        if delta > 10:  # 新增超过10小时
            key_changes.append(f"大量游玩《{game_info['name']}》({delta}h)")
            requires_analysis = True

        # 检测标签变化
        for tag in game_info.get("tags", []):
            old_weight = current_state.get_tag_weight(tag)
            new_weight = old_weight + delta * self._tag_weight_factor(tag)
            if abs(new_weight - old_weight) > 0.2:
                key_changes.append(f"标签【{tag}】权重变化显著")

        return Perception(
            summary=f"用户游玩了{game_info['name']}，新增{delta}小时",
            key_changes=key_changes,
            metadata={
                "game_id": game_id,
                "game_name": game_info["name"],
                "tags": game_info.get("tags", []),
                "playtime_delta": delta,
                "genre": game_info.get("genre", "")
            },
            requires_deep_analysis=requires_analysis
        )
```

---

## 5. 决策层设计

### 5.1 决策类型

```python
class DecisionType(Enum):
    """决策类型"""
    NO_ACTION = "no_action"                    # 无需行动
    PROFILE_UPDATE = "profile_update"           # 更新画像
    TAG_WEIGHT_ADJUST = "tag_weight_adjust"    # 调整标签权重
    ANOMALY_DETECTED = "anomaly_detected"      # 异常检测
    PREFERENCE_DRIFT = "preference_drift"      # 偏好漂移
    NEW_PATTERN = "new_pattern"               # 新模式识别
    PUSH_NOTIFICATION = "push_notification"    # 推送通知
    GENERATE_RECOMMENDATION = "generate_rec"   # 生成推荐
    REQUEST_CONFIRMATION = "request_confirm"  # 请求确认

@dataclass
class Decision:
    """决策结果"""
    type: DecisionType
    confidence: float
    reasoning: str
    actions: List[dict]           # 行动列表
    insight: str = ""            # LLM洞察（如果有）

    @property
    def action_needed(self) -> bool:
        return self.type != DecisionType.NO_ACTION
```

### 5.2 规则引擎

```python
class RuleEngine:
    """规则引擎"""

    RULES = {
        # 画像更新规则
        "profile_update": [
            {
                "name": "大量新增游戏",
                "condition": lambda p: len(p.metadata.get("new_games", [])) > 5,
                "decision": DecisionType.PROFILE_UPDATE,
                "priority": 10
            },
            {
                "name": "游玩时长显著变化",
                "condition": lambda p: p.metadata.get("playtime_delta", 0) > 10,
                "decision": DecisionType.PROFILE_UPDATE,
                "priority": 8
            },
            {
                "name": "周期性更新",
                "condition": lambda p: p.metadata.get("is_periodic", False),
                "decision": DecisionType.PROFILE_UPDATE,
                "priority": 5
            }
        ],

        # 异常检测规则
        "anomaly": [
            {
                "name": "单游戏极端游玩",
                "condition": lambda p: p.metadata.get("playtime_delta", 0) > 50,
                "decision": DecisionType.ANOMALY_DETECTED,
                "priority": 20
            },
            {
                "name": "突然大量游玩陌生类型",
                "condition": self._check_unusual_genre,
                "decision": DecisionType.PREFERENCE_DRIFT,
                "priority": 15
            }
        ],

        # 推荐触发规则
        "recommendation": [
            {
                "name": "新促销匹配高偏好",
                "condition": lambda p: p.metadata.get("new_deal_matched", False),
                "decision": DecisionType.GENERATE_RECOMMENDATION,
                "priority": 12
            }
        ]
    }

    def evaluate(self, perception: Perception,
                 state: 'PreferenceState') -> List[Decision]:
        """评估规则并返回决策列表"""
        decisions = []

        for category, rules in self.RULES.items():
            for rule in rules:
                if rule["condition"](perception):
                    decisions.append(Decision(
                        type=rule["decision"],
                        confidence=0.9,
                        reasoning=f"规则触发: {rule['name']}",
                        actions=self._get_actions_for_decision(rule["decision"]),
                        priority=rule["priority"]
                    ))

        # 按优先级排序
        decisions.sort(key=lambda x: x.priority, reverse=True)
        return decisions
```

### 5.3 混合决策器

```python
class PreferenceDecider:
    """
    混合决策器：规则引擎 + LLM判断
    """

    def __init__(self, llm_manager):
        self.rule_engine = RuleEngine()
        self.llm_manager = llm_manager

    def decide(self, perception: Perception,
               state: 'PreferenceState') -> Decision:
        """决策入口"""

        # Step 1: 规则快速判断
        rule_decisions = self.rule_engine.evaluate(perception, state)

        if rule_decisions and rule_decisions[0].confidence >= 0.9:
            # 规则已明确，直接返回
            return rule_decisions[0]

        # Step 2: LLM深度分析（复杂情况）
        if perception.requires_deep_analysis:
            llm_decision = self._llm_analyze(perception, state)
            return self._merge_decisions(rule_decisions, llm_decision)

        # Step 3: 默认决策
        return rule_decisions[0] if rule_decisions else Decision(
            type=DecisionType.NO_ACTION,
            confidence=1.0,
            reasoning="无需行动",
            actions=[]
        )

    def _llm_analyze(self, perception: Perception,
                     state: 'PreferenceState') -> Decision:
        """使用LLM进行深度分析"""

        prompt = f"""分析以下用户行为，决定是否需要更新偏好画像：

## 用户行为
{perception.summary}

## 关键变化
{chr(10).join(f"- {c}" for c in perception.key_changes)}

## 当前画像摘要
{state.get_profile_summary()}

## 游戏详情
- 游戏名称: {perception.metadata.get('game_name')}
- 类型: {perception.metadata.get('genre')}
- 标签: {', '.join(perception.metadata.get('tags', []))}
- 新增游玩时长: {perception.metadata.get('playtime_delta')}小时

请分析：
1. 这是否代表用户偏好的重要变化？
2. 需要更新哪些标签的权重？调整幅度是多少？
3. 是否检测到异常模式？
4. 是否发现新的偏好模式？
5. 是否应该主动推荐相关游戏？

以JSON格式输出决策：
{{
    "type": "profile_update/anomaly_detected/new_pattern/no_action",
    "confidence": 0.0-1.0,
    "reasoning": "分析理由",
    "insight": "关键洞察（中文）",
    "tag_changes": [
        {{"tag": "开放世界", "delta": +0.15}},
        {{"tag": "RPG", "delta": +0.10}}
    ],
    "recommend_action": true/false
}}"""

        response = self.llm_manager.chat(prompt)

        return self._parse_llm_response(response)
```

---

## 6. 行动层设计

### 6.1 行动类型

```python
class ActionType(Enum):
    """行动类型"""
    UPDATE_PROFILE = "update_profile"           # 更新画像
    UPDATE_TAG_WEIGHT = "update_tag_weight"     # 更新标签权重
    LOG_INSIGHT = "log_insight"                # 记录洞察
    PUSH_NOTIFICATION = "push_notification"   # 推送通知
    GENERATE_RECOMMENDATION = "generate_rec"   # 生成推荐
    TRIGGER_SYNC = "trigger_sync"             # 触发同步

@dataclass
class Action:
    """行动定义"""
    type: ActionType
    params: dict
    priority: int = 0
    async_execute: bool = True   # 是否异步执行
```

### 6.2 执行器

```python
class PreferenceActor:
    """偏好Agent执行器"""

    def __init__(self, db, notification_service, llm_manager):
        self.db = db
        self.notification = notification_service
        self.llm_manager = llm_manager

    def execute(self, decision: Decision,
                state: 'PreferenceState') -> List[dict]:
        """执行决策"""
        results = []

        for action_def in decision.actions:
            action_type = action_def["type"]
            params = action_def.get("params", {})

            try:
                result = self._execute_action(action_type, params, decision)
                results.append({
                    "action": action_type,
                    "success": True,
                    "result": result
                })

                # 更新状态
                state.record_action(action_type, result)

            except Exception as e:
                results.append({
                    "action": action_type,
                    "success": False,
                    "error": str(e)
                })

        # 记录洞察
        if decision.insight:
            self._log_insight(decision.insight, state)

        return results

    def _execute_action(self, action_type: ActionType,
                       params: dict, decision: Decision) -> dict:
        """执行单个行动"""

        if action_type == ActionType.UPDATE_TAG_WEIGHT:
            tag_changes = params.get("tag_changes", [])
            for change in tag_changes:
                self.db.update_tag_weight(
                    user_id=params["user_id"],
                    tag=change["tag"],
                    delta=change["delta"]
                )
            return {"updated_tags": len(tag_changes)}

        elif action_type == ActionType.PUSH_NOTIFICATION:
            message = params.get("message", decision.insight)
            return self.notification.send(
                user_id=params["user_id"],
                title="偏好更新提示",
                body=message
            )

        elif action_type == ActionType.GENERATE_RECOMMENDATION:
            # 生成推荐（调用推荐引擎）
            return self._generate_recommendation(params)

        return {}

    def _log_insight(self, insight: str, state: 'PreferenceState'):
        """记录洞察日志"""
        self.db.log_insight(
            user_id=state.user_id,
            insight=insight,
            timestamp=datetime.now()
        )
```

---

## 7. 状态管理

### 7.1 状态类

```python
from typing import Dict, List, Optional
from datetime import datetime

class PreferenceState:
    """偏好Agent状态"""

    def __init__(self, user_id: int):
        self.user_id = user_id
        self.current_profile: Dict[str, float] = {}  # 当前画像
        self.change_history: List[dict] = []          # 变化历史
        self.insights: List[str] = []                 # 洞察记录
        self.last_update: Optional[datetime] = None
        self.anomaly_flags: List[str] = []             # 异常标记

    def get_tag_weight(self, tag: str) -> float:
        """获取标签权重"""
        return self.current_profile.get(tag, 0.0)

    def update_profile(self, tag_changes: List[dict]):
        """更新画像"""
        for change in tag_changes:
            tag = change["tag"]
            delta = change["delta"]
            self.current_profile[tag] = max(0,
                self.current_profile.get(tag, 0) + delta)

        self.last_update = datetime.now()
        self._trim_history()

    def record_action(self, action_type: str, result: dict):
        """记录行动"""
        self.change_history.append({
            "type": action_type,
            "result": result,
            "timestamp": datetime.now()
        })

    def get_profile_summary(self) -> str:
        """获取画像摘要"""
        sorted_tags = sorted(
            self.current_profile.items(),
            key=lambda x: x[1],
            reverse=True
        )[:10]
        return "\n".join(
            f"- {tag}: {weight:.2f}" for tag, weight in sorted_tags
        )

    def _trim_history(self):
        """清理历史记录（保留最近100条）"""
        if len(self.change_history) > 100:
            self.change_history = self.change_history[-100:]
```

---

## 8. Agent主类

### 8.1 完整实现

```python
class PreferenceAgent:
    """
    自主偏好分析Agent
    持续观察数据变化，自主决策并执行行动
    """

    def __init__(self, config: dict):
        self.user_id = config["user_id"]
        self.llm_manager = config["llm_manager"]
        self.db = config["database"]
        self.notification = config["notification_service"]

        # 初始化组件
        self.state = PreferenceState(self.user_id)
        self.perceiver = self._create_perceiver()
        self.decider = PreferenceDecider(self.llm_manager)
        self.actor = PreferenceActor(self.db, self.notification, self.llm_manager)

        # 加载当前画像
        self._load_current_profile()

    def on_event(self, event: TriggerEvent):
        """
        事件入口 - Agent被事件触发
        """
        if event.user_id != self.user_id:
            return

        # 1. 感知
        perception = self.perceiver.perceive(event, self.state)

        # 2. 决策
        decision = self.decider.decide(perception, self.state)

        # 3. 执行行动
        if decision.action_needed:
            results = self.actor.execute(decision, self.state)

            # 4. 更新状态
            self.state.update(perception, decision)

            return {
                "perception": perception.summary,
                "decision": decision.type.value,
                "reasoning": decision.reasoning,
                "insight": decision.insight,
                "actions": len(results)
            }

        return None

    def _create_perceiver(self) -> PreferencePerceiver:
        """创建感知器工厂"""
        perceivers = {
            TriggerType.PLAYTIME_UPDATE: PlaytimePerceiver(self.db),
            TriggerType.STEAM_EMAIL_SYNC: SteamEmailPerceiver(self.db),
            TriggerType.USER_FEEDBACK: UserFeedbackPerceiver(self.db),
            TriggerType.PERIODIC_CHECK: PeriodicPerceiver(self.db),
        }
        return CompositePerceiver(perceivers)

    def _load_current_profile(self):
        """加载当前画像"""
        profile = self.db.get_user_preference_profile(self.user_id)
        if profile:
            self.state.current_profile = profile

    def get_insights(self, limit: int = 10) -> List[dict]:
        """获取近期洞察"""
        return self.db.get_user_insights(self.user_id, limit)
```

---

## 9. 自主性等级

| 等级 | 能力 | 实现复杂度 | 说明 |
|------|------|-----------|------|
| **L1 被动** | 回答用户问题 | ⭐ | 当前实现 |
| **L2 半自动** | 规则触发 + 自动更新 | ⭐⭐ | 仅规则引擎 |
| **L3 条件自主** | 基于规则的自主决策 | ⭐⭐⭐ | 规则 + 行动 |
| **L4 智能自主** | 规则 + LLM混合决策 | ⭐⭐⭐⭐ | 复杂情况用LLM |
| **L5 完全自主** | 完全自主 + 持续学习 | ⭐⭐⭐⭐⭐ | 强化学习优化 |

### 推荐目标

**MVP: L3 (条件自主)**
- 实现规则引擎
- 定义明确触发规则
- 自动画像更新

**增强: L4 (智能自主)**
- 复杂情况用LLM判断
- 生成自然语言洞察
- 异常情况请求确认

---

## 10. 实现路线图

### Phase 1: 基础规则引擎 (1周)

```
目标: 实现L3级别，规则驱动的自主更新

任务:
1. 定义触发事件类型
2. 实现规则引擎
3. 实现感知器
4. 实现执行器
5. 集成到游戏库同步流程

验收:
- 游戏库同步后自动更新画像
- 游玩时长变化 > 10h 自动调整权重
- 有决策日志可追溯
```

### Phase 2: LLM增强 (1周)

```
目标: 实现L4级别，规则+LLM混合决策

任务:
1. 实现LLM决策器
2. 添加洞察生成
3. 异常检测增强
4. 自主推荐触发

验收:
- 复杂情况LLM辅助决策
- 生成可读的洞察描述
- 异常情况标记并通知
```

### Phase 3: 高级功能 (待定)

```
目标: L5级别高级能力

可能的方向:
- 偏好漂移预测
- 时段模式识别（周末vs工作日）
- 多人游戏偏好分析
- 推荐效果预测
```

---

## 11. 风险与控制

### 11.1 风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 过度更新画像 | 画像不稳定 | 中 | 设置更新冷却时间 |
| LLM误判 | 错误洞察 | 中 | 规则兜底，保留人工确认 |
| 行动边界失控 | 意外行为 | 低 | 严格限制行动类型 |
| Token成本过高 | 成本增加 | 中 | 规则优先，LLM仅复杂情况 |

### 11.2 控制措施

```python
# 行动边界控制
ALLOWED_ACTIONS = [
    ActionType.UPDATE_PROFILE,
    ActionType.UPDATE_TAG_WEIGHT,
    ActionType.LOG_INSIGHT,
    ActionType.PUSH_NOTIFICATION,
    ActionType.GENERATE_RECOMMENDATION,
]

# 禁止的行动
FORBIDDEN_ACTIONS = [
    "delete_data",
    "send_external_request",
    "modify_user_settings",
    "financial_transaction",
]

# 更新冷却时间（防止频繁更新）
MIN_UPDATE_INTERVAL = timedelta(hours=2)
```

---

## 12. 总结

### 核心价值

1. **主动性** - 不再等待用户请求，自动观察数据变化
2. **持续性** - 持续分析，而非单次请求
3. **智能性** - 规则 + LLM混合决策，处理复杂情况
4. **透明性** - 决策日志可追溯，用户可理解

### 实现优先级

| 优先级 | 功能 | 价值 |
|--------|------|------|
| P0 | 规则引擎 + 自动更新 | 核心价值 |
| P0 | 游玩时长变化感知 | 高频触发 |
| P1 | LLM决策增强 | 复杂情况 |
| P1 | 洞察生成 | 用户可解释性 |
| P2 | 异常检测 | 质量保障 |

### 下一步行动

1. **评审设计** - 确认架构是否符合预期
2. **实现Phase 1** - 规则引擎 + 基础感知器
3. **集成测试** - 接入游戏库同步流程
4. **用户反馈** - 验证自主性是否带来更好体验

---

*文档版本：v1.0*
*创建日期：2026-04-22*
