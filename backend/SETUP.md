# Setup Guide - Relay Voice Server (JavaScript)

This guide walks you through setting up the JavaScript-based Relay Voice Server using the OpenAI Agent SDK.

## System Requirements

- **Node.js**: 18.0 or later
- **OpenAI API**: Account with Realtime API access
- **Git**: Latest version
- **Operating System**: macOS, Linux, or Windows (with WSL2 recommended)

## Step-by-Step Setup

### 1. Clone and Navigate

```bash
cd voice-server-js
```

### 2. Install Dependencies

```bash
npm install
```

This will install:
- `@openai/agents` - OpenAI Agent SDK for voice AI
- `express` - Web server framework
- `socket.io` - Real-time WebSocket communication
- Additional utilities and development tools

### 3. Environment Configuration

```bash
cp .env.example .env
```

Edit `.env` file:
```env
# Required
OPENAI_API_KEY=sk-your-actual-openai-api-key-here

# Optional but recommended
GH_TOKEN=ghp_your-github-personal-access-token
PORT=8080
NODE_ENV=development
LOG_LEVEL=info
```

### 4. OpenAI API Setup

1. **Get API Key**:
   - Visit [OpenAI API Keys](https://platform.openai.com/api-keys)
   - Create a new secret key
   - Copy the key to your `.env` file

2. **Verify Realtime API Access**:
   ```bash
   curl -H "Authorization: Bearer $OPENAI_API_KEY" \
        https://api.openai.com/v1/models/gpt-4o-realtime-preview-2024-10-01
   ```

3. **Check Usage Limits**:
   - Realtime API requires higher tier access
   - Verify your account has sufficient credits

### 5. GitHub Integration (Optional)

1. **Install GitHub CLI**:
   ```bash
   # macOS
   brew install gh
   
   # Linux
   curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
   echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
   sudo apt update && sudo apt install gh
   ```

2. **Authenticate**:
   ```bash
   gh auth login
   ```

3. **Create Personal Access Token** (if not using GitHub CLI):
   - Go to GitHub Settings → Developer settings → Personal access tokens
   - Create token with `repo`, `issues`, and `user` scopes
   - Add token to `.env` file

### 6. Relay Binary Setup

The voice server works best with the existing Relay binary. It will automatically detect it in these locations:

```bash
# Check if relay is in PATH
which relay

# Or place binary in one of these locations:
../server/relay          # Relative to voice-server-js/
../server/tmp/relay      # Build output location
./relay                  # Current directory
```

If Relay binary is not found, the server will fall back to basic git operations and project discovery.

### 7. Project Discovery Setup

The server will look for projects in these directories:
- `~/Code`
- `~/Projects`
- `~/Development`
- `~/src`

Ensure your development projects are in one of these locations, or they have `.git` directories to be auto-discovered.

### 8. Test Installation

```bash
# Start the server
npm run dev

# In another terminal, test the health endpoint
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "relay-voice-server-js",
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

### 9. Web Interface (Optional)

If you have a web client:

1. **Place web files** in `web/build/` directory
2. **Access interface** at `http://localhost:8080`
3. **Test WebSocket** connection for voice functionality

### 10. Docker Setup (Alternative)

For containerized deployment:

```bash
# Build image
docker build -t relay-voice-server .

# Run container
docker run -p 8080:8080 \
  -e OPENAI_API_KEY=your-key \
  -v $HOME/Code:/home/relay/projects:ro \
  -v $HOME/.gitconfig:/home/relay/.gitconfig:ro \
  relay-voice-server
```

Or use Docker Compose:
```bash
export OPENAI_API_KEY=your-key
docker-compose up -d
```

## Verification Checklist

- [ ] Node.js 18+ installed (`node --version`)
- [ ] Dependencies installed (`npm list`)
- [ ] OpenAI API key configured and valid
- [ ] Server starts without errors (`npm run dev`)
- [ ] Health endpoint responds (`curl localhost:8080/health`)
- [ ] Projects are discoverable (`curl localhost:8080/api/projects`)
- [ ] GitHub CLI authenticated (if using GitHub features)
- [ ] Relay binary detected (check server logs)

## Common Setup Issues

### 1. OpenAI API Key Issues
```bash
# Test API key manually
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/models
```

### 2. Port Already in Use
```bash
# Find what's using port 8080
lsof -i :8080

# Use different port
PORT=3000 npm run dev
```

### 3. Node.js Version Issues
```bash
# Check version
node --version

# Install Node.js 18+ using nvm
nvm install 18
nvm use 18
```

### 4. Permission Issues (Linux/macOS)
```bash
# Fix npm permissions
sudo chown -R $(whoami) ~/.npm
sudo chown -R $(whoami) /usr/local/lib/node_modules
```

### 5. Git Configuration
```bash
# Ensure git is configured
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

## Development Mode

For active development:

```bash
# Install nodemon for auto-restart
npm install -g nodemon

# Start with auto-reload
npm run dev

# Or manually with debugging
DEBUG=* npm run dev
```

## Production Deployment

For production use:

1. **Environment**:
   ```bash
   NODE_ENV=production
   LOG_LEVEL=warn
   ```

2. **Process Management**:
   ```bash
   # Using PM2
   npm install -g pm2
   pm2 start src/server.js --name relay-voice
   
   # Using systemd (Linux)
   sudo systemctl enable relay-voice
   sudo systemctl start relay-voice
   ```

3. **Reverse Proxy** (nginx example):
   ```nginx
   server {
       listen 80;
       server_name your-domain.com;
       
       location / {
           proxy_pass http://localhost:8080;
           proxy_http_version 1.1;
           proxy_set_header Upgrade $http_upgrade;
           proxy_set_header Connection 'upgrade';
           proxy_set_header Host $host;
           proxy_cache_bypass $http_upgrade;
       }
   }
   ```

## Next Steps

Once setup is complete:

1. **Test voice commands** using the web interface
2. **Configure your projects** for voice control
3. **Customize tools** in `src/tools/` as needed
4. **Set up monitoring** and logging for production use

For usage instructions, see the main [README.md](README.md).