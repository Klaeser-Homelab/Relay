# Relay Voice Frontend

A modern React frontend for voice-controlled GitHub repository management, built with TypeScript and Tailwind CSS.

## Features

- 🎤 **Voice Recording**: Push-to-talk or toggle recording modes
- 📱 **Responsive Design**: Works on desktop and mobile devices
- 🔄 **Real-time Updates**: Live transcription and function results
- 🎨 **Modern UI**: Clean, GitHub-inspired interface
- 🔊 **Audio Playback**: Hear AI responses directly in the browser
- ⚡ **Real-time Communication**: WebSocket integration with voice server

## Getting Started

### Prerequisites

- Node.js 18+ 
- The voice server running on port 8080
- A GitHub token configured in the voice server

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

The frontend will be available at `http://localhost:3000` and will proxy API calls to the voice server at `http://localhost:8080`.

### Build for Production

```bash
npm run build
npm run preview
```

## Usage

1. **Select Repository**: Choose a GitHub repository from your available projects
2. **Connect**: Click "Connect to Voice Assistant" 
3. **Start Recording**: Use push-to-talk (hold) or toggle mode
4. **Voice Commands**: Say things like:
   - "Create an issue for adding user authentication"
   - "Show me the open issues"
   - "List recent commits"
   - "Get repository information"

## Architecture

```
Frontend (React/TypeScript)
├── Components/
│   ├── ProjectSelector - GitHub repo selection
│   ├── VoiceChat - Recording interface
│   ├── StatusDisplay - Connection status
│   ├── TranscriptionView - Speech-to-text
│   └── FunctionResults - GitHub operation results
├── Hooks/
│   ├── useGitHubProjects - Project management
│   ├── useWebSocket - Socket.io integration
│   └── useAudioRecording - Web Audio API
└── Types/
    └── api.ts - TypeScript interfaces
```

## Development

- **Hot Reload**: Changes are reflected instantly
- **TypeScript**: Full type safety throughout
- **ESLint**: Code quality enforcement
- **Tailwind CSS**: Utility-first styling

## Browser Compatibility

- Chrome/Edge 88+
- Firefox 84+
- Safari 14+

Requires modern browser with:
- WebSocket support
- Web Audio API
- MediaRecorder API
- ES2020 features