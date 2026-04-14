# 今日任务计划

> 日期：2026-04-09
> 预计耗时：2.5-3小时
> 目标：完成邮件同步核心功能

---

## 昨日完成情况

| 任务 | 状态 | 说明 |
|------|------|------|
| Phase 1 项目基础搭建 | ✅ 完成 | 三端项目初始化 |
| email-backend Clean Architecture | ✅ 完成 | Handler/Service/Repository分层 |
| Skills文档整理 | ✅ 完成 | 4个开发规范文档 |
| 代码推送GitHub | ✅ 完成 | Pt0819/email-agent |

---

## 今日任务拆分 (3小时)

### 任务1: 126邮箱Provider实现 (60分钟)

**目标**: 实现IMAP邮件拉取
- [ ] `email-backend/pkg/email/provider.go` - Provider接口定义
- [ ] `email-backend/pkg/email/net126.go` - Net126Provider实现
- [ ] `email-backend/pkg/email/parser.go` - MIME解析工具

**验收标准**: 能连接126邮箱并拉取邮件列表

---

### 任务2: 邮件同步服务 (60分钟)

**目标**: 定时同步邮件到数据库
- [ ] `email-backend/server/sync/sync.go` - 同步服务
- [ ] `email-backend/server/sync/scheduler.go` - 定时任务
- [ ] 更新Repository支持批量创建

**验收标准**: 能将126邮件同步到MySQL

---

### 任务3: API完善与联调 (45分钟)

**目标**: 完善邮件API
- [ ] 邮件列表接口支持分页/筛选
- [ ] 邮件详情接口
- [ ] 触发同步接口

**验收标准**: API返回正确数据

---

### 任务4: 数据库初始化 (15分钟)

**目标**: 创建MySQL表结构
- [ ] `sql/001_init.sql` - 初始化表
- [ ] 更新go.mod添加依赖

**验收标准**: 数据库可正常连接

---

## 启动依赖

```bash
# 需要先启动
docker run -d --name email-mysql -e MYSQL_ROOT_PASSWORD=root -p 3306:3306 mysql:8.0
docker run -d --name email-redis -p 6379:6379 redis:7-alpine
```

---

## 时间分配

| 任务 | 预计时间 | 优先级 |
|------|---------|--------|
| 任务1 | 60分钟 | P0 |
| 任务2 | 60分钟 | P0 |
| 任务3 | 45分钟 | P1 |
| 任务4 | 15分钟 | P1 |

---

*计划生成时间：2026-04-09*
