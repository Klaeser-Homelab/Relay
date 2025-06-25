from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime
from decimal import Decimal


class UsageRecord(BaseModel):
    """Pydantic model for API usage records"""
    id: Optional[int] = None
    timestamp: datetime
    prompt: str
    response: str
    input_tokens: int
    output_tokens: int
    model: str
    price_per_input_token: Decimal
    price_per_output_token: Decimal
    total_input_cost: Decimal = Field(default=Decimal("0"))
    total_output_cost: Decimal = Field(default=Decimal("0"))
    total_cost: Decimal = Field(default=Decimal("0"))
    processing_time: float
    success: bool

    class Config:
        from_attributes = True

    def calculate_costs(self):
        """Calculate the cost fields based on tokens and prices"""
        self.total_input_cost = Decimal(str(self.input_tokens)) * self.price_per_input_token
        self.total_output_cost = Decimal(str(self.output_tokens)) * self.price_per_output_token
        self.total_cost = self.total_input_cost + self.total_output_cost


class UsageStats(BaseModel):
    """Statistics for API usage"""
    total_requests: int
    successful_requests: int
    failed_requests: int
    total_input_tokens: int
    total_output_tokens: int
    total_cost: Decimal
    average_processing_time: float
    period_start: Optional[datetime] = None
    period_end: Optional[datetime] = None


class UsageHistory(BaseModel):
    """Response model for usage history"""
    records: list[UsageRecord]
    total: int
    page: int
    per_page: int