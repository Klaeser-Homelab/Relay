from fastapi import APIRouter
from app.api.v1.endpoints import agents, health, usage, conversations, models, mcp, model_configs, repositories


api_router = APIRouter()

# Include routers
api_router.include_router(agents.router, prefix="/agent", tags=["agents"])
api_router.include_router(health.router, prefix="/health", tags=["health"])
api_router.include_router(usage.router, prefix="/usage", tags=["usage"])
api_router.include_router(conversations.router, prefix="/conversations", tags=["conversations"])
api_router.include_router(models.router, prefix="/models", tags=["models"])
api_router.include_router(model_configs.router, prefix="/model-configs", tags=["model-configs"])
api_router.include_router(mcp.router, prefix="/mcp", tags=["mcp"])
api_router.include_router(repositories.router, prefix="/repositories", tags=["repositories"])