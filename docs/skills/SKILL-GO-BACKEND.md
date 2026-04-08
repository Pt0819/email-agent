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
│   │   └── core.go              # InitConfig, InitDB, Close
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
│   ├── repository/              # 数据访问层
│   │   └── repository.go
│   │
│   ├── router/                 # 路由注册
│   │   └── router.go
│   │
│   └── service/                # 业务逻辑层
│       └── service.go
│
├── config/                      # 配置文件
│   └── config.yaml
│
├── sql/                         # 数据库脚本
│   └── *.sql
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
```

### 环境变量使用
```go
// 从环境变量读取敏感信息
os.Getenv("DB_PASSWORD")
os.Getenv("CREDENTIAL_KEY")
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
    Provider string `json:"provider" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password"`
}
```

## 7. 快速开始命令

```bash
# 1. 初始化项目
cd email-backend
go mod init email-backend
go mod tidy

# 2. 编译运行
go build -o server.exe .
go run cmd/server/main.go

# 3. 开发模式 (热重载)
# 需要安装air: go install github.com/air-verse/air@latest
air

# 4. 测试
go test ./...
go test -v ./server/...
```

## 8. 常用依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| gin | v1.9.x | HTTP框架 |
| viper | v1.18.x | 配置管理 |
| gorm | v1.25.x | ORM框架 |
| mysql | v1.5.x | MySQL驱动 |
| go-redis | v9.3.x | Redis客户端 |
| logrus | v1.9.x | 日志 |

---

> 生成时间: 2026-04-08
> 更新: 2026-04-08 (Clean Architecture结构)
> 适用于: Go后端开发
