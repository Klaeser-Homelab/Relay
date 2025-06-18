import express from 'express';
import { createServer } from 'http';
import { Server as SocketIOServer } from 'socket.io';
import cors from 'cors';
import dotenv from 'dotenv';
import path from 'path';
import { fileURLToPath } from 'url';
import * as pty from 'node-pty';

import { VoiceSession } from './voice-session.js';
import { GitHubManager } from './github-manager.js';
import { GitManager } from './git-manager.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

dotenv.config();

// Obscure and log API key presence for debugging
function obscure(key) {
  if (!key) return '(not set)';
  if (key.length <= 5) return key[0] + '***' + key[key.length-1];
  return key.substring(0,3) + '***' + key.substring(key.length-2);
}
console.log(`OPENAI_API_KEY: ${obscure(process.env.OPENAI_API_KEY)}`);
console.log(`GEMINI_API_KEY: ${obscure(process.env.GEMINI_API_KEY)}`);
console.log(`GH_TOKEN: ${obscure(process.env.GH_TOKEN)}`);

class VoiceServer {
  constructor() {
    this.app = express();
    this.server = createServer(this.app);
    this.io = new SocketIOServer(this.server, {
      cors: {
        origin: "*",
        methods: ["GET", "POST"]
      }
    });
    
    this.openaiAPIKey = process.env.OPENAI_API_KEY;
    this.port = process.env.PORT || 8080;
    this.sessions = new Map();
    this.terminalSessions = new Map();
    
    if (!this.openaiAPIKey) {
      throw new Error('OPENAI_API_KEY environment variable is required');
    }
    
    this.githubManager = new GitHubManager();
    this.gitManager = new GitManager();
    this.setupMiddleware();
    this.setupRoutes();
    this.setupWebSocket();
  }

  setupMiddleware() {
    this.app.use(cors());
    this.app.use(express.json());
    this.app.use(express.static(path.join(__dirname, '../web/build')));
    
    this.app.use((req, res, next) => {
      console.log(`${new Date().toISOString()} ${req.method} ${req.path}`);
      next();
    });
  }

  setupRoutes() {
    this.app.get('/health', (req, res) => {
      res.json({
        status: 'healthy',
        service: 'relay-backend',
        timestamp: new Date().toISOString()
      });
    });

    const api = express.Router();
    
    api.get('/projects', async (req, res) => {
      try {
        const projects = await this.githubManager.listProjects();
        const projectsWithStatus = await this.gitManager.checkRepositoryStatus(projects);
        res.json({ projects: projectsWithStatus });
      } catch (error) {
        console.error('Failed to list projects:', error);
        res.status(500).json({ error: error.message });
      }
    });

    api.post('/projects/:name/select', async (req, res) => {
      try {
        const projectName = req.params.name;
        const repository = await this.githubManager.selectProject(projectName);
        res.json({
          message: 'Repository selected successfully',
          repository: repository
        });
      } catch (error) {
        console.error('Failed to select repository:', error);
        res.status(400).json({ error: error.message });
      }
    });

    api.get('/projects/:name/status', async (req, res) => {
      try {
        const projectName = req.params.name;
        const status = await this.githubManager.getProjectStatus(projectName);
        res.json(status);
      } catch (error) {
        console.error('Failed to get repository status:', error);
        res.status(400).json({ error: error.message });
      }
    });

    // Git configuration endpoints
    api.get('/config/git', async (req, res) => {
      try {
        const config = await this.gitManager.getConfig();
        res.json(config);
      } catch (error) {
        console.error('Failed to get Git configuration:', error);
        res.status(500).json({ error: error.message });
      }
    });

    api.post('/config/git', async (req, res) => {
      try {
        const config = await this.gitManager.setConfig(req.body);
        res.json(config);
      } catch (error) {
        console.error('Failed to set Git configuration:', error);
        res.status(400).json({ error: error.message });
      }
    });

    // Repository management endpoints
    api.post('/repositories/clone', async (req, res) => {
      try {
        const { repository } = req.body;
        if (!repository) {
          return res.status(400).json({ error: 'Repository data required' });
        }
        const result = await this.gitManager.cloneRepository(repository);
        res.json(result);
      } catch (error) {
        console.error('Failed to clone repository:', error);
        res.status(500).json({ error: error.message });
      }
    });

    api.post('/repositories/:name/pull', async (req, res) => {
      try {
        const repositoryName = req.params.name;
        const repository = { name: repositoryName };
        const result = await this.gitManager.pullRepository(repository);
        res.json(result);
      } catch (error) {
        console.error('Failed to pull repository:', error);
        res.status(500).json({ error: error.message });
      }
    });

    api.delete('/repositories/:name', async (req, res) => {
      try {
        const repositoryName = req.params.name;
        const result = await this.gitManager.removeRepository(repositoryName);
        res.json(result);
      } catch (error) {
        console.error('Failed to remove repository:', error);
        res.status(500).json({ error: error.message });
      }
    });

    api.get('/repositories/:name/info', async (req, res) => {
      try {
        const repositoryName = req.params.name;
        const result = await this.gitManager.getRepositoryInfo(repositoryName);
        res.json(result);
      } catch (error) {
        console.error('Failed to get repository info:', error);
        res.status(500).json({ error: error.message });
      }
    });

    this.app.use('/api', api);

    this.app.get('*', (req, res) => {
      res.sendFile(path.join(__dirname, '../web/build/index.html'));
    });
  }

  setupWebSocket() {
    this.io.on('connection', (socket) => {
      const sessionId = this.generateSessionId();
      console.log(`New connection: ${sessionId}`);

      socket.on('disconnect', () => {
        console.log(`Session ended: ${sessionId}`);
        if (this.sessions.has(sessionId)) {
          this.sessions.get(sessionId).close();
          this.sessions.delete(sessionId);
        }
        if (this.terminalSessions.has(sessionId)) {
          const terminalSession = this.terminalSessions.get(sessionId);
          if (terminalSession.ptyProcess) {
            terminalSession.ptyProcess.kill();
          }
          this.terminalSessions.delete(sessionId);
        }
      });

      // Handle voice session initialization
      socket.on('voice_session', () => {
        const session = new VoiceSession(
          sessionId,
          socket,
          this.openaiAPIKey,
          this.githubManager
        );
        this.sessions.set(sessionId, session);
        session.start();
      });

      // Handle terminal session initialization
      socket.on('terminal_init', (data) => {
        this.handleTerminalInit(socket, sessionId, data);
      });

      // Handle terminal commands
      socket.on('terminal_command', (data) => {
        this.handleTerminalCommand(socket, sessionId, data);
      });

      // Handle terminal resize
      socket.on('terminal_resize', (data) => {
        this.handleTerminalResize(sessionId, data);
      });
    });
  }

  handleTerminalInit(socket, sessionId, data) {
    const { workingDirectory } = data;
    console.log(`Initializing terminal session: ${sessionId} in ${workingDirectory}`);

    // Create PTY process
    const ptyProcess = pty.spawn('/bin/bash', [], {
      name: 'xterm-256color',
      cols: 80,
      rows: 24,
      cwd: workingDirectory || process.cwd(),
      env: { 
        ...process.env,
        TERM: 'xterm-256color'
      }
    });

    const terminalSession = {
      ptyProcess,
      workingDirectory: workingDirectory || process.cwd()
    };

    this.terminalSessions.set(sessionId, terminalSession);

    // Handle PTY output
    ptyProcess.onData((data) => {
      socket.emit('terminal_output', { data });
    });

    ptyProcess.onExit((exitCode, signal) => {
      console.log(`Terminal session ${sessionId} exited with code ${exitCode}, signal ${signal}`);
      socket.emit('terminal_closed', { code: exitCode, signal });
      this.terminalSessions.delete(sessionId);
    });

    socket.emit('terminal_ready', { sessionId });
  }

  handleTerminalCommand(socket, sessionId, data) {
    const terminalSession = this.terminalSessions.get(sessionId);
    if (!terminalSession || !terminalSession.ptyProcess) {
      socket.emit('terminal_error', { message: 'Terminal session not found' });
      return;
    }

    const { command } = data;
    terminalSession.ptyProcess.write(command);
  }

  handleTerminalResize(sessionId, data) {
    const terminalSession = this.terminalSessions.get(sessionId);
    if (!terminalSession || !terminalSession.ptyProcess) {
      return;
    }

    const { cols, rows } = data;
    terminalSession.ptyProcess.resize(cols, rows);
  }

  generateSessionId() {
    return `session_${Math.random().toString(36).substring(2, 10)}`;
  }

  start() {
    this.server.listen(this.port, () => {
      console.log(`Voice server starting on port ${this.port}`);
      console.log(`Health check: http://localhost:${this.port}/health`);
      console.log(`WebSocket endpoint: ws://localhost:${this.port}/socket.io/`);
    });
  }

  async stop() {
    for (const session of this.sessions.values()) {
      session.close();
    }
    this.sessions.clear();
    
    return new Promise((resolve) => {
      this.server.close(resolve);
    });
  }
}

const server = new VoiceServer();

process.on('SIGINT', async () => {
  console.log('\\nShutting down voice server...');
  await server.stop();
  process.exit(0);
});

process.on('uncaughtException', (error) => {
  console.error('Uncaught exception:', error);
  process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled rejection at:', promise, 'reason:', reason);
  process.exit(1);
});

server.start();

export { VoiceServer };