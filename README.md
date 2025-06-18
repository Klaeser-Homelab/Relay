# Relay

[![CI Status](https://github.com/yourusername/relay/workflows/Relay%20Server%20CI/badge.svg)](https://github.com/yourusername/relay/actions)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Code from anywhere, control with your voice**

A mobile friendly voice-controlled development proof of concept.

## Overview

Relay consists of two components:

- **Relay Server**: Runs on your development machine, executing MCP tools and development commands
- **Relay Mobile**: Voice-enabled mobile website for issuing commands and receiving audio feedback.

## How It Works

1. **Voice Command**: Speak development instructions to Relay Mobile
2. **Command Processing**: Your voice is converted to actionable development commands
3. **Tool Execution**: Relay Server executes commands using MCP tools (Claude Code, Git, npm, etc.)
4. **Smart Summaries**: Server generates concise summaries of results and changes
5. **Audio Feedback**: Relay Mobile reads summaries aloud, keeping you informed

## Key Features

### Voice-Driven Development
- Natural language commands: "Add user authentication with JWT"
- Context-aware responses: "What if we used Sequelize instead?"
- Hands-free coding workflow

### Asynchronous Command Queue
- Queue multiple commands while others execute
- Dependency-aware execution: "Run tests, and if they pass, commit"
- Think ahead while previous commands complete

### MCP Tool Integration
- **Claude Code**: AI-powered code generation and modification
- **Git**: Version control operations
- **Execution Logs**: Monitor build processes, test results, dev servers
- **Extensible**: Add custom tools through MCP protocol

### Smart Context Management
- Maintains awareness of current project state
- Rollback points for safe experimentation
- Command history and undo capabilities

## Architecture

```
┌─────────────────┐    WebSocket/SSE    ┌──────────────────┐
│                 │◄─────────────────►  │                  │
│  Relay Mobile   │                     │   Relay Server   │
│                 │                     │                  │
│ • Voice Input   │    Text Summaries   │ • MCP Tools      │
│ • Audio Output  │◄─────────────────── │ • Command Queue  │
│ • Command Queue │                     │ • State Manager  │
│                 │                     │                  │
└─────────────────┘                     └──────────────────┘
                                                    │
                                                    ▼
                                        ┌──────────────────┐
                                        │  Dev Environment │
                                        │                  │
                                        │ • File System    │
                                        │ • Git Repository │
                                        │ • Build Tools    │
                                        │ • Test Runners   │
                                        └──────────────────┘
```

## Example Workflow

```
You: "Create a new React component for user profiles"
Relay: "Creating UserProfile component with JSX structure, PropTypes, and basic styling. Done. Component created at src/components/UserProfile.jsx"

You: "Add a form for editing user details"
Relay: "Adding form with name, email, and bio fields. Includes validation and submit handler. Complete. Would you like me to add any specific validation rules?"

You: "Run the development server and run tests"
Relay: "Development server starting on port 3000. Running test suite... 12 tests passing, 0 failures. Server ready at localhost:3000"
```

## Getting Started

### Prerequisites
- Development machine with Node.js, Git, and your preferred development tools
- iOS/Android device for Relay Mobile
- Network connectivity between devices

### Installation

1. **Set up Relay Server**:
   ```bash
   git clone https://github.com/yourusername/relay
   cd relay/server
   npm install
   npm run setup
   ```

2. **Configure MCP Tools**:
   ```bash
   relay configure claude-code
   relay configure git
   relay configure npm
   ```

3. **Install Relay Mobile**:
   - Download from App Store/Google Play
   - Connect to your Relay Server using the pairing code

### Configuration

Create a `relay.config.json` in your project root:

```json
{
  "tools": {
    "claude-code": {
      "enabled": true,
      "apiKey": "your-api-key"
    },
    "git": {
      "enabled": true,
      "autoCommit": false
    },
    "npm": {
      "enabled": true,
      "scripts": ["dev", "test", "build"]
    }
  },
  "voice": {
    "language": "en-US",
    "confirmDestructive": true
  }
}
```

## Voice Commands

### Code Management
- "Create a new [component/function/class] called [name]"
- "Add [feature] to [file/component]"
- "Refactor [component] to use [pattern/library]"
- "Fix the error in [file]"

### Git Operations
- "Commit these changes with message [message]"
- "Create a new branch called [name]"
- "Merge [branch] into main"
- "Show me the diff"

### Execution
- "Run the development server"
- "Run tests"
- "Build the project"
- "Install [package]"

### Queue Management
- "After that's done, [next command]"
- "Hold on, stop the current command"
- "What's in the queue?"
- "Clear the queue"

## Development

### Building from Source

```bash
git clone https://github.com/yourusername/relay
cd relay/server
go build -o relay .
```

### Development with Hot Reload

For faster development iterations, use Air for automatic rebuilding:

```bash
# Install Air (one-time setup)
go install github.com/air-verse/air@latest

# Start development server with hot reload
cd server
air

# Air will automatically rebuild and restart when you modify Go files
```

Air is configured to:
- Watch all `.go` files in the server directory
- Exclude test files and temporary directories  
- Automatically restart with `start Relay` arguments
- Show colored build output and logs

### Running Tests

```bash
# Run full test suite
cd server
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestSpecificFunction ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Continuous Integration

Relay uses GitHub Actions for automated testing:

- **Tests**: Run on every push and pull request
- **Multiple Go versions**: Tested against Go 1.21.x and 1.22.x
- **Cross-platform builds**: Linux, macOS, and Windows
- **Security scanning**: Vulnerability and security checks
- **Performance testing**: Automated performance benchmarks

All tests must pass before merging to main branch.

## Security

- All communication encrypted with TLS
- Authentication tokens required for server access
- Recommend VPN or local network usage
- No code or credentials stored on mobile device