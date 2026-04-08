# 项目Skills索引

> 本目录包含邮件分类Agent系统的可复用开发Skills

---

## Skills列表

| 文件 | 适用场景 | 说明 |
|------|---------|------|
| [SKILL-GO-BACKEND.md](./SKILL-GO-BACKEND.md) | Go后端开发 | 项目结构、命名规范、配置、错误处理 |
| [SKILL-PYTHON-AGENT.md](./SKILL-PYTHON-AGENT.md) | Python Agent开发 | FastAPI项目、LLM集成、配置管理 |
| [SKILL-REACT-WEB.md](./SKILL-REACT-WEB.md) | React前端开发 | 组件开发、API调用、Tailwind配置 |
| [SKILL-COMMON.md](./SKILL-COMMON.md) | 通用开发规范 | API规范、数据库设计、安全、Git规范 |

---

## 快速开始

### Go后端
```bash
cd email-backend
go mod tidy
go run cmd/server/main.go
```

### Python Agent
```bash
cd email-agent
pip install -r requirements.txt
python app/main.py
```

### React Web
```bash
cd email-web
npm install
npm run dev
```

---

## 开发流程

1. **开始开发前** - 阅读对应端点的Skill文档
2. **开发过程中** - 遵循命名规范和代码风格
3. **开发完成后** - 运行自测脚本验证
4. **提交代码前** - 确保符合Git提交规范

---

*生成时间: 2026-04-08*