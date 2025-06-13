# Relay MCP Tools

This directory contains MCP (Model Context Protocol) tools that extend Claude Code's capabilities for Relay project management.

## Available Tools

### Issue Planner (`tools/issue-planner`)

Updates GitHub issues with plans extracted from Claude Code conversations.

**Features:**
- Auto-detects issue number from current git branch (`feature/issue-16` → issue #16)
- Extracts and formats plan content
- Updates GitHub issue body with the plan
- Preserves existing issue content

**Usage:**
```json
{
  "name": "update_issue_plan",
  "arguments": {
    "plan": "1. Analyze codebase\n2. Implement feature\n3. Test functionality",
    "workingDir": "/path/to/project",
    "issueNumber": 16
  }
}
```

**Parameters:**
- `plan` (required): The plan content to add to the issue
- `workingDir` (optional): Working directory, defaults to current directory
- `issueNumber` (optional): Issue number, auto-detected from branch if not provided

## Setup

1. **Build the tool:**
   ```bash
   cd mcp/tools/issue-planner
   go build -o issue-planner main.go
   ```

2. **Configure Claude Code:**
   Add the MCP server configuration to your Claude Code settings:
   ```json
   {
     "mcpServers": {
       "issue-planner": {
         "command": "go",
         "args": ["run", "main.go"],
         "cwd": "./mcp/tools/issue-planner"
       }
     }
   }
   ```

3. **Verify GitHub CLI setup:**
   ```bash
   gh auth status
   ```

## Development

### Adding New Tools

1. Create a new directory under `tools/`
2. Implement the MCP server interface
3. Add configuration to `mcp_server.json`
4. Update this README

### Testing

```bash
# Test issue planner tool
cd mcp/tools/issue-planner
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | go run main.go
```

## Architecture

- **`shared/`**: Common utilities shared across MCP tools
- **`tools/`**: Individual MCP tool implementations
- **JSON-RPC**: Standard MCP protocol for communication with Claude Code

## Requirements

- Go 1.21+
- GitHub CLI (`gh`) authenticated
- Git repository with GitHub remote