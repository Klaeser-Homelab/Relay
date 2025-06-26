from agents import Agent, Runner, function_tool
from agents.mcp import MCPServerStreamableHttp
from typing import Tuple, Optional, Dict
import os
import logging
from app.services.model_service import ModelService
from app.core.database import async_session_maker

logger = logging.getLogger(__name__)


@function_tool
def get_weather(location: str) -> str:
    return f"The weather in {location} is sunny with a high of 75°F."


def extract_metadata(result, prompt: str, response: str, routing_model: str, planning_model: str) -> Dict:
    """Extract metadata from agent result"""
    metadata = {
        "routing_model": routing_model,
        "planning_model": planning_model,
        "model": f"{routing_model}→{planning_model}",
        "input_tokens": getattr(result, 'input_tokens', 0),
        "output_tokens": getattr(result, 'output_tokens', 0),
        "total_tokens": getattr(result, 'total_tokens', 0)
    }
    
    # If token counts aren't available, estimate them
    if metadata["input_tokens"] == 0:
        metadata["input_tokens"] = len(prompt) // 4
        metadata["output_tokens"] = len(response) // 4
        metadata["total_tokens"] = metadata["input_tokens"] + metadata["output_tokens"]
    
    return metadata


async def run_without_github(prompt: str, routing_model: str, planning_model: str, repository_name: Optional[str] = None) -> Tuple[str, Optional[Dict]]:
    """Run agents without GitHub MCP server (fallback)"""
    logger.info("Running without GitHub MCP server")
    
    weather_agent = Agent(
        name="Weather Agent",
        instructions="You are a cheery weather man. Provide weather information based on the location provided. Reply informally and use emojis in your response.",
        model=planning_model,
        tools=[get_weather]
    )
    
    routing_agent = Agent(
        name="Routing Agent",
        instructions=(
            "Decide which agent to use based on the prompt. "
            "If about weather, use the weather agent. "
            "For other requests, respond directly as a helpful assistant."
        ),
        model=routing_model,
        handoffs=[weather_agent]
    )
    
    result = await Runner.run(routing_agent, prompt)
    
    if result:
        response = result.final_output or "No response from agent"
        metadata = extract_metadata(result, prompt, response, routing_model, planning_model)
        return response, metadata
    
    return "No response from agent", None



async def process_agent_request(prompt: str, routing_model: str = "gpt-4.1-nano", planning_model: str = "gpt-4.1-mini", repository_name: Optional[str] = None) -> Tuple[str, Optional[dict]]:
    """
    Process an agent request with the given prompt using specified models.
    
    Args:
        prompt: The user's input prompt
        routing_model: Model name to use for the routing agent (e.g., "gpt-4.1-nano")
        planning_model: Model name to use for the planning agent (e.g., "gpt-4.1-mini")
        repository_name: Optional repository name for context (e.g., "Klaeser-Homelab/Relay")
        
    Returns:
        Tuple of (response, metadata) where metadata includes token counts
    """
    
    logger.info(f"Using models: routing={routing_model}, planning={planning_model}")
    
    # Check for GitHub token
    github_token = os.getenv("GH_TOKEN")
    
    if not github_token:
        logger.warning("GH_TOKEN not found, GitHub features will be unavailable")
        return await run_without_github(prompt, routing_model, planning_model, repository_name)
    
    try:
        async with MCPServerStreamableHttp(
            name="GitHub MCP Server",
            params={
                "url": "https://api.githubcopilot.com/mcp/",
                "headers": {
                    "Authorization": f"Bearer {github_token}",
                    "Content-Type": "application/json"
                }
            }
        ) as github_mcp_server:
            logger.info("Connected to GitHub MCP server")
            
            # Create all agents inside the context manager
            github_instructions = "You are a helpful assistant that can use the GitHub MCP server to manage repositories, issues, and pull requests."
            if repository_name:
                github_instructions += f" The current repository context is '{repository_name}'. When the user refers to 'this repo', 'add an issue', 'create a PR', or similar repository operations without specifying a repository, use '{repository_name}' as the repository."
            
            github_agent = Agent(
                name="GitHub Agent",
                instructions=github_instructions,
                model=planning_model,
                mcp_servers=[github_mcp_server]  # Use mcp_servers instead of tools
            )

            weather_agent = Agent(
                name="Weather Agent", 
                instructions="You are a cheery weather man. Provide weather information based on the location provided. Reply informally and use emojis in your response.",
                model=planning_model,
                tools=[get_weather]
            )

            routing_instructions = (
                "Decide which agent to use based on the prompt. "
                "If about weather, use the weather agent. "
                "If about planning, GitHub repositories, issues, pull requests, or code management, use the GitHub agent."
            )

            routing_agent = Agent(
                name="Routing Agent", 
                instructions=routing_instructions,
                model=routing_model,
                handoffs=[weather_agent, github_agent]
            )

            # Run the agent inside the context manager
            result = await Runner.run(routing_agent, prompt)
            
            if result:
                response = result.final_output or "No response from agent"
                metadata = extract_metadata(result, prompt, response, routing_model, planning_model)
                return response, metadata
            
            return "No response from agent", None
            
    except Exception as e:
        logger.error(f"MCP Server error: {e}")
        logger.info("Falling back to non-GitHub agents")
        return await run_without_github(prompt, routing_model, planning_model, repository_name)