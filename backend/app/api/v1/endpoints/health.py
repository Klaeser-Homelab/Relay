from fastapi import APIRouter
from app.models.common import HealthResponse
from datetime import datetime


router = APIRouter()


@router.get("/", response_model=HealthResponse)
async def health():
    """Health check endpoint"""
    return HealthResponse(
        status="healthy",
        timestamp=datetime.now().isoformat()
    )


@router.get("/ready")
async def readiness():
    """Readiness check endpoint"""
    return {
        "status": "ready",
        "timestamp": datetime.now().isoformat()
    }