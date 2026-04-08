# 个人邮件分类汇总 Agent 系统 - 设计文档

> 版本：v1.0
> 日期：2026-04-07
> 状态：**设计阶段**

---

## 1. 项目概述

### 1.1 项目目标

开发一个个人邮件智能分类汇总系统，实现以下核心功能：
- 读取邮件内容（优先支持网易126邮箱）
- 对邮件进行智能分类
- 提取关键信息并生成摘要
- 支持个人部署，确保数据安全

### 1.2 核心设计原则

1. **模块独立性**：Web端、服务端、Agent端完全解耦，通过标准API通信
2. **数据安全性**：邮箱凭证加密存储，不泄露用户敏感信息
3. **可扩展性**：预留多邮箱支持接口，便于后续扩展
4. **个人部署友好**：Docker一键部署，适合个人服务器/PC运行

---

## 2. 系统整体架构

### 2.1 三端架构图

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                              系统整体架构                                        │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                                │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                         Web 前端 (email-web)                              │  │
│  │                                                                          │  │
│  │   功能模块：                                                              │  │
│  │   • 邮件列表展示 & 筛选                                                   │  │
│  │   • 邮件详情 & 分类信息                                                   │  │
│  │   • 每日摘要 & 待办事项                                                   │  │
│  │   • 统计分析 & 可视化                                                     │  │
│  │   • 账户设置 & LLM配置                                                    │  │
│  │                                                                          │  │
│  │   技术栈：React 18 + Vite + TypeScript + Tailwind CSS                   │  │
│  └──────────────────────────────────┬───────────────────────────────────────────┘  │
│                                     │ HTTP REST API (JSON)                         │
│                                     ▼                                              │
│  ┌──────────────────────────────────▼───────────────────────────────────────────┐  │
│  │                         服务端 (email-backend)                               │  │
│  │                                                                          │  │
│  │   核心模块：                                                              │  │
│  │   • API网关 (Gin Router)                                                  │  │
│  │   • 业务服务层                                                            │  │
│  │   • 邮件采集器 (126邮箱 + 通用IMAP)                                       │  │
│  │   • 数据存储层 (MySQL + Redis)                                            │  │
│  │   • Agent通信客户端                                                       │  │
│  │   • 凭证安全管理                                                          │  │
│  │                                                                          │  │
│  │   技术栈：Go 1.21+ + Gin + GORM + MySQL 8.0 + Redis                     │  │
│  └──────────────────────────────────┬───────────────────────────────────────────┘  │
│                                     │ HTTP REST API / Redis Queue                    │
│                                     ▼                                              │
│  ┌──────────────────────────────────▼───────────────────────────────────────────┐  │
│  │                          Agent 端 (email-agent)                              │  │
│  │                                                                          │  │
│  │   核心模块：                                                              │  │
│  │   • FastAPI 服务层                                                       │  │
│  │   • Orchestrator 编排器                                                   │  │
│  │   • Classification Agent (分类)                                           │  │
│  │   • Extraction Agent (信息提取)                                           │  │
│  │   • Summary Agent (摘要生成)                                              │  │
│  │   • LLM 适配层 (支持多Provider)                                           │  │
│  │                                                                          │  │
│  │   技术栈：Python 3.11+ + LangChain + FastAPI + ChromaDB                 │  │
│  └────────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                              数据层                                        │  │
│  │                                                                          │  │
│  │   ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐        │  │
│  │   │   MySQL    │  │   Redis    │  │  ChromaDB  │  │  邮件服务  │        │  │
│  │   │  业务数据  │  │ 缓存 & 队列 │  │  向量存储  │  │  126邮箱   │        │  │
│  │   └────────────┘  └────────────┘  └────────────┘  └────────────┘        │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
└────────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 技术栈选型

| 模块 | 技术选型 | 选型理由 |
|------|---------|----------|
| **Web前端** | React 18 + Vite + TypeScript + Tailwind CSS | 现代化前端框架，开发效率高，类型安全 |
| **服务端** | Go 1.21+ + Gin + GORM | 高性能、低内存、交叉编译方便个人部署 |
| **Agent端** | Python 3.11+ + LangChain + FastAPI | LLM集成生态成熟，开发灵活 |
| **关系数据库** | MySQL 8.0 | 稳定可靠，个人部署简单 |
| **缓存/消息队列** | Redis 7 | 高性能，支持多种数据结构 |
| **向量数据库** | ChromaDB | 轻量级，适合个人部署 |
| **LLM Provider** | DeepSeek / 智谱GLM / 通义千问 | 国产大模型，性价比高，支持私有部署 |

---

## 3. 网易126邮箱支持方案

### 3.1 邮箱接入方式

网易126邮箱支持以下接入方式：

| 接入方式 | 协议 | 端口 | 加密 | 推荐度 |
|---------|------|------|------|--------|
| **IMAP** | IMAP4 | 993 | SSL | ⭐⭐⭐⭐⭐ |
| IMAP | IMAP4 | 143 | 无 | ⭐⭐ |
| SMTP | SMTP | 465/994 | SSL | ⭐⭐⭐⭐⭐ |

**推荐配置**：
```yaml
126邮箱IMAP配置:
  服务器地址: imap.126.com
  端口: 993 (SSL)
  用户名: 你的邮箱地址 (@126.com)
  授权密码: 网易邮箱专用授权码  # 注意：不是登录密码！
```

### 3.2 授权码获取流程

用户需要为应用生成专用授权码，而不是使用登录密码：

```
用户操作流程：
1. 登录 126 邮箱网页版
2. 进入 设置 → POP3/SMTP/IMAP
3. 开启 IMAP/SMTP 服务
4. 生成"授权码"（16位字母数字）
5. 在系统中输入授权码（系统将加密存储）
```

### 3.3 邮箱凭证安全方案

**核心原则：永不存储明文密码/授权码**

#### 3.3.1 凭证存储架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                        凭证安全存储方案                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  用户输入凭证                                                          │
│       │                                                              │
│       ▼                                                              │
│  ┌─────────────────────────────────────────────────────────────┐     │
│  │  1. 服务端接收授权码                                          │     │
│  │  2. 使用 AES-256-GCM 加密                                    │     │
│  │  3. 加密后的密文存储到 MySQL                                  │     │
│  │  4. 密钥存储在环境变量或配置中心                              │     │
│  └─────────────────────────────────────────────────────────────┘     │
│       │                                                              │
│       ▼                                                              │
│  ┌─────────────────────────────────────────────────────────────┐     │
│  │  MySQL 存储内容示例：                                         │     │
│  │  {                                                           │     │
│  │    "account_email": "user@126.com",                         │     │
│  │    "encrypted_credential": "aGVsbG8gd29ybGQ...",            │     │
│  │    "iv": "random_iv_value_16bytes",                         │     │
│  │    "provider": "126"                                         │     │
│  │  }                                                           │     │
│  └─────────────────────────────────────────────────────────────┘     │
│       │                                                              │
│       ▼                                                              │
│  ┌─────────────────────────────────────────────────────────────┐     │
│  │  运行时解密流程：                                              │     │
│  │  1. 从数据库读取加密数据                                       │     │
│  │  2. 使用环境变量中的密钥 + IV 解密                            │     │
│  │  3. 解密后的凭证仅存在于内存中                                 │     │
│  │  4. 使用后立即清零内存                                         │     │
│  └─────────────────────────────────────────────────────────────┘     │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

#### 3.3.2 密钥管理方案

| 环境 | 密钥存储方式 | 说明 |
|------|-------------|------|
| **本地开发** | `.env` 文件 | 密钥不提交到版本控制 |
| **Docker部署** | Docker Secret / 环境变量注入 | 通过 docker-compose 注入 |
| **生产环境** | 专用密钥管理服务 | 如阿里云KMS、本地Vault |

#### 3.3.3 凭证加密实现

```go
// internal/pkg/crypto/credential.go

package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
)

// CredentialEncryptor 凭证加密器
type CredentialEncryptor struct {
    key []byte // 32字节，AES-256
}

// NewCredentialEncryptor 创建加密器
func NewCredentialEncryptor(masterKey string) (*CredentialEncryptor, error) {
    key := deriveKey(masterKey) // 使用KDF函数从主密钥派生
    return &CredentialEncryptor{key: key}, nil
}

// Encrypt 加密凭证
func (e *CredentialEncryptor) Encrypt(plaintext string) (encrypted, iv string, err error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", "", err
    }

    // GCM模式
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", "", err
    }

    // 生成随机IV
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", "", err
    }

    // 加密
    ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

    return base64.StdEncoding.EncodeToString(ciphertext),
           base64.StdEncoding.EncodeToString(nonce),
           nil
}

// Decrypt 解密凭证
func (e *CredentialEncryptor) Decrypt(encrypted, iv string) (string, error) {
    ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
    if err != nil {
        return "", err
    }

    nonce, err := base64.StdEncoding.DecodeString(iv)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}
```

### 3.4 邮件获取模块设计

#### 3.4.1 Provider接口设计

```go
// internal/pkg/email/provider/provider.go

package provider

import (
    "context"
    "time"
)

// EmailProvider 邮件提供商接口
type EmailProvider interface {
    // Provider名称
    Name() string

    // 连接邮箱
    Connect(ctx context.Context, credential string) error

    // 获取邮件列表摘要
    FetchEmailList(ctx context.Context, since time.Time, limit int) ([]*EmailSummary, error)

    // 获取邮件详情
    FetchEmailDetail(ctx context.Context, messageID string) (*Email, error)

    // 断开连接
    Disconnect() error

    // 检查连接状态
    IsConnected() bool
}

// EmailSummary 邮件摘要
type EmailSummary struct {
    MessageID   string
    Subject     string
    Sender      string
    ReceivedAt  time.Time
    HasAttachment bool
    Size        int
}

// Email 完整邮件
type Email struct {
    MessageID      string
    Subject        string
    SenderName     string
    SenderEmail    string
    To             string
    CC             []string
    Content        string
    ContentHTML    string
    ContentType    string
    ReceivedAt     time.Time
    HasAttachment  bool
    Attachments    []Attachment
}
```

#### 3.4.2 126邮箱Provider实现

```go
// internal/pkg/email/provider/net126.go

package provider

import (
    "context"
    "crypto/tls"
    "fmt"
    "imap"
    "net/mail"
    "strings"
    "time"
)

// Net126Provider 网易126邮箱Provider
type Net126Provider struct {
    server     string
    port       int
    username   string
    credential string  // 授权码（已解密）
    client     *imap.Client
}

// NewNet126Provider 创建126邮箱Provider
func NewNet126Provider() *Net126Provider {
    return &Net126Provider{
        server: "imap.126.com",
        port:   993,
    }
}

func (p *Net126Provider) Name() string {
    return "126"
}

func (p *Net126Provider) Connect(ctx context.Context, credential string) error {
    p.credential = credential

    // 连接IMAP服务器（SSL）
    conn, err := tls.Dial("tcp",
        fmt.Sprintf("%s:%d", p.server, p.port),
        &tls.Config{ServerName: p.server},
    )
    if err != nil {
        return fmt.Errorf("连接126邮箱失败: %w", err)
    }

    p.client = imap.NewClient(conn)

    // 登录
    username := p.username
    if err := p.client.Login(username, p.credential); err != nil {
        return fmt.Errorf("登录126邮箱失败: %w", err)
    }

    return nil
}

func (p *Net126Provider) FetchEmailList(ctx context.Context, since time.Time, limit int) ([]*EmailSummary, error) {
    // 选择收件箱
    mailbox, err := p.client.Select("INBOX", true)
    if err != nil {
        return nil, fmt.Errorf("选择收件箱失败: %w", err)
    }

    // 构建查询条件：since指定日期之后的新邮件
    criteria := fmt.Sprintf("SINCE %s", since.Format("02-Jan-2006"))

    // 搜索邮件
    ids, err := p.client.Search(criteria)
    if err != nil {
        return nil, fmt.Errorf("搜索邮件失败: %w", err)
    }

    // 限制数量
    if len(ids) > limit {
        ids = ids[len(ids)-limit:]
    }

    // 获取邮件摘要
    summaries := make([]*EmailSummary, 0, len(ids))
    for _, id := range ids {
        summary, err := p.fetchHeader(id, mailbox)
        if err != nil {
            continue
        }
        summaries = append(summaries, summary)
    }

    return summaries, nil
}

func (p *Net126Provider) FetchEmailDetail(ctx context.Context, messageID string) (*Email, error) {
    // 根据MessageID获取邮件详情
    seqSet := &imap.SeqSet{}
    seqSet.AddNum(messageIDToUID(messageID))

    // 获取邮件内容
    section := &imap.BodySectionName{
        Header: true,
        Body:   true,
    }

    msg := &imap.Message{}
    if err := p.client.Fetch(seqSet, section, msg); err != nil {
        return nil, err
    }

    return p.parseEmail(msg)
}

func (p *Net126Provider) parseEmail(msg *imap.Message) (*Email, error) {
    // 解析邮件头
    header := msg.Envelope

    email := &Email{
        MessageID:   header.MessageID,
        Subject:     header.Subject,
        SenderName:  header.From[0].PersonalName,
        SenderEmail: header.From[0].MailboxName + "@" + header.From[0].HostName,
        ReceivedAt:  header.InternalDate,
    }

    // 解析正文
    for _, part := range msg.Body {
        contentType := part.Header.Get("Content-Type")
        if strings.Contains(contentType, "text/plain") {
            email.Content = string(part.Body)
            email.ContentType = "text/plain"
        } else if strings.Contains(contentType, "text/html") {
            email.ContentHTML = string(part.Body)
        }
    }

    // 检查附件
    if len(msg.Envelope.Envelope) > 0 {
        email.HasAttachment = true
    }

    return email, nil
}
```

---

## 4. Web端设计

### 4.1 功能模块划分

| 模块 | 功能描述 | 核心组件 |
|------|---------|---------|
| **邮件列表** | 邮件列表展示、分类筛选、搜索、批量操作 | EmailList, EmailCard, FilterBar |
| **邮件详情** | 邮件内容查看、分类标签、行动项展示 | EmailDetail, ClassificationBadge |
| **每日摘要** | 今日邮件汇总、重要邮件、待办事项 | DailySummary, ActionItemsList |
| **统计分析** | 分类趋势、发件人统计、处理效率 | StatsDashboard, CategoryChart |
| **设置中心** | 邮箱账户管理、LLM配置 | SettingsPage, AccountForm |

### 4.2 页面结构

```
src/
├── pages/
│   ├── Dashboard/          # 首页仪表盘
│   │   ├── Dashboard.tsx
│   │   ├── DailySummary.tsx
│   │   ├── QuickStats.tsx
│   │   └── RecentEmails.tsx
│   │
│   ├── EmailList/          # 邮件列表
│   │   ├── EmailList.tsx
│   │   ├── EmailCard.tsx
│   │   └── FilterBar.tsx
│   │
│   ├── EmailDetail/        # 邮件详情
│   │   ├── EmailDetail.tsx
│   │   └── ActionItems.tsx
│   │
│   ├── Stats/              # 统计分析
│   │   ├── StatsPage.tsx
│   │   └── Charts.tsx
│   │
│   └── Settings/           # 设置中心
│       ├── SettingsPage.tsx
│       └── AccountForm.tsx
│
├── components/
│   ├── ui/                 # shadcn/ui 组件
│   ├── layout/             # 布局组件
│   └── email/              # 邮件专用组件
│
└── api/
    ├── client.ts           # API客户端
    └── types.ts            # 类型定义
```

### 4.3 核心API调用

```typescript
// src/api/client.ts
import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: 30000,
});

// 请求拦截：添加Token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截：统一错误处理
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // 跳转登录
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;

// src/api/email.ts
export const emailApi = {
  list: (params: { page?: number; category?: string }) =>
    api.get('/emails', { params }),

  getById: (id: string) =>
    api.get(`/emails/${id}`),

  classify: (id: string) =>
    api.post(`/emails/${id}/classify`),

  getSummary: (date: string) =>
    api.get('/summary/daily', { params: { date } }),
};
```

---

## 5. 服务端设计

### 5.1 API接口设计

#### 5.1.1 邮件管理接口

| 方法 | 路径 | 描述 | 请求体/参数 |
|------|------|------|------------|
| GET | `/api/v1/emails` | 获取邮件列表 | `page`, `pageSize`, `category`, `status` |
| GET | `/api/v1/emails/:id` | 获取邮件详情 | - |
| POST | `/api/v1/emails/:id/classify` | 分类邮件 | - |
| POST | `/api/v1/emails/sync` | 同步邮件 | - |

#### 5.1.2 摘要接口

| 方法 | 路径 | 描述 | 请求体/参数 |
|------|------|------|------------|
| GET | `/api/v1/summary/daily` | 每日摘要 | `date` |
| GET | `/api/v1/summary/weekly` | 周报 | `startDate`, `endDate` |

#### 5.1.3 账户接口

| 方法 | 路径 | 描述 | 请求体/参数 |
|------|------|------|------------|
| GET | `/api/v1/accounts` | 获取账户列表 | - |
| POST | `/api/v1/accounts` | 添加账户 | `{ email, provider, credential }` |
| DELETE | `/api/v1/accounts/:id` | 删除账户 | - |
| POST | `/api/v1/accounts/:id/test` | 测试连接 | - |

### 5.2 数据库设计

```sql
-- 用户表
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 邮箱账户表
CREATE TABLE email_accounts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    provider VARCHAR(20) NOT NULL,          -- 126, gmail, outlook, imap
    account_email VARCHAR(255) NOT NULL,
    encrypted_credential TEXT NOT NULL,      -- AES加密的授权码
    credential_iv VARCHAR(64) NOT NULL,      -- 加密IV
    last_sync_at DATETIME,
    sync_enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_email (user_id, account_email)
);

-- 邮件表
CREATE TABLE emails (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    message_id VARCHAR(255) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,

    -- 发件人信息
    sender_name VARCHAR(255),
    sender_email VARCHAR(255) NOT NULL,

    -- 邮件内容
    subject VARCHAR(512),
    content TEXT,
    content_html TEXT,
    content_type VARCHAR(20) DEFAULT 'text/plain',

    -- 分类信息
    category VARCHAR(50) DEFAULT 'unclassified',
    priority VARCHAR(20) DEFAULT 'medium',
    confidence_score DECIMAL(5,4) DEFAULT 0,
    classification_reason TEXT,

    -- 状态
    status VARCHAR(20) DEFAULT 'unread',
    is_processed BOOLEAN DEFAULT FALSE,
    has_attachment BOOLEAN DEFAULT FALSE,

    -- 时间
    received_at DATETIME NOT NULL,
    processed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (account_id) REFERENCES email_accounts(id) ON DELETE CASCADE,

    INDEX idx_user_category (user_id, category),
    INDEX idx_user_received (user_id, received_at DESC),
    INDEX idx_status (status)
);

-- 行动项表
CREATE TABLE action_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    task TEXT NOT NULL,
    task_type VARCHAR(50),                  -- reply, review, deadline, meeting
    deadline DATETIME,
    priority VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_status (user_id, status),
    INDEX idx_deadline (deadline)
);

-- 每日摘要表
CREATE TABLE daily_summaries (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    summary_date DATE NOT NULL,
    summary_content TEXT,
    important_emails JSON,
    pending_actions JSON,
    statistics JSON,
    generated_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_date (user_id, summary_date)
);

-- LLM配置表
CREATE TABLE llm_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    provider VARCHAR(50) NOT NULL,          -- deepseek, zhipu, qwen
    model_name VARCHAR(100),
    api_key_encrypted TEXT,
    api_key_iv VARCHAR(64),
    base_url VARCHAR(255),
    config_json JSON,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### 5.3 服务端目录结构

```
email-backend/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
│
├── internal/
│   ├── handler/                 # HTTP处理器
│   │   ├── email.go
│   │   ├── account.go
│   │   ├── summary.go
│   │   └── auth.go
│   │
│   ├── service/                 # 业务逻辑层
│   │   ├── email_service.go
│   │   ├── account_service.go
│   │   ├── sync_service.go
│   │   └── agent_client.go      # 与Agent通信
│   │
│   ├── pkg/
│   │   ├── email/
│   │   │   ├── provider/       # 邮件Provider
│   │   │   │   ├── provider.go  # 接口定义
│   │   │   │   ├── net126.go   # 126邮箱
│   │   │   │   ├── gmail.go    # Gmail
│   │   │   │   └── imap.go     # 通用IMAP
│   │   │   ├── fetcher.go      # 邮件拉取
│   │   │   └── parser.go       # 邮件解析
│   │   │
│   │   ├── crypto/             # 加密工具
│   │   │   └── credential.go
│   │   │
│   │   └── response/           # 统一响应
│   │       └── response.go
│   │
│   ├── model/                   # 数据模型
│   │   ├── user.go
│   │   ├── email.go
│   │   └── account.go
│   │
│   ├── repository/              # 数据访问层
│   │   ├── email_repo.go
│   │   └── account_repo.go
│   │
│   └── middleware/              # 中间件
│       ├── auth.go
│       └── cors.go
│
├── config/
│   └── config.yaml              # 配置文件
│
├── migrations/                  # 数据库迁移
│   └── 001_init.sql
│
├── go.mod
├── go.sum
└── Dockerfile
```

---

## 6. Agent端设计

### 6.1 Agent架构

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                            Agent 端架构                                         │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                                │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                          FastAPI 服务层                                   │  │
│  │                                                                          │  │
│  │   POST /api/v1/classify       - 单封邮件分类                              │  │
│  │   POST /api/v1/batch-classify - 批量邮件分类                              │  │
│  │   POST /api/v1/extract        - 信息提取                                   │  │
│  │   POST /api/v1/summary        - 生成摘要                                   │  │
│  │   POST /api/v1/process        - 完整处理流程                               │  │
│  │                                                                          │  │
│  └──────────────────────────────────┬───────────────────────────────────────────┘  │
│                                     │                                              │
│  ┌──────────────────────────────────▼───────────────────────────────────────────┐  │
│  │                          Orchestrator 编排器                                │  │
│  │                                                                          │  │
│  │   process_email(email) → 分类 → 信息提取 → 向量化 → 结果组装              │  │
│  │                                                                          │  │
│  └──────────────────────────────────┬───────────────────────────────────────────┘  │
│                                     │                                              │
│  ┌──────────────────────────────────▼───────────────────────────────────────────┐  │
│  │                           Agent 执行层                                      │  │
│  │                                                                          │  │
│  │   ┌───────────────┐  ┌───────────────┐  ┌───────────────┐                │  │
│  │   │ Classification │  │  Extraction   │  │    Summary    │                │  │
│  │   │     Agent      │  │     Agent     │  │     Agent     │                │  │
│  │   │               │  │               │  │               │                │  │
│  │   │ 邮件分类      │  │ 行动项提取     │  │ 每日摘要      │                │  │
│  │   │ 优先级判断    │  │ 会议信息      │  │ 重要邮件汇总  │                │  │
│  │   │ 置信度评估    │  │ 截止日期识别   │  │ 统计数据生成  │                │  │
│  │   └───────────────┘  └───────────────┘  └───────────────┘                │  │
│  │                                                                          │  │
│  └──────────────────────────────────┬───────────────────────────────────────────┘  │
│                                     │                                              │
│  ┌──────────────────────────────────▼───────────────────────────────────────────┐  │
│  │                           LLM 适配层                                        │  │
│  │                                                                          │  │
│  │   ┌─────────────────────────────────────────────────────────────────┐     │  │
│  │   │                      LLMManager                                  │     │  │
│  │   │   ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐     │     │  │
│  │   │   │ DeepSeek  │ │  智谱GLM  │ │ 通义千问  │ │  Kimi    │     │     │  │
│  │   │   │ Provider  │ │ Provider  │ │ Provider  │ │ Provider│     │     │  │
│  │   │   └───────────┘ └───────────┘ └───────────┘ └───────────┘     │     │  │
│  │   │                                                                  │     │  │
│  │   │   routing: { classification: deepseek, extraction: zhipu }     │     │  │
│  │   └─────────────────────────────────────────────────────────────────┘     │  │
│  │                                                                          │  │
│  └────────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                           ChromaDB 向量存储                                │  │
│  │   • 邮件内容向量化                                                        │  │
│  │   • 语义相似邮件检索                                                      │  │
│  │   • Embedding Model: BAAI/bge-m3 (支持中文)                               │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
└────────────────────────────────────────────────────────────────────────────────┘
```

### 6.2 分类策略

#### 6.2.1 分类类别体系

| 类别 | 名称 | 触发特征 | 优先级建议 |
|------|------|---------|-----------|
| `work_urgent` | 紧急工作 | 截止今天/明天、领导邮件、紧急关键词 | critical/high |
| `work_normal` | 普通工作 | 工作相关、同事邮件、常规沟通 | medium |
| `personal` | 个人邮件 | 私人话题、朋友家人、社交邀请 | medium/low |
| `subscription` | 订阅邮件 | 新闻简报、技术博客、定期推送 | low |
| `notification` | 系统通知 | GitHub、Jira、CI/CD通知 | low |
| `promotion` | 营销推广 | 广告、促销、商业推广 | low |
| `spam` | 垃圾邮件 | 可疑内容、诈骗信息 | low |

#### 6.2.2 分类Agent提示词

```python
# app/prompts/classification.py

CLASSIFICATION_SYSTEM_PROMPT = """你是一个专业的邮件分类助手。请根据以下规则准确分析邮件并进行分类。

## 分类标准

### 类别定义
| 类别 | 描述 | 典型场景 |
|------|------|---------|
| work_urgent | 紧急工作事务 | 项目截止通知、领导紧急指令 |
| work_normal | 普通工作事务 | 常规工作沟通、会议通知 |
| personal | 个人事务 | 家人朋友联系、私人邀请 |
| subscription | 订阅推送 | 技术博客、新闻简报 |
| notification | 系统通知 | GitHub PR、CI/CD、告警 |
| promotion | 营销推广 | 广告、促销、推广内容 |
| spam | 垃圾邮件 | 诈骗、可疑内容 |

### 优先级判断
- critical: 今天截止，或来自直接领导的紧急邮件
- high: 本周截止，有明确行动请求
- medium: 有价值信息，本周内处理即可
- low: 信息性内容，可延后处理

## 输出格式
返回JSON格式的分类结果，包含：
- category: 分类结果
- priority: 优先级
- confidence: 置信度 (0-1)
- reasoning: 判断理由

请严格按JSON格式输出，不要包含其他内容。"""

CLASSIFICATION_USER_PROMPT = """请分析以下邮件并分类：

## 邮件信息
- 发件人: {sender_name}
- 发件人邮箱: {sender_email}
- 主题: {subject}
- 接收时间: {received_at}
- 正文:
{content}

## 输出要求
请严格按以下JSON格式输出：
{{"category": "...", "priority": "...", "confidence": 0.0, "reasoning": "..."}}"""
```

### 6.3 信息提取Agent

```python
# app/schemas/extraction.py

from pydantic import BaseModel, Field
from typing import List, Optional
from enum import Enum

class EmailIntent(str, Enum):
    REQUEST = "request"           # 请求行动
    INFORMATION = "information"  # 提供信息
    INVITATION = "invitation"    # 邀请参会
    APPROVAL = "approval"         # 需要审批
    DISCUSSION = "discussion"     # 需要讨论
    NOTIFICATION = "notification" # 系统通知

class ActionItem(BaseModel):
    task: str                     # 任务描述
    task_type: str               # reply/review/submit/attend/prepare
    deadline: Optional[str]      # 截止时间
    priority: str                # high/medium/low
    confidence: float = Field(ge=0, le=1)

class MeetingInfo(BaseModel):
    title: str
    time: str
    location: Optional[str]
    attendees: List[str]
    meeting_url: Optional[str]

class ExtractionOutput(BaseModel):
    action_items: List[ActionItem] = []
    meetings: List[MeetingInfo] = []
    key_entities: List[str] = []
    summary: str
    intent: EmailIntent
```

### 6.4 Agent端目录结构

```
email-agent/
├── app/
│   ├── main.py                  # FastAPI入口
│   │
│   ├── api/
│   │   └── routes/
│   │       ├── classify.py
│   │       ├── extract.py
│   │       └── summary.py
│   │
│   ├── agents/
│   │   ├── orchestrator.py      # 任务编排器
│   │   ├── classification.py    # 分类Agent
│   │   ├── extraction.py        # 提取Agent
│   │   └── summary.py           # 摘要Agent
│   │
│   ├── llm/
│   │   ├── manager.py           # LLM管理器
│   │   ├── deepseek.py
│   │   ├── zhipu.py
│   │   └── qwen.py
│   │
│   ├── embeddings/
│   │   └── bge.py               # BGE向量化
│   │
│   ├── vectorstore/
│   │   └── chroma.py            # ChromaDB集成
│   │
│   ├── prompts/
│   │   ├── classification.py
│   │   └── extraction.py
│   │
│   └── schemas/
│       ├── request.py
│       └── response.py
│
├── config/
│   └── config.yaml
│
├── requirements.txt
└── Dockerfile
```

---

## 7. 部署方案

### 7.1 部署架构概览

系统支持**分离部署**，各服务可独立运行在不同服务器上，通过HTTP REST API进行通信。

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                            分离部署架构                                         │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                                │
│   ┌──────────────────────────────────────────────────────────────────────┐    │
│   │                    Nginx 反向代理 (可选)                               │    │
│   │                       https://your-domain.com                         │    │
│   └──────────────────────────────┬───────────────────────────────────────────┘    │
│                                   │                                              │
│            ┌──────────────────────┼──────────────────────┐                      │
│            │                      │                      │                      │
│            ▼                      ▼                      ▼                      │
│   ┌────────────────┐    ┌────────────────┐    ┌────────────────┐              │
│   │   Web 前端     │    │   服务端       │    │   Agent 端     │              │
│   │                │    │                │    │                │              │
│   │  服务器 A      │    │  服务器 B      │    │  服务器 C      │              │
│   │  Port: 80/443  │    │  Port: 8080    │    │  Port: 8001    │              │
│   │                │    │                │    │                │              │
│   │  Nginx静态托管 │    │  Go + Gin      │    │  Python+FastAPI│              │
│   │  或CDN加速     │    │                │    │                │              │
│   └────────┬───────┘    └────────┬───────┘    └────────┬───────┘              │
│            │                     │                     │                       │
│            │    HTTP REST API    │    HTTP REST API    │                       │
│            └────────────────────►│◄────────────────────┘                       │
│                                  │                                              │
│                                  ▼                                              │
│                        ┌────────────────────────┐                              │
│                        │       数据层           │                              │
│                        │  MySQL + Redis + ChromaDB                              │
│                        │  (可部署在服务器B或独立)                                │
│                        └────────────────────────┘                              │
│                                                                                │
└────────────────────────────────────────────────────────────────────────────────┘
```

### 7.2 服务间通信协议

#### 7.2.1 Web → Server 通信

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/emails` | GET | 获取邮件列表 |
| `/api/v1/emails/:id` | GET | 获取邮件详情 |
| `/api/v1/emails/:id/classify` | POST | 触发邮件分类 |
| `/api/v1/accounts` | GET/POST | 邮箱账户管理 |
| `/api/v1/summary/daily` | GET | 获取每日摘要 |

#### 7.2.2 Server → Agent 通信

| 接口 | 方法 | 说明 | 调用方式 |
|------|------|------|---------|
| `/api/v1/classify` | POST | 单封邮件分类 | 同步HTTP |
| `/api/v1/batch-classify` | POST | 批量邮件分类 | 异步(Redis队列) |
| `/api/v1/extract` | POST | 信息提取 | 同步HTTP |
| `/api/v1/summary` | POST | 生成摘要 | 异步(Redis队列) |
| `/health` | GET | 健康检查 | 同步HTTP |

#### 7.2.3 通信时序图

```
┌────────┐          ┌────────┐          ┌────────┐          ┌────────┐
│  Web   │          │ Server │          │ Agent  │          │ Redis  │
└───┬────┘          └───┬────┘          └───┬────┘          └───┬────┘
    │                   │                   │                   │
    │  GET /emails      │                   │                   │
    │──────────────────►│                   │                   │
    │                   │                   │                   │
    │    邮件列表        │                   │                   │
    │◄──────────────────│                   │                   │
    │                   │                   │                   │
    │  POST /classify   │                   │                   │
    │──────────────────►│                   │                   │
    │                   │  POST /classify   │                   │
    │                   │──────────────────►│                   │
    │                   │                   │                   │
    │                   │  分类结果          │                   │
    │                   │◄──────────────────│                   │
    │    分类完成        │                   │                   │
    │◄──────────────────│                   │                   │
    │                   │                   │                   │
    │                   │  批量任务推送      │                   │
    │                   │─────────────────────────────────────►│
    │                   │                   │                   │
    │                   │                   │  拉取任务         │
    │                   │                   │◄──────────────────│
    │                   │                   │                   │
    │                   │  结果回调          │                   │
    │                   │◄──────────────────│                   │
    │                   │                   │                   │
```

### 7.3 Docker Compose配置

#### 7.3.1 开发环境 (单机部署)

```yaml
# docker-compose.yml - 开发环境单机部署
version: '3.8'

services:
  # Web前端
  web:
    build:
      context: ./email-web
      dockerfile: Dockerfile
    container_name: email-web
    ports:
      - "80:80"
    depends_on:
      - server
    networks:
      - email-network
    environment:
      - VITE_API_BASE_URL=http://server:8080/api/v1

  # Go服务端
  server:
    build:
      context: ./email-backend
      dockerfile: Dockerfile
    container_name: email-server
    ports:
      - "8080:8080"
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_started
      agent:
        condition: service_started
    networks:
      - email-network
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USER=root
      - DB_PASSWORD=${MYSQL_ROOT_PASSWORD}
      - DB_NAME=email_system
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - AGENT_URL=http://agent:8001
      - AGENT_API_KEY=${AGENT_API_KEY}
      - CREDENTIAL_KEY=${CREDENTIAL_ENCRYPTION_KEY}
    volumes:
      - ./email-backend/config:/app/config

  # Python Agent
  agent:
    build:
      context: ./email-agent
      dockerfile: Dockerfile
    container_name: email-agent
    ports:
      - "8001:8001"
    depends_on:
      redis:
        condition: service_started
      chroma:
        condition: service_started
    networks:
      - email-network
    environment:
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - CHROMA_HOST=chroma
      - CHROMA_PORT=8000
      - DEEPSEEK_API_KEY=${DEEPSEEK_API_KEY}
    volumes:
      - ./email-agent/config:/app/config

  # MySQL数据库
  mysql:
    image: mysql:8.0
    container_name: email-mysql
    ports:
      - "3306:3306"
    networks:
      - email-network
    environment:
      - MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
      - MYSQL_DATABASE=email_system
    volumes:
      - mysql-data:/var/lib/mysql
      - ./sql/init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis
  redis:
    image: redis:7-alpine
    container_name: email-redis
    ports:
      - "6379:6379"
    networks:
      - email-network
    volumes:
      - redis-data:/data

  # ChromaDB向量数据库
  chroma:
    image: ghcr.io/chroma-core/chroma:latest
    container_name: email-chroma
    ports:
      - "8000:8000"
    networks:
      - email-network
    volumes:
      - chroma-data:/chroma/chroma-data

networks:
  email-network:
    driver: bridge

volumes:
  mysql-data:
  redis-data:
  chroma-data:
```

#### 7.3.2 生产环境 (分离部署)

**服务器A - Web前端 (docker-compose.web.yml)**
```yaml
# docker-compose.web.yml - Web前端独立部署
version: '3.8'

services:
  web:
    build:
      context: ./email-web
      dockerfile: Dockerfile
    container_name: email-web
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/ssl:/etc/nginx/ssl:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
    environment:
      - SERVER_URL=https://api.your-domain.com
      - AGENT_URL=https://agent.your-domain.com
    restart: unless-stopped
```

**服务器B - 服务端 (docker-compose.server.yml)**
```yaml
# docker-compose.server.yml - 服务端独立部署
version: '3.8'

services:
  server:
    build:
      context: ./email-backend
      dockerfile: Dockerfile
    container_name: email-server
    ports:
      - "8080:8080"
    environment:
      # 数据库配置 (指向独立数据库服务器)
      - DB_HOST=${DB_HOST}
      - DB_PORT=${DB_PORT:-3306}
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=email_system
      # Redis配置
      - REDIS_HOST=${REDIS_HOST}
      - REDIS_PORT=${REDIS_PORT:-6379}
      # Agent配置 (指向Agent服务器)
      - AGENT_URL=${AGENT_URL}
      - AGENT_API_KEY=${AGENT_API_KEY}
      # 安全配置
      - CREDENTIAL_KEY=${CREDENTIAL_ENCRYPTION_KEY}
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

**服务器C - Agent端 (docker-compose.agent.yml)**
```yaml
# docker-compose.agent.yml - Agent端独立部署
version: '3.8'

services:
  agent:
    build:
      context: ./email-agent
      dockerfile: Dockerfile
    container_name: email-agent
    ports:
      - "8001:8001"
    environment:
      # Redis配置
      - REDIS_HOST=${REDIS_HOST}
      - REDIS_PORT=${REDIS_PORT:-6379}
      # ChromaDB配置
      - CHROMA_HOST=${CHROMA_HOST}
      - CHROMA_PORT=${CHROMA_PORT:-8000}
      # LLM配置
      - DEEPSEEK_API_KEY=${DEEPSEEK_API_KEY}
      - ZHIPU_API_KEY=${ZHIPU_API_KEY}
    volumes:
      - ./config:/app/config
      - chroma-data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8001/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ChromaDB (可与Agent部署在同一服务器)
  chroma:
    image: ghcr.io/chroma-core/chroma:latest
    container_name: email-chroma
    ports:
      - "8000:8000"
    volumes:
      - chroma-data:/chroma/chroma-data
    restart: unless-stopped

volumes:
  chroma-data:
```

### 7.4 Nginx配置示例

```nginx
# nginx/conf.d/default.conf

# Web前端
server {
    listen 80;
    listen 443 ssl;
    server_name your-domain.com;

    # SSL配置
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    # 前端静态资源
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }

    # API代理到服务端
    location /api/ {
        proxy_pass http://${SERVER_IP}:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Agent API (可选，如果需要直接访问)
server {
    listen 8001;
    server_name agent.your-domain.com;

    location / {
        proxy_pass http://${AGENT_IP}:8001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 7.5 环境变量配置

#### 7.5.1 开发环境

```bash
# .env.example - 开发环境配置模板

# MySQL配置
MYSQL_ROOT_PASSWORD=your_secure_mysql_password

# 凭证加密密钥 (32字节hex字符串，使用 openssl rand -hex 32 生成)
CREDENTIAL_ENCRYPTION_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef

# JWT密钥
JWT_SECRET=your_jwt_secret_key

# Agent API认证
AGENT_API_KEY=your_agent_api_key

# LLM API Keys
DEEPSEEK_API_KEY=sk-your-deepseek-key
ZHIPU_API_KEY=your-zhipu-api-key

# 服务URL (开发环境使用容器名)
SERVER_URL=http://server:8080
AGENT_URL=http://agent:8001
```

#### 7.5.2 生产环境 (分离部署)

**服务器A - Web前端 (.env.web)**
```bash
# 服务端API地址
SERVER_URL=https://api.your-domain.com
AGENT_URL=https://agent.your-domain.com
```

**服务器B - 服务端 (.env.server)**
```bash
# 数据库配置 (指向独立数据库服务器)
DB_HOST=your-db-server.com
DB_PORT=3306
DB_USER=email_user
DB_PASSWORD=secure_db_password
DB_NAME=email_system

# Redis配置
REDIS_HOST=your-redis-server.com
REDIS_PORT=6379

# Agent配置 (指向Agent服务器)
AGENT_URL=https://agent.your-domain.com
AGENT_API_KEY=your_agent_api_key

# 安全配置
CREDENTIAL_ENCRYPTION_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
JWT_SECRET=your_jwt_secret_key
```

**服务器C - Agent端 (.env.agent)**
```bash
# Redis配置
REDIS_HOST=your-redis-server.com
REDIS_PORT=6379

# ChromaDB (本地或远程)
CHROMA_HOST=localhost
CHROMA_PORT=8000

# LLM配置
DEEPSEEK_API_KEY=sk-your-deepseek-key
ZHIPU_API_KEY=your-zhipu-api-key
```

### 7.6 部署脚本

#### 7.6.1 开发环境一键部署

```bash
#!/bin/bash
# deploy-dev.sh - 开发环境一键部署

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
echo_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# 检查环境
check_env() {
    echo_info "检查系统环境..."
    command -v docker >/dev/null 2>&1 || { echo "Docker未安装"; exit 1; }
    command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose未安装"; exit 1; }
    echo_info "Docker: $(docker --version)"
}

# 初始化配置
init_config() {
    if [ ! -f ".env" ]; then
        echo_warn "创建.env配置文件..."
        cat > .env << 'EOF'
MYSQL_ROOT_PASSWORD=change_this_password
CREDENTIAL_ENCRYPTION_KEY=$(openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 16)
AGENT_API_KEY=$(openssl rand -hex 16)
DEEPSEEK_API_KEY=your_deepseek_api_key
ZHIPU_API_KEY=your_zhipu_api_key
EOF
        echo_warn "请编辑.env文件配置API密钥"
        exit 1
    fi
    mkdir -p sql logs
}

# 部署
deploy() {
    check_env
    init_config
    echo_info "拉取镜像..."
    docker-compose pull
    echo_info "构建服务..."
    docker-compose build
    echo_info "启动服务..."
    docker-compose up -d
    echo_info "等待服务就绪..."
    sleep 15
    docker-compose ps
    echo_info "部署完成！访问 http://localhost"
}

case "$1" in
    deploy) deploy ;;
    start) docker-compose up -d ;;
    stop) docker-compose down ;;
    restart) docker-compose restart ;;
    status) docker-compose ps ;;
    logs) docker-compose logs -f $2 ;;
    *) echo "用法: $0 {deploy|start|stop|restart|status|logs [service]}" ;;
esac
```

#### 7.6.2 生产环境分离部署

```bash
#!/bin/bash
# deploy-prod.sh - 生产环境分离部署

set -e

# 部署Web前端
deploy_web() {
    echo "部署Web前端到服务器A..."
    scp docker-compose.web.yml user@web-server:/opt/email-agent/
    scp .env.web user@web-server:/opt/email-agent/.env
    ssh user@web-server "cd /opt/email-agent && docker-compose -f docker-compose.web.yml up -d"
}

# 部署服务端
deploy_server() {
    echo "部署服务端到服务器B..."
    scp docker-compose.server.yml user@server:/opt/email-agent/
    scp .env.server user@server:/opt/email-agent/.env
    ssh user@server "cd /opt/email-agent && docker-compose -f docker-compose.server.yml up -d"
}

# 部署Agent端
deploy_agent() {
    echo "部署Agent端到服务器C..."
    scp docker-compose.agent.yml user@agent:/opt/email-agent/
    scp .env.agent user@agent:/opt/email-agent/.env
    ssh user@agent "cd /opt/email-agent && docker-compose -f docker-compose.agent.yml up -d"
}

case "$1" in
    web) deploy_web ;;
    server) deploy_server ;;
    agent) deploy_agent ;;
    all) deploy_web && deploy_server && deploy_agent ;;
    *) echo "用法: $0 {web|server|agent|all}" ;;
esac
```

---

## 8. 项目目录结构

```
mail-agent/
├── docs/                        # 本文档目录
│   └── DESIGN.md
│
├── email-web/                   # Web前端
│   ├── src/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── api/
│   │   └── App.tsx
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile
│
├── email-backend/               # Go服务端
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── pkg/
│   │   └── model/
│   ├── config/
│   ├── migrations/
│   ├── go.mod
│   └── Dockerfile
│
├── email-agent/                 # Python Agent
│   ├── app/
│   │   ├── api/
│   │   ├── agents/
│   │   ├── llm/
│   │   └── schemas/
│   ├── config/
│   ├── requirements.txt
│   └── Dockerfile
│
├── sql/
│   └── init.sql                 # 数据库初始化脚本
│
├── docker-compose.yml
├── deploy.sh                    # 一键部署脚本
├── .env.example                 # 环境变量示例
└── README.md                    # 项目说明
```

---

## 9. 实现计划

### Phase 1: 基础框架 (2-3天)
- [ ] 搭建三端项目结构
- [ ] 配置Docker Compose
- [ ] 实现126邮箱Provider
- [ ] 基础API连通性测试

### Phase 2: 核心功能 (3-4天)
- [ ] 邮件获取和存储
- [ ] 基础分类Agent
- [ ] Web端邮件列表展示
- [ ] 分类结果展示

### Phase 3: 智能功能 (4-5天)
- [ ] 完整分类Agent
- [ ] 信息提取Agent
- [ ] 每日摘要生成
- [ ] 向量存储和检索

### Phase 4: 完善优化 (2-3天)
- [ ] 用户认证系统
- [ ] 统计分析功能
- [ ] 部署文档完善

---

## 10. 风险评估

| 风险项 | 影响程度 | 缓解措施 |
|--------|---------|---------|
| 126邮箱API限制 | 中 | 使用标准IMAP协议，控制请求频率 |
| LLM服务不稳定 | 中 | 实现多Provider自动切换 |
| 凭证泄露 | 高 | 严格加密存储，密钥隔离管理 |
| 数据丢失 | 中 | 定期备份MySQL数据 |
| Agent处理超时 | 低 | 设置合理超时时间，队列异步处理 |

---

*文档版本：v1.0*
*最后更新：2026-04-07*
