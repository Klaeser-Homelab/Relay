from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from app.core.database import ModelConfig, Model
from typing import Optional, List, Dict
import uuid


class ModelConfigService:
    """Service for managing model configurations"""
    
    @staticmethod
    async def get_model_configs(db: AsyncSession) -> List[ModelConfig]:
        """Get all model configurations with their associated models"""
        query = (
            select(ModelConfig)
            .options(selectinload(ModelConfig.model_ref))
            .order_by(ModelConfig.model_role)
        )
        
        result = await db.execute(query)
        return result.scalars().all()
    
    @staticmethod
    async def get_model_config_by_role(
        db: AsyncSession, 
        role: str
    ) -> Optional[ModelConfig]:
        """Get model configuration by role (e.g., 'triage', 'planning')"""
        query = (
            select(ModelConfig)
            .options(selectinload(ModelConfig.model_ref))
            .where(ModelConfig.model_role == role)
        )
        
        result = await db.execute(query)
        return result.scalar_one_or_none()
    
    @staticmethod
    async def get_models_by_roles(
        db: AsyncSession
    ) -> Dict[str, Optional[Model]]:
        """Get models mapped by their roles"""
        configs = await ModelConfigService.get_model_configs(db)
        
        models_by_role = {}
        for config in configs:
            models_by_role[config.model_role] = config.model_ref
        
        return models_by_role
    
    @staticmethod
    async def update_model_config(
        db: AsyncSession,
        role: str,
        new_model_id: str
    ) -> Optional[ModelConfig]:
        """Update model configuration for a specific role"""
        # First check if the new model exists
        model_query = select(Model).where(Model.name == new_model_id)
        model_result = await db.execute(model_query)
        model = model_result.scalar_one_or_none()
        
        if not model:
            raise ValueError(f"Model with name {new_model_id} not found")
        
        # Get existing config for this role
        config = await ModelConfigService.get_model_config_by_role(db, role)
        
        if config:
            # Update existing config
            config.model_name = new_model_id
        else:
            # Create new config if it doesn't exist
            config = ModelConfig(
                id=str(uuid.uuid4()),
                model_role=role,
                model_name=new_model_id
            )
            db.add(config)
        
        await db.commit()
        await db.refresh(config)
        
        # Return config with model relationship loaded
        return await ModelConfigService.get_model_config_by_role(db, role)