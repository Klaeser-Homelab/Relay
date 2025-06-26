from fastapi import APIRouter, HTTPException, Depends
from sqlalchemy.orm import Session
from typing import List, Dict, Any, Optional
import os
import aiohttp
from datetime import datetime

from app.core.database import get_db
from app.core.config import settings

router = APIRouter()

# In-memory storage for selected repository (in production, use database)
selected_repository: Optional[str] = None

async def get_github_headers():
    """Get GitHub API headers with authentication"""
    token = os.getenv("GH_TOKEN") or os.getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
    if not token:
        raise HTTPException(status_code=500, detail="GitHub token not configured")
    
    return {
        "Authorization": f"Bearer {token}",
        "Accept": "application/vnd.github.v3+json"
    }

@router.get("/")
async def list_repositories():
    """Get list of all GitHub repositories"""
    headers = await get_github_headers()
    
    async with aiohttp.ClientSession() as session:
        # Get user's repositories
        async with session.get(
            "https://api.github.com/user/repos",
            headers=headers,
            params={"per_page": 100, "sort": "updated"}
        ) as response:
            if response.status != 200:
                raise HTTPException(
                    status_code=response.status,
                    detail=f"GitHub API error: {await response.text()}"
                )
            
            repos = await response.json()
            
            # Format repository data
            repositories = []
            for repo in repos:
                repositories.append({
                    "id": repo["id"],
                    "name": repo["name"],
                    "full_name": repo["full_name"],
                    "description": repo["description"],
                    "private": repo["private"],
                    "language": repo["language"],
                    "default_branch": repo["default_branch"],
                    "created_at": repo["created_at"],
                    "updated_at": repo["updated_at"],
                    "pushed_at": repo["pushed_at"],
                    "html_url": repo["html_url"],
                    "clone_url": repo["clone_url"],
                    "ssh_url": repo["ssh_url"],
                    "is_cloned": check_if_cloned(repo["name"]),  # Check local clone status
                    "stargazers_count": repo["stargazers_count"],
                    "open_issues_count": repo["open_issues_count"]
                })
            
            return {
                "repositories": repositories,
                "total": len(repositories),
                "selected": selected_repository
            }

@router.post("/select")
async def select_repository(repository_name: str):
    """Select a repository as the active repository"""
    global selected_repository
    
    # Verify repository exists
    headers = await get_github_headers()
    
    async with aiohttp.ClientSession() as session:
        # Try to get repository info to verify it exists
        async with session.get(
            f"https://api.github.com/repos/{repository_name}",
            headers=headers
        ) as response:
            if response.status == 404:
                raise HTTPException(status_code=404, detail=f"Repository '{repository_name}' not found")
            elif response.status != 200:
                raise HTTPException(
                    status_code=response.status,
                    detail=f"GitHub API error: {await response.text()}"
                )
            
            repo_data = await response.json()
            selected_repository = repo_data["full_name"]
            
            return {
                "success": True,
                "selected": selected_repository,
                "repository": {
                    "name": repo_data["name"],
                    "full_name": repo_data["full_name"],
                    "description": repo_data["description"],
                    "language": repo_data["language"]
                }
            }

@router.get("/status")
async def get_repository_status(repository_name: str):
    """Get detailed status of a repository"""
    headers = await get_github_headers()
    
    async with aiohttp.ClientSession() as session:
        # Get repository info
        async with session.get(
            f"https://api.github.com/repos/{repository_name}",
            headers=headers
        ) as response:
            if response.status == 404:
                raise HTTPException(status_code=404, detail=f"Repository '{repository_name}' not found")
            elif response.status != 200:
                raise HTTPException(
                    status_code=response.status,
                    detail=f"GitHub API error: {await response.text()}"
                )
            
            repo_data = await response.json()
            
            # Get additional status info
            status_info = {
                "name": repo_data["name"],
                "full_name": repo_data["full_name"],
                "is_selected": selected_repository == repo_data["full_name"],
                "is_cloned": check_if_cloned(repo_data["name"]),
                "local_path": get_local_path(repo_data["name"]) if check_if_cloned(repo_data["name"]) else None,
                "default_branch": repo_data["default_branch"],
                "last_push": repo_data["pushed_at"],
                "open_issues": repo_data["open_issues_count"],
                "language": repo_data["language"],
                "size": repo_data["size"],  # KB
                "visibility": "private" if repo_data["private"] else "public"
            }
            
            # Get branch info
            async with session.get(
                f"https://api.github.com/repos/{repository_name}/branches",
                headers=headers
            ) as branch_response:
                if branch_response.status == 200:
                    branches = await branch_response.json()
                    status_info["branches"] = [b["name"] for b in branches]
                    status_info["branch_count"] = len(branches)
            
            return status_info

@router.get("/info")
async def get_repository_info(repository_name: str):
    """Get comprehensive repository information"""
    headers = await get_github_headers()
    
    async with aiohttp.ClientSession() as session:
        # Get repository info
        async with session.get(
            f"https://api.github.com/repos/{repository_name}",
            headers=headers
        ) as response:
            if response.status == 404:
                raise HTTPException(status_code=404, detail=f"Repository '{repository_name}' not found")
            elif response.status != 200:
                raise HTTPException(
                    status_code=response.status,
                    detail=f"GitHub API error: {await response.text()}"
                )
            
            repo_data = await response.json()
            
            # Get contributors
            contributors = []
            async with session.get(
                f"https://api.github.com/repos/{repository_name}/contributors",
                headers=headers,
                params={"per_page": 10}
            ) as contrib_response:
                if contrib_response.status == 200:
                    contributors = await contrib_response.json()
            
            # Get languages
            languages = {}
            async with session.get(
                f"https://api.github.com/repos/{repository_name}/languages",
                headers=headers
            ) as lang_response:
                if lang_response.status == 200:
                    languages = await lang_response.json()
            
            # Get recent commits
            commits = []
            async with session.get(
                f"https://api.github.com/repos/{repository_name}/commits",
                headers=headers,
                params={"per_page": 5}
            ) as commit_response:
                if commit_response.status == 200:
                    commit_data = await commit_response.json()
                    commits = [{
                        "sha": c["sha"][:7],
                        "message": c["commit"]["message"].split("\n")[0],
                        "author": c["commit"]["author"]["name"],
                        "date": c["commit"]["author"]["date"]
                    } for c in commit_data]
            
            return {
                "repository": {
                    "name": repo_data["name"],
                    "full_name": repo_data["full_name"],
                    "description": repo_data["description"],
                    "created_at": repo_data["created_at"],
                    "updated_at": repo_data["updated_at"],
                    "pushed_at": repo_data["pushed_at"],
                    "size": repo_data["size"],
                    "stargazers_count": repo_data["stargazers_count"],
                    "watchers_count": repo_data["watchers_count"],
                    "forks_count": repo_data["forks_count"],
                    "open_issues_count": repo_data["open_issues_count"],
                    "default_branch": repo_data["default_branch"],
                    "private": repo_data["private"],
                    "archived": repo_data["archived"],
                    "disabled": repo_data["disabled"],
                    "topics": repo_data["topics"],
                    "html_url": repo_data["html_url"],
                    "clone_url": repo_data["clone_url"],
                    "ssh_url": repo_data["ssh_url"]
                },
                "languages": languages,
                "contributors": [
                    {
                        "login": c["login"],
                        "contributions": c["contributions"],
                        "avatar_url": c["avatar_url"]
                    } for c in contributors[:5]
                ],
                "recent_commits": commits,
                "local_status": {
                    "is_cloned": check_if_cloned(repo_data["name"]),
                    "local_path": get_local_path(repo_data["name"]) if check_if_cloned(repo_data["name"]) else None
                }
            }

@router.get("/issues")
async def get_repository_issues(
    repository_name: str,
    state: str = "open",
    labels: Optional[str] = None,
    sort: str = "created",
    direction: str = "desc",
    per_page: int = 30,
    page: int = 1
):
    """Get issues for a specific repository"""
    headers = await get_github_headers()
    
    params = {
        "state": state,
        "sort": sort,
        "direction": direction,
        "per_page": per_page,
        "page": page
    }
    
    if labels:
        params["labels"] = labels
    
    async with aiohttp.ClientSession() as session:
        async with session.get(
            f"https://api.github.com/repos/{repository_name}/issues",
            headers=headers,
            params=params
        ) as response:
            if response.status == 404:
                raise HTTPException(status_code=404, detail=f"Repository '{repository_name}' not found")
            elif response.status != 200:
                raise HTTPException(
                    status_code=response.status,
                    detail=f"GitHub API error: {await response.text()}"
                )
            
            issues_data = await response.json()
            
            # Format issues
            issues = []
            for issue in issues_data:
                # Skip pull requests (they appear in issues endpoint too)
                if "pull_request" in issue:
                    continue
                
                issues.append({
                    "number": issue["number"],
                    "title": issue["title"],
                    "body": issue["body"],
                    "state": issue["state"],
                    "created_at": issue["created_at"],
                    "updated_at": issue["updated_at"],
                    "closed_at": issue["closed_at"],
                    "user": {
                        "login": issue["user"]["login"],
                        "avatar_url": issue["user"]["avatar_url"]
                    },
                    "labels": [label["name"] for label in issue["labels"]],
                    "assignees": [a["login"] for a in issue["assignees"]],
                    "comments": issue["comments"],
                    "html_url": issue["html_url"],
                    "milestone": issue["milestone"]["title"] if issue["milestone"] else None
                })
            
            # Get total count from headers
            link_header = response.headers.get("Link", "")
            total_pages = parse_link_header_pages(link_header, page)
            
            return {
                "issues": issues,
                "pagination": {
                    "page": page,
                    "per_page": per_page,
                    "total_pages": total_pages,
                    "count": len(issues)
                },
                "repository": repository_name,
                "filters": {
                    "state": state,
                    "labels": labels,
                    "sort": sort,
                    "direction": direction
                }
            }

def check_if_cloned(repo_name: str) -> bool:
    """Check if repository is cloned locally"""
    # Check common locations
    home_dir = os.path.expanduser("~")
    possible_paths = [
        f"{home_dir}/Code/{repo_name}",
        f"{home_dir}/Projects/{repo_name}",
        f"{home_dir}/repos/{repo_name}",
        f"{home_dir}/github/{repo_name}",
        f"/tmp/{repo_name}"
    ]
    
    for path in possible_paths:
        if os.path.exists(os.path.join(path, ".git")):
            return True
    
    return False

def get_local_path(repo_name: str) -> Optional[str]:
    """Get local path of cloned repository"""
    home_dir = os.path.expanduser("~")
    possible_paths = [
        f"{home_dir}/Code/{repo_name}",
        f"{home_dir}/Projects/{repo_name}",
        f"{home_dir}/repos/{repo_name}",
        f"{home_dir}/github/{repo_name}",
        f"/tmp/{repo_name}"
    ]
    
    for path in possible_paths:
        if os.path.exists(os.path.join(path, ".git")):
            return path
    
    return None

def parse_link_header_pages(link_header: str, current_page: int) -> int:
    """Parse GitHub Link header to get total pages"""
    if not link_header:
        return current_page
    
    # Look for 'last' link
    links = link_header.split(",")
    for link in links:
        if 'rel="last"' in link:
            # Extract page number from URL
            import re
            match = re.search(r'[?&]page=(\d+)', link)
            if match:
                return int(match.group(1))
    
    return current_page