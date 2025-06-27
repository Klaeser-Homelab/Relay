from typing import Dict, List
import logging
from .base import AgentFramework
from .openai_agents import OpenAIAgentsFramework
from .langgraph import LangGraphFramework

logger = logging.getLogger(__name__)


class FrameworkRegistry:
    """Registry for managing available agent frameworks"""
    
    def __init__(self):
        self._frameworks: Dict[str, AgentFramework] = {}
        self._register_default_frameworks()
    
    def _register_default_frameworks(self):
        """Register the default frameworks"""
        frameworks = [
            OpenAIAgentsFramework(),
            LangGraphFramework(),
        ]
        
        for framework in frameworks:
            if framework.is_available():
                self._frameworks[framework.name] = framework
                logger.info(f"Registered framework: {framework.name}")
            else:
                logger.warning(f"Framework {framework.name} is not available (missing dependencies)")
    
    def register_framework(self, framework: AgentFramework):
        """Register a custom framework"""
        if framework.is_available():
            self._frameworks[framework.name] = framework
            logger.info(f"Registered custom framework: {framework.name}")
        else:
            logger.warning(f"Cannot register framework {framework.name}: not available")
    
    def get_framework(self, name: str) -> AgentFramework:
        """Get a framework by name"""
        if name not in self._frameworks:
            available = list(self._frameworks.keys())
            raise ValueError(f"Framework '{name}' not found. Available frameworks: {available}")
        return self._frameworks[name]
    
    def list_frameworks(self) -> List[Dict[str, str]]:
        """List all available frameworks"""
        return [
            {
                "name": framework.name,
                "description": framework.description,
                "available": framework.is_available()
            }
            for framework in self._frameworks.values()
        ]
    
    def get_default_framework(self) -> str:
        """Get the default framework name"""
        if "openai_agents" in self._frameworks:
            return "openai_agents"
        elif self._frameworks:
            return next(iter(self._frameworks.keys()))
        else:
            raise RuntimeError("No frameworks available")


# Global registry instance
framework_registry = FrameworkRegistry()