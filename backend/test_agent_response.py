import asyncio
import json
from app.main import app
from fastapi.testclient import TestClient

def test_agent_run_response():
    """Test that the agent/run endpoint returns complete Chat object"""
    client = TestClient(app)
    
    # Test request payload
    test_request = {
        "prompt": "Hello, this is a test message",
        "routing_model": "gpt-4.1-nano",
        "planning_model": "gpt-4.1-mini",
        "conversation_id": None  # Test without conversation first
    }
    
    # Make request to agent/run endpoint
    response = client.post("/api/v1/agent/run", json=test_request)
    
    print(f"Status code: {response.status_code}")
    print(f"Response: {json.dumps(response.json(), indent=2)}")
    
    if response.status_code == 200:
        data = response.json()
        
        # Check that all Chat object fields are present
        expected_fields = [
            'id', 'conversation_id', 'timestamp', 'prompt', 'response',
            'model_id', 'input_tokens', 'output_tokens', 'input_cost',
            'output_cost', 'total_cost', 'processing_time', 'success'
        ]
        
        print("\nField validation:")
        for field in expected_fields:
            if field in data:
                print(f"✓ {field}: {data[field]}")
            else:
                print(f"✗ Missing field: {field}")
    else:
        print(f"Request failed: {response.text}")

if __name__ == "__main__":
    test_agent_run_response()