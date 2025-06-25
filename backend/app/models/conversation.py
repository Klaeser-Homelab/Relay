from pydantic import BaseModel
from typing import Optional, List
from datetime import datetime
from decimal import Decimal


class MessageBase(BaseModel):
    prompt: str
    response: Optional[str] = None
    model_name: str


class MessageCreate(MessageBase):
    conversation_id: str


class Message(MessageBase):
    id: str
    conversation_id: str
    timestamp: datetime
    
    # Usage tracking data
    input_tokens: int = 0
    output_tokens: int = 0
    input_cost: Decimal = Decimal("0")
    output_cost: Decimal = Decimal("0")
    total_cost: Decimal = Decimal("0")
    processing_time: float = 0
    success: bool = False
    error_message: Optional[str] = None
    
    class Config:
        from_attributes = True


class ConversationBase(BaseModel):
    title: str
    project_name: Optional[str] = None


class ConversationCreate(ConversationBase):
    pass


class ConversationRunRequest(BaseModel):
    prompt: str


class ConversationStats(BaseModel):
    total_chats: int
    successful_chats: int
    failed_chats: int
    total_tokens: int
    input_tokens: int
    output_tokens: int
    total_cost: float
    avg_processing_time: float
    models_used: dict


class Conversation(ConversationBase):
    id: str
    created_at: datetime
    updated_at: datetime
    messages: Optional[List[Message]] = []
    stats: Optional[ConversationStats] = None
    
    class Config:
        from_attributes = True


class ConversationList(BaseModel):
    conversations: List[Conversation]
    total: int
    page: int
    per_page: int