from fastapi import APIRouter
from app.services.mcp_service import mcp_service

router = APIRouter()


@router.get("/status")
async def get_mcp_status():
    """Get status of available MCP servers"""
    servers = await mcp_service.get_available_mcp_servers()
    return {
        "mcp_available": any(server["available"] for server in servers),
        "servers": servers
    }


@router.get("/github/test")
async def test_github_connection():
    """Test GitHub MCP server connection"""
    result = await mcp_service.test_github_connection()
    return result