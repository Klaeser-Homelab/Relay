from fastapi import APIRouter, Depends, Query
from app.models.usage import UsageStats, UsageHistory
from app.services.usage_service import UsageService
from app.core.database import get_db
from sqlalchemy.ext.asyncio import AsyncSession
from datetime import datetime
from typing import Optional


router = APIRouter()


@router.get("/stats", response_model=UsageStats)
async def get_usage_stats(
    start_date: Optional[datetime] = Query(None, description="Start date for filtering stats"),
    end_date: Optional[datetime] = Query(None, description="End date for filtering stats"),
    model: Optional[str] = Query(None, description="Filter by specific model"),
    db: AsyncSession = Depends(get_db)
):
    """
    Get aggregated usage statistics.
    
    Returns total requests, tokens used, costs, and average processing time.
    Optionally filter by date range.
    """
    return await UsageService.get_usage_stats(db, start_date, end_date, model)


@router.get("/history", response_model=UsageHistory)
async def get_usage_history(
    page: int = Query(1, ge=1, description="Page number"),
    per_page: int = Query(50, ge=1, le=100, description="Items per page"),
    start_date: Optional[datetime] = Query(None, description="Start date for filtering history"),
    end_date: Optional[datetime] = Query(None, description="End date for filtering history"),
    db: AsyncSession = Depends(get_db)
):
    """
    Get paginated usage history.
    
    Returns detailed records of all API calls with token counts and costs.
    Optionally filter by date range.
    """
    return await UsageService.get_usage_history(db, page, per_page, start_date, end_date)