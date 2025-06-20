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
import { initializeDatabase } from './database.js';
import { sessionManager } from './session-manager.js';

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
console.log(`ANTHROPIC_API_KEY: ${obscure(process.env.ANTHROPIC_API_KEY)}`);
console.log(`GEMINI_API_KEY: ${obscure(process.env.GEMINI_API_KEY)}`);
console.log(`GH_TOKEN: ${obscure(process.env.GH_TOKEN)}`);
console.log(`DEFAULT_CODE_PATH: ${process.env.DEFAULT_CODE_PATH || '(not set, using default)'}`);

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
    this.databaseInitialized = false;
    
    if (!this.openaiAPIKey) {
      throw new Error('OPENAI_API_KEY environment variable is required');
    }
    
    this.githubManager = new GitHubManager();
    this.gitManager = new GitManager();
    this.setupMiddleware();
    this.setupRoutes();
    this.setupWebSocket();
    
    // Initialize database
    this.initializeDatabase();
  }
  
  async initializeDatabase() {
    try {
      this.databaseInitialized = await initializeDatabase();
      if (!this.databaseInitialized) {
        console.warn('Database initialization failed - session persistence will be disabled');
      }
    } catch (error) {
      console.error('Database initialization error:', error);
      this.databaseInitialized = false;
    }
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

    // Execute GitHub function
    api.post('/github/execute', async (req, res) => {
      try {
        const { projectName, functionName, args } = req.body;
        
        if (!functionName) {
          return res.status(400).json({ 
            success: false, 
            error: 'Function name is required' 
          });
        }

        const result = await this.githubManager.executeFunction(projectName, functionName, args);
        res.json(result);
      } catch (error) {
        console.error('Failed to execute GitHub function:', error);
        res.status(500).json({ 
          success: false, 
          error: error.message 
        });
      }
    });

    // Session management endpoints
    api.post('/sessions', async (req, res) => {
      try {
        const { repositoryName, repositoryFullName, title } = req.body;
        
        if (!repositoryName || !repositoryFullName) {
          return res.status(400).json({ 
            success: false, 
            error: 'Repository name and full name are required' 
          });
        }

        const session = await sessionManager.createSession(
          repositoryName, 
          repositoryFullName,
          title
        );

        res.json({
          success: true,
          session: session
        });
      } catch (error) {
        console.error('Failed to create session:', error);
        res.status(500).json({ 
          success: false, 
          error: error.message 
        });
      }
    });

    api.get('/sessions', async (req, res) => {
      try {
        const { repository } = req.query;
        const sessions = await sessionManager.listSessions(repository);
        
        res.json({
          success: true,
          sessions: sessions
        });
      } catch (error) {
        console.error('Failed to list sessions:', error);
        res.status(500).json({ 
          success: false, 
          error: error.message 
        });
      }
    });

    api.get('/sessions/:id', async (req, res) => {
      try {
        const session = await sessionManager.getSession(req.params.id);
        
        if (!session) {
          return res.status(404).json({ 
            success: false, 
            error: 'Session not found' 
          });
        }

        res.json({
          success: true,
          session: session
        });
      } catch (error) {
        console.error('Failed to get session:', error);
        res.status(500).json({ 
          success: false, 
          error: error.message 
        });
      }
    });

    api.put('/sessions/:id', async (req, res) => {
      try {
        const session = await sessionManager.updateSession(
          req.params.id,
          req.body
        );

        res.json({
          success: true,
          session: session
        });
      } catch (error) {
        console.error('Failed to update session:', error);
        res.status(500).json({ 
          success: false, 
          error: error.message 
        });
      }
    });

    api.post('/sessions/:id/resume', async (req, res) => {
      try {
        const sessionData = await sessionManager.resumeSession(req.params.id);
        
        res.json({
          success: true,
          ...sessionData
        });
      } catch (error) {
        console.error('Failed to resume session:', error);
        res.status(500).json({ 
          success: false, 
          error: error.message 
        });
      }
    });

    api.delete('/sessions/:id', async (req, res) => {
      try {
        const deleted = await sessionManager.deleteSession(req.params.id);
        
        res.json({
          success: deleted,
          message: deleted ? 'Session deleted' : 'Session not found'
        });
      } catch (error) {
        console.error('Failed to delete session:', error);
        res.status(500).json({ 
          success: false, 
          error: error.message 
        });
      }
    });

    // Save plan as GitHub issue
    api.post('/plans/save-as-issue', async (req, res) => {
      try {
        const { plan, title, repository } = req.body;
        
        if (!plan || !title) {
          return res.status(400).json({ 
            success: false, 
            error: 'Plan content and title are required' 
          });
        }

        // Select the repository if provided
        if (repository) {
          await this.githubManager.selectProject(repository);
        }

        // Create the GitHub issue
        const result = await this.githubManager.createGitHubIssue({
          title: title,
          body: plan,
          labels: ['plan', 'claude-generated']
        });

        if (result.success) {
          console.log(`Plan saved as GitHub issue: ${result.data.url}`);
          res.json({
            success: true,
            message: result.message,
            issueUrl: result.data.url,
            issueNumber: result.data.number
          });
        } else {
          res.status(500).json({
            success: false,
            error: result.message
          });
        }
      } catch (error) {
        console.error('Failed to save plan as GitHub issue:', error);
        res.status(500).json({ 
          success: false, 
          error: error.message 
        });
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
          // Remove session association
          sessionManager.removeActiveSession(sessionId);
        }
        if (this.terminalSessions.has(sessionId)) {
          const terminalSession = this.terminalSessions.get(sessionId);
          if (terminalSession.ptyProcess) {
            terminalSession.ptyProcess.kill();
          }
          this.terminalSessions.delete(sessionId);
        }
      });

      // Handle voice session initialization with optional chat ID
      socket.on('voice_session', async (data) => {
        console.log('🎙️ [DEBUG] Voice session requested for session:', sessionId);
        const chatId = data?.chatId || null;
        
        const session = new VoiceSession(
          sessionId,
          socket,
          this.openaiAPIKey,
          this.githubManager,
          sessionManager,
          chatId
        );
        this.sessions.set(sessionId, session);
        
        // Associate the socket session with the chat if provided
        if (chatId) {
          sessionManager.setActiveSession(sessionId, chatId);
        }
        
        console.log('🎙️ [DEBUG] Voice session created, starting...');
        await session.start();
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
    console.log('=== SERVER: Terminal command received ===');
    console.log('Session ID:', sessionId);
    console.log('Command data:', JSON.stringify(data, null, 2));
    
    const terminalSession = this.terminalSessions.get(sessionId);
    if (!terminalSession || !terminalSession.ptyProcess) {
      console.log('❌ Terminal session not found for:', sessionId);
      socket.emit('terminal_error', { message: 'Terminal session not found' });
      return;
    }

    const { command } = data;
    console.log('✓ Terminal session found, writing command to PTY');
    console.log('Command being written:', JSON.stringify(command));
    console.log('Working directory:', terminalSession.workingDirectory);
    
    try {
      terminalSession.ptyProcess.write(command);
      console.log('✓ Command written to PTY successfully');
    } catch (error) {
      console.error('❌ Error writing to PTY:', error);
      socket.emit('terminal_error', { message: `Failed to execute command: ${error.message}` });
    }
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