# FastAPI Agent Example

A simple FastAPI application with an `/agent/run` POST endpoint that mimics the behavior of your Fastify endpoint.

## Features

- **POST /agent/run** - Main agent endpoint for processing prompts
- **GET /** - Root endpoint with API information
- **GET /health** - Health check endpoint
- **GET /agent/status** - Agent status and capabilities
- **GET /docs** - Auto-generated API documentation (Swagger UI)
- **GET /redoc** - Alternative API documentation

## Quick Start

1. **Install dependencies using uv:**
   ```bash
   uv sync
   ```

2. **Run the server:**
   ```bash
   uv run python main.py
   ```
   
   Or using the configured script:
   ```bash
   uv run start
   ```
   
   Or directly with uvicorn:
   ```bash
   uv run uvicorn main:app --reload --host 0.0.0.0 --port 8080
   ```

### Alternative: Using pip (if you prefer)
1. **Install dependencies:**
   ```bash
   pip install -r requirements.txt
   ```

2. **Run the server:**
   ```bash
   python main.py
   ```

3. **Access the API:**
   - API: http://localhost:8080
   - Interactive docs: http://localhost:8080/docs
   - Alternative docs: http://localhost:8080/redoc

## API Usage

### POST /agent/run

Send a prompt to the agent for processing.

**Request:**
```json
{
  "prompt": "Hello, how are you?",
  "max_tokens": 1000,
  "temperature": 0.7,
  "stream": false
}
```

**Response:**
```json
{
  "id": "uuid-string",
  "response": "Agent response text",
  "prompt": "Original prompt",
  "tokens_used": 25,
  "processing_time": 0.523,
  "timestamp": "2024-01-01T12:00:00.000Z",
  "success": true
}
```

### Example cURL Commands

```bash
# Basic request
curl -X POST "http://localhost:8080/agent/run" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello, how are you?"}'

# Request with parameters
curl -X POST "http://localhost:8080/agent/run" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Generate a Python function",
    "max_tokens": 500,
    "temperature": 0.5
  }'

# Health check
curl http://localhost:8080/health

# Agent status
curl http://localhost:8080/agent/status
```

## Development

### Using uv for development:

```bash
# Install with dev dependencies
uv sync --dev

# Run tests (when added)
uv run pytest

# Format code
uv run black .

# Sort imports
uv run isort .

# Lint code
uv run flake8 .
```

## Mock Behavior

The current implementation includes mock responses based on prompt content:

- Prompts containing "error" → Simulates processing error
- Prompts containing "hello" → Friendly greeting response
- Prompts containing "code" → Code generation example
- Other prompts → Generic acknowledgment response

## Customization

To integrate with a real AI model or agent:

1. Replace the `mock_agent_processing()` function in `main.py`
2. Install additional dependencies (OpenAI, Anthropic, etc.)
3. Add your API keys and configuration
4. Update the response logic as needed

## CORS

The application includes CORS middleware configured to allow all origins for development. In production, update the `allow_origins` list to include only your specific domains.

## Error Handling

The API includes comprehensive error handling:
- Input validation (empty prompts, length limits)
- Processing errors with detailed messages
- HTTP status codes and structured error responses