# Skill: Python Agent Project Setup

> 本Skill用于快速初始化Python Agent项目结构，基于FastAPI

## 1. 项目结构

```
email-agent/
├── app/
│   ├── main.py                    # 🚀 应用入口 (仅启动)
│   │
│   ├── api/v1/                    # 🌐 API路由层 (轻量化路由)
│   │   ├── health/                # 💚 健康检查
│   │   │   ├── router.py         # 路由定义
│   │   │   └── schemas.py        # 请求/响应模型
│   │   ├── classify/              # 📊 邮件分类
│   │   │   ├── router.py
│   │   │   └── schemas.py
│   │   ├── extract/               # 🔍 信息提取
│   │   └── summary/               # 📝 摘要生成
│   │
│   ├── services/                  # 🔧 业务逻辑层 (核心业务)
│   │   ├── base_service.py        # 🏗️ 服务基类
│   │   ├── classify_service.py    # 📊 分类服务
│   │   └── extract_service.py     # 🔍 提取服务
│   │
│   ├── repositories/              # 🗄️ 数据访问层 (可选)
│   │
│   ├── schemas/                   # ✅ 数据验证层
│   │   ├── enums.py              # 📝 枚举定义
│   │   ├── request.py             # 📥 请求模型
│   │   └── response.py            # 📤 响应模型
│   │
│   ├── agents/                    # 🤖 Agent实现
│   │   ├── orchestrator.py        # 任务编排器
│   │   └── classification_agent.py
│   │
│   ├── llm/                       # 💡 LLM适配层
│   │   ├── manager.py             # LLM管理器
│   │   ├── deepseek.py
│   │   └── zhipu.py
│   │
│   ├── prompts/                   # 💬 提示词模板
│   │   ├── classification.py
│   │   └── extraction.py
│   │
│   ├── core/                      # ⚙️ 核心功能
│   │   ├── config.py             # 📋 配置管理
│   │   └── dependency.py          # 🔗 依赖注入
│   │
│   └── utils/                     # 🔧 工具函数
│
├── config/
│   └── config.yaml                # 📋 配置文件
│
├── tests/                         # 🧪 测试文件
├── requirements.txt
└── main.py                        # 入口文件
```

## 2. 命名规范

### 文件命名
- 小写字母+下划线: `classify_service.py`
- 测试文件: `test_classify_service.py`

### 目录命名
- 小写字母+下划线: `api/`, `services/`

### 模块/类命名
- 模块: 小写下划线
- 类: 大驼峰: `ClassifyService`
- 函数: 小写下划线: `classify_email()`

### 常量命名
- 全大写下划线: `MAX_RETRIES = 3`

## 3. 开发规范

### 分层职责

| 层级 | 职责 | 示例 |
|------|------|------|
| api/v1/ | 路由定义、参数验证 | router.py |
| services/ | 业务逻辑处理 | ClassifyService |
| repositories/ | 数据库操作 | EmailRepository |
| schemas/ | 数据模型定义 | ClassifyRequest |
| core/ | 配置、依赖注入 | config.py |

### 路由层示例

```python
# app/api/v1/classify/router.py
from fastapi import APIRouter
from app.schemas import ClassifyRequest, ClassifyResponse
from app.services.classify_service import ClassifyService

router = APIRouter(prefix="/classify", tags=["分类"])

@router.post("", response_model=ClassifyResponse)
async def classify_email(request: ClassifyRequest):
    """分类邮件"""
    service = ClassifyService()
    return await service.classify(request)
```

### 服务层示例

```python
# app/services/classify_service.py
from app.services.base_service import BaseService

class ClassifyService(BaseService):
    async def classify(self, request: ClassifyRequest) -> ClassifyResponse:
        """执行分类业务逻辑"""
        self.log_info(f"分类邮件: {request.email_id}")
        # TODO: 调用LLM
        return result
```

### 依赖注入

```python
from app.core.dependency import get_classify_service

@router.post("/classify")
async def classify(
    request: ClassifyRequest,
    service: ClassifyService = Depends(get_classify_service)
):
    return await service.classify(request)
```

## 4. 错误处理

### 服务基类日志

```python
from app.services.base_service import BaseService

class MyService(BaseService):
    async def do_something(self):
        self.log_info("开始处理")
        try:
            # 业务逻辑
            self.log_info("处理成功")
        except Exception as e:
            self.log_error(f"处理失败: {e}")
            raise
```

### API异常处理

```python
from fastapi import HTTPException

@router.post("/classify")
async def classify_email(request: ClassifyRequest):
    try:
        return await service.classify(request)
    except LLMError as e:
        raise HTTPException(status_code=500, detail="LLM服务不可用")
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
```

## 5. 快速开始命令

```bash
# 1. 创建虚拟环境
cd email-agent
python -m venv venv
.\venv\Scripts\Activate.ps1

# 2. 安装依赖
pip install -r requirements.txt

# 3. 运行开发服务器
python app/main.py

# 4. 或使用uvicorn
uvicorn app.main:app --reload --port 8001

# 5. 访问API文档
# http://localhost:8001/docs
```

## 6. 常用依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| fastapi | 0.110.x | Web框架 |
| uvicorn | 0.27.x | ASGI服务器 |
| pydantic | 2.6.x | 数据验证 |
| langchain | 0.1.x | LLM框架 |
| loguru | 0.7.x | 日志库 |
| pyyaml | 6.0.x | YAML配置 |

---

> 更新时间: 2026-04-08
> 适用于: Python Agent开发