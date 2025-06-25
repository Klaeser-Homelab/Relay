from pydantic import BaseModel
from typing import Optional
from datetime import datetime
from decimal import Decimal


class ModelBase(BaseModel):
    name: str  # Model name, e.g., "gpt-4.1-mini"
    provider: str  # e.g., "openai", "anthropic"
    price_per_input_token: Decimal  # Price per token (not per million)
    price_per_output_token: Decimal  # Price per token (not per million)
    max_tokens: Optional[int] = None
    is_active: bool = True


class ModelCreate(ModelBase):
    pass


class ModelUpdate(BaseModel):
    name: Optional[str] = None
    provider: Optional[str] = None
    price_per_input_token: Optional[Decimal] = None
    price_per_output_token: Optional[Decimal] = None
    max_tokens: Optional[int] = None
    is_active: Optional[bool] = None


class Model(ModelBase):
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class ModelList(BaseModel):
    models: list[Model]
    total: int