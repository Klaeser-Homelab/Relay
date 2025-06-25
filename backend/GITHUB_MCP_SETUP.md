# GitHub MCP Integration Setup

This application now supports GitHub integration via Model Context Protocol (MCP), allowing agents to interact with GitHub repositories, issues, and pull requests.

## Setup Instructions

### 1. Get a GitHub Personal Access Token

1. Go to [GitHub Settings > Developer Settings > Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select the following scopes:
   - `repo` - Full control of private repositories
   - `read:org` - Read org and team membership
   - `read:user` - Read user profile data

### 2. Configure Environment Variables

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and add your GitHub token:
   ```
   GITHUB_PERSONAL_ACCESS_TOKEN=your_github_token_here
   ```

### 3. Install Dependencies

If you haven't already, install the dependencies:
```bash
pip install -e .
```

### 4. Test the Integration

1. Start the FastAPI server:
   ```bash
   python -m uvicorn app.main:app --host 0.0.0.0 --port 8080 --reload
   ```

2. Check MCP status:
   ```bash
   curl http://localhost:8080/api/v1/mcp/status
   ```

3. Test GitHub connection:
   ```bash
   curl http://localhost:8080/api/v1/mcp/github/test
   ```

## Using GitHub Integration

### In the Frontend

1. Open the application in your browser
2. In the model selector, toggle "GitHub Integration" to enabled
3. Now you can ask questions about GitHub repositories, such as:
   - "Show me the issues in my repository"
   - "What are the recent pull requests?"
   - "Create a new issue in the repo"

### Available GitHub Capabilities

When GitHub integration is enabled, the agents can:
- List repositories
- Read and create issues
- Manage pull requests
- Access repository content
- Search code and repositories

## Troubleshooting

### GitHub Token Issues
- Ensure your token has the correct scopes
- Check that the token is not expired
- Verify the token is correctly set in your `.env` file

### MCP Connection Issues
- Check the server logs for MCP-related errors
- Ensure you have a stable internet connection for the GitHub API
- Try restarting the server after changing environment variables

## Security Notes

- Never commit your `.env` file with real tokens
- Use environment-specific tokens (development vs production)
- Regularly rotate your GitHub tokens
- Consider using GitHub App authentication for production deployments