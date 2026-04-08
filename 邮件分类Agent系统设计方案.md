# 个人邮件消息分类汇总 Agent 系统设计方案

## 1. 项目概述

### 1.1 系统目标

构建一个智能邮件管理系统，能够：
- 自动分类邮件（工作、个人、订阅、通知、垃圾邮件等）
- 提取关键信息（发件人、主题、截止日期、行动项等）
- 生成每日/每周邮件摘要
- 智能回复建议
- 重要邮件提醒

### 1.2 核心功能

| 功能模块 | 描述 |
|---------|------|
| **邮件采集** | 对接邮箱API，定时拉取新邮件 |
| **智能分类** | 基于内容的自动分类 |
| **信息提取** | 提取关键实体和行动项 |
| **摘要生成** | 生成结构化摘要报告 |
| **智能回复** | 生成回复建议 |
| **通知推送** | 重要邮件即时通知 |

## 2. LangChain 框架选择说明

### 2.1 为什么文档中写"自研框架"？

**原因：LangChain 官方没有 Go 语言实现**

| 语言 | LangChain 支持 | 生态成熟度 |
|------|---------------|-----------|
| Python | ✅ 官方支持 | ⭐⭐⭐⭐⭐ 最成熟 |
| JavaScript/TypeScript | ✅ 官方支持 | ⭐⭐⭐⭐ 较成熟 |
| Go | ❌ 无官方支持 | ⭐ 有第三方实验性项目 |

Go 生态中的 Agent 框架选择：
- **langchain-go** (第三方)：实验性项目，功能不完整，不建议生产使用
- **自研轻量框架**：根据需求定制，更灵活可控
- **继续用 Python LangChain**：Agent 服务独立，与 Go 后端通信

### 2.2 推荐方案：项目拆分

**最佳实践：将系统拆分为两个独立项目**

```
┌─────────────────────────────────────────────────────────────┐
│                     整体系统架构                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌─────────────────┐         ┌─────────────────┐          │
│   │  后端服务 (Go)   │  RPC/   │  Agent服务      │          │
│   │                 │  HTTP   │  (Python)       │          │
│   │  - Gin API      │◄───────►│  - LangChain    │          │
│   │  - MySQL        │         │  - LLM调用      │          │
│   │  - 业务逻辑     │         │  - Agent编排    │          │
│   │  - 邮件采集     │         │  - 工具执行     │          │
│   └─────────────────┘         └─────────────────┘          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**优势：**
1. **后端用 Go**：高性能、高并发、部署简单
2. **Agent 用 Python + LangChain**：生态成熟、开发效率高
3. **职责分离**：后端专注业务，Agent专注推理
4. **独立扩展**：Agent 服务可独立扩容
5. **技术最优**：各取所长

## 3. 拆分项目架构设计

### 3.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        Web 前端                                  │
│                   React + Vite + Tailwind                        │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP
┌────────────────────────────▼────────────────────────────────────┐
│                    后端服务 (Go + Gin)                           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │  API Handler│ │  邮件服务   │ │  数据存储   │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │  用户认证   │ │  定时任务   │ │  缓存管理   │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP / gRPC
┌────────────────────────────▼────────────────────────────────────┐
│                    Agent 服务 (Python + LangChain)              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │ 分类 Agent  │ │ 提取 Agent  │ │ 摘要 Agent  │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │ 回复 Agent  │ │ 检索 Agent  │ │ 分析 Agent  │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
│  ┌─────────────┐ ┌─────────────┐                               │
│  │ Orchestrator│ │  LLM适配层  │                               │
│  └─────────────┘ └─────────────┘                               │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                      数据层                                      │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐               │
│  │ MySQL   │ │ Redis   │ │ Chroma  │ │ 邮件API │               │
│  │ 业务数据│ │ 缓存    │ │ 向量DB  │ │ Gmail   │               │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘               │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 服务通信方式

| 方式 | 适用场景 | 说明 |
|------|---------|------|
| **HTTP REST** | 简单调用 | Go 调用 Python Agent API |
| **gRPC** | 高频调用 | 性能更高，适合批量处理 |
| **消息队列** | 异步任务 | Redis Streams / RabbitMQ |

**推荐：HTTP REST + 消息队列**

```
Go 后端 ───► HTTP POST ───► Python Agent 服务
    │                              │
    │◄─── HTTP Response ──────────◄│
    │                              │
    │                              │
    ▼                              ▼
Redis Streams (异步任务队列)
```

### 3.3 两个项目职责划分

#### 后端服务 (Go) 职责

| 模块 | 职责 |
|------|------|
| **API层** | 提供Web API，处理用户请求 |
| **邮件采集** | 定时拉取邮件，存储到数据库 |
| **数据存储** | MySQL业务数据、Redis缓存 |
| **用户管理** | 认证、权限、用户配置 |
| **任务调度** | 定时任务触发Agent处理 |
| **结果整合** | 接收Agent结果，更新数据库 |

#### Agent服务 (Python) 职责

| 模块 | 职责 |
|------|------|
| **LLM调用** | 调用DeepSeek/豆包等大模型 |
| **Agent编排** | 协调各Agent执行任务 |
| **分类推理** | 邮件智能分类 |
| **信息提取** | 提取行动项、实体等 |
| **摘要生成** | 生成各类摘要报告 |
| **向量处理** | 邮件向量化、语义检索 |

### 3.4 数据流设计

```
新邮件到达
    │
    ▼
┌─────────────┐
│ Go 邮件采集 │ ─── 拉取邮件，存MySQL
└─────────────┘
    │
    ▼
┌─────────────┐
│ Redis队列   │ ─── 发送处理任务
└─────────────┘
    │
    ▼
┌─────────────────────┐
│ Python Agent服务    │
│  1. 分类 Agent      │ ─── 调用LLM分类
│  2. 提取 Agent      │ ─── 提取关键信息
│  3. 向量化          │ ─── 存入向量DB
└─────────────────────┘
    │
    ▼
┌─────────────┐
│ 返回Go后端  │ ─── HTTP回调/队列通知
└─────────────┘
    │
    ▼
┌─────────────┐
│ 更新MySQL   │ ─── 保存分类、提取结果
└─────────────┘
    │
    ▼
┌─────────────┐
│ 生成摘要    │ ─── 定时触发Agent生成
└─────────────┘
```

## 4. 项目结构设计

### 4.1 项目一：后端服务 (Go)

```
email-backend/                     # Go后端项目
├── cmd/
│   └── server/
│       └── main.go               # 程序入口
├── config/
│   ├── config.yaml               # 配置文件
│   └── config.go
├── internal/
│   ├── api/
│   │   ├── handler/              # HTTP Handler
│   │   │   ├── email.go
│   │   │   ├── user.go
│   │   │   └── settings.go
│   │   ├── middleware/           # 中间件
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   └── logger.go
│   │   └── router.go             # 路由定义
│   ├── service/
│   │   ├── email_service.go      # 邮件业务
│   │   ├── user_service.go       # 用户业务
│   │   └── agent_client.go       # Agent服务客户端
│   ├── repository/
│   │   ├── email_repo.go         # 邮件数据访问
│   │   ├── user_repo.go          # 用户数据访问
│   │   └── action_repo.go        # 行动项数据访问
│   ├── model/
│   │   ├── email.go
│   │   ├── user.go
│   │   └── action_item.go
│   ├── worker/
│   │   ├── email_fetcher.go      # 邮件拉取Worker
│   │   ├── task_scheduler.go     # 任务调度
│   │   └── result_processor.go   # Agent结果处理
│   └── pkg/
│       ├── email/                # 邮件解析
│       ├── queue/                # 消息队列
│       └── utils/
├── pkg/
│   └── shared/
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

### 4.2 项目二：Agent服务 (Python + LangChain)

```
email-agent/                       # Python Agent项目
├── app/
│   ├── main.py                   # FastAPI入口
│   ├── config.py                 # 配置管理
│   ├── api/
│   │   ├── routes/
│   │   │   ├── classification.py # 分类API
│   │   │   ├── extraction.py     # 提取API
│   │   │   ├── summary.py        # 摘要API
│   │   │   └── reply.py          # 回复API
│   │   └── dependencies.py
│   ├── agents/
│   │   ├── __init__.py
│   │   ├── orchestrator.py       # Agent编排器
│   │   ├── classification.py     # 分类Agent
│   │   ├── extraction.py         # 提取Agent
│   │   ├── summary.py            # 摘要Agent
│   │   ├── reply.py              # 回复Agent
│   │   ├── retrieval.py          # 检索Agent
│   │   └── analysis.py           # 分析Agent
│   ├── tools/
│   │   ├── __init__.py
│   │   ├── email_tools.py        # 邮件工具
│   │   ├── vector_tools.py       # 向量工具
│   │   └── search_tools.py       # 搜索工具
│   ├── llm/
│   │   ├── __init__.py
│   │   ├── provider.py           # Provider接口
│   │   ├── deepseek.py           # DeepSeek实现
│   │   ├── doubao.py             # 豆包实现
│   │   ├── zhipu.py              # 智谱实现
│   │   └── manager.py            # Provider管理
│   ├── prompts/
│   │   ├── __init__.py
│   │   ├── classification.py     # 分类提示词
│   │   ├── extraction.py         # 提取提示词
│   │   ├── summary.py            # 摘要提示词
│   │   └── reply.py              # 回复提示词
│   ├── memory/
│   │   ├── __init__.py
│   │   └── conversation.py       # 对话记忆
│   ├── parsers/
│   │   ├── __init__.py
│   │   ├── json_parser.py        # JSON输出解析
│   │   └── structured.py         # 结构化输出
│   ├── vector/
│   │   ├── __init__.py
│   │   ├── embeddings.py         # 向量嵌入
│   │   └── store.py              # 向量存储
│   └── models/
│       ├── __init__.py
│       ├── email.py
│       ├── classification.py
│       ├── extraction.py
│       └── summary.py
├── tests/
│   ├── agents/
│   └── llm/
├── requirements.txt
├── pyproject.toml
├── Dockerfile
└── docker-compose.yml
```

## 5. 技术栈推荐

### 5.1 后端服务 (Go) 技术栈

| 层级 | 技术选型 | 说明 |
|------|---------|------|
| **语言** | Go 1.21+ | 高性能、高并发 |
| **Web框架** | Gin | 轻量级、高性能 |
| **数据库** | MySQL 8.0 | 业务数据存储 |
| **缓存** | Redis | 结果缓存、会话、消息队列 |
| **定时任务** | asynq / cron | 定时拉取邮件 |
| **HTTP客户端** | resty | 调用Agent服务 |

### 5.2 Agent服务 (Python) 技术栈

| 层级 | 技术选型 | 说明 |
|------|---------|------|
| **语言** | Python 3.11+ | LangChain生态友好 |
| **Web框架** | FastAPI | 提供Agent API |
| **Agent框架** | LangChain + LangGraph | 主流Agent框架 |
| **LLM** | DeepSeek / 豆包 / 智谱 | 国产大模型，多模型切换 |
| **Embedding** | BGE / M3E | 中文向量化模型 |
| **向量数据库** | Chroma (开发) / Milvus (生产) | 邮件向量检索 |

### 5.3 前端技术栈

| 技术 | 说明 |
|------|------|
| React 18 + Vite | 现代化前端框架 |
| Tailwind CSS | 快速UI开发 |
| Chart.js / ECharts | 数据可视化 |
| React Query | 数据请求管理 |

### 5.4 部署技术栈

| 技术 | 说明 |
|------|------|
| Docker + Docker Compose | 容器化部署 |
| Nginx | 反向代理 |
| GitHub Actions | CI/CD |

## 4. 数据库选型建议

### 4.1 推荐：MySQL 8.0

**优势：**
- 国内使用最广泛，文档和工具链成熟
- 运维人员熟悉，出问题容易解决
- 兼容性最好，各种框架支持完善
- 阿里云/腾讯云等云服务支持好，托管成本低
- 8.0版本支持JSON字段，满足半结构化数据存储

**适用场景：**
- 团队熟悉MySQL
- 需要云托管服务
- 运维成本低优先

**表结构设计：**
```sql
-- 邮件主表
CREATE TABLE emails (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    message_id VARCHAR(255) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    sender_name VARCHAR(255),
    sender_email VARCHAR(255) NOT NULL,
    subject VARCHAR(512),
    content TEXT,
    category ENUM('work_urgent', 'work_normal', 'personal', 'subscription', 'notification', 'promotion', 'spam'),
    priority ENUM('critical', 'high', 'medium', 'low') DEFAULT 'medium',
    status ENUM('unread', 'read', 'processed', 'archived') DEFAULT 'unread',
    has_attachment BOOLEAN DEFAULT FALSE,
    received_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_category (user_id, category),
    INDEX idx_user_received (user_id, received_at),
    INDEX idx_user_status (user_id, status)
);

-- 行动项表
CREATE TABLE action_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email_id BIGINT NOT NULL,
    task TEXT NOT NULL,
    deadline DATETIME,
    priority ENUM('high', 'medium', 'low'),
    status ENUM('pending', 'in_progress', 'completed', 'cancelled') DEFAULT 'pending',
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE
);

-- 邮件向量索引表（用于与向量DB关联）
CREATE TABLE email_vectors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email_id BIGINT NOT NULL,
    vector_id VARCHAR(255) NOT NULL,  -- 对应向量数据库的ID
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE
);
```

### 4.2 备选：PostgreSQL + pgvector

**优势：**
- pgvector扩展支持向量存储，可以统一存储
- JSONB字段更强大，适合复杂数据
- 全文检索能力更强
- 并发性能更好

**适用场景：**
- 希望数据库存储向量和业务数据
- 复杂查询需求多
- 团队熟悉PostgreSQL

### 4.3 建议

**个人项目推荐 MySQL**，原因：
1. 国内云服务MySQL托管更便宜
2. 大多数Go项目默认使用MySQL
3. 出问题网上资料更多

如果需要向量存储，可以 MySQL + Chroma/Milvus 组合。

## 5. LLM 多模型适配设计

### 5.1 支持的国产大模型

| 模型 | 提供商 | 特点 | 适用场景 |
|------|--------|------|----------|
| **DeepSeek** | DeepSeek AI | 性价比高，推理能力强 | 主力模型 |
| **豆包** | 字节跳动 | 中文理解好，API稳定 | 分类、摘要 |
| **智谱GLM** | 智谱AI | 开源可私有化部署 | 隐私要求高 |
| **通义千问** | 阿里云 | 生态完善，企业支持好 | 企业场景 |
| **文心一言** | 百度 | 知识问答强 | 信息提取 |
| **Kimi** | 月之暗面 | 长上下文 | 长邮件处理 |

### 5.2 多模型切换架构

```go
// LLM Provider 接口定义
type LLMProvider interface {
    Name() string
    Chat(ctx context.Context, messages []Message, options ChatOptions) (*ChatResponse, error)
    Embed(ctx context.Context, texts []string) ([][]float32, error)
    Classify(ctx context.Context, content string, categories []string) (string, float64, error)
}

// 具体实现
type DeepSeekProvider struct { /* ... */ }
type DoubaoProvider struct { /* ... */ }
type ZhipuProvider struct { /* ... */ }

// Provider Manager
type LLMManager struct {
    providers map[string]LLMProvider
    defaultProvider string
}

func (m *LLMManager) GetProvider(name string) (LLMProvider, error)
func (m *LLMManager) SetDefault(name string)
```

### 5.3 配置文件

```yaml
# config.yaml
llm:
  default: deepseek  # 默认模型
  
  providers:
    deepseek:
      enabled: true
      api_key: ${DEEPSEEK_API_KEY}
      base_url: https://api.deepseek.com/v1
      model: deepseek-chat
      max_tokens: 4096
      temperature: 0.3
      
    doubao:
      enabled: true
      api_key: ${DOUBAO_API_KEY}
      model: doubao-pro-32k
      max_tokens: 4096
      temperature: 0.3
      
    zhipu:
      enabled: false
      api_key: ${ZHIPU_API_KEY}
      model: glm-4
      max_tokens: 4096

  # 任务-模型映射
  routing:
    classification: deepseek  # 分类任务用DeepSeek
    extraction: doubao        # 信息提取用豆包
    summary: deepseek         # 摘要用DeepSeek
    reply: doubao             # 回复建议用豆包
```

### 5.4 客户端实现

```go
// pkg/llm/deepseek.go
package llm

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
)

type DeepSeekClient struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func NewDeepSeekClient(apiKey string) *DeepSeekClient {
    return &DeepSeekClient{
        apiKey:  apiKey,
        baseURL: "https://api.deepseek.com/v1",
        client:  &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *DeepSeekClient) Chat(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
    reqBody := map[string]interface{}{
        "model":       opts.Model,
        "messages":    messages,
        "max_tokens":  opts.MaxTokens,
        "temperature": opts.Temperature,
    }
    
    // 调用API
    resp, err := c.doRequest(ctx, "/chat/completions", reqBody)
    if err != nil {
        return nil, err
    }
    
    return c.parseResponse(resp)
}

func (c *DeepSeekClient) ClassifyEmail(ctx context.Context, subject, content string) (*ClassificationResult, error) {
    prompt := buildClassificationPrompt(subject, content)
    
    messages := []Message{
        {Role: "system", Content: ClassificationSystemPrompt},
        {Role: "user", Content: prompt},
    }
    
    resp, err := c.Chat(ctx, messages, ChatOptions{
        Model:       "deepseek-chat",
        MaxTokens:   512,
        Temperature: 0.1,  // 分类用低temperature
    })
    if err != nil {
        return nil, err
    }
    
    return parseClassificationResponse(resp.Content)
}
```

### 5.5 Go后端调用Agent服务

```go
// internal/service/agent_client.go
package service

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
)

type AgentClient struct {
    baseURL string
    client  *http.Client
    apiKey  string
}

func NewAgentClient(baseURL, apiKey string) *AgentClient {
    return &AgentClient{
        baseURL: baseURL,
        apiKey:  apiKey,
        client:  &http.Client{Timeout: 60 * time.Second},
    }
}

// ClassifyEmail 调用Agent服务进行邮件分类
func (c *AgentClient) ClassifyEmail(ctx context.Context, req *ClassificationRequest) (*ClassificationResponse, error) {
    body, _ := json.Marshal(req)

    httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/classify", bytes.NewReader(body))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("X-API-Key", c.apiKey)

    resp, err := c.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result ClassificationResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

// ExtractInfo 调用Agent服务提取邮件信息
func (c *AgentClient) ExtractInfo(ctx context.Context, req *ExtractionRequest) (*ExtractionResponse, error) {
    body, _ := json.Marshal(req)

    httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/extract", bytes.NewReader(body))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("X-API-Key", c.apiKey)

    resp, err := c.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result ExtractionResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

// GenerateSummary 调用Agent服务生成摘要
func (c *AgentClient) GenerateSummary(ctx context.Context, req *SummaryRequest) (*SummaryResponse, error) {
    body, _ := json.Marshal(req)

    httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/summary", bytes.NewReader(body))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("X-API-Key", c.apiKey)

    resp, err := c.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result SummaryResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}
```

### 5.6 Agent服务API定义 (Python FastAPI)

```python
# app/api/routes/classification.py
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional, List

router = APIRouter(prefix="/api/v1", tags=["classification"])

class ClassificationRequest(BaseModel):
    email_id: str
    subject: str
    content: str
    sender: str
    received_at: str

class ClassificationResponse(BaseModel):
    email_id: str
    category: str  # work_urgent, work_normal, personal, subscription, notification, promotion, spam
    priority: str   # critical, high, medium, low
    confidence: float
    reasoning: str
    action_required: bool
    deadline: Optional[str] = None

@router.post("/classify", response_model=ClassificationResponse)
async def classify_email(request: ClassificationRequest):
    """对单封邮件进行分类"""
    agent = ClassificationAgent()
    result = await agent.execute(
        subject=request.subject,
        content=request.content,
        sender=request.sender
    )
    return result

@router.post("/batch-classify")
async def batch_classify(requests: List[ClassificationRequest]):
    """批量分类邮件"""
    agent = ClassificationAgent()
    results = await agent.batch_execute([(r.subject, r.content) for r in requests])
    return {"results": results}
```

## 6. 数据库选型建议

### 6.1 推荐：MySQL 8.0

**优势：**
- 国内使用最广泛，文档和工具链成熟
- 运维人员熟悉，出问题容易解决
- 阿里云/腾讯云等云服务支持好
- 8.0版本支持JSON字段

**表结构设计：**
```sql
-- 邮件主表
CREATE TABLE emails (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    message_id VARCHAR(255) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    sender_name VARCHAR(255),
    sender_email VARCHAR(255) NOT NULL,
    subject VARCHAR(512),
    content TEXT,
    category VARCHAR(50) DEFAULT 'unclassified',
    priority VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'unread',
    has_attachment BOOLEAN DEFAULT FALSE,
    received_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_category (user_id, category),
    INDEX idx_user_received (user_id, received_at),
    INDEX idx_user_status (user_id, status)
);

-- 行动项表
CREATE TABLE action_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email_id BIGINT NOT NULL,
    task TEXT NOT NULL,
    deadline DATETIME,
    priority VARCHAR(20),
    status VARCHAR(20) DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE
);

-- 邮件向量索引表
CREATE TABLE email_vectors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email_id BIGINT NOT NULL,
    vector_id VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE
);
```

### 6.2 建议

**个人项目推荐 MySQL 8.0**，向量存储用 Chroma/Milvus 分离存储。
```

## 7. Agent 设计详细方案 (Python + LangChain)

### 7.1 Agent 角色划分

```
┌─────────────────────────────────────────────────────────────┐
│                    Orchestrator Agent                        │
│              (主控Agent，负责任务分发和协调)                   │
└───────────────────────────┬─────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ 分类 Agent    │   │ 提取 Agent    │   │ 摘要 Agent    │
│ 邮件分类打标  │   │ 信息抽取      │   │ 摘要生成      │
└───────────────┘   └───────────────┘   └───────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ 回复 Agent    │   │ 检索 Agent    │   │ 分析 Agent    │
│ 回复建议生成  │   │ 邮件搜索      │   │ 统计分析      │
└───────────────┘   └───────────────┘   └───────────────┘
```

### 7.2 LangChain Agent 基础实现

```python
# app/agents/base.py
from abc import ABC, abstractmethod
from typing import Any, Dict, Optional
from langchain_core.language_models import BaseChatModel
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import PydanticOutputParser
from pydantic import BaseModel

class AgentResult(BaseModel):
    """Agent执行结果基类"""
    success: bool
    data: Optional[Dict[str, Any]] = None
    error: Optional[str] = None

class BaseAgent(ABC):
    """Agent基类"""

    def __init__(
        self,
        llm: BaseChatModel,
        temperature: float = 0.3,
    ):
        self.llm = llm
        self.temperature = temperature
        self._setup_prompts()

    @abstractmethod
    def _setup_prompts(self):
        """设置提示词模板"""
        pass

    @abstractmethod
    async def execute(self, input_data: Any) -> AgentResult:
        """执行Agent任务"""
        pass

    def create_chain(self, prompt: ChatPromptTemplate, output_parser=None):
        """创建处理链"""
        if output_parser:
            return prompt | self.llm | output_parser
        return prompt | self.llm
```

### 7.3 分类Agent实现

```python
# app/agents/classification.py
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import PydanticOutputParser
from pydantic import BaseModel, Field
from typing import Optional
from .base import BaseAgent, AgentResult

class ClassificationOutput(BaseModel):
    """分类结果输出"""
    category: str = Field(description="邮件类别")
    priority: str = Field(description="优先级")
    confidence: float = Field(description="置信度 0-1")
    reasoning: str = Field(description="分类理由")
    action_required: bool = Field(description="是否需要行动")
    deadline: Optional[str] = Field(description="截止日期", default=None)

CLASSIFICATION_SYSTEM_PROMPT = """你是一个专业的邮件分类助手。你的任务是分析邮件内容并对其进行分类。

分类类别：
- work_urgent: 紧急工作邮件（有截止日期、领导发送、需要立即处理）
- work_normal: 普通工作邮件（常规工作沟通）
- personal: 个人邮件（朋友、家人）
- subscription: 订阅邮件（新闻简报、技术博客）
- notification: 系统通知（GitHub、Jira等系统消息）
- promotion: 营销推广（广告、促销信息）
- spam: 垃圾邮件（诈骗、无意义内容）

优先级判断：
- critical: 有紧急截止日期、来自重要联系人
- high: 需要今日处理、有行动请求
- medium: 本周需处理、有价值信息
- low: 可稍后处理、信息性内容

请根据邮件的发件人、主题、内容进行分析，给出分类结果。"""

class ClassificationAgent(BaseAgent):
    """邮件分类Agent"""

    def _setup_prompts(self):
        self.parser = PydanticOutputParser(pydantic_object=ClassificationOutput)
        self.prompt = ChatPromptTemplate.from_messages([
            ("system", CLASSIFICATION_SYSTEM_PROMPT),
            ("human", """请分析以下邮件：

发件人: {sender}
主题: {subject}
内容: {content}

{format_instructions}""")
        ])

    async def execute(self, email_data: dict) -> AgentResult:
        """执行分类任务"""
        try:
            chain = self.prompt | self.llm | self.parser

            result = await chain.ainvoke({
                "sender": email_data.get("sender", ""),
                "subject": email_data.get("subject", ""),
                "content": email_data.get("content", ""),
                "format_instructions": self.parser.get_format_instructions()
            })

            return AgentResult(
                success=True,
                data=result.model_dump()
            )
        except Exception as e:
            return AgentResult(
                success=False,
                error=str(e)
            )

    async def batch_execute(self, emails: list) -> list:
        """批量分类"""
        results = []
        for email in emails:
            result = await self.execute(email)
            results.append(result)
        return results
```

### 7.4 提取Agent实现

```python
# app/agents/extraction.py
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import PydanticOutputParser
from pydantic import BaseModel, Field
from typing import List, Optional
from .base import BaseAgent, AgentResult

class ActionItem(BaseModel):
    """行动项"""
    task: str = Field(description="任务描述")
    deadline: Optional[str] = Field(description="截止时间")
    priority: str = Field(description="优先级 high/medium/low")
    assignee: Optional[str] = Field(description="负责人")

class Meeting(BaseModel):
    """会议信息"""
    title: str = Field(description="会议标题")
    time: str = Field(description="会议时间")
    location: Optional[str] = Field(description="会议地点")
    attendees: List[str] = Field(description="参会人员", default_factory=list)

class ExtractionOutput(BaseModel):
    """信息提取结果"""
    sender_info: dict = Field(description="发件人信息", default_factory=dict)
    action_items: List[ActionItem] = Field(description="行动项列表", default_factory=list)
    meetings: List[Meeting] = Field(description="会议列表", default_factory=list)
    deadlines: List[dict] = Field(description="截止日期列表", default_factory=list)
    key_entities: List[str] = Field(description="关键实体", default_factory=list)
    summary: str = Field(description="邮件摘要")

EXTRACTION_SYSTEM_PROMPT = """你是一个信息提取专家。请从邮件中提取以下信息：

1. 发件人信息：姓名、邮箱、组织、职位
2. 行动项：需要执行的任务、截止时间、优先级
3. 会议信息：标题、时间、地点、参会人员
4. 截止日期：重要时间节点
5. 关键实体：人名、公司、项目、金额等
6. 邮件摘要：一句话概括邮件内容

请准确提取，如果信息不存在则留空。"""

class ExtractionAgent(BaseAgent):
    """信息提取Agent"""

    def _setup_prompts(self):
        self.parser = PydanticOutputParser(pydantic_object=ExtractionOutput)
        self.prompt = ChatPromptTemplate.from_messages([
            ("system", EXTRACTION_SYSTEM_PROMPT),
            ("human", """请从以下邮件中提取信息：

主题: {subject}
内容: {content}
发件人: {sender}
日期: {date}

{format_instructions}""")
        ])

    async def execute(self, email_data: dict) -> AgentResult:
        """执行提取任务"""
        try:
            chain = self.prompt | self.llm | self.parser

            result = await chain.ainvoke({
                "subject": email_data.get("subject", ""),
                "content": email_data.get("content", ""),
                "sender": email_data.get("sender", ""),
                "date": email_data.get("date", ""),
                "format_instructions": self.parser.get_format_instructions()
            })

            return AgentResult(
                success=True,
                data=result.model_dump()
            )
        except Exception as e:
            return AgentResult(
                success=False,
                error=str(e)
            )
```

### 7.5 摘要Agent实现

```python
# app/agents/summary.py
from langchain_core.prompts import ChatPromptTemplate
from .base import BaseAgent, AgentResult
from typing import List

SUMMARY_SYSTEM_PROMPT = """你是一个邮件摘要专家。请根据当日邮件生成结构化摘要报告。

要求：
1. 按重要程度分组展示
2. 突出需要行动的事项
3. 简洁明了，突出重点
4. 使用Markdown格式输出"""

DAILY_SUMMARY_PROMPT = """请根据以下今日邮件生成摘要：

日期: {date}
邮件总数: {total_count}

邮件列表：
{emails_text}

请生成包含以下内容的摘要：
1. 重要待办事项（带截止时间）
2. 工作邮件概览
3. 订阅与通知汇总
4. 今日统计数据"""

class SummaryAgent(BaseAgent):
    """摘要生成Agent"""

    def _setup_prompts(self):
        self.prompt = ChatPromptTemplate.from_messages([
            ("system", SUMMARY_SYSTEM_PROMPT),
            ("human", DAILY_SUMMARY_PROMPT)
        ])

    async def execute(self, summary_data: dict) -> AgentResult:
        """生成摘要"""
        try:
            chain = self.prompt | self.llm

            # 格式化邮件列表
            emails_text = self._format_emails(summary_data.get("emails", []))

            result = await chain.ainvoke({
                "date": summary_data.get("date", ""),
                "total_count": len(summary_data.get("emails", [])),
                "emails_text": emails_text
            })

            return AgentResult(
                success=True,
                data={"summary": result.content}
            )
        except Exception as e:
            return AgentResult(
                success=False,
                error=str(e)
            )

    def _format_emails(self, emails: List[dict]) -> str:
        """格式化邮件列表"""
        lines = []
        for i, email in enumerate(emails, 1):
            lines.append(f"""
{i}. [{email.get('category', 'unknown')}] {email.get('subject', '无主题')}
   发件人: {email.get('sender', '未知')}
   时间: {email.get('time', '')}
   摘要: {email.get('preview', '')[:100]}...
""")
        return "\n".join(lines)

    async def generate_weekly_report(self, weekly_data: dict) -> AgentResult:
        """生成周报"""
        weekly_prompt = ChatPromptTemplate.from_messages([
            ("system", """你是一个邮件分析专家。请根据本周邮件数据生成周报。
包含：趋势分析、活跃联系人、完成事项、待处理事项、优化建议。"""),
            ("human", """请根据本周数据生成周报：

周范围: {week_range}
邮件总数: {total_count}
日均值: {daily_avg}

每日统计：
{daily_stats}

请生成详细的周报。""")
        ])

        chain = weekly_prompt | self.llm
        result = await chain.ainvoke(weekly_data)

        return AgentResult(
            success=True,
            data={"report": result.content}
        )
```

### 7.6 回复Agent实现

```python
# app/agents/reply.py
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import PydanticOutputParser
from pydantic import BaseModel, Field
from typing import List
from .base import BaseAgent, AgentResult

class ReplyOption(BaseModel):
    """回复选项"""
    type: str = Field(description="回复类型: confirm/decline/ask_details/accept_meeting")
    label: str = Field(description="选项标签")
    content: str = Field(description="回复内容")
    confidence: float = Field(description="推荐程度 0-1")

class ReplyOutput(BaseModel):
    """回复输出"""
    intent: str = Field(description="邮件意图: request/notification/invitation/approval/discussion")
    reply_options: List[ReplyOption] = Field(description="回复选项列表")
    suggested_action: str = Field(description="建议的操作")

REPLY_SYSTEM_PROMPT = """你是一个专业的邮件回复助手。请分析邮件意图并生成合适的回复建议。

回复类型：
- confirm: 确认回复
- decline: 婉拒回复
- ask_details: 询问详情
- accept_meeting: 接受会议
- custom: 自定义回复

要求：
1. 回复要礼貌得体
2. 直接回应请求
3. 提供多个选项供用户选择
4. 保持专业语气"""

class ReplyAgent(BaseAgent):
    """回复建议Agent"""

    def _setup_prompts(self):
        self.parser = PydanticOutputParser(pydantic_object=ReplyOutput)
        self.prompt = ChatPromptTemplate.from_messages([
            ("system", REPLY_SYSTEM_PROMPT),
            ("human", """请分析以下邮件并生成回复建议：

发件人: {sender}
主题: {subject}
内容: {content}

用户上下文: {user_context}

{format_instructions}""")
        ])

    async def execute(self, email_data: dict) -> AgentResult:
        """生成回复建议"""
        try:
            chain = self.prompt | self.llm | self.parser

            result = await chain.ainvoke({
                "sender": email_data.get("sender", ""),
                "subject": email_data.get("subject", ""),
                "content": email_data.get("content", ""),
                "user_context": email_data.get("user_context", ""),
                "format_instructions": self.parser.get_format_instructions()
            })

            return AgentResult(
                success=True,
                data=result.model_dump()
            )
        except Exception as e:
            return AgentResult(
                success=False,
                error=str(e)
            )
```

### 7.7 Agent编排器

```python
# app/agents/orchestrator.py
from typing import Dict, Any, List
from .classification import ClassificationAgent
from .extraction import ExtractionAgent
from .summary import SummaryAgent
from .reply import ReplyAgent
from ..llm.manager import LLMManager

class AgentOrchestrator:
    """Agent编排器 - 协调各Agent执行任务"""

    def __init__(self, llm_manager: LLMManager):
        self.llm_manager = llm_manager

        # 初始化各Agent，根据任务类型选择不同的LLM
        self.classification_agent = ClassificationAgent(
            llm=llm_manager.get_provider("classification")
        )
        self.extraction_agent = ExtractionAgent(
            llm=llm_manager.get_provider("extraction")
        )
        self.summary_agent = SummaryAgent(
            llm=llm_manager.get_provider("summary")
        )
        self.reply_agent = ReplyAgent(
            llm=llm_manager.get_provider("reply")
        )

    async def process_email(self, email: Dict[str, Any]) -> Dict[str, Any]:
        """处理单封邮件完整流程"""
        result = {
            "email_id": email.get("id"),
            "classification": None,
            "extraction": None,
        }

        # 1. 分类
        class_result = await self.classification_agent.execute(email)
        if class_result.success:
            result["classification"] = class_result.data

            # 2. 如果不是垃圾邮件，提取信息
            if class_result.data.get("category") != "spam":
                extract_result = await self.extraction_agent.execute(email)
                if extract_result.success:
                    result["extraction"] = extract_result.data

        return result

    async def generate_daily_summary(
        self,
        emails: List[Dict[str, Any]],
        date: str
    ) -> Dict[str, Any]:
        """生成每日摘要"""
        summary_result = await self.summary_agent.execute({
            "emails": emails,
            "date": date
        })

        return summary_result.data if summary_result.success else None

    async def get_reply_suggestions(
        self,
        email: Dict[str, Any],
        user_context: str = ""
    ) -> List[Dict[str, Any]]:
        """获取回复建议"""
        email["user_context"] = user_context
        reply_result = await self.reply_agent.execute(email)

        if reply_result.success:
            return reply_result.data.get("reply_options", [])
        return []

    async def batch_process(self, emails: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """批量处理邮件"""
        results = []
        for email in emails:
            result = await self.process_email(email)
            results.append(result)
        return results
```

### 7.8 LLM Provider管理

```python
# app/llm/manager.py
from typing import Dict, Optional
from langchain_core.language_models import BaseChatModel
from .deepseek import DeepSeekProvider
from .doubao import DoubaoProvider
from .zhipu import ZhipuProvider

class LLMManager:
    """LLM Provider管理器"""

    def __init__(self, config: Dict):
        self.providers: Dict[str, BaseChatModel] = {}
        self.default_provider: str = config.get("default", "deepseek")
        self.routing: Dict[str, str] = config.get("routing", {})

        # 初始化所有启用的Provider
        for name, cfg in config.get("providers", {}).items():
            if cfg.get("enabled", False):
                self.providers[name] = self._create_provider(name, cfg)

    def _create_provider(self, name: str, config: Dict) -> BaseChatModel:
        """创建Provider实例"""
        if name == "deepseek":
            return DeepSeekProvider(
                api_key=config["api_key"],
                model=config.get("model", "deepseek-chat"),
                temperature=config.get("temperature", 0.3),
            )
        elif name == "doubao":
            return DoubaoProvider(
                api_key=config["api_key"],
                model=config.get("model", "doubao-pro-32k"),
                temperature=config.get("temperature", 0.3),
            )
        elif name == "zhipu":
            return ZhipuProvider(
                api_key=config["api_key"],
                model=config.get("model", "glm-4"),
                temperature=config.get("temperature", 0.3),
            )
        else:
            raise ValueError(f"Unknown provider: {name}")

    def get_provider(self, task: str) -> BaseChatModel:
        """根据任务类型获取Provider"""
        # 查找任务路由配置
        provider_name = self.routing.get(task, self.default_provider)
        return self.providers.get(provider_name)

    def get_provider_by_name(self, name: str) -> Optional[BaseChatModel]:
        """根据名称获取Provider"""
        return self.providers.get(name)

    def list_providers(self) -> list:
        """列出所有可用Provider"""
        return list(self.providers.keys())
```

```python
# app/llm/deepseek.py
from langchain_openai import ChatOpenAI

class DeepSeekProvider(ChatOpenAI):
    """DeepSeek Provider"""

    def __init__(
        self,
        api_key: str,
        model: str = "deepseek-chat",
        temperature: float = 0.3,
        **kwargs
    ):
        super().__init__(
            api_key=api_key,
            model=model,
            temperature=temperature,
            base_url="https://api.deepseek.com/v1",
            **kwargs
        )
```
    }
    
    return parseExtractionResponse(response), nil
}
```

    
    // 条件搜索
    return t.emailRepo.Search(ctx, filters, limit)
}
```

## 9. API 接口设计

### 9.1 Go后端API (Gin)

```go
// internal/api/router.go
package api

import (
    "github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, handlers *Handlers) {
    // 健康检查
    r.GET("/health", handlers.Health.Check)

    v1 := r.Group("/api/v1")
    {
        // 邮件管理
        emails := v1.Group("/emails")
        {
            emails.GET("", handlers.Email.List)
            emails.GET("/:id", handlers.Email.Get)
            emails.POST("/:id/classify", handlers.Email.Classify)
            emails.GET("/:id/summary", handlers.Email.GetSummary)
            emails.GET("/:id/reply-suggestions", handlers.Email.GetReplySuggestions)
            emails.POST("/:id/mark-read", handlers.Email.MarkRead)
        }

        // 设置
        settings := v1.Group("/settings")
        {
            settings.GET("/llm/providers", handlers.Settings.GetLLMProviders)
            settings.GET("/llm/current", handlers.Settings.GetCurrentLLM)
            settings.POST("/llm", handlers.Settings.SetLLMProvider)

            settings.GET("/email-accounts", handlers.Settings.GetEmailAccounts)
            settings.POST("/email-accounts", handlers.Settings.AddEmailAccount)
            settings.DELETE("/email-accounts/:id", handlers.Settings.DeleteEmailAccount)
        }

        // 统计分析
        stats := v1.Group("/stats")
        {
            stats.GET("/dashboard", handlers.Stats.Dashboard)
            stats.GET("/trends", handlers.Stats.Trends)
            stats.GET("/senders", handlers.Stats.TopSenders)
        }
    }
}
```

### 9.2 Python Agent服务API (FastAPI)

```python
# app/main.py
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Optional

app = FastAPI(title="Email Agent Service", version="1.0.0")

# CORS配置
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 请求模型
class ClassificationRequest(BaseModel):
    email_id: str
    subject: str
    content: str
    sender: str

class ExtractionRequest(BaseModel):
    email_id: str
    subject: str
    content: str
    sender: str
    date: Optional[str] = None

class SummaryRequest(BaseModel):
    date: str
    emails: List[dict]

class ReplyRequest(BaseModel):
    email_id: str
    subject: str
    content: str
    sender: str
    user_context: Optional[str] = None

# API路由
@app.post("/api/v1/classify")
async def classify_email(request: ClassificationRequest):
    """邮件分类"""
    from .agents.classification import ClassificationAgent
    from .llm.manager import llm_manager

    agent = ClassificationAgent(llm_manager.get_provider("classification"))
    result = await agent.execute(request.dict())

    if not result.success:
        raise HTTPException(status_code=500, detail=result.error)
    return result.data

@app.post("/api/v1/extract")
async def extract_info(request: ExtractionRequest):
    """信息提取"""
    from .agents.extraction import ExtractionAgent
    from .llm.manager import llm_manager

    agent = ExtractionAgent(llm_manager.get_provider("extraction"))
    result = await agent.execute(request.dict())

    if not result.success:
        raise HTTPException(status_code=500, detail=result.error)
    return result.data

@app.post("/api/v1/summary")
async def generate_summary(request: SummaryRequest):
    """生成摘要"""
    from .agents.summary import SummaryAgent
    from .llm.manager import llm_manager

    agent = SummaryAgent(llm_manager.get_provider("summary"))
    result = await agent.execute(request.dict())

    if not result.success:
        raise HTTPException(status_code=500, detail=result.error)
    return result.data

@app.post("/api/v1/reply")
async def get_reply_suggestions(request: ReplyRequest):
    """获取回复建议"""
    from .agents.reply import ReplyAgent
    from .llm.manager import llm_manager

    agent = ReplyAgent(llm_manager.get_provider("reply"))
    result = await agent.execute(request.dict())

    if not result.success:
        raise HTTPException(status_code=500, detail=result.error)
    return result.data

@app.post("/api/v1/process")
async def process_email(request: dict):
    """处理单封邮件完整流程（分类+提取）"""
    from .agents.orchestrator import orchestrator

    result = await orchestrator.process_email(request)
    return result

@app.get("/health")
async def health_check():
    """健康检查"""
    return {"status": "healthy"}
```

## 10. 开发计划

### 10.1 阶段一：基础框架 (1-2周)

**Go后端：**
- [ ] 项目初始化（Go module + Gin）
- [ ] MySQL 数据库设计与迁移
- [ ] 基础API框架搭建
- [ ] 用户认证中间件
- [ ] Agent服务客户端封装

**Python Agent服务：**
- [ ] 项目初始化（FastAPI + LangChain）
- [ ] LLM Provider接口设计（DeepSeek优先）
- [ ] 配置管理
- [ ] 基础Agent框架

### 10.2 阶段二：核心功能 (2-3周)

**Go后端：**
- [ ] 邮件API集成（Gmail OAuth）
- [ ] 邮件拉取Worker
- [ ] Agent服务调用封装
- [ ] 结果存储与更新

**Python Agent服务：**
- [ ] 分类Agent实现
- [ ] 提取Agent实现
- [ ] 摘要Agent实现
- [ ] Agent编排器
- [ ] 向量化与检索

### 10.3 阶段三：扩展功能 (1-2周)

**Go后端：**
- [ ] 回复建议API
- [ ] 统计分析API
- [ ] 定时任务（摘要生成）

**Python Agent服务：**
- [ ] 回复Agent实现
- [ ] 多模型切换功能
- [ ] 向量搜索（Chroma/Milvus）

**前端：**
- [ ] Web界面开发
- [ ] 数据可视化

### 10.4 阶段四：完善与部署 (1-2周)

- [ ] 其他国产模型接入（豆包、智谱）
- [ ] Docker容器化
- [ ] 性能优化
- [ ] 测试与部署

## 11. 总结

本系统采用 **前后端分离 + Agent服务独立** 的架构：

### 架构优势

1. **职责分离**：Go处理业务逻辑，Python专注AI推理
2. **技术最优**：Go高性能并发，Python + LangChain生态成熟
3. **独立扩展**：Agent服务可独立扩容
4. **灵活切换**：LLM Provider可随时切换

### 技术栈总结

| 服务 | 技术栈 | 职责 |
|------|--------|------|
| **Go后端** | Go + Gin + MySQL | 业务逻辑、邮件采集、数据存储 |
| **Python Agent** | Python + LangChain + FastAPI | LLM调用、Agent编排、向量检索 |
| **前端** | React + Vite + Tailwind | Web界面 |
| **存储** | MySQL + Redis + Chroma | 业务数据、缓存、向量 |

### 数据库选型

**推荐 MySQL 8.0**：
- 国内生态成熟，运维成本低
- 云服务支持完善
- 8.0版本支持JSON字段
4. **仅Web界面**：聚焦核心功能，开发更高效

技术栈选择理由：
- **Go + Gin**：高并发、部署简单、性能优秀
- **MySQL**：国内最流行，云服务支持好
- **DeepSeek**：国产模型中性价比高
- **BGE/M3E**：中文向量化效果好的开源模型
