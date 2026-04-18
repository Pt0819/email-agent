# 后续开发计划

> 版本：v1.0
> 日期：2026-04-18
> 状态：规划中

---

## 1. 当前进展

| Phase | 内容 | 状态 | 说明 |
|-------|------|------|------|
| Phase 1 | 三端项目结构搭建 | ✅ 已完成 | Go/Python/React 项目骨架 |
| Phase 2 | 126邮箱IMAP接入 | ✅ 已完成 | IMAP Provider、ID命令、字符集处理 |
| Phase 3 | Go后端API | ✅ 基本完成 | Clean Architecture、CRUD、Agent通信 |
| Phase 4 | Agent分类/提取/摘要 | ✅ 基本完成 | LLM集成（智谱/DeepSeek）、Mock测试 |
| Phase 5 | React前端核心页面 | ✅ 基本完成 | 列表、详情、设置、仪表盘 |

### 已实现功能清单

**Go后端 (email-backend)**
- 邮件CRUD + 分类触发 + 状态管理
- 账户管理（添加/删除/测试连接）
- 邮件同步（手动触发 + 并发同步 + 自动分类）
- Agent HTTP客户端
- AES-256-GCM凭证加密

**Python Agent (email-agent)**
- 单封/批量邮件分类
- 信息提取（行动项、会议、实体）
- 每日摘要生成
- LLM Manager（智谱GLM、DeepSeek、Mock）

**React前端 (email-web)**
- 仪表盘（统计概览）
- 邮件列表（筛选、分页、关键词搜索）
- 邮件详情（分类信息、正文展示）
- 设置页（账户管理、同步触发）

---

## 2. 待开发任务

### P0 - 核心缺失功能

> 缺少这些功能，系统无法安全地多用户使用

#### 2.1 用户认证系统

| 项目 | 内容 |
|------|------|
| **模块** | Backend + Web |
| **现状** | 所有Handler硬编码 `userID = 1`，无鉴权中间件 |
| **目标** | JWT登录注册，多用户隔离 |

**后端任务：**

- [ ] 用户模型（`users`表已定义在`sql/init.sql`）
- [ ] 注册/登录API：`POST /api/v1/auth/register`、`POST /api/v1/auth/login`
- [ ] JWT中间件：Token生成、验证、续期
- [ ] 所有Handler替换硬编码userID，改为从Token上下文获取
- [ ] CORS中间件完善

**前端任务：**

- [ ] 登录页面 `Login.tsx`
- [ ] 注册页面 `Register.tsx`
- [ ] Token管理（localStorage存储、Axios拦截器自动携带）
- [ ] 路由守卫（未登录跳转登录页）

**涉及文件：**

```
email-backend/server/
├── api/v1/auth.go              # 新增
├── middleware/auth.go           # 新增
├── model/user.go               # 新增
├── repository/user_repo.go     # 新增
├── service/user_service.go     # 新增
└── router/router.go            # 修改

email-web/src/
├── pages/Login.tsx             # 新增
├── pages/Register.tsx          # 新增
├── api/authApi.ts              # 新增
└── components/ProtectedRoute.tsx  # 新增
```

---

#### 2.2 Action Items API + UI

| 项目 | 内容 |
|------|------|
| **模块** | 全栈 |
| **现状** | `action_items`表已定义，Agent提取逻辑已实现，缺后端API和前端展示 |
| **目标** | 展示提取的行动项，支持状态管理（待办/完成） |

**后端任务：**

- [ ] ActionItem Model + Repository
- [ ] API端点：
  - `GET /api/v1/emails/:id/actions` - 获取邮件关联行动项
  - `GET /api/v1/actions` - 获取用户所有行动项（按状态/优先级筛选）
  - `PUT /api/v1/actions/:id` - 更新行动项状态
  - `DELETE /api/v1/actions/:id` - 删除行动项
- [ ] Agent提取结果自动写入action_items表

**前端任务：**

- [ ] 邮件详情页增加"行动项"卡片
- [ ] 行动项列表页（可筛选、可标记完成）
- [ ] 仪表盘增加待办事项概览

**涉及文件：**

```
email-backend/server/
├── api/v1/action.go            # 新增
├── model/action_item.go        # 新增
├── repository/action_repo.go   # 新增
├── service/action_service.go   # 新增

email-web/src/
├── pages/ActionItems.tsx       # 新增
├── components/ActionCard.tsx   # 新增
└── api/actionApi.ts            # 新增
```

---

#### 2.3 LLM配置管理

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

### P1 - 智能功能增强

> 提升系统的智能化水平，从基础分类走向语义理解

#### 2.4 ChromaDB向量存储

| 项目 | 内容 |
|------|------|
| **模块** | Agent |
| **现状** | `vectorstore/`目录为空，无向量存储能力 |
| **目标** | 邮件内容向量化存储，支持语义相似检索 |

**任务：**

- [ ] ChromaDB客户端封装（连接管理、Collection管理）
- [ ] 邮件内容向量化存储（分类后自动入库）
- [ ] 语义相似邮件检索API：`POST /api/v1/search/similar`
- [ ] 相关邮件推荐（查看某封邮件时推荐相似邮件）

**涉及文件：**

```
email-agent/app/
├── vectorstore/
│   ├── __init__.py             # 新增
│   └── chroma_client.py        # 新增
├── embeddings/
│   ├── __init__.py             # 新增
│   └── bge.py                  # 新增
└── api/routes/search.py        # 新增
```

---

#### 2.5 BGE Embedding集成

| 项目 | 内容 |
|------|------|
| **模块** | Agent |
| **现状** | 未实现 |
| **目标** | 使用BAAI/bge-m3模型对中文邮件内容向量化 |

**任务：**

- [ ] BGE-M3模型加载（支持本地部署和API调用）
- [ ] 邮件内容Embedding生成
- [ ] 批量向量化处理
- [ ] 向量维度和相似度阈值配置

---

#### 2.6 Orchestrator编排器

| 项目 | 内容 |
|------|------|
| **模块** | Agent |
| **现状** | `agents/`目录为空，各Agent独立调用 |
| **目标** | 统一编排 分类→提取→向量化 完整流水线 |

**任务：**

- [ ] Orchestrator实现（Pipeline模式）
- [ ] 完整处理流程：`POST /api/v1/process` → 分类 → 提取行动项 → 向量化存储
- [ ] 步骤失败处理和重试策略
- [ ] 处理进度回调

**涉及文件：**

```
email-agent/app/
├── agents/
│   ├── __init__.py             # 新增
│   ├── orchestrator.py         # 新增
│   ├── classification.py       # 新增（从service迁移）
│   ├── extraction.py           # 新增（从service迁移）
│   └── summary.py              # 新增（从service迁移）
```

---

#### 2.7 周报摘要API

| 项目 | 内容 |
|------|------|
| **模块** | Backend + Agent |
| **现状** | 仅实现每日摘要 |
| **目标** | 支持周报摘要生成 |

**任务：**

- [ ] 后端API：`GET /api/v1/summary/weekly?start_date=&end_date=`
- [ ] Agent端周报生成Prompt
- [ ] 前端周报展示页面

---

### P2 - 功能完善

> 提升用户体验和系统可靠性

#### 2.8 统计分析页面

| 项目 | 内容 |
|------|------|
| **模块** | Web |
| **现状** | 仪表盘仅有基础数字统计 |
| **目标** | 分类趋势图、发件人统计、处理效率可视化 |

**任务：**

- [ ] 后端统计API：`GET /api/v1/stats/overview`
- [ ] 分类趋势折线图（近7/30天）
- [ ] 发件人Top10柱状图
- [ ] 分类分布饼图
- [ ] 处理效率指标（分类数量、平均置信度）

**涉及文件：**

```
email-web/src/
├── pages/Stats.tsx             # 新增
├── components/charts/
│   ├── CategoryTrend.tsx       # 新增
│   ├── SenderRanking.tsx       # 新增
│   └── CategoryPie.tsx         # 新增
└── api/statsApi.ts             # 新增
```

---

#### 2.9 Redis异步队列

| 项目 | 内容 |
|------|------|
| **模块** | Backend + Agent |
| **现状** | 批量分类为同步HTTP调用，大数量会超时 |
| **目标** | Redis消息队列异步处理 |

**任务：**

- [ ] Backend任务发布到Redis队列
- [ ] Agent消费者拉取任务处理
- [ ] 任务状态跟踪（pending/processing/done/failed）
- [ ] 失败重试机制（最多3次）

---

#### 2.10 定时自动同步

| 项目 | 内容 |
|------|------|
| **模块** | Backend |
| **现状** | 仅支持手动触发同步 |
| **目标** | Cron定时拉取新邮件并自动分类 |

**任务：**

- [ ] Go Cron调度器集成
- [ ] 可配置同步间隔（默认5分钟）
- [ ] 同步后自动触发分类
- [ ] 同步状态记录和展示

**涉及文件：**

```
email-backend/server/
├── service/sync_scheduler.go   # 新增
└── core/cron.go                # 新增
```

---

#### 2.11 附件处理

| 项目 | 内容 |
|------|------|
| **模块** | Backend + Web |
| **现状** | 仅标记是否有附件，无下载功能 |
| **目标** | 附件信息提取和下载 |

**任务：**

- [ ] IMAP附件提取
- [ ] 附件元数据存储（文件名、大小、类型）
- [ ] 附件下载API：`GET /api/v1/emails/:id/attachments/:aid`
- [ ] 前端附件列表展示和下载按钮

---

### P3 - 生产化

> 满足部署上线要求

#### 2.12 Docker Compose部署

| 项目 | 内容 |
|------|------|
| **模块** | DevOps |
| **现状** | 各服务本地手动启动 |
| **目标** | 一键部署全套环境 |

**任务：**

- [ ] 各服务Dockerfile编写
- [ ] `docker-compose.yml`（开发环境，单机部署）
- [ ] `docker-compose.server.yml`（生产环境，分离部署）
- [ ] `deploy.sh`一键部署脚本

---

#### 2.13 Nginx反向代理

| 项目 | 内容 |
|------|------|
| **模块** | DevOps |
| **现状** | 未配置 |
| **目标** | SSL终止、静态托管、API代理 |

**任务：**

- [ ] Nginx配置文件
- [ ] 前端静态资源托管
- [ ] API请求代理到Go后端
- [ ] SSL证书配置模板

---

#### 2.14 安全加固

| 项目 | 内容 |
|------|------|
| **模块** | DevOps |
| **现状** | 基础安全措施已实现（凭证加密） |
| **目标** | 生产级安全 |

**任务：**

- [ ] `.env.example`完善
- [ ] API限流（Gin Rate Limit中间件）
- [ ] 请求日志审计
- [ ] 敏感数据脱敏（API响应中不返回授权码等）

---

#### 2.15 通义千问Provider

| 项目 | 内容 |
|------|------|
| **模块** | Agent |
| **现状** | 设计文档提到但未实现 |
| **目标** | 支持通义千问作为LLM Provider |

**任务：**

- [ ] `qwen.py` Provider实现
- [ ] API对接和测试
- [ ] 配置选项

---

## 3. 开发路线图

```
2026-04 ──────────────────────────────────────────────
  │
  │  第一阶段：核心完善（P0）
  │  ├── 用户认证系统
  │  ├── Action Items API + UI
  │  └── LLM配置管理
  │
2026-05 上旬 ──────────────────────────────────────────
  │
  │  第二阶段：智能增强（P1）
  │  ├── Orchestrator编排器
  │  ├── ChromaDB + BGE Embedding
  │  └── 周报摘要
  │
2026-05 中旬 ──────────────────────────────────────────
  │
  │  第三阶段：体验优化（P2）
  │  ├── 统计分析页面
  │  ├── Redis异步队列
  │  ├── 定时自动同步
  │  └── 附件处理
  │
2026-05 下旬 ──────────────────────────────────────────
  │
  │  第四阶段：生产部署（P3）
  │  ├── Docker Compose
  │  ├── Nginx配置
  │  ├── 安全加固
  │  └── 通义千问Provider
  │
  ▼
上线
```

---

## 4. 验收标准

### 第一阶段验收

| 序号 | 验收项 | 通过标准 |
|------|--------|---------|
| 1 | 用户注册登录 | 新用户可注册、登录后看到自己的数据 |
| 2 | 数据隔离 | 不同用户只能看到自己的邮件和账户 |
| 3 | 行动项展示 | 邮件详情页显示提取的行动项，可标记完成 |
| 4 | LLM配置 | 用户可切换LLM Provider，配置生效 |

### 第二阶段验收

| 序号 | 验收项 | 通过标准 |
|------|--------|---------|
| 5 | 完整流水线 | 同步邮件后自动完成分类→提取→向量化 |
| 6 | 语义搜索 | 输入关键词可检索语义相关邮件 |
| 7 | 周报生成 | 可生成指定日期范围的邮件周报 |

### 第三阶段验收

| 序号 | 验收项 | 通过标准 |
|------|--------|---------|
| 8 | 统计图表 | 展示分类趋势、发件人排行、分类分布 |
| 9 | 异步处理 | 50封邮件批量分类不超时，可查看进度 |
| 10 | 自动同步 | 5分钟间隔自动拉取新邮件 |
| 11 | 附件下载 | 可查看并下载邮件附件 |

### 第四阶段验收

| 序号 | 验收项 | 通过标准 |
|------|--------|---------|
| 12 | Docker部署 | `docker-compose up -d` 一键启动全部服务 |
| 13 | HTTPS访问 | 通过Nginx提供HTTPS访问 |
| 14 | 安全审计 | 无明文密钥、API限流生效、日志可追溯 |

---

*文档版本：v1.0*
*最后更新：2026-04-18*
