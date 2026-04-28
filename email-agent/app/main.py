"""
FastAPI应用入口
仅负责应用初始化和启动
"""
import sys
from contextlib import asynccontextmanager
from pathlib import Path

# 添加项目根目录到路径，使 from app.* 导入生效
sys.path.insert(0, str(Path(__file__).parent.parent))

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from loguru import logger

from app.core import get_config
from app.api.v1.health import router as health_router
from app.api.v1.classify import router as classify_router
from app.api.v1.extract import router as extract_router
from app.api.v1.summary import router as summary_router
from app.api.v1.steam import router as steam_router
from app.api.v1.preference.router import router as preference_router


@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理"""
    # 启动时执行
    config = get_config()
    logger.info(f"应用启动中... Agent服务端口: {config.server.port}")
    yield
    # 关闭时执行
    logger.info("应用关闭中...")


def create_app() -> FastAPI:
    """创建FastAPI应用实例"""
    config = get_config()

    # 配置日志
    logger.remove()
    logger.add(
        sys.stderr,
        level=config.logging.level,
        format=config.logging.format
    )

    # 创建应用
    app = FastAPI(
        title="Email Agent API",
        description="邮件智能分类Agent服务",
        version="1.0.0",
        docs_url="/docs",
        redoc_url="/redoc",
        lifespan=lifespan,
    )

    # CORS中间件
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # 注册路由
    app.include_router(health_router, prefix="/api/v1")
    app.include_router(classify_router, prefix="/api/v1")
    app.include_router(extract_router, prefix="/api/v1")
    app.include_router(summary_router, prefix="/api/v1")
    app.include_router(steam_router, prefix="/api/v1")
    app.include_router(preference_router, prefix="/api/v1")

    logger.info("FastAPI应用创建完成")
    return app


# 创建应用实例
app = create_app()


if __name__ == "__main__":
    import uvicorn

    config = get_config()
    uvicorn.run(
        app,
        host=config.server.host,
        port=config.server.port,
        log_level="info"
    )