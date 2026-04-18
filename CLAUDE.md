# Email Agent 项目开发指南

> 本文件为 Claude Code 提供项目上下文，帮助 AI 更好地理解项目结构和开发规范

## 项目概述

个人邮件智能分类汇总 Agent 系统，支持网易126邮箱，三端独立架构。

## 项目结构

```
mail-agent/
├── email-backend/     # Go服务端 (API + 邮件同步)
├── email-agent/       # Python Agent (AI分类 + 提取)
├── email-web/         # React前端 (用户界面)
├── docs/              # 项目文档
│   ├── DESIGN.md     # 设计文档
│   └── REQUIREMENTS.md # 需求文档
├── .claude/skills/    # 开发Skills (AI专用规范)
└── sql/               # 数据库脚本
```

## 技术栈

| 端 | 技术 | 端口 |
|---|------|------|
| Go后端 | Go 1.21+ / Gin / GORM | 8080 |
| Python Agent | Python 3.11+ / FastAPI / LangChain | 8001 |
| React前端 | React 18 / Vite / TypeScript / Tailwind | 5173 |
| 数据库 | MySQL 8.0 / Redis / ChromaDB | 3306/6379/8000 |

## 开发规范

开发时请参考 `.claude/skills/` 目录下的规范：

- **SKILL-GO-BACKEND.md** - Go后端开发规范
- **SKILL-PYTHON-AGENT.md** - Python Agent开发规范
- **SKILL-REACT-WEB.md** - React前端开发规范
- **SKILL-COMMON.md** - 通用开发规范（API、数据库、安全）

## API规范

### 统一响应格式
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 服务间通信
- Web → Server: HTTP REST API (localhost:8080/api/v1)
- Server → Agent: HTTP REST API (localhost:8001/api/v1)

## 启动命令

```bash
# Go后端
cd email-backend && go run cmd/server/main.go

# Python Agent
cd email-agent && python app/main.py

# React前端
cd email-web && npm run dev
```

## 当前开发阶段

- [x] Phase 1: 项目基础搭建
- [x] Phase 2: 126邮箱接入 (结构就绪，待实现Provider)
- [x] Phase 3: 服务端API (Clean Architecture结构完成)
- [ ] Phase 4: Agent开发
- [ ] Phase 5: 前端开发

## email-backend 项目结构 (Clean Architecture)

```
email-backend/server/
├── api/v1/          # API处理器
├── config/          # 配置
├── core/            # 核心初始化
├── global/          # 全局对象
├── middleware/      # 中间件
├── model/           # 数据模型
│   ├── request/    # 请求DTO
│   └── response/   # 响应DTO
├── repository/      # 数据访问层
├── router/          # 路由
└── service/         # 业务逻辑层
```

## 注意事项

1. **配置安全**: 敏感信息使用环境变量，不硬编码
2. **凭证加密**: 邮箱授权码使用AES-256加密存储
3. **日志规范**: 使用各端统一的日志格式
4. **错误处理**: 统一错误码和错误信息格式