#!/bin/bash

# Relay Voice Server Startup Script

set -e

echo "🎤 Relay Voice Server Setup"
echo "=========================="

# Load .env file if it exists (before other checks)
if [ -f ".env" ]; then
    echo "📋 Loading environment from .env file..."
    export $(grep -v '^#' .env | xargs)
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

echo "✅ Go version: $(go version)"

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    echo "❌ Please run this script from the voice-server directory"
    exit 1
fi

# Check for OpenAI API key
if [ -z "$OPENAI_API_KEY" ]; then
    echo "⚠️  OpenAI API key not set in environment"
    echo "Please set OPENAI_API_KEY environment variable or add it to .env file"
    echo ""
    echo "Example:"
    echo "export OPENAI_API_KEY='your-api-key-here'"
    echo "./start.sh"
    echo ""
    echo "Or create .env file (copy from .env.example)"
    if [ ! -f ".env" ]; then
        echo "Creating .env file from template..."
        cp .env.example .env
        echo "📝 Please edit .env file and add your OpenAI API key"
    fi
    echo ""
    read -p "Do you want to continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo ""
echo "📦 Installing dependencies..."
go mod download
go mod tidy

echo ""
echo "🧪 Running tests..."
if go test -v ./...; then
    echo "✅ All tests passed!"
else
    echo "⚠️  Some tests failed, but continuing..."
fi

echo ""
echo "🔨 Building server..."
go build -o relay-voice .

echo ""
echo "🚀 Starting Relay Voice Server..."
echo "Server will be available at: http://localhost:${PORT:-8080}"
echo "Health check: http://localhost:${PORT:-8080}/health"
echo "WebSocket endpoint: ws://localhost:${PORT:-8080}/voice"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Start the server
exec ./relay-voice
