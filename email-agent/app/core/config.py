"""
配置管理模块
"""
import os
from typing import Any, Dict, List, Optional
from pathlib import Path

import yaml
from pydantic import BaseModel, Field
from pydantic_settings import BaseSettings


class ServerConfig(BaseModel):
    """服务配置"""
    host: str = "0.0.0.0"
    port: int = 8001


class RedisConfig(BaseModel):
    """Redis配置"""
    host: str = "localhost"
    port: int = 6379
    db: int = 0
    password: Optional[str] = None


class ChromaConfig(BaseModel):
    """ChromaDB配置"""
    host: str = "localhost"
    port: int = 8000
    collection_name: str = "emails"


class LLMProviderConfig(BaseModel):
    """LLM Provider配置"""
    enabled: bool = True
    api_key: str = ""
    model: str = ""
    temperature: float = 0.3
    max_tokens: int = 4096
    base_url: Optional[str] = None


class LLMConfig(BaseModel):
    """LLM配置"""
    default_provider: str = "deepseek"
    providers: Dict[str, LLMProviderConfig] = Field(default_factory=dict)


class CategoryConfig(BaseModel):
    """分类配置"""
    code: str
    name: str
    priority: str = "medium"


class ClassificationConfig(BaseModel):
    """分类配置"""
    auto_classify: bool = True
    categories: List[CategoryConfig] = Field(default_factory=list)


class ExtractionConfig(BaseModel):
    """提取配置"""
    extract_actions: bool = True
    extract_meetings: bool = True
    extract_deadlines: bool = True
    extract_entities: bool = True


class LoggingConfig(BaseModel):
    """日志配置"""
    level: str = "INFO"
    format: str = "{time:YYYY-MM-DD HH:mm:ss} | {level} | {message}"


class Config(BaseModel):
    """全局配置"""
    server: ServerConfig = Field(default_factory=ServerConfig)
    redis: RedisConfig = Field(default_factory=RedisConfig)
    chroma: ChromaConfig = Field(default_factory=ChromaConfig)
    llm: LLMConfig = Field(default_factory=LLMConfig)
    classification: ClassificationConfig = Field(default_factory=ClassificationConfig)
    extraction: ExtractionConfig = Field(default_factory=ExtractionConfig)
    logging: LoggingConfig = Field(default_factory=LoggingConfig)


def _substitute_env_vars(value: Any) -> Any:
    """递归替换环境变量"""
    if isinstance(value, str):
        if value.startswith("${") and value.endswith("}"):
            env_var = value[2:-1]
            return os.getenv(env_var, "")
        return value
    elif isinstance(value, dict):
        return {k: _substitute_env_vars(v) for k, v in value.items()}
    elif isinstance(value, list):
        return [_substitute_env_vars(item) for item in value]
    return value


def load_config(config_path: Optional[str] = None) -> Config:
    """
    加载配置文件

    Args:
        config_path: 配置文件路径，默认为 ./config/config.yaml

    Returns:
        Config对象
    """
    if config_path is None:
        config_path = os.getenv("CONFIG_PATH", "config/config.yaml")

    config_file = Path(config_path)

    if not config_file.exists():
        print(f"配置文件 {config_path} 不存在，使用默认配置")
        return Config()

    with open(config_file, "r", encoding="utf-8") as f:
        config_dict = yaml.safe_load(f)

    # 替换环境变量
    config_dict = _substitute_env_vars(config_dict)

    return Config(**config_dict)


# 全局配置实例
_config: Optional[Config] = None


def get_config() -> Config:
    """获取全局配置实例"""
    global _config
    if _config is None:
        _config = load_config()
    return _config


def reload_config(config_path: Optional[str] = None) -> Config:
    """重新加载配置"""
    global _config
    _config = load_config(config_path)
    return _config