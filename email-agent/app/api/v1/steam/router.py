"""
Steam信息提取API路由
"""
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional

from app.services.steam_extract_service import SteamExtractService

router = APIRouter(prefix="/steam", tags=["Steam提取"])


class SteamExtractRequest(BaseModel):
    """Steam提取请求"""
    email_id: str
    subject: str
    sender_email: str
    content: str
    content_html: Optional[str] = None


@router.post("/extract")
async def extract_steam_info(request: SteamExtractRequest):
    """
    从Steam邮件中提取游戏信息

    Args:
        request: Steam提取请求

    Returns:
        提取的游戏信息列表
    """
    try:
        service = SteamExtractService()
        result = await service.extract(
            email_id=request.email_id,
            subject=request.subject,
            sender_email=request.sender_email,
            content=request.content,
            content_html=request.content_html or "",
        )
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
