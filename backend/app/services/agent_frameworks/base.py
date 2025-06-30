from abc import ABC, abstractmethod
from typing import Tuple, Optional, Dict, TypedDict


class ModelInfo(TypedDict):
    name: str
    provider: str


class AgentFramework(ABC):
    """Abstract base class for agent frameworks"""
    
    @property
    @abstractmethod
    def name(self) -> str:
        """Return the name of this framework"""
        pass
    
    @property
    @abstractmethod
    def description(self) -> str:
        """Return a description of this framework"""
        pass
    
    @abstractmethod
    async def process_request(
        self, 
        prompt: str, 
        routing_model: ModelInfo, 
        planning_model: ModelInfo, 
        repository_name: Optional[str] = None,
        selected_issue: Optional[Dict] = None
    ) -> Tuple[str, Optional[Dict]]:
        """
        Process an agent request using this framework.
        
        Args:
            prompt: The user's input prompt
            routing_model: Model info for the routing agent
            planning_model: Model info for the planning agent
            repository_name: Optional repository name for context
            selected_issue: Optional selected issue dict for context
            
        Returns:
            Tuple of (response, metadata) where metadata includes token counts
        """
        pass
    
    @abstractmethod
    def is_available(self) -> bool:
        """Check if this framework is available (dependencies installed, etc.)"""
        pass