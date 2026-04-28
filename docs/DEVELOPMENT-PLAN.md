# 后续开发计划

> 版本：v2.3
> 日期：2026-04-22
> 状态：进行中
> **战略方向：以 Steam 游戏资讯为核心的智能邮件 Agent**

---

## 0. 战略方向调整

### 0.1 定位转变

项目从**通用邮件分类汇总**转向**Steam 游戏资讯智能推荐平台**。

| 维度 | 原定位 | 新定位 |
|------|--------|--------|
| **核心价值** | 邮件智能分类 | Steam 游戏资讯聚合 + 个性化推荐 |
| **用户场景** | 管理所有邮件 | 获取最新游戏咨询、促销活动 |
| **AI 侧重** | 通用分类 | Steam 邮件解析 → 偏好分析 → 智能推荐 |
| **数据来源** | 全部邮件 | Steam 邮件 + Steam Web API |

### 0.2 迭代路径

```
Phase 6.1  Steam邮件解析          ← 当前优先级最高
    │       识别促销/资讯，提取游戏信息
    ▼
Phase 6.2  Steam数据集成
    │       绑定Steam账号，获取30天游玩记录
    ▼
Phase 6.3  用户偏好分析
    │       LLM分析游戏偏好，构建用户画像
    ▼
Phase 6.4  智能游戏推荐
    │       匹配资讯与偏好，生成个性化推荐
    ▼
Phase 6.5  推荐反馈闭环
            用户反馈优化，推荐效果迭代
```

### 0.3 可行性评估

> 详细分析见 [STEAM-RECOMMENDATION-FEASIBILITY.md](./STEAM-RECOMMENDATION-FEASIBILITY.md)

| 模块 | 可行度 | 核心依赖 | 风险点 |
|------|--------|----------|--------|
| Steam邮件解析 | ⭐⭐⭐⭐⭐ (95%) | 现有IMAP + LLM | HTML邮件模板多变 |
| Steam数据集成 | ⭐⭐⭐⭐ (80%) | Steam Web API + 用户API Key | API限流、需用户配置 |
| 用户偏好分析 | ⭐⭐⭐⭐⭐ (90%) | LLM + 游玩数据 | 偏好标签覆盖率 |
| 智能游戏推荐 | ⭐⭐⭐⭐⭐ (95%) | 偏好画像 + 促销数据 | 推荐多样性 |

**总体可行性：⭐⭐⭐⭐ (85/100)，预计工期：6-10周**

---

## 1. 当前进展

| Phase | 内容 | 状态 | 说明 |
|-------|------|------|------|
| Phase 1 | 三端项目结构搭建 | ✅ 已完成 | Go/Python/React 项目骨架 |
| Phase 2 | 126邮箱IMAP接入 | ✅ 已完成 & 已测试 | IMAP Provider、ID命令、字符集处理（2026-04-21通过实际邮箱测试） |
| Phase 3 | Go后端API | ✅ 基本完成 | Clean Architecture、CRUD、Agent通信 |
| Phase 4 | Agent分类/提取/摘要 | ✅ 基本完成 | LLM集成（智谱/DeepSeek）、Mock测试 |
| Phase 5 | React前端核心页面 | ✅ 基本完成 | 列表、详情、设置、仪表盘 |
| Phase 6 | Steam游戏推荐系统 | 🚀 **核心方向** | 邮件解析 → 偏好分析 → 智能推荐 |
| └ 6.1 | Steam邮件解析 | ✅ 已完成 | 正则预筛选、HTML解析、游戏信息提取 |
| └ 6.2 | Steam数据集成 | ✅ 已完成 | Web API客户端、账号绑定、游戏库同步 |
| └ 6.3 | 用户偏好分析 | ✅ 已完成 | 感知-决策-行动闭环、偏好画像、洞察日志 |
| └ 6.4 | 智能游戏推荐 | ✅ 已完成 | 匹配度算法、推荐理由生成、用户反馈 |

### 已实现功能清单

**Go后端 (email-backend)**
- 邮件CRUD + 分类触发 + 状态管理
- 账户管理（添加/删除/测试连接）
- 邮件同步（手动触发 + 并发同步 + 自动分类）
- Agent HTTP客户端
- AES-256-GCM凭证加密
- **用户认证系统（JWT登录注册、多用户隔离）**
- **126邮箱IMAP Provider（已通过实际邮箱测试）**
  - IMAP ID命令支持（避免Unsafe Login）
  - 中文字符集自动解码
  - 连接重试机制（最多3次）
  - 并发安全保护（sync.Mutex）
  - 邮件列表搜索和分页
  - 邮件正文解析（text/plain和text/html）

**Python Agent (email-agent)**
- 单封/批量邮件分类
- 信息提取（行动项、会议、实体）
- 每日摘要生成
- LLM Manager（智谱GLM、DeepSeek、Mock）
- **正则预筛选分类引擎（`classify_rules.py`）**
  - Steam邮件快速识别（发件人域名 + 主题关键词）
  - 7类普通邮件正则规则覆盖
  - 高置信度直接返回，减少LLM调用成本
  - 低置信度降级到LLM，支持参考预判结果
- **Steam信息提取Agent**（HTML邮件解析、游戏信息结构化提取）

**React前端 (email-web)**
- 仪表盘（统计概览）
- 邮件列表（筛选、分页、关键词搜索）
- 邮件详情（分类信息、正文展示）
- 设置页（账户管理、同步触发）
- **登录/注册页面、路由守卫、用户菜单**

---

## 2. 待开发任务

### P0 - Steam核心功能（最高优先级）

> 项目核心方向，直接决定产品价值

#### 2.1 Steam邮件解析（Phase 6.1）

| 项目 | 内容 |
|------|------|
| **模块** | Backend + Agent |
| **优先级** | **P0（最高）** |
| **目标** | 识别并提取Steam邮件中的游戏信息 |
| **工期** | 第1-2周 |

**后端任务：**

- [x] 扩展邮件分类类别
  - `steam_promotion` (促销邮件)
  - `steam_wishlist` (愿望单通知)
  - `steam_news` (游戏资讯)
  - `steam_update` (游戏更新)
- [x] 数据库表创建
  - `steam_games` 表（游戏基础信息：名称、开发商、标签、类型）
  - `steam_deals` 表（促销信息：原价、折扣价、折扣率、起止日期）
- [x] API端点
  - `GET /api/v1/steam/emails` - 获取Steam分类邮件列表
  - `GET /api/v1/steam/games` - 获取已提取的游戏列表
  - `GET /api/v1/steam/deals` - 获取当前促销列表（支持筛选排序）
  - `GET /api/v1/steam/deals/:id` - 获取促销详情
- [x] **同步后自动触发Steam信息提取**（`sync_service.go`）

**Agent任务：**

- [x] Steam信息提取Agent（`steam_extract_service.py`）
  - 解析HTML邮件结构（适配Steam邮件模板）
  - 提取游戏名称、原价、折扣价、折扣率
  - 识别游戏标签和类型
  - 识别促销截止日期
- [x] Steam分类Prompt更新
  - 增加Steam邮件分类规则
  - 提取结果结构化输出（JSON Schema）
- [x] **正则预筛选引擎**（`classify_rules.py`）
  - Steam发件人域名匹配
  - 主题关键词快速识别
  - Steam子分类细分（促销/愿望单/资讯/更新）

**前端任务：**

- [ ] Steam邮件列表页（与普通邮件列表区分）
- [ ] 促销信息卡片展示（游戏封面、价格、折扣率）
- [ ] 促销时间倒计时

**涉及文件：**

```
email-backend/server/
├── model/steam.go                  # 已实现
├── model/preference.go             # 已实现（偏好数据模型）
├── repository/steam_repo.go         # 已实现
├── repository/preference_repo.go    # 已实现（偏好仓库）
├── service/steam_service.go         # 已实现
└── api/v1/steam.go                  # 已实现

email-agent/app/
├── prompts/classify_rules.py        # 已实现（正则预筛选引擎）
├── prompts/steam_extraction.py      # 已实现
├── services/classify_service.py     # 已实现（集成正则筛选）
└── services/steam_extract_service.py # 已实现

email-web/src/
├── pages/SteamDeals.tsx             # 已实现
├── components/DealCard.tsx          # 已实现
└── api/steamApi.ts                  # 已实现
```

**验收标准：**
- 收到Steam促销邮件后，自动提取游戏信息并存储
- 游戏名称、价格、折扣提取准确率 > 90%
- 前端可展示促销列表，支持按折扣率/价格排序

---

#### 2.2 Steam数据集成（Phase 6.2）

| 项目 | 内容 |
|------|------|
| **模块** | Backend + Web |
| **优先级** | **P0** |
| **目标** | 获取用户Steam 30天游玩记录 |
| **工期** | 第3周 |
| **前置依赖** | Phase 6.1 完成 |

**Steam Web API能力：**

| API | 端点 | 功能 | 说明 |
|-----|------|------|------|
| GetOwnedGames | `IPlayerService/GetOwnedGames` | 用户游戏库 | 含总游玩时长 |
| GetRecentlyPlayedGames | `IPlayerService/GetRecentlyPlayedGames` | 最近2周游玩 | 含近期时长 |
| GetPlayerSummaries | `ISteamUser/GetPlayerSummaries` | 用户资料 | 头像、昵称 |
| GetSchemaForGame | `ISteamUserStats/GetSchemaForGame` | 游戏成就 | 可选 |

**后端任务：**

- [x] Steam Web API客户端（`pkg/steam/client.go`）
  - HTTP客户端封装
  - 请求限流（300次/分钟以内）
  - 响应缓存（游戏元数据本地缓存24h）
  - 错误处理和重试
  - **默认Mock模式，无需真实API Key**
- [x] Steam账号管理
  - 用户绑定Steam ID
  - 系统使用统一API Key（或用户自备）
  - 验证账号有效性
- [x] 数据同步服务
  - 定期同步用户游戏库（每日）
  - 同步最近游玩记录
  - 同步游戏元数据到 `steam_games` 表（标签、类型、封面图）
- [x] API端点
  - `POST /api/v1/steam/bind` - 绑定Steam账号
  - `GET /api/v1/steam/profile` - 获取Steam资料
  - `DELETE /api/v1/steam/unbind` - 解绑账号
  - `GET /api/v1/steam/library` - 获取游戏库
  - `GET /api/v1/steam/library/recent` - 获取最近游玩
  - `POST /api/v1/steam/sync` - 手动同步

**前端任务：**

- [x] Steam游戏库页面（账号绑定/解绑、游戏库展示）
- [x] 游戏库列表展示（名称、游玩时长、最近游玩时间）
- [x] 近两周游玩卡片展示

**涉及文件：**

```
email-backend/server/
├── pkg/steam/
│   ├── client.go               # Steam API客户端（支持Mock）
│   └── mock.go                  # Mock测试数据
├── model/steam_account.go      # Steam账号和游戏库模型
├── service/steam_service.go    # 扩展（账号绑定、游戏库同步）
└── api/v1/steam.go             # 扩展（新增API端点）

email-web/src/
├── pages/SteamLibrary.tsx      # 游戏库页面
├── pages/SteamDeals.tsx        # 促销页面（更新）
├── api/steamApi.ts             # 扩展
└── components/layout/AppLayout.tsx # Steam导航下拉菜单
```

**验收标准：**
- 用户可绑定Steam账号（输入Steam ID）
- 显示最近30天游玩的游戏列表及游玩时长
- 游戏元数据（标签、类型）完整度 > 80%

> **注意**：当前使用Mock模式，绑定任意Steam ID即可获得示例游戏库数据

---

#### 2.3 用户偏好分析（Phase 6.3）

| 项目 | 内容 |
|------|------|
| **模块** | Agent + Backend |
| **优先级** | **P0** |
| **目标** | **自主性Agent** - 持续观察数据变化，自主决策并执行行动 |
| **工期** | 第4周 |
| **前置依赖** | Phase 6.2 完成 |
| **设计文档** | [PREFERENCE-AGENT-AUTONOMOUS-DESIGN.md](./PREFERENCE-AGENT-AUTONOMOUS-DESIGN.md) |

> **核心升级**: 从被动问答升级为感知-决策-行动闭环（Perceive-Decide-Act Loop）

**自主性设计：**

```
┌─────────────────────────────────────────────────────────┐
│                  PreferenceAgent                        │
│                                                          │
│  感知层(Perceive) ──▶ 决策层(Decide) ──▶ 行动层(Act)     │
│       │                  │                   │           │
│  数据变化感知          规则+LLM判断         更新画像     │
│  触发事件检测          自主模式识别         推送通知     │
│                        异常情况发现         生成推荐     │
└─────────────────────────────────────────────────────────┘
```

**自主性等级目标：**

| 等级 | 能力 | 实现 |
|------|------|------|
| **L3 条件自主** | 基于规则的自主决策 | 规则引擎 + 自动画像更新 |
| **L4 智能自主** | 规则 + LLM混合决策 | 复杂情况用LLM判断 |

**后端任务：**

- [x] 用户偏好数据结构
  - `user_game_preferences` 表（偏好标签及权重）
  - `recommendation_feedback` 表（推荐反馈记录）
  - `model/preference.go` + `repository/preference_repo.go`
- [x] 自主性偏好分析Agent集成
  - 偏好分析触发器（`PreferenceAgent`）
  - 触发事件定义（Steam邮件同步、游玩时长更新、用户反馈等）
  - 规则引擎决策
  - 自动画像更新执行
- [x] API端点
  - `GET /api/v1/steam/profile/preference` - 获取偏好画像
  - `POST /api/v1/steam/profile/analyze` - 重新分析
  - `GET /api/v1/steam/profile/insights` - 获取Agent洞察记录

**Agent任务：**

- [x] 偏好分析Agent（`PreferenceAgent`类）
  - **感知器（Perceiver）**
    - `PlaytimePerceiver` - 游玩时长变化感知
    - `SteamEmailPerceiver` - Steam邮件变化感知
    - `UserFeedbackPerceiver` - 用户反馈感知
    - `PeriodicPerceiver` - 定时检查感知
  - **决策器（Decider）**
    - `RuleEngine` - 规则快速判断
    - `PreferenceDecider` - 规则+LLM混合决策
  - **执行器（Actor）**
    - `PreferenceActor` - 决策执行
  - **状态管理（State）**
    - `PreferenceState` - 画像快照、变化历史、洞察记录
- [x] 触发事件系统
  - `TriggerEvent` - 事件数据结构
  - `TriggerType` - 事件类型枚举
- [x] 自主行动能力
  - 自动检测偏好变化并更新画像
  - 异常模式识别（极端游玩、新类型探索）
  - 生成自然语言洞察
  - 主动推送匹配推荐

**前端任务：**

- [x] 偏好画像展示页
  - 偏好标签云可视化
  - Top类型、风格展示
  - 偏好权重雷达图
- [x] Agent洞察日志展示
  - 决策理由展示
  - 变化历史时间线

**涉及文件：**

```
email-backend/server/
├── model/user_gaming_profile.go    # 新增
├── repository/profile_repo.go      # 新增
├── service/preference_service.go   # 新增
├── api/v1/preference.go           # 新增
├── model/trigger_event.go          # 新增（触发事件模型）
├── model/preference_insight.go     # 新增（洞察记录模型）

email-agent/app/
├── agents/preference_agent.py      # 新增（主Agent类）
├── agents/perceiver/               # 新增（感知器目录）
│   ├── base.py                      # 感知器基类
│   ├── playtime.py                  # 游玩感知器
│   ├── steam_email.py               # Steam邮件感知器
│   └── user_feedback.py             # 用户反馈感知器
├── agents/decider/                 # 新增（决策器目录）
│   ├── rule_engine.py               # 规则引擎
│   └── llm_decider.py              # LLM决策器
├── agents/actor.py                  # 新增（执行器）
├── agents/state.py                  # 新增（状态管理）
└── models/trigger_event.py         # 新增（事件数据模型）

email-web/src/
├── pages/PreferenceAnalysis.tsx    # 新增
├── components/PreferenceChart.tsx   # 新增
├── components/AgentInsightLog.tsx  # 新增（洞察日志组件）
└── api/steamApi.ts                 # 扩展
```

**验收标准：**
- 游戏库同步后自动触发偏好分析
- 自动生成用户偏好画像，覆盖率 > 85%
- Agent决策可追溯（洞察日志、决策理由）
- 异常模式自动检测并通知用户
- 用户反馈驱动画像动态更新

---

#### 2.4 智能游戏推荐（Phase 6.4）

| 项目 | 内容 |
|------|------|
| **模块** | Backend + Agent + Web |
| **优先级** | **P0** |
| **目标** | **自主推荐触发** - Agent主动匹配促销与偏好，生成个性化推荐 |
| **工期** | 第5-6周 |
| **前置依赖** | Phase 6.3 完成 |

> **与自主性Agent整合**: 推荐生成作为Agent的执行能力之一，可主动触发

**后端任务：**

- [x] 推荐算法引擎
  - 标签匹配度计算（用户偏好标签 vs 游戏标签）
  - 多维度评分（类型、风格、开发商历史偏好）
  - 个性化排序算法
  - 过滤已拥有游戏
- [x] 推荐结果存储（`game_recommendations` 表）
  - 匹配分数、推荐理由
  - 用户反馈追踪（like/dislike/ignore）
- [x] **自主推荐触发机制**
  - 新促销邮件到达时自动匹配偏好
  - 偏好变化后重新生成推荐
  - 定时检查高匹配度促销
- [x] API端点
  - `GET /api/v1/recommendations` - 获取推荐列表
  - `GET /api/v1/recommendations/deals` - 仅推荐促销游戏
  - `POST /api/v1/recommendations/:id/feedback` - 用户反馈

**Agent任务：**

- [x] 推荐生成作为Agent执行能力
  - 集成到 `PreferenceActor` 的 `ActionType.GENERATE_RECOMMENDATION`
  - 规则触发：`新促销匹配高偏好` → 自动生成推荐
  - LLM辅助生成推荐理由
- [x] 推荐理由生成（`recommendation_reason.py`）
  - LLM生成个性化推荐理由
  - 结合用户游玩历史和偏好
  - 强调匹配点（如："基于你最近大量游玩《艾尔登法环》..."）
- [x] 主动推送决策
  - 高匹配度促销（>80%）自动推送通知
  - 中等匹配度（50-80%）纳入推荐列表
  - 低匹配度过滤

**前端任务：**

- [x] 推荐列表页面（卡片式布局）
  - 游戏封面 + 价格 + 折扣
  - 匹配度评分可视化
  - 推荐理由展开/收起
- [x] 用户反馈交互（点赞/点踩）
- [x] 推荐筛选（仅促销/全部/按类型）

**涉及文件：**

```
email-backend/server/
├── model/recommendation.go           # 新增
├── repository/recommendation_repo.go # 新增
├── service/recommendation_service.go # 新增
└── api/v1/recommendation.go          # 新增

email-agent/app/
├── agents/recommendation.py          # 新增
└── prompts/recommendation_reason.py  # 新增

email-web/src/
├── pages/Recommendations.tsx         # 新增
├── components/RecommendationCard.tsx # 新增
└── api/recommendationApi.ts          # 新增
```

**验收标准：**
- 生成个性化推荐列表（非已拥有游戏）
- 推荐理由自然流畅、有说服力、结合用户历史
- 用户满意度 > 70%
- 推荐多样性 > 60%（不只推荐同一类型）

---

#### 2.5 推荐反馈闭环（Phase 6.5）

| 项目 | 内容 |
|------|------|
| **模块** | Backend + Agent |
| **优先级** | **P1** |
| **目标** | 用户反馈驱动推荐优化 |
| **工期** | 第7周 |
| **前置依赖** | Phase 6.4 完成 |

**任务：**

- [ ] 反馈数据收集
  - 点赞/点踩记录存储
  - 忽略行为追踪
  - 点击/查看率统计
- [ ] 偏好画像动态更新
  - 正反馈增强相关标签权重
  - 负反馈降低标签权重
  - 定期重新分析偏好人
- [ ] 推荐多样性策略
  - 引入"惊喜推荐"（低匹配但可能有兴趣）
  - 探索-利用平衡（Exploration vs Exploitation）
- [ ] 推荐效果指标
  - 点击率 (CTR)
  - 正反馈率
  - 推荐覆盖率

**验收标准：**
- 用户反馈自动影响后续推荐
- 推荐结果随反馈逐步改善
- 无"信息茧房"问题（推荐多样性不下降）

---

### P1 - 基础功能完善

> 支撑Steam核心功能所需的基础能力

#### 2.6 LLM配置管理

| 项目 | 内容 |
|------|------|
| **模块** | Backend + Web + Agent |
| **现状** | Agent的LLM配置硬编码在环境变量，用户无法自定义 |
| **目标** | 每个用户可配置自己的LLM Provider和API Key |

**后端任务：**

- [ ] LLM Config Model（`llm_configs`表已定义）
- [ ] API端点：
  - `GET /api/v1/llm-configs` - 获取用户LLM配置
  - `POST /api/v1/llm-configs` - 添加LLM配置
  - `PUT /api/v1/llm-configs/:id` - 更新配置
  - `DELETE /api/v1/llm-configs/:id` - 删除配置
- [ ] 调用Agent时传递用户LLM配置

**Agent任务：**

- [ ] 接受动态LLM配置参数（覆盖默认配置）
- [ ] 按用户配置路由到不同LLM Provider

**前端任务：**

- [ ] 设置页增加"LLM配置"Tab
- [ ] LLM配置表单（Provider选择、API Key输入、模型选择）
- [ ] API Key掩码显示和编辑

**涉及文件：**

```
email-backend/server/
├── api/v1/llm_config.go           # 新增
├── model/llm_config.go            # 新增
├── repository/llm_config_repo.go  # 新增
├── service/llm_config_service.go  # 新增

email-agent/app/
├── llm/manager.py                 # 修改（支持动态配置）
├── api/routes/classify.py         # 修改（接受配置参数）

email-web/src/
├── components/LLMConfigForm.tsx   # 新增
└── api/llmApi.ts                  # 新增
```

---

#### 2.7 头像存储迁移（待办）

| 项目 | 内容 |
|------|------|
| **模块** | Backend + Web |
| **现状** | 头像图片存储在本地 `./uploads/avatars/` 目录，通过Nginx/静态文件服务访问 |
| **目标** | 迁移到云存储（COS/OSS），支持CDN加速，全球可访问 |
| **优先级** | P3（后续阶段） |

**现状实现：**
- 头像上传到 `uploads/avatars/` 目录
- 通过 `r.Static("/uploads", "./uploads")` 提供访问
- Vite开发环境通过代理配置访问

**待升级内容：**
- 接入腾讯云COS或阿里云OSS
- 上传接口改为返回云存储URL
- 支持CDN域名配置
- 考虑头像裁剪和压缩优化

**后端修改：**
- 集成云SDK（cos-go-sdk 或 oss-sdk）
- 新增头像上传服务层
- 保留本地存储作为fallback

**前端修改：**
- avatar_url 改为完整CDN地址
- 移除Vite代理配置（Nginx处理）

---

---

## 3. 开发路线图

```
2026-04 下旬 ──────────────────────────────────────────────
  │
  │  ★ 第一阶段：Steam邮件解析（P0 - 核心）
  │  ├── 扩展邮件分类类别（steam_promotion/wishlist/news）
  │  ├── Steam信息提取Agent（HTML解析、游戏信息提取）
  │  ├── 游戏信息存储（steam_games + steam_deals表）
  │  └── 前端促销列表展示
  │
2026-05 上旬 ──────────────────────────────────────────────
  │
  │  ★ 第二阶段：Steam数据集成（P0 - 核心）
  │  ├── Steam Web API客户端
  │  ├── Steam账号绑定/解绑
  │  ├── 用户游戏库同步（30天游玩记录）
  │  └── 前端游戏库展示
  │
2026-05 中旬 ──────────────────────────────────────────────
  │
  │  ★ 第三阶段：偏好分析（自主性Agent）+ 智能推荐（P0 - 核心）
  │  ├── 感知-决策-行动闭环实现（PreferenceAgent）
  │  ├── 规则引擎 + LLM混合决策
  │  ├── 自动画像更新 + 异常检测
  │  ├── 自主推荐触发
  │  ├── 推荐算法引擎
  │  ├── LLM推荐理由生成
  │  └── 前端偏好画像 + 推荐列表
  │
2026-05 下旬 ──────────────────────────────────────────────
  │
  │  第四阶段：反馈闭环 + 基础完善（P1）
  │  ├── 推荐反馈闭环
  │  ├── LLM配置管理
  │  ├── Action Items API
  │  └── Orchestrator编排器
  │
2026-06 ──────────────────────────────────────────────
  │
  │  第五阶段：体验优化 + 生产部署（P2/P3）
  │  ├── 统计分析（Steam游戏维度）
  │  ├── Redis异步队列
  │  ├── Docker Compose
  │  ├── Nginx + 安全加固
  │  └── 通义千问Provider
  │
  ▼
上线
```

---

## 4. 验收标准

### 已完成验收

| 序号 | 验收项 | 状态 | 通过标准 |
|------|--------|------|---------|
| 1 | 用户注册登录 | ✅ 已完成 | 新用户可注册、登录后看到自己的数据 |
| 2 | 数据隔离 | ✅ 已完成 | 不同用户只能看到自己的邮件和账户 |
| 3 | 126邮箱连接 | ✅ 已完成 | IMAP连接、ID命令、中文字符集解码正常 |
| 4 | 邮件同步 | ✅ 已完成 | 手动触发同步 + 定时自动同步 |
| 5 | Agent分类/摘要 | ✅ 已完成 | 智谱GLM/DeepSeek集成、每日摘要生成 |

### 第一阶段验收：Steam邮件解析

| 序号 | 验收项 | 状态 | 通过标准 |
|------|--------|------|---------|
| 6 | Steam邮件识别 | ✅ 已完成 | 正则预筛选 + LLM双重识别，准确率 > 90% |
| 7 | 游戏信息提取 | ✅ 已完成 | 正确提取游戏名称、价格、折扣率 |
| 8 | 促销展示 | ✅ 已完成 | 前端展示促销列表，支持按折扣/价格排序 |
| 9 | 自动触发提取 | ✅ 已完成 | Steam分类后自动异步提取游戏信息 |

### 第二阶段验收：Steam数据集成

| 序号 | 验收项 | 状态 | 通过标准 |
|------|--------|------|---------|
| 10 | Steam绑定 | ✅ 已完成 | 用户可输入Steam ID绑定账号（Mock模式） |
| 11 | 游戏库同步 | ✅ 已完成 | 显示最近30天游玩游戏及游玩时长 |
| 12 | 元数据完整 | ✅ 已完成 | 游戏标签、类型完整度 > 80%（Mock数据） |

### 第三阶段验收：偏好分析 + 推荐（自主性Agent）

| 序号 | 验收项 | 状态 | 通过标准 |
|------|--------|------|---------|
| 12 | **自主感知** | ✅ 已完成 | 游戏库同步后自动触发感知，不需用户手动 |
| 13 | **规则决策** | ✅ 已完成 | 游玩时长>10h自动调整权重，有日志可查 |
| 14 | **LLM增强决策** | ✅ 已完成 | 复杂情况LLM分析，生成可读洞察 |
| 15 | **异常检测** | ✅ 已完成 | 极端游玩/新类型探索自动标记 |
| 16 | 偏好画像 | ✅ 已完成 | 自动生成偏好标签，覆盖率 > 85% |
| 17 | 个性化推荐 | ✅ 已完成 | 推荐非已拥有游戏，理由自然有说服力 |
| 18 | 用户满意度 | 部分完成 | 推荐列表生成，用户可反馈（待实际数据验证） |

### 第四阶段验收：反馈闭环 + 基础完善

| 序号 | 验收项 | 通过标准 |
|------|--------|---------|
| 19 | 反馈优化 | 用户反馈影响后续推荐，无信息茧房 |
| 20 | LLM配置 | 用户可切换LLM Provider，配置生效 |
| 21 | Action Items | 邮件详情页显示提取的行动项，可标记完成 |

### 第五阶段验收：生产部署

| 序号 | 验收项 | 通过标准 |
|------|--------|---------|
| 22 | Docker部署 | `docker-compose up -d` 一键启动全部服务 |
| 23 | HTTPS访问 | 通过Nginx提供HTTPS访问 |
| 24 | 安全审计 | 无明文密钥、API限流生效、日志可追溯 |

---

## 5. 风险评估

### Steam方向特有风险

| 风险项 | 影响 | 概率 | 缓解措施 |
|--------|------|------|----------|
| Steam邮件HTML模板频繁变动 | 高 | 中 | LLM容错 + 多模板适配策略 |
| Steam Web API限流 | 中 | 低 | 本地缓存24h、批量定时同步 |
| 游戏元数据缺失 | 中 | 中 | Steam Store API + 多源补全（IGDB） |
| 用户无Steam账号 | 高 | 低 | 不影响原有邮件分类功能 |
| Steam邮件数量不足 | 中 | 中 | 支持关注Curator获取更多资讯 |
| 推荐同质化 | 低 | 中 | 多样性策略 + 惊喜推荐机制 |

### 自主性Agent风险

| 风险项 | 影响 | 概率 | 缓解措施 |
|--------|------|------|----------|
| 过度更新画像 | 画像不稳定 | 中 | 设置更新冷却时间（MIN_UPDATE_INTERVAL） |
| LLM误判 | 错误洞察 | 中 | 规则兜底，保留人工确认机制 |
| 行动边界失控 | 意外行为 | 低 | 严格限制ActionType白名单 |
| Token成本过高 | 成本增加 | 中 | 规则优先，LLM仅复杂情况调用 |
| 洞察质量不一 | 用户困惑 | 中 | 洞察格式标准化 + 人工审核 |

### 通用技术风险

| 风险项 | 影响 | 缓解措施 |
|--------|------|----------|
| 126邮箱API限制 | 中 | 标准IMAP协议，控制请求频率 |
| LLM服务不稳定 | 中 | 多Provider自动切换 |
| 凭证泄露 | 高 | AES-256-GCM加密，密钥隔离管理 |
| 数据丢失 | 中 | 定期备份MySQL数据 |

---

*文档版本：v2.3*
*最后更新：2026-04-22*
