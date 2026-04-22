---
name: 日志机制现状
description: 项目三端日志机制检查结果及待补全项
type: project
---

## 日志机制现状（2026-04-22）

### 当前状态

| 端 | 日志库 | 状态 | 说明 |
|---|---|---|---|
| **Go后端** | `log` (标准库) | ⚠️ 部分完善 | 仅main.go和scheduler.go有日志，service/api/repository层缺失 |
| **Python Agent** | `loguru` | ✅ 较好 | 配置完整，支持文件输出，base_service提供封装 |
| **React前端** | 无 | ❌ 缺失 | 无任何日志机制 |

### 待补全项（P2优先级）

**Go后端：**
- [ ] 引入结构化日志库（推荐slog或zap）
- [ ] Service层添加业务日志（关键操作、耗时统计）
- [ ] API层添加请求日志（请求ID、参数、响应时间）
- [ ] Repository层添加数据访问日志（慢查询警告）
- [ ] 中间件完善请求追踪

**Python Agent：**
- [ ] classify_service添加分类日志
- [ ] steam_extract_service添加提取日志
- [ ] 统一日志格式和级别

**React前端：**
- [ ] API调用错误日志
- [ ] 前端错误边界
- [ ] 用户行为埋点（可选）

### 实施建议

**Why:** 生产环境问题排查需要完整日志链路

**How to apply:** 在功能开发间隙补全，不阻塞Phase 6核心功能开发
