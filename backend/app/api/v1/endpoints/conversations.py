from fastapi import APIRouter, HTTPException, Depends, Query
from app.models.conversation import (
    ConversationCreate, Conversation, ConversationList, 
    MessageCreate, Message, ConversationStats, ConversationRunRequest
)
from app.services.conversation_service import ConversationService
from app.core.database import get_db
from sqlalchemy.ext.asyncio import AsyncSession
from typing import Optional, List


router = APIRouter()


@router.post("/", response_model=Conversation)
async def create_conversation(
    conversation: ConversationCreate,
    db: AsyncSession = Depends(get_db)
):
    """Create a new conversation"""
    return await ConversationService.create_conversation(db, conversation)


@router.get("/", response_model=ConversationList)
async def list_conversations(
    page: int = Query(1, ge=1, description="Page number"),
    per_page: int = Query(20, ge=1, le=100, description="Items per page"),
    db: AsyncSession = Depends(get_db)
):
    """List all conversations with pagination"""
    conversations, total = await ConversationService.list_conversations(db, page, per_page)
    
    return ConversationList(
        conversations=conversations,
        total=total,
        page=page,
        per_page=per_page
    )


@router.get("/{conversation_id}", response_model=Conversation)
async def get_conversation(
    conversation_id: str,
    include_messages: bool = Query(True, description="Include messages in response"),
    db: AsyncSession = Depends(get_db)
):
    """Get a specific conversation"""
    conversation = await ConversationService.get_conversation(db, conversation_id, include_messages)
    
    if not conversation:
        raise HTTPException(status_code=404, detail="Conversation not found")
    
    return conversation



@router.get("/{conversation_id}/stats", response_model=ConversationStats)
async def get_conversation_stats(
    conversation_id: str,
    db: AsyncSession = Depends(get_db)
):
    """Get usage statistics for a specific conversation"""
    stats = await ConversationService.get_conversation_stats(db, conversation_id)
    
    if not stats:
        raise HTTPException(status_code=404, detail="Conversation not found or has no messages")
    
    return stats


@router.post("/{conversation_id}/run", response_model=Message)
async def run_conversation(
    conversation_id: str,
    request: ConversationRunRequest,
    db: AsyncSession = Depends(get_db)
):
    """Process a prompt through AI agent and create a chat record in the conversation"""
    import time
    from app.services.agent_service import process_agent_request
    
    start_time = time.time()
    
    # Verify conversation exists
    conversation = await ConversationService.get_conversation(db, conversation_id, include_messages=False)
    if not conversation:
        raise HTTPException(status_code=404, detail="Conversation not found")
    
    # Validate input
    if not request.prompt or not request.prompt.strip():
        raise HTTPException(status_code=400, detail="Prompt cannot be empty")
    
    if len(request.prompt) > 10000:
        raise HTTPException(status_code=400, detail="Prompt too long (max 10000 characters)")
    
    # Use provided models or fallback to ModelConfig lookup
    if request.triage_model and request.planning_model:
        # Use models provided in request
        routing_model = request.triage_model
        planning_model = request.planning_model
    else:
        # Fallback to database lookup
        from app.core.database import ModelConfig
        from sqlalchemy import select
        
        # Get routing model (triage role)
        triage_config_query = select(ModelConfig).where(ModelConfig.model_role == "triage")
        triage_result = await db.execute(triage_config_query)
        triage_config = triage_result.scalar_one_or_none()
        routing_model = triage_config.model_name if triage_config else "gpt-4.1-nano"
        
        # Get planning model (planning role)
        planning_config_query = select(ModelConfig).where(ModelConfig.model_role == "planning")
        planning_result = await db.execute(planning_config_query)
        planning_config = planning_result.scalar_one_or_none()
        planning_model = planning_config.model_name if planning_config else "gpt-4.1-mini"
    
    try:
        # Process the request through the agent service
        agent_response, metadata = await process_agent_request(
            request.prompt, 
            routing_model, 
            planning_model,
            repository_name=request.repository_name
        )
        
        # Calculate processing time
        processing_time = time.time() - start_time
        
        # Get token counts from metadata or estimate
        if metadata:
            input_tokens = metadata.get("input_tokens", 0)
            output_tokens = metadata.get("output_tokens", 0)
            model = metadata.get("model", "unknown")
        else:
            # Fallback estimation
            input_tokens = len(request.prompt.split())
            output_tokens = len(agent_response.split()) if agent_response else 0
            model = "unknown"
        
        # Create chat record in the conversation
        chat = await ConversationService.create_chat(
            db=db,
            conversation_id=conversation_id,
            prompt=request.prompt,
            model_name=f"{routing_model}→{planning_model}",
            response=agent_response,
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            processing_time=processing_time,
            success=True
        )
        
        return chat
        
    except Exception as e:
        # Handle processing errors
        processing_time = time.time() - start_time
        
        # Create failed chat record
        failed_chat = await ConversationService.create_chat(
            db=db,
            conversation_id=conversation_id,
            prompt=request.prompt,
            model_name=f"{routing_model}→{planning_model}",
            response=None,
            input_tokens=len(request.prompt.split()),
            output_tokens=0,
            processing_time=processing_time,
            success=False,
            error_message=str(e)
        )
        
        return failed_chat


@router.delete("/{conversation_id}")
async def delete_conversation(
    conversation_id: str,
    db: AsyncSession = Depends(get_db)
):
    """Delete a conversation and all its messages"""
    success = await ConversationService.delete_conversation(db, conversation_id)
    
    if not success:
        raise HTTPException(status_code=404, detail="Conversation not found")
    
    return {"message": "Conversation deleted successfully"}