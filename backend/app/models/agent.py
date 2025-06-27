from pydantic import BaseModel
from typing import Optional


class AgentRunRequest(BaseModel):
    prompt: str
    routing_model: Optional[str] = "gpt-4.1-nano"
    planning_model: Optional[str] = "gpt-4.1-mini"
    agent_framework: Optional[str] = "openai_agents"
    model: Optional[str] = None  # Keep for backward compatibility
    max_tokens: Optional[int] = 1000
    temperature: Optional[float] = 0.7
    stream: Optional[bool] = False
    conversation_id: Optional[str] = None


class AgentRunResponse(BaseModel):
    id: str
    conversation_id: str
    timestamp: str
    prompt: str
    response: Optional[str] = None
    model_id: str
    input_tokens: int
    output_tokens: int
    input_cost: float
    output_cost: float
    total_cost: float
    processing_time: float
    success: bool
    error_message: Optional[str] = None