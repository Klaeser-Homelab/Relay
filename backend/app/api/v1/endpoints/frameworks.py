from fastapi import APIRouter
from app.services.agent_service import list_available_frameworks, get_default_framework
from typing import List, Dict

router = APIRouter()


@router.get("/", response_model=List[Dict[str, str]])
async def list_frameworks():
    """List all available agent frameworks"""
    return list_available_frameworks()


@router.get("/default")
async def get_default():
    """Get the default framework name"""
    return {"framework": get_default_framework()}