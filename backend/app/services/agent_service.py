from typing import Tuple, Optional, Dict
import logging
from app.services.agent_frameworks.registry import framework_registry
from app.services.agent_frameworks.base import ModelInfo

logger = logging.getLogger(__name__)


async def process_agent_request(
    prompt: str, 
    routing_model: ModelInfo, 
    planning_model: ModelInfo, 
    framework: str = "openai_agents",
    repository_name: Optional[str] = None
) -> Tuple[str, Optional[Dict]]:
    """
    Process an agent request using the specified framework.
    
    Args:
        prompt: The user's input prompt
        routing_model: Model info dict with 'name' and 'provider' for the routing agent
        planning_model: Model info dict with 'name' and 'provider' for the planning agent
        framework: Name of the agent framework to use
        repository_name: Optional repository name for context
        
    Returns:
        Tuple of (response, metadata) where metadata includes token counts
    """
    
    logger.info(f"Using framework: {framework}, routing={routing_model['name']}, planning={planning_model['name']}")
    
    # Get the framework implementation
    framework_impl = framework_registry.get_framework(framework)
    
    # Process the request using the selected framework
    return await framework_impl.process_request(
        prompt=prompt,
        routing_model=routing_model,
        planning_model=planning_model,
        repository_name=repository_name
    )


def list_available_frameworks():
    """List all available agent frameworks"""
    return framework_registry.list_frameworks()


def get_default_framework() -> str:
    """Get the default framework name"""
    return framework_registry.get_default_framework()