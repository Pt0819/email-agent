# Skill: Project Development Standards

> 本项目通用开发规范，适用于邮件分类Agent系统三端开发

## 1. 通用项目结构

```
mail-agent/
├── docs/                    # 文档
│   ├── skills/             # 可复用Skills
│   │   ├── SKILL-GO-BACKEND.md
│   │   ├── SKILL-PYTHON-AGENT.md
│   │   ├── SKILL-REACT-WEB.md
│   │   └── SKILL-COMMON.md
│   ├── DESIGN.md            # 设计文档
│   ├── REQUIREMENTS.md      # 需求文档
│   └── DAILY-*.md          # 每日报告
│
├── email-backend/           # Go服务端
├── email-agent/             # Python Agent
├── email-web/               # React前端
│
├── sql/                     # SQL脚本
├── docker/                  # Docker配置
├── configs/                 # 共享配置
│
├── docker-compose.yml       # Docker编排
├── test-dev.ps1             # 自测脚本
└── README.md
```

## 2. API通信规范

### HTTP REST API

| 路径 | 方法 | 说明 |
|------|------|------|
| `/api/v1/emails` | GET | 获取邮件列表 |
| `/api/v1/emails/:id` | GET | 获取邮件详情 |
| `/api/v1/emails/:id/classify` | POST | 分类邮件 |
| `/api/v1/accounts` | GET/POST | 账户管理 |
| `/api/v1/sync` | POST | 触发同步 |

### 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误码定义

| 错误码 | 含义 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 3. 数据库设计规范

### 表命名
- 使用小写下划线: `email_accounts`
- 时间戳字段: `created_at`, `updated_at`

### 主键
- 使用BIGINT自增主键
- 外键命名: `user_id`, `account_id`

### 索引
- 常用查询字段建立索引
- 复合索引: `idx_user_category (user_id, category)`
- 软删除字段索引: `idx_deleted_at (deleted_at)`

### 数据库视图
```sql
-- 邮件统计视图
CREATE OR REPLACE VIEW v_email_stats AS
SELECT
    user_id,
    category,
    COUNT(*) as total_count,
    SUM(CASE WHEN status = 'unread' THEN 1 ELSE 0 END) as unread_count,
    SUM(CASE WHEN DATE(received_at) = CURDATE() THEN 1 ELSE 0 END) as today_count
FROM emails
WHERE deleted_at IS NULL
GROUP BY user_id, category;

-- 账户同步状态视图
CREATE OR REPLACE VIEW v_account_sync_status AS
SELECT
    a.id as account_id,
    a.account_email,
    a.last_sync_at,
    COUNT(e.id) as total_emails
FROM email_accounts a
LEFT JOIN emails e ON a.id = e.account_id AND e.deleted_at IS NULL
WHERE a.deleted_at IS NULL
GROUP BY a.id, a.account_email, a.last_sync_at;
```

```sql
CREATE TABLE email_accounts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    account_email VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_user_email (user_id, account_email)
);
```

## 4. 配置管理规范

### 配置文件格式
- 开发环境: `config.yaml`
- 生产环境: 使用环境变量覆盖

### 环境变量命名
```bash
# 前缀+下划线+名称
DB_PASSWORD=xxx
CREDENTIAL_KEY=xxx
AGENT_API_KEY=xxx
```

### 敏感信息处理
- 敏感配置不硬编码
- 使用环境变量或密钥管理服务
- 凭证加密存储

## 5. 日志规范

### 日志格式
```
时间 | 级别 | 消息 | 上下文
2026-04-08 10:00:00 | INFO | 用户登录成功 | user_id=1
```

### 日志级别
- DEBUG: 开发调试
- INFO: 正常流程
- WARN: 警告信息
- ERROR: 错误信息

### 敏感信息
- 禁止在日志中记录密码、Token等
- 脱敏处理用户敏感信息

## 6. 安全规范

### 输入验证
- 所有用户输入必须验证
- SQL参数化查询
- XSS防护

### 凭证安全
- 邮箱授权码必须加密存储
- 使用AES-256-GCM加密
- 密钥从环境变量读取
- 每次加密生成随机IV，确保相同明文产生不同密文
- 加密密文和IV分别存储（密文TEXT + IV VARCHAR(64)）
- 禁止在日志、API响应中返回凭证明文

### API安全
- 关键API添加认证
- 限流保护
- CORS配置

### 敏感数据传输
- 前端只传输授权码明文（HTTPS保护）
- 后端接收后立即加密存储
- API响应中不返回加密凭证相关字段

## 7. 测试规范

### 自测脚本
```powershell
# 运行自测
.\test-dev.ps1
```

### 单元测试
- 核心业务逻辑必须测试
- 覆盖率 > 60%

### 集成测试
- API接口测试
- 数据库操作测试

## 8. Git提交规范

### 提交格式
```
<type>: <subject>

<body>

<footer>
```

### Type类型
- feat: 新功能
- fix: 修复bug
- docs: 文档变更
- style: 代码格式
- refactor: 重构
- test: 测试相关
- chore: 构建/工具

### 示例
```
feat: 添加126邮箱IMAP支持

- 实现Net126Provider
- 添加凭证加密模块
- 支持增量同步

Closes #1
```

## 9. 部署规范

### Docker环境
- 基础镜像选择
- 多阶段构建优化
- 非root用户运行

### 环境隔离
- 开发环境
- 测试环境
- 生产环境

## 10. 性能规范

### API响应时间
- 简单查询: < 100ms
- 复杂查询: < 500ms
- 文件操作: < 2s

### 并发处理
- 连接池复用
- 异步处理非核心逻辑
- 缓存热点数据

---

> 生成时间: 2026-04-08
> 更新: 2026-04-14 (补充安全凭证规范和数据库视图模式)
> 适用于: 全栈开发