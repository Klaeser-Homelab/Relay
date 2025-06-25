from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.api.v1.api import api_router
from app.core.database import init_db
from datetime import datetime
from contextlib import asynccontextmanager
import os
from dotenv import load_dotenv

# Load environment variables
load_dotenv()


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    await init_db()
    
    # Seed default models
    # Database seeding is handled by init_db() in database.py
    
    yield
    # Shutdown


app = FastAPI(
    title="Agent API Example",
    description="FastAPI example with /agent/run endpoint",
    version="1.0.0",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, specify actual origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API router
app.include_router(api_router, prefix="/api/v1")


@app.get("/")
async def root():
    """Health check endpoint"""
    return {
        "message": "FastAPI Agent API is running",
        "timestamp": datetime.now().isoformat(),
        "endpoints": {
            "agent_run": "/api/v1/agent/run",
            "agent_status": "/api/v1/agent/status",
            "health": "/api/v1/health",
            "usage_stats": "/api/v1/usage/stats",
            "usage_history": "/api/v1/usage/history",
            "docs": "/docs"
        }
    }