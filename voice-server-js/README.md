# Relay Voice Server (JavaScript)

A voice-controlled interface for the Relay development tool, rewritten in JavaScript using the OpenAI Agent SDK for enhanced voice AI capabilities.

## Features

- **🎤 Advanced Voice Control**: Powered by OpenAI Agent SDK with built-in audio processing
- **🔄 Real-time Communication**: WebSocket-based bidirectional communication via Socket.io
- **🛠️ GitHub Integration**: Create, update, and manage GitHub issues via voice commands
- **📦 Git Operations**: Perform git commits, status checks, and branch operations
- **🎯 Project Management**: Switch between and manage multiple development projects
- **🤖 Intelligent Function Calling**: Automatic schema generation and validation
- **📊 Built-in Tracing**: Debug voice interactions with comprehensive logging

## Prerequisites

- Node.js 18.0 or later
- OpenAI API key with Realtime API access
- Git installed and configured
- GitHub CLI (`gh`) installed and authenticated (optional but recommended)
- Existing Relay installation (for full functionality)

## Quick Start

### 1. Installation

```bash
cd voice-server-js
npm install
```

### 2. Configuration

```bash
cp .env.example .env
# Edit .env and add your OpenAI API key
```

Required environment variables:
```env
OPENAI_API_KEY=your-openai-api-key-here
```

Optional:
```env
GITHUB_TOKEN=your-github-token-here
PORT=8080
NODE_ENV=development
```

### 3. Development

```bash
# Start development server with hot reload
npm run dev

# Or start production server
npm start
```

### 4. Testing

Open your browser to `http://localhost:8080` to access the web interface, or connect directly via WebSocket to test voice functionality.

## API Reference

### REST Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/projects` | List available projects |
| `POST` | `/api/projects/:name/select` | Select a project |
| `GET` | `/api/projects/:name/status` | Get project status |

### WebSocket Events

#### Client → Server
- `audio` - Send audio data for processing
- `start_recording` - Begin voice recording session
- `stop_recording` - End voice recording session
- `select_project` - Switch to a different project

#### Server → Client
- `status` - Session status updates
- `audio_response` - AI-generated audio response
- `transcription` - Speech-to-text results
- `function_result` - Results from executed commands

## Voice Commands

Once connected, you can use natural voice commands like:

- *"Create a new issue titled 'Add user authentication'"*
- *"Update issue 23 to mark it as completed"*
- *"Show me the git status"*
- *"Commit my changes with a smart message"*
- *"List all open issues"*
- *"Switch to the relay project"*
- *"Push my changes to the remote repository"*

## Architecture Improvements

### OpenAI Agent SDK Benefits

The JavaScript rewrite leverages the OpenAI Agent SDK for significant improvements:

- **Simplified Audio Handling**: No manual base64 encoding/decoding
- **Automatic Schema Generation**: Function definitions auto-generate OpenAI schemas
- **Built-in Error Handling**: Better connection management and error recovery
- **Context Management**: Improved conversation context and interruption handling
- **Tracing & Debugging**: Built-in logging for voice interaction debugging

### Project Structure

```
voice-server-js/
├── package.json          # Dependencies and scripts
├── src/
│   ├── server.js         # Express + Socket.io server
│   ├── voice-session.js  # RealtimeAgent session management
│   ├── relay-manager.js  # Relay CLI integration
│   └── tools/            # Function definitions
│       ├── index.js      # Tool registry
│       ├── github-tools.js
│       ├── git-tools.js
│       └── project-tools.js
├── web/                  # Static web client (optional)
├── Dockerfile           # Container configuration
├── docker-compose.yml   # Multi-service setup
└── README.md            # This file
```

## Docker Deployment

### Build and Run

```bash
# Build the image
docker build -t relay-voice-server .

# Run with environment variables
docker run -p 8080:8080 \
  -e OPENAI_API_KEY=your-key \
  -v $HOME/Code:/home/relay/projects:ro \
  relay-voice-server
```

### Docker Compose

```bash
# Set environment variables
export OPENAI_API_KEY=your-key-here
export GITHUB_TOKEN=your-token-here

# Start services
docker-compose up -d

# View logs
docker-compose logs -f relay-voice

# Stop services
docker-compose down
```

## Development

### Available Scripts

```bash
npm start       # Start production server
npm run dev     # Start development server with hot reload
npm test        # Run tests (when implemented)
npm run build   # No build step needed for Node.js
```

### Adding New Voice Commands

1. **Define the tool** in `src/tools/`:
```javascript
export const myNewTool = {
  name: 'my_new_tool',
  description: 'Description of what this tool does',
  parameters: {
    type: 'object',
    properties: {
      param1: {
        type: 'string',
        description: 'Parameter description'
      }
    },
    required: ['param1']
  },
  async execute(relayManager, projectName, args) {
    // Implementation
    return { success: true, message: 'Done!' };
  }
};
```

2. **Register the tool** in `src/tools/index.js`
3. **Implement the backend** in `relay-manager.js` if needed

### Integration with Existing Relay

The voice server automatically detects and integrates with your existing Relay installation:

1. **Binary Detection**: Searches for `relay` binary in common locations
2. **Project Discovery**: Uses `relay list` or falls back to git repository scanning
3. **Command Execution**: Executes Relay commands like `relay commit`, `relay open`
4. **GitHub Integration**: Leverages existing GitHub CLI authentication

## Troubleshooting

### Common Issues

1. **OpenAI Connection Failed**
   - Verify API key has Realtime API access
   - Check network connectivity and firewall settings
   - Ensure you're using a supported model

2. **Audio Not Working**
   - Check browser microphone permissions
   - Verify WebSocket connection is established
   - Look for audio format compatibility issues

3. **Relay Commands Not Working**
   - Ensure `relay` binary is in PATH or expected locations
   - Verify project configuration
   - Check GitHub CLI authentication: `gh auth status`

### Debugging

Enable detailed logging:
```bash
export LOG_LEVEL=debug
npm run dev
```

Monitor WebSocket connections and function calls through server logs.

## Migration from Go Version

Key differences from the original Go implementation:

- **Simplified codebase**: ~50% reduction in lines of code
- **Better error handling**: Built-in SDK error management
- **Improved audio**: No manual audio format handling
- **Enhanced debugging**: Built-in tracing and logging
- **Same API compatibility**: Existing clients work unchanged

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run tests: `npm test`
5. Commit changes: `git commit -m 'Add amazing feature'`
6. Push to branch: `git push origin feature/amazing-feature`
7. Create a Pull Request

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
- Check the [troubleshooting section](#troubleshooting)
- Review server logs for error details
- Open an issue on the repository with:
  - Environment details (Node.js version, OS)
  - Error logs and steps to reproduce
  - Expected vs actual behavior