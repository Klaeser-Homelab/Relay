# 🎤 Relay Voice Control Server

## What We Built

I've created a complete voice-controlled web server that integrates with your existing Relay TUI project. Here's what's included:

### Architecture Overview
```
Voice Web Client ↔ Voice Server ↔ OpenAI Realtime API
                        ↓
                 Relay Core (your existing code)
                        ↓
                 GitHub/Git Operations
```

## Files Created

### Core Server Files
- `main.go` - Web server with Fiber framework and WebSocket handling
- `voice_session.go` - WebSocket session management and voice processing
- `openai_client.go` - OpenAI Realtime API integration
- `relay_manager.go` - Wrapper for your existing Relay functionality

### Configuration & Deployment
- `go.mod` - Go dependencies
- `Makefile` - Build and development commands
- `.env.example` - Environment variable template
- `Dockerfile` - Container build configuration
- `docker-compose.yml` - Multi-service deployment
- `Caddyfile` - Reverse proxy configuration
- `.air.toml` - Hot reload configuration

### Web Client
- `web/build/index.html` - Test web interface for voice control

### Scripts & Documentation
- `start.sh` - Quick start script
- `README.md` - Comprehensive documentation
- `main_test.go` - Basic tests

## Quick Start

1. **Setup Environment**:
```bash
cd voice-server
cp .env.example .env
# Edit .env and add your OpenAI API key
```

2. **Get OpenAI API Key**:
   - Go to https://platform.openai.com/
   - Create an API key with Realtime API access
   - Add it to `.env`: `OPENAI_API_KEY=your-key-here`

3. **Run the Server**:
```bash
chmod +x start.sh
./start.sh
```

Or manually:
```bash
export OPENAI_API_KEY="your-api-key-here"
make run
```

4. **Test the Interface**:
   - Open http://localhost:8080 in your browser
   - Allow microphone access
   - Select a project from the dropdown
   - Hold the microphone button and speak commands

## Voice Commands You Can Try

Once the server is running and connected:

- *"Create a new issue titled 'Add voice control feature'"*
- *"Show me the git status"*
- *"Commit my changes"*
- *"List the open issues"*
- *"Update issue 5 to mark it as completed"*

## Integration with Your Existing Relay

The voice server automatically integrates with your existing Relay setup:

1. **Binary Detection**: Looks for your `relay` binary in `../server/relay` and other locations
2. **Project Discovery**: Uses `relay list` command or discovers git repositories
3. **Command Execution**: Executes existing Relay commands like `relay commit`
4. **GitHub Integration**: Uses existing GitHub CLI authentication

## Development Mode

For development with hot reload:
```bash
make install-air  # One time setup
make dev
```

This will automatically rebuild and restart the server when you modify Go files.

## Docker Deployment

For deployment on your Proxmox server:

1. **Build and run with Docker**:
```bash
make docker-build
export OPENAI_API_KEY="your-key"
make docker-run
```

2. **Or use Docker Compose**:
```bash
export OPENAI_API_KEY="your-key"
docker-compose up -d
```

This includes Caddy reverse proxy for SSL termination and better WebSocket handling.

## How It Works

### Voice Processing Flow
1. **Browser** captures audio via Web Audio API
2. **WebSocket** streams audio to voice server
3. **Voice Server** forwards audio to OpenAI Realtime API
4. **OpenAI** processes speech → text → function calls → speech
5. **Voice Server** executes Relay commands and streams responses back

### Function Calling
The server configures OpenAI with these tools:
- `create_github_issue` - Create new GitHub issues
- `update_github_issue` - Update existing issues
- `git_commit` - Perform smart commits
- `git_status` - Check repository status
- `list_issues` - List project issues

### Real-time Features
- **Low-latency audio streaming** using WebSocket binary frames
- **Bidirectional communication** for status updates and responses
- **Session management** with automatic reconnection
- **Project context** maintained across voice interactions

## Security Considerations

- ✅ **API Key Security**: OpenAI key stays on server, never exposed to browser
- ✅ **Local Network**: Recommended for home network deployment
- ✅ **HTTPS/WSS**: Supported via Caddy reverse proxy
- ⚠️ **Microphone Access**: Requires user permission in browser

## Customization

### Adding New Voice Commands
1. Add function definition in `voice_session.go` `configureOpenAITools()`
2. Implement handler in `relay_manager.go` `ExecuteFunction()`
3. Test with voice commands

### Extending Relay Integration
The `RelayManager` can be extended to call more of your existing Relay functionality:
- Issue management from your GitHub service
- Git operations from your git wrapper
- Project management from your project manager

## Next Steps

1. **Test the basic functionality** with the provided web interface
2. **Customize voice commands** based on your workflow
3. **Deploy to your Proxmox server** using Docker Compose
4. **Build a more advanced React frontend** for mobile devices
5. **Add authentication** for production use

## Troubleshooting

### Common Issues

1. **"Connection failed"**: Check if server is running on port 8080
2. **"OpenAI error"**: Verify API key and Realtime API access
3. **"No projects found"**: Ensure `relay` binary is accessible or projects exist in common directories
4. **"Microphone access denied"**: Allow microphone permissions in browser

### Debug Mode
```bash
export LOG_LEVEL=debug
./start.sh
```

The voice server is ready to use! It provides a solid foundation for voice-controlled development and can be extended with additional Relay functionality as needed.
