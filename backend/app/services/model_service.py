from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func
from app.core.database import Model as DBModel
from app.models.model import ModelCreate, ModelUpdate
from typing import Optional, List, Tuple
from decimal import Decimal
from datetime import datetime


class ModelService:
    """Service for managing AI model pricing and metadata"""
    
    @staticmethod
    async def create_model(db: AsyncSession, model_data: ModelCreate) -> DBModel:
        """Create a new model"""
        db_model = DBModel(
            id=model_data.id,
            name=model_data.name,
            provider=model_data.provider,
            price_per_input_token=model_data.price_per_input_token,
            price_per_output_token=model_data.price_per_output_token,
            max_tokens=model_data.max_tokens,
            is_active=model_data.is_active,
            created_at=datetime.utcnow(),
            updated_at=datetime.utcnow()
        )
        
        db.add(db_model)
        await db.commit()
        await db.refresh(db_model)
        return db_model
    
    @staticmethod
    async def get_model(db: AsyncSession, model_id: str) -> Optional[DBModel]:
        """Get a model by name"""
        result = await db.execute(
            select(DBModel).where(DBModel.name == model_id)
        )
        return result.scalar_one_or_none()
    
    @staticmethod
    async def list_models(
        db: AsyncSession, 
        active_only: bool = False
    ) -> List[DBModel]:
        """List all models"""
        query = select(DBModel).order_by(DBModel.name)
        
        if active_only:
            query = query.where(DBModel.is_active == True)
        
        result = await db.execute(query)
        return result.scalars().all()
    
    @staticmethod
    async def update_model(
        db: AsyncSession, 
        model_id: str, 
        model_update: ModelUpdate
    ) -> Optional[DBModel]:
        """Update a model"""
        db_model = await ModelService.get_model(db, model_id)
        if not db_model:
            return None
        
        update_data = model_update.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(db_model, field, value)
        
        db_model.updated_at = datetime.utcnow()
        await db.commit()
        await db.refresh(db_model)
        return db_model
    
    @staticmethod
    async def delete_model(db: AsyncSession, model_id: str) -> bool:
        """Delete a model (soft delete by setting is_active=False)"""
        db_model = await ModelService.get_model(db, model_id)
        if not db_model:
            return False
        
        db_model.is_active = False
        db_model.updated_at = datetime.utcnow()
        await db.commit()
        return True
    
    @staticmethod
    async def calculate_cost(
        db: AsyncSession, 
        model_id: str, 
        input_tokens: int, 
        output_tokens: int
    ) -> Decimal:
        """Calculate the cost for a given model and token usage"""
        model = await ModelService.get_model(db, model_id)
        if not model:
            raise ValueError(f"Model {model_id} not found")
        
        input_cost = Decimal(input_tokens) * model.price_per_input_token
        output_cost = Decimal(output_tokens) * model.price_per_output_token
        return input_cost + output_cost
    
