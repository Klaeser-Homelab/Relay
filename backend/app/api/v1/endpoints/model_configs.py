from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from typing import List, Dict, Any
from app.core.database import get_db
from app.services.model_config_service import ModelConfigService

router = APIRouter()


@router.get("/", response_model=Dict[str, Any])
async def get_model_configs(db: AsyncSession = Depends(get_db)):
    """Get all model configurations mapped by role"""
    try:
        models_by_role = await ModelConfigService.get_models_by_roles(db)
        
        # Convert to API response format
        response_data = {}
        for role, model in models_by_role.items():
            if model:
                response_data[role] = {
                    "id": model.name,
                    "name": model.name,
                    "provider": model.provider,
                    "is_active": model.is_active
                }
            else:
                response_data[role] = None
        
        return {
            "success": True,
            "data": response_data,
            "message": "Model configurations retrieved successfully"
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to retrieve model configurations: {str(e)}")


@router.get("/{role}")
async def get_model_config_by_role(
    role: str,
    db: AsyncSession = Depends(get_db)
):
    """Get model configuration for a specific role"""
    try:
        config = await ModelConfigService.get_model_config_by_role(db, role)
        
        if not config:
            raise HTTPException(status_code=404, detail=f"Model configuration for role '{role}' not found")
        
        return {
            "success": True,
            "data": {
                "role": config.model_role,
                "model": {
                    "name": config.model_ref.name,
                    "provider": config.model_ref.provider,
                    "is_active": config.model_ref.is_active
                } if config.model_ref else None
            },
            "message": f"Model configuration for role '{role}' retrieved successfully"
        }
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to retrieve model configuration: {str(e)}")


@router.put("/{role}")
async def update_model_config(
    role: str,
    model_id: str,
    db: AsyncSession = Depends(get_db)
):
    """Update model configuration for a specific role"""
    try:
        config = await ModelConfigService.update_model_config(db, role, model_id)
        
        return {
            "success": True,
            "data": {
                "role": config.model_role,
                "model": {
                    "name": config.model_ref.name,
                    "provider": config.model_ref.provider,
                    "is_active": config.model_ref.is_active
                } if config.model_ref else None
            },
            "message": f"Model configuration for role '{role}' updated successfully"
        }
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to update model configuration: {str(e)}")