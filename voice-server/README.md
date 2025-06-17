# Relay Voice Server

A voice-controlled interface for the Relay development tool using OpenAI's Realtime API.

## Features

- **Voice Control**: Speak commands to control Relay functionality
- **Real-time Audio**: Low-latency voice processing using OpenAI Realtime API
- **GitHub Integration**: Create and manage GitHub issues via voice
- **Git Operations**: Perform git commits and status checks
- **Project Management**: Switch between and manage multiple projects
- **WebSocket Communication**: Real-time bidirectional communication

## Prerequisites

- Go 1.21 or later
- OpenAI API key with Realtime API access
- Git installed and configured
- GitHub CLI (`gh`) installed and authenticated (optional but recommended)
- Existing Relay installation (for full functionality)

## Quick Start

1. **Clone and Setup**:
```bash
cd voice-server
cp .env.example .env
# Edit .env and add your OpenAI API key
```

2. **Install Dependencies**:
```bash
make deps
```

3. **Run Development Server**:
```bash
export OPENAI_API_KEY="your-api-key-here"
make run
```

4. **Or use live reload**:
```bash
make install-air  # One time setup
make dev
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

- `OPENAI_API_KEY`: Your OpenAI API key (required)
- `PORT`: Server port (default: 8080)
- `GITHUB_TOKEN`: GitHub token for issue management (optional)

## API Endpoints

### REST API
- `GET /health` - Health check
- `GET /api/projects` - List available projects
- `POST /api/projects/:name/select` - Select a project
- `GET /api/projects/:name/status` - Get project status

### WebSocket
- `GET /voice` - Voice communication endpoint

## Voice Commands

Once connected via WebSocket, you can use voice commands like:

- *"Create a new issue titled 'Add user authentication'"*
- *"Update issue 23 to mark it as completed"*
- *"Show me the git status"*
- *"Commit my changes"*
- *"List the open issues"*

## WebSocket Message Format

### Client to Server
```json
{
  "type": "audio|start_recording|stop_recording|select_project",
  "data": "base64_audio_data_or_project_name"
}
```

### Server to Client
```json
{
  "type": "status|audio_response|transcription|function_result",
  "data": {
    "status": "connected|recording|processing|completed|error",
    "message": "Human readable message",
    "audio_data": "base64_encoded_audio",
    "function": "function_name",
    "result": {...}
  }
}
```

## Development

### Project Structure
```
voice-server/
├── main.go              # Server entry point
├── voice_session.go     # WebSocket session handling
├── openai_client.go     # OpenAI Realtime API client
├── relay_manager.go     # Relay integration wrapper
├── Makefile            # Build commands
├── Dockerfile          # Container build
├── .air.toml           # Live reload config
└── .env.example        # Environment template
```

### Building

```bash
# Development build
make build

# Linux build (for Docker)
make build-linux

# Docker build
make docker-build

# Run tests
make test
```

### Hot Reload Development

```bash
# Install air (one time)
make install-air

# Start with hot reload
make dev
```

## Docker Deployment

### Build and Run
```bash
# Build image
make docker-build

# Run container
make docker-run OPENAI_API_KEY=your-key-here
```

### Docker Compose
```yaml
version: '3.8'
services:
  relay-voice:
    build: .
    ports:
      - "8080:8080"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - GITHUB_TOKEN=${GITHUB_TOKEN}
    volumes:
      - ./projects:/root/projects
      - ~/.gitconfig:/root/.gitconfig:ro
      - ~/.ssh:/root/.ssh:ro
    restart: unless-stopped
```

## Integration with Existing Relay

The voice server automatically detects and integrates with your existing Relay installation:

1. **Binary Detection**: Looks for `relay` binary in common locations
2. **Project Discovery**: Uses `relay list` command or falls back to git repository discovery
3. **Command Execution**: Executes Relay commands like `relay commit`, `relay open`, etc.
4. **GitHub Integration**: Leverages existing GitHub CLI authentication

## Architecture

```
Voice Web Client ↔ Voice Server ↔ OpenAI Realtime API
                        ↓
                 Relay Core (existing)
                        ↓
                 GitHub/Git Operations
```

## Troubleshooting

### Common Issues

1. **OpenAI Connection Failed**
   - Verify your API key has Realtime API access
   - Check network connectivity
   - Ensure you're using the correct model

2. **Relay Commands Not Working**
   - Make sure `relay` binary is in PATH or in expected locations
   - Verify project configuration
   - Check GitHub CLI authentication: `gh auth status`

3. **Audio Issues**
   - Ensure proper WebSocket connection
   - Check audio format compatibility (PCM16, 16kHz)
   - Verify browser microphone permissions

### Debugging

Enable debug logging:
```bash
export LOG_LEVEL=debug
make run
```

Monitor WebSocket connections and OpenAI interactions through server logs.

## Security

- OpenAI API keys are kept server-side (never exposed to browser)
- WebSocket connections should be secured with HTTPS/WSS in production
- Consider network access controls for voice endpoints
- GitHub tokens are optional and only used for enhanced functionality

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run `make test`
6. Submit a pull request

## License

MIT License - see LICENSE file for details.
