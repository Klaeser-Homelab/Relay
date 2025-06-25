from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func, case
from app.core.database import Chat, Model as DBModel
from app.models.usage import UsageRecord, UsageStats, UsageHistory
from app.services.model_service import ModelService
from datetime import datetime
from typing import Optional
from decimal import Decimal


class UsageService:
    """Service for retrieving API usage statistics from messages"""
    
    @staticmethod
    async def get_usage_stats(
        db: AsyncSession,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        model: Optional[str] = None
    ) -> UsageStats:
        """Get usage statistics for a given period from exchanges"""
        
        query = select(
            func.count(Chat.id).label("total_requests"),
            func.sum(case((Chat.success == True, 1), else_=0)).label("successful_requests"),
            func.coalesce(func.sum(Chat.input_tokens), 0).label("total_input_tokens"),
            func.coalesce(func.sum(Chat.output_tokens), 0).label("total_output_tokens"),
            func.coalesce(func.sum(Chat.total_cost), 0).label("total_cost"),
            func.coalesce(func.avg(Chat.processing_time), 0).label("avg_processing_time")
        )
        
        # Apply filters if provided
        if start_date:
            query = query.where(Chat.timestamp >= start_date)
        if end_date:
            query = query.where(Chat.timestamp <= end_date)
        if model:
            query = query.where(Chat.model_name == model)
        
        result = await db.execute(query)
        row = result.one()
        
        return UsageStats(
            total_requests=row.total_requests or 0,
            successful_requests=row.successful_requests or 0,
            failed_requests=(row.total_requests or 0) - (row.successful_requests or 0),
            total_input_tokens=row.total_input_tokens or 0,
            total_output_tokens=row.total_output_tokens or 0,
            total_cost=Decimal(str(row.total_cost or 0)),
            average_processing_time=float(row.avg_processing_time or 0),
            period_start=start_date,
            period_end=end_date
        )
    
    @staticmethod
    async def get_usage_history(
        db: AsyncSession,
        page: int = 1,
        per_page: int = 50,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None
    ) -> UsageHistory:
        """Get paginated usage history from exchanges"""
        
        query = select(Chat)
        
        # Apply date filters
        if start_date:
            query = query.where(Chat.timestamp >= start_date)
        if end_date:
            query = query.where(Chat.timestamp <= end_date)
        
        # Get total count
        count_query = select(func.count()).select_from(query.subquery())
        total_result = await db.execute(count_query)
        total = total_result.scalar() or 0
        
        # Apply pagination
        query = query.order_by(Chat.timestamp.desc())
        query = query.offset((page - 1) * per_page).limit(per_page)
        
        # Execute query
        result = await db.execute(query)
        db_messages = result.scalars().all()
        
        # Convert to UsageRecord format for compatibility
        records = []
        for message in db_messages:
            # Get model info for price display
            model = await ModelService.get_model(db, message.model_name) if message.model_name else None
            
            record = UsageRecord(
                id=message.id,
                timestamp=message.timestamp,
                prompt=message.prompt,
                response=message.response or "",
                input_tokens=message.input_tokens,
                output_tokens=message.output_tokens,
                model=message.model_name,
                price_per_input_token=model.price_per_input_token if model else Decimal("0"),
                price_per_output_token=model.price_per_output_token if model else Decimal("0"),
                total_input_cost=message.input_cost,
                total_output_cost=message.output_cost,
                total_cost=message.total_cost,
                processing_time=message.processing_time,
                success=message.success
            )
            records.append(record)
        
        return UsageHistory(
            records=records,
            total=total,
            page=page,
            per_page=per_page
        )