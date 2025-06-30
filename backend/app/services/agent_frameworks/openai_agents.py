from agents import Agent, Runner, function_tool
from agents.mcp import MCPServerStreamableHttp
from agents.extensions.models.litellm_model import LitellmModel
from typing import Tuple, Optional, Dict
import os
import logging
from .base import AgentFramework, ModelInfo

logger = logging.getLogger(__name__)


class OpenAIAgentsFramework(AgentFramework):
    """Implementation using OpenAI Agents SDK"""
    
    @property
    def name(self) -> str:
        return "openai_agents"
    
    @property
    def description(self) -> str:
        return "OpenAI Agents SDK with routing and planning agents"
    
    def is_available(self) -> bool:
        try:
            import agents
            return True
        except ImportError:
            return False
    
    def _get_model_for_agent(self, model: ModelInfo):
        """Get the appropriate model based on the model info"""
        if model["provider"] == "OPENAI":
            # For OpenAI, use the model name directly
            return model["name"]
        else:
            # For now, fallback to OpenAI models due to LiteLLM compatibility issues
            # TODO: Fix LiteLLM integration with agents library
            logger.warning(f"LiteLLM model {model['name']} not fully supported, falling back to gpt-4o-mini")
            return "gpt-4o-mini"

    @function_tool
    def get_weather(self, location: str) -> str:
        """Get weather information for a location."""
        return f"The weather in {location} is sunny with a high of 75°F."

    def _extract_metadata(self, result, prompt: str, response: str, routing_model: ModelInfo, planning_model: ModelInfo) -> Dict:
        """Extract metadata from agent result"""
        metadata = {
            "routing_model": routing_model["name"],
            "planning_model": planning_model["name"],
            "model": f"{routing_model['name']}→{planning_model['name']}",
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

    async def _run_without_github(self, prompt: str, routing_model: ModelInfo, planning_model: ModelInfo, repository_name: Optional[str] = None) -> Tuple[str, Optional[Dict]]:
        """Run agents without GitHub MCP server (fallback)"""
        logger.info("Running without GitHub MCP server")
        
        # Simple planning agent for non-GitHub operations
        planning_agent = Agent(
            name="Planning Agent",
            instructions="You are a helpful assistant. Respond to any requests directly with detailed and thoughtful answers.",
            model=self._get_model_for_agent(planning_model)
        )
        
        routing_agent = Agent(
            name="Routing Agent",
            instructions=(
                "You are a routing assistant. For complex questions or requests, use the planning agent. "
                "For simple questions, respond directly."
            ),
            model=self._get_model_for_agent(routing_model),
            handoffs=[planning_agent]
        )
        
        result = await Runner.run(routing_agent, prompt)
        
        if result:
            response = result.final_output or "No response from agent"
            metadata = self._extract_metadata(result, prompt, response, routing_model, planning_model)
            return response, metadata
        
        return "No response from agent", None

    async def process_request(
        self, 
        prompt: str, 
        routing_model: ModelInfo, 
        planning_model: ModelInfo, 
        repository_name: Optional[str] = None,
        selected_issue: Optional[Dict] = None
    ) -> Tuple[str, Optional[Dict]]:
        """Process an agent request using OpenAI Agents SDK"""
        
        logger.info(f"Using OpenAI Agents framework: routing={routing_model['name']}, planning={planning_model['name']}")
        
        # Check for GitHub token
        github_token = os.getenv("GH_TOKEN")
        
        if not github_token:
            logger.warning("GH_TOKEN not found, GitHub features will be unavailable")
            return await self._run_without_github(prompt, routing_model, planning_model, repository_name)
        
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
                
                # Create GitHub agent
                github_instructions = """You are an efficient non-chatty assistant that can use the GitHub MCP server to manage repositories, issues, and pull requests. Do your best to implement the user's request immediately without asking for clarification.

When creating GitHub issues:
- If the user gives a brief request like "Add a github issue, add authentication", create the issue with the title "Add Authentication" and no description.
- Only add a description if the user explicitly asks for one like "Add a github issue, ad authentication, with the description decide between oauth or jwt". In this case, create the issue with the title "Add Authentication" and the description "Decide between oauth or jwt". """
                
                if repository_name:
                    github_instructions += f"\n\nThe current repository context is '{repository_name}'. When the user refers to 'this repo', 'add an issue', 'create a PR', or similar repository operations without specifying a repository, use '{repository_name}' as the repository."
                
                if selected_issue:
                    github_instructions += f"\n\nThe currently selected issue is #{selected_issue.get('number')}: '{selected_issue.get('title')}'. When the user refers to 'this issue', 'the issue', or similar, they are referring to this specific issue."
                
                github_agent = Agent(
                    name="GitHub Agent",
                    instructions=github_instructions,
                    model=self._get_model_for_agent(planning_model),
                    mcp_servers=[github_mcp_server]
                )

                routing_instructions = (
                    "Decide which agent to use based on the prompt. "
                    "Handoff all requests to the GitHub agent. "
                    "For other requests, apologize and say you are not able to help with that."
                )

                routing_agent = Agent(
                    name="Routing Agent", 
                    instructions=routing_instructions,
                    model=self._get_model_for_agent(routing_model),
                    handoffs=[github_agent]
                )

                # Run the agent inside the context manager
                result = await Runner.run(routing_agent, prompt)
                
                if result:
                    response = result.final_output or "No response from agent"
                    metadata = self._extract_metadata(result, prompt, response, routing_model, planning_model)
                    return response, metadata
                
                return "No response from agent", None
                
        except Exception as e:
            logger.error(f"MCP Server error: {e}")
            logger.info("Falling back to non-GitHub agents")
            return await self._run_without_github(prompt, routing_model, planning_model, repository_name)