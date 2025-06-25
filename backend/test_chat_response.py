import asyncio
from app.models.agent import AgentRunResponse
from app.core.database import Chat, async_session_maker
from app.services.conversation_service import ConversationService
from datetime import datetime

async def test_chat_response_structure():
    """Test that AgentRunResponse matches Chat object structure"""
    
    # Create a mock Chat object (similar to what would be in database)
    mock_chat = type('MockChat', (), {
        'id': 'chat_1234567890_abc123',
        'conversation_id': 'conv_1234567890_def456',
        'timestamp': datetime.now(),
        'prompt': 'Test prompt',
        'response': 'Test response',
        'model_id': 'gpt-4.1-nano→gpt-4.1-mini',
        'input_tokens': 10,
        'output_tokens': 15,
        'input_cost': 0.000001,
        'output_cost': 0.000002,
        'total_cost': 0.000003,
        'processing_time': 1.5,
        'success': True,
        'error_message': None
    })()
    
    # Create AgentRunResponse from Chat object
    response = AgentRunResponse(
        id=mock_chat.id,
        conversation_id=mock_chat.conversation_id,
        timestamp=mock_chat.timestamp.isoformat(),
        prompt=mock_chat.prompt,
        response=mock_chat.response,
        model_id=mock_chat.model_name,
        input_tokens=mock_chat.input_tokens,
        output_tokens=mock_chat.output_tokens,
        input_cost=float(mock_chat.input_cost),
        output_cost=float(mock_chat.output_cost),
        total_cost=float(mock_chat.total_cost),
        processing_time=mock_chat.processing_time,
        success=mock_chat.success,
        error_message=mock_chat.error_message
    )
    
    print("AgentRunResponse structure:")
    print(f"  id: {response.id}")
    print(f"  conversation_id: {response.conversation_id}")
    print(f"  timestamp: {response.timestamp}")
    print(f"  prompt: {response.prompt}")
    print(f"  response: {response.response}")
    print(f"  model_id: {response.model_name}")
    print(f"  input_tokens: {response.input_tokens}")
    print(f"  output_tokens: {response.output_tokens}")
    print(f"  input_cost: {response.input_cost}")
    print(f"  output_cost: {response.output_cost}")
    print(f"  total_cost: {response.total_cost}")
    print(f"  processing_time: {response.processing_time}")
    print(f"  success: {response.success}")
    print(f"  error_message: {response.error_message}")
    
    # Convert to dict (like JSON response)
    response_dict = response.dict()
    print(f"\nAs JSON structure:")
    import json
    print(json.dumps(response_dict, indent=2))
    
    print("\n✅ AgentRunResponse successfully created with all Chat object fields!")

if __name__ == "__main__":
    asyncio.run(test_chat_response_structure())