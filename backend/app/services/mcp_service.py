import os
import logging
from typing import Optional, List
from contextlib import asynccontextmanager

try:
    from agents.mcp.server import MCPServerStdio, MCPServerStreamableHttp
except ImportError:
    logging.warning("MCP server not available in openai-agents version")
    MCPServerStdio = None
    MCPServerStreamableHttp = None

logger = logging.getLogger(__name__)


class MCPService:
    """Service for managing MCP (Model Context Protocol) servers"""
    
    def __init__(self):
        self.github_token = os.getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
        self.mcp_servers = []
    
    @asynccontextmanager
    async def get_github_mcp_server(self):
        """Get GitHub MCP server instance"""
        if not MCPServerStreamableHttp:
            logger.warning("MCP not available - GitHub integration disabled")
            yield None
            return
            
        if not self.github_token:
            logger.warning("No GitHub token found - GitHub MCP server disabled")
            yield None
            return
        
        try:
            async with MCPServerStreamableHttp(
                params={
                    "url": "https://api.githubcopilot.com/mcp/",
                    "headers": {
                        "Authorization": f"Bearer {self.github_token}",
                        "Content-Type": "application/json"
                    }
                },
                cache_tools_list=True
            ) as server:
                logger.info("GitHub MCP server connected successfully")
                yield server
        except Exception as e:
            logger.error(f"Failed to connect to GitHub MCP server: {e}")
            yield None
    
    @asynccontextmanager 
    async def get_filesystem_mcp_server(self, base_path: str = "."):
        """Get filesystem MCP server instance"""
        if not MCPServerStdio:
            logger.warning("MCP not available - filesystem integration disabled") 
            yield None
            return
            
        try:
            async with MCPServerStdio(
                params={
                    "command": "npx",
                    "args": ["-y", "@modelcontextprotocol/server-filesystem", base_path]
                },
                cache_tools_list=True
            ) as server:
                logger.info(f"Filesystem MCP server connected for path: {base_path}")
                yield server
        except Exception as e:
            logger.error(f"Failed to connect to filesystem MCP server: {e}")
            yield None
    
    async def get_available_mcp_servers(self) -> List[dict]:
        """Get list of available MCP servers and their status"""
        servers = []
        
        # GitHub MCP Server
        github_status = {
            "name": "github",
            "type": "http",
            "available": bool(self.github_token and MCPServerStreamableHttp),
            "reason": None
        }
        
        if not MCPServerStreamableHttp:
            github_status["reason"] = "MCP not available in agents library"
        elif not self.github_token:
            github_status["reason"] = "No GitHub token configured"
            
        servers.append(github_status)
        
        # Filesystem MCP Server  
        filesystem_status = {
            "name": "filesystem",
            "type": "stdio", 
            "available": bool(MCPServerStdio),
            "reason": None
        }
        
        if not MCPServerStdio:
            filesystem_status["reason"] = "MCP not available in agents library"
            
        servers.append(filesystem_status)
        
        return servers
    
    async def test_github_connection(self) -> dict:
        """Test GitHub MCP server connection"""
        async with self.get_github_mcp_server() as server:
            if not server:
                return {"connected": False, "error": "Server not available"}
                
            try:
                tools = await server.list_tools()
                return {
                    "connected": True, 
                    "tools_count": len(tools) if tools else 0,
                    "tools": [tool.get("name") for tool in tools[:5]] if tools else []
                }
            except Exception as e:
                return {"connected": False, "error": str(e)}


# Global instance
mcp_service = MCPService()