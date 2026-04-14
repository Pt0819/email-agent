# Skill: Go Backend Project Setup

> 本Skill用于快速初始化Go后端项目结构（Clean Architecture）

## 1. 项目结构

```
email-backend/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
│
├── server/
│   ├── api/
│   │   └── v1/                  # API v1版本
│   │       ├── account.go      # 账户接口
│   │       ├── email.go         # 邮件接口
│   │       ├── sync.go          # 同步接口
│   │       └── response.go      # 响应辅助函数
│   │
│   ├── config/                  # 配置管理
│   │   └── config.go
│   │
│   ├── core/                    # 核心初始化
│   │   └── core.go              # InitConfig, InitDB, Close, InitEncryptor, InitProviders
│   │
│   ├── global/                  # 全局对象
│   │   └── global.go
│   │
│   ├── middleware/               # 中间件
│   │   └── middleware.go
│   │
│   ├── model/                   # 数据模型
│   │   ├── email.go            # 邮件/账户模型
│   │   ├── request/            # 请求DTO
│   │   │   └── request.go
│   │   └── response/           # 响应DTO
│   │       └── response.go
│   │
│   ├── pkg/                     # 公共包
│   │   ├── crypto/              # 加密工具
│   │   │   └── credential.go    # AES-256-GCM凭证加密
│   │   └── email/
│   │       └── provider/        # 邮件Provider接口
│   │           ├── provider.go  # Provider接口定义
│   │           ├── mock.go      # Mock实现(测试用)
│   │           └── net126.go    # 126邮箱实现
│   │
│   ├── repository/              # 数据访问层
│   │   └── repository.go
│   │
│   ├── router/                 # 路由注册
│   │   └── router.go
│   │
│   └── service/                # 业务逻辑层
│       ├── service.go          # 基础服务
│       └── sync_service.go     # 同步服务
│
├── config/                      # 配置文件
│   └── config.yaml
│
├── sql/                         # 数据库脚本
│   └── init.sql
│
├── go.mod
├── go.sum
└── server.exe                   # 编译产物
```

## 2. 分层架构

### Handler → Service → Repository

```
请求 → Handler (api/v1) → Service (service) → Repository (repository) → Database
       ↓
    响应格式化
```

### 依赖关系

- `api` 依赖 `service` 和 `model`
- `service` 依赖 `repository` 和 `model`
- `repository` 依赖 `model` 和 `gorm`
- `model` 无外部依赖

## 3. 命名规范

### 文件命名
- 使用小写字母+下划线: `email_service.go`
- 测试文件: `email_service_test.go`

### 包命名
- 使用小写: `package service`
- 避免使用下划线或混合大小写
- 导入别名避免冲突: `respModel "email-backend/server/model/response"`

### 变量/函数命名
- 驼峰命名: `GetEmailList()`
- 缩写保持大写: `HTTP`, `URL`, `API`
- 私有变量: `userName` (小写开头)
- 公开变量: `UserName` (大写开头)

### 常量命名
- 全大写下划线: `MAX_RETRY_COUNT`
- 枚举类型: `StatusPending` (类型+值)

## 4. 开发规范

### 模块初始化
```bash
# 创建项目
cd email-backend
go mod init email-backend

# 添加依赖
go get github.com/gin-gonic/gin
go get github.com/spf13/viper
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/redis/go-redis/v9
go get github.com/emersion/go-imap
```

### 配置文件格式 (YAML)
```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 3306
  username: root
  password: ${DB_PASSWORD}  # 从环境变量读取
  dbname: email_system

security:
  credential_key: ${CREDENTIAL_KEY}  # 32字节密钥用于凭证加密
```

### 环境变量使用
```go
// 从环境变量读取敏感信息
os.Getenv("DB_PASSWORD")
os.Getenv("CREDENTIAL_KEY")  // 32字节密钥
```

## 5. API响应规范

### 统一响应格式 (server/model/response/response.go)
```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    TraceID string      `json:"trace_id,omitempty"`
}

// 错误码
const (
    CodeSuccess       = 0
    CodeBadRequest    = 400
    CodeUnauthorized  = 401
    CodeForbidden     = 403
    CodeNotFound      = 404
    CodeInternalError = 500
)
```

### 响应辅助函数 (server/api/v1/response.go)
```go
// 成功响应
func success(c *gin.Context, data interface{})

// 创建成功
func created(c *gin.Context, data interface{})

// 请求参数错误
func badRequest(c *gin.Context, message string)

// 资源不存在
func notFound(c *gin.Context, message string)

// 通用错误
func errorResp(c *gin.Context, status int, message string)
```

### 使用示例
```go
func (h *EmailHandler) ListEmails(c *gin.Context) {
    emails, total, err := h.emailService.List(c.Request.Context(), req)
    if err != nil {
        errorResp(c, 500, err.Error())
        return
    }
    success(c, &model.EmailListResponse{
        List:  emails,
        Total: total,
    })
}
```

## 6. 请求DTO规范

### 文件位置
`server/model/request/request.go`

### 示例
```go
type ListRequest struct {
    Page     int    `form:"page" json:"page"`
    PageSize int    `form:"page_size" json:"page_size"`
    UserID   int64  `form:"-" json:"-"`
    Category string `form:"category" json:"category"`
    Status   string `form:"status" json:"status"`
}

type CreateAccountRequest struct {
    Provider   string `json:"provider" binding:"required"`
    Email      string `json:"email" binding:"required,email"`
    Credential string `json:"credential" binding:"required"`  # 授权码
}
```

## 7. Provider插件架构

### 实现新的邮件Provider

```go
// 1. 在 server/pkg/email/provider/ 下创建新文件
// 例如: outlook.go

package provider

type OutlookProvider struct {
    // 配置
}

func NewOutlookProvider(config *ProviderConfig) EmailProvider {
    return &OutlookProvider{...}
}

func (p *OutlookProvider) Name() string { return "outlook" }
func (p *OutlookProvider) Connect(ctx context.Context, email, credential string) error { ... }
func (p *OutlookProvider) TestConnection(ctx context.Context) (*ConnectionResult, error) { ... }
func (p *OutlookProvider) FetchEmailList(ctx context.Context, since time.Time, limit int) ([]*EmailSummary, error) { ... }
func (p *OutlookProvider) FetchEmailDetail(ctx context.Context, messageID string) (*Email, error) { ... }
func (p *OutlookProvider) FetchEmails(ctx context.Context, since time.Time, limit int) (*SyncResult, error) { ... }
func (p *OutlookProvider) Disconnect() error { ... }
func (p *OutlookProvider) IsConnected() bool { ... }

func init() {
    Register("outlook", NewOutlookProvider)  // 注册到工厂
}
```

### 使用Provider
```go
// 创建Provider
provider, ok := provider.Create("126", &provider.ProviderConfig{
    Server: "imap.126.com",
    Port:   993,
    UseSSL: true,
})
if !ok {
    return errors.New("不支持的邮件提供商")
}

// 连接
err := provider.Connect(ctx, email, credential)

// 获取邮件
result, err := provider.FetchEmails(ctx, since, limit)
```

## 8. 凭证加密使用

### 初始化加密器
```go
// 在 core/core.go 中
func InitEncryptor() error {
    key := GlobalConfig.Security.CredentialKey
    if key == "" {
        return fmt.Errorf("凭证加密密钥未配置")
    }
    enc, err := crypto.NewCredentialEncryptor(key)
    if err != nil {
        return err
    }
    GlobalEncryptor = enc
    return nil
}
```

### 加密/解密凭证
```go
// 加密
encrypted, iv, err := global.Encryptor().Encrypt(credential)

// 解密
credential, err := global.Encryptor().Decrypt(encrypted, iv)
```

## 9. 快速开始命令

```bash
# 1. 初始化项目
cd email-backend
go mod init email-backend
go mod tidy

# 2. 编译运行
go build -o server.exe .
go run .

# 3. 开发模式 (热重载)
# 需要安装air: go install github.com/air-verse/air@latest
air

# 4. 测试
go test ./...
go test -v ./server/...
```

## 10. 常用依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| gin | v1.9.x | HTTP框架 |
| viper | v1.18.x | 配置管理 |
| gorm | v1.25.x | ORM框架 |
| mysql | v1.5.x | MySQL驱动 |
| go-imap | v1.2.x | IMAP客户端 |
| go-redis | v9.3.x | Redis客户端 |

---

> 生成时间: 2026-04-08
> 更新: 2026-04-09 (新增Provider插件架构、凭证加密、同步服务)
> 适用于: Go后端开发