from pydantic import BaseModel
from typing import Optional


class ErrorResponse(BaseModel):
    error: str
    detail: Optional[str] = None
    timestamp: str


class HealthResponse(BaseModel):
    status: str
    timestamp: str


class AgentStatusResponse(BaseModel):
    status: str
    capabilities: list[str]
    max_prompt_length: int
    default_max_tokens: int
    supported_temperatures: str