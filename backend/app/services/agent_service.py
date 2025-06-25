from agents import Agent, Runner, function_tool
from agents.mcp.server import MCPServerStdio, MCPServerSse, MCPServerStreamableHttp
from typing import Tuple, Optional
import os
import logging
from app.services.model_service import ModelService
from app.core.database import async_session_maker

logger = logging.getLogger(__name__)


@function_tool
def get_weather(location: str) -> str:
    return f"The weather in {location} is sunny with a high of 75°F."


async def resolve_model_id(model_id: str) -> str:
    """
    Resolve a model ID to its actual name for API calls.
    If the model_id looks like a UUID, fetch from database.
    Otherwise, return as-is.
    """
    # Check if it looks like a UUID (contains dashes and is long)
    if len(model_id) > 20 and '-' in model_id:
        try:
            async with async_session_maker() as db:
                model = await ModelService.get_model(db, model_id)
                if model:
                    return model.id  # Return the actual model name (e.g., "gpt-4.1-mini")
                else:
                    logger.warning(f"Model {model_id} not found in database, using as-is")
                    return model_id
        except Exception as e:
            logger.error(f"Error resolving model ID {model_id}: {e}")
            return model_id
    else:
        return model_id


async def process_agent_request(prompt: str, routing_model: str = "gpt-4.1-nano", planning_model: str = "gpt-4.1-mini") -> Tuple[str, Optional[dict]]:
    """
    Process an agent request with the given prompt using specified models.
    
    Args:
        prompt: The user's input prompt
        routing_model: Model to use for the routing agent (can be UUID or model name)
        planning_model: Model to use for the weather agent (can be UUID or model name)
        
    Returns:
        Tuple of (response, metadata) where metadata includes token counts
    """
    
    # Resolve model IDs to actual model names
    resolved_routing_model = await resolve_model_id(routing_model)
    resolved_planning_model = await resolve_model_id(planning_model)
    
    logger.info(f"Resolved models: {routing_model} -> {resolved_routing_model}, {planning_model} -> {resolved_planning_model}")
    
    # Set up GitHub MCP server (always enabled)
    # github_token = os.getenv("GH_TOKEN")
    
    # github_mcp_server = MCPServerStreamableHttp(
    #     params={
    #         "url": "https://api.githubcopilot.com/mcp/",
    #         "headers": {
    #             "Authorization": f"Bearer {github_token}",
    #             "Content-Type": "application/json"
    #         }
    #     }
    # )

    # tools = await github_mcp_server.list_tools()
    # print(f"DEBUG: Available tools from GitHub MCP server: {tools}")

    weather_agent = Agent(
        name="Weather Agent", 
        instructions="You are a cheery weather man. Provide weather information based on the location provided. Reply informally and use emojis in your response.",
        model=resolved_planning_model,
        tools=[get_weather]
    )
    
    # Create planning agent with MCP servers
    # planning_agent = Agent(
    #     name="Planning Agent",
    #     instructions="You are a helpful planning assistant. Provide detailed plans based on the user's request. You have access to GitHub tools for repository management, issues, and pull requests.",
    #     model=planning_model,  # Use same model as weather agent
    #     mcp_servers=[github_mcp_server]
    # )

    # Set up routing instructions based on available agents
    routing_instructions = (
        "Decide which agent to use based on the prompt. "
        "If about weather, use the weather agent. "
        "If about planning, GitHub repositories, issues, pull requests, or code management, use the planning agent."
    )

    routing_agent = Agent(
        name="Routing Agent", 
        instructions=routing_instructions,
        model=resolved_routing_model,
        handoffs=[weather_agent]
    )

    result = await Runner.run(routing_agent, prompt)
    
    if result:
        response = result.final_output or "No response from agent"
        
        # Extract token usage from result
        metadata = {
            "routing_model": resolved_routing_model,
            "planning_model": resolved_planning_model,
            "model": f"{resolved_routing_model}→{resolved_planning_model}",
            "input_tokens": getattr(result, 'input_tokens', 0),
            "output_tokens": getattr(result, 'output_tokens', 0),
            "total_tokens": getattr(result, 'total_tokens', 0)
        }
        
        # If token counts aren't available in result, estimate them
        if metadata["input_tokens"] == 0:
            # Rough estimation: ~1 token per 4 characters
            metadata["input_tokens"] = len(prompt) // 4
            metadata["output_tokens"] = len(response) // 4
            metadata["total_tokens"] = metadata["input_tokens"] + metadata["output_tokens"]
            
        return response, metadata
    
    return "No response from agent", None