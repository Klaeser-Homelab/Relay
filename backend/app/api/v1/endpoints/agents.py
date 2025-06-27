from fastapi import APIRouter, HTTPException, Depends
from app.models.agent import AgentRunRequest, AgentRunResponse
from app.models.common import AgentStatusResponse
from app.services.agent_service import process_agent_request
from app.services.conversation_service import ConversationService
from app.core.database import get_db
from sqlalchemy.ext.asyncio import AsyncSession
import uuid
import time
from datetime import datetime


router = APIRouter()


@router.post("/run", response_model=AgentRunResponse)
async def agent_run(request: AgentRunRequest, db: AsyncSession = Depends(get_db)):
    """
    Main agent endpoint that processes prompts and returns responses.
    
    This endpoint accepts a prompt and optional parameters, processes it through
    an agent (AI model, LLM, or custom logic), and returns the response.
    """
    start_time = time.time()
    
    try:
        print(f"DEBUG: Received request - prompt: {request.prompt[:50]}..., model: {request.model}, conversation_id: {request.conversation_id}")
        
        # Validate input
        if not request.prompt or not request.prompt.strip():
            raise HTTPException(
                status_code=400,
                detail="Prompt cannot be empty"
            )
        
        if len(request.prompt) > 10000:
            raise HTTPException(
                status_code=400,
                detail="Prompt too long (max 10000 characters)"
            )
        
        # Handle conversation context
        conversation_id = request.conversation_id
        exchange_id = None
        
        if conversation_id:
            # Verify conversation exists
            conversation = await ConversationService.get_conversation(db, conversation_id, include_messages=False)
            if not conversation:
                raise HTTPException(status_code=404, detail="Conversation not found")
        
        # Process the request through the agent service
        print(f"DEBUG: About to call process_agent_request")
        # Use new model parameters or fall back to legacy model parameter
        from app.core.database import Model
        from sqlalchemy import select
        
        routing_model_name = request.routing_model or request.model or "gpt-4.1-nano"
        planning_model_name = request.planning_model or request.model or "gpt-4.1-mini"
        
        # Fetch model details
        routing_model_query = select(Model).where(Model.name == routing_model_name)
        routing_result = await db.execute(routing_model_query)
        routing_model_obj = routing_result.scalar_one_or_none()
        
        planning_model_query = select(Model).where(Model.name == planning_model_name)
        planning_result = await db.execute(planning_model_query)
        planning_model_obj = planning_result.scalar_one_or_none()
        
        if not routing_model_obj or not planning_model_obj:
            raise HTTPException(status_code=400, detail="Invalid model name provided")
        
        routing_model = {"name": routing_model_obj.name, "provider": routing_model_obj.provider}
        planning_model = {"name": planning_model_obj.name, "provider": planning_model_obj.provider}
        
        agent_response, metadata = await process_agent_request(
            request.prompt, 
            routing_model, 
            planning_model,
            framework=request.agent_framework
        )
        print(f"DEBUG: Agent response received: {agent_response[:50] if agent_response else 'None'}...")
        
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
            output_tokens = len(agent_response.split())
            model = "unknown"
        
        tokens_used = input_tokens + output_tokens
        
        # Create exchange record if part of a conversation
        if conversation_id:
            exchange = await ConversationService.create_chat(
                db=db,
                conversation_id=conversation_id,
                prompt=request.prompt,
                model_id=f"{routing_model['name']}→{planning_model['name']}",
                response=agent_response,
                input_tokens=input_tokens,
                output_tokens=output_tokens,
                processing_time=processing_time,
                success=True
            )
            
            # Return the complete Chat object
            return AgentRunResponse(
                id=exchange.id,
                conversation_id=exchange.conversation_id,
                timestamp=exchange.timestamp.isoformat(),
                prompt=exchange.prompt,
                response=exchange.response,
                model_id=exchange.model_name,
                input_tokens=exchange.input_tokens,
                output_tokens=exchange.output_tokens,
                input_cost=float(exchange.input_cost),
                output_cost=float(exchange.output_cost),
                total_cost=float(exchange.total_cost),
                processing_time=exchange.processing_time,
                success=exchange.success,
                error_message=exchange.error_message
            )
        else:
            # If no conversation, create a temporary response object
            return AgentRunResponse(
                id=str(uuid.uuid4()),
                conversation_id="temp",
                timestamp=datetime.now().isoformat(),
                prompt=request.prompt,
                response=agent_response,
                model_id=f"{routing_model['name']}→{planning_model['name']}",
                input_tokens=input_tokens,
                output_tokens=output_tokens,
                input_cost=0.0,
                output_cost=0.0,
                total_cost=0.0,
                processing_time=round(processing_time, 3),
                success=True,
                error_message=None
            )
        
    except HTTPException:
        # Re-raise HTTP exceptions as-is
        raise
    except Exception as e:
        # Handle unexpected errors
        print(f"ERROR: Exception occurred: {e}")
        print(f"ERROR: Exception type: {type(e)}")
        import traceback
        traceback.print_exc()
        processing_time = time.time() - start_time
        
        # Get model names for error case
        try:
            routing_model = request.routing_model or request.model or "gpt-4.1-nano"
            planning_model = request.planning_model or request.model or "gpt-4.1-mini"
        except:
            routing_model = "unknown"
            planning_model = "unknown"
        
        # Record failed exchange if part of conversation
        if conversation_id:
            try:
                failed_exchange = await ConversationService.create_chat(
                    db=db,
                    conversation_id=conversation_id,
                    prompt=request.prompt,
                    model_id=f"{routing_model['name']}→{planning_model['name']}",
                    response=None,
                    input_tokens=len(request.prompt.split()) if hasattr(request, 'prompt') else 0,
                    output_tokens=0,
                    processing_time=processing_time,
                    success=False,
                    error_message=str(e)
                )
                
                # Return failed Chat object instead of raising exception
                raise HTTPException(
                    status_code=500,
                    detail=AgentRunResponse(
                        id=failed_exchange.id,
                        conversation_id=failed_exchange.conversation_id,
                        timestamp=failed_exchange.timestamp.isoformat(),
                        prompt=failed_exchange.prompt,
                        response=failed_exchange.response,
                        model_id=failed_exchange.model_name,
                        input_tokens=failed_exchange.input_tokens,
                        output_tokens=failed_exchange.output_tokens,
                        input_cost=float(failed_exchange.input_cost),
                        output_cost=float(failed_exchange.output_cost),
                        total_cost=float(failed_exchange.total_cost),
                        processing_time=failed_exchange.processing_time,
                        success=failed_exchange.success,
                        error_message=failed_exchange.error_message
                    ).dict()
                )
            except:
                pass  # Don't fail the response if exchange recording fails
        
        raise HTTPException(
            status_code=500,
            detail={
                "error": "Agent processing failed",
                "detail": str(e),
                "processing_time": round(processing_time, 3),
                "timestamp": datetime.now().isoformat()
            }
        )


@router.get("/status", response_model=AgentStatusResponse)
async def agent_status():
    """Get agent status and capabilities"""
    return AgentStatusResponse(
        status="active",
        capabilities=[
            "text_processing",
            "code_generation", 
            "question_answering"
        ],
        max_prompt_length=10000,
        default_max_tokens=1000,
        supported_temperatures="0.0 - 2.0"
    )