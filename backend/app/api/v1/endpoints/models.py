from fastapi import APIRouter, HTTPException, Depends, Query
from app.models.model import ModelCreate, ModelUpdate, Model, ModelList
from app.services.model_service import ModelService
from app.core.database import get_db
from sqlalchemy.ext.asyncio import AsyncSession
from typing import List


router = APIRouter()


@router.post("/", response_model=Model)
async def create_model(
    model: ModelCreate,
    db: AsyncSession = Depends(get_db)
):
    """Create a new AI model"""
    # Check if model ID already exists
    existing = await ModelService.get_model(db, model.id)
    if existing:
        raise HTTPException(status_code=400, detail="Model with this ID already exists")
    
    return await ModelService.create_model(db, model)


@router.get("/", response_model=ModelList)
async def list_models(
    active_only: bool = Query(False, description="Only return active models"),
    db: AsyncSession = Depends(get_db)
):
    """List all AI models"""
    models = await ModelService.list_models(db, active_only)
    return ModelList(models=models, total=len(models))


@router.get("/{model_id}", response_model=Model)
async def get_model(
    model_id: str,
    db: AsyncSession = Depends(get_db)
):
    """Get a specific AI model"""
    model = await ModelService.get_model(db, model_id)
    if not model:
        raise HTTPException(status_code=404, detail="Model not found")
    return model


@router.put("/{model_id}", response_model=Model)
async def update_model(
    model_id: str,
    model_update: ModelUpdate,
    db: AsyncSession = Depends(get_db)
):
    """Update an AI model"""
    model = await ModelService.update_model(db, model_id, model_update)
    if not model:
        raise HTTPException(status_code=404, detail="Model not found")
    return model


@router.delete("/{model_id}")
async def delete_model(
    model_id: str,
    db: AsyncSession = Depends(get_db)
):
    """Delete (deactivate) an AI model"""
    success = await ModelService.delete_model(db, model_id)
    if not success:
        raise HTTPException(status_code=404, detail="Model not found")
    return {"message": "Model deactivated successfully"}


