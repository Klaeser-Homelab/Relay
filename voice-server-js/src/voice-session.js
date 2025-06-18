import OpenAI from 'openai';
import WebSocket from 'ws';
import { v4 as uuidv4 } from 'uuid';

export class VoiceSession {
  constructor(sessionId, socket, openaiAPIKey, githubManager) {
    this.sessionId = sessionId;
    this.socket = socket;
    this.openaiAPIKey = openaiAPIKey;
    this.githubManager = githubManager;
    
    this.currentProject = null;
    this.isRecording = false;
    this.lastActivity = new Date();
    
    this.openaiWs = null;
    this.openaiClient = new OpenAI({ apiKey: openaiAPIKey });
    this.closed = false;
    this.processedCallIds = new Set(); // Track processed function calls to prevent duplicates
    
    this.setupSocketHandlers();
  }

  setupSocketHandlers() {
    this.socket.on('audio', (data) => this.handleAudioMessage(data));
    this.socket.on('start_recording', () => this.handleStartRecording());
    this.socket.on('stop_recording', () => this.handleStopRecording());
    this.socket.on('select_project', (data) => this.handleSelectProject(data));
    this.socket.on('test_function', (data) => this.handleTestFunction(data));
    this.socket.on('disconnect', () => this.close());
  }

  async start() {
    try {
      this.sendStatusMessage('connected', 'Voice session started. Press record to begin.', '');
      console.log(`Voice session ${this.sessionId} started`);
    } catch (error) {
      console.error('Failed to start voice session:', error);
      this.sendStatusMessage('error', 'Failed to initialize voice session', '');
    }
  }

  async initializeOpenAIRealtime() {
    if (this.openaiWs && this.openaiWs.readyState === WebSocket.OPEN) {
      return; // Already connected
    }

    const wsUrl = 'wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview-2024-10-01';
    
    this.openaiWs = new WebSocket(wsUrl, {
      headers: {
        'Authorization': `Bearer ${this.openaiAPIKey}`,
        'OpenAI-Beta': 'realtime=v1'
      }
    });

    this.openaiWs.on('open', () => {
      console.log('OpenAI Realtime WebSocket connected');
      this.initializeSession();
    });

    this.openaiWs.on('message', (data) => {
      try {
        const message = JSON.parse(data.toString());
        this.handleOpenAIMessage(message);
      } catch (error) {
        console.error('Failed to parse OpenAI message:', error);
      }
    });

    this.openaiWs.on('error', (error) => {
      console.error('OpenAI WebSocket error:', error);
      this.sendStatusMessage('error', 'Voice connection failed', this.currentProject);
    });

    this.openaiWs.on('close', () => {
      console.log('OpenAI WebSocket closed');
      this.openaiWs = null;
    });
  }

  initializeSession() {
    if (!this.openaiWs || this.openaiWs.readyState !== WebSocket.OPEN) {
      return;
    }

    const sessionUpdate = {
      type: 'session.update',
      session: {
        model: 'gpt-4o-realtime-preview-2024-10-01',
        instructions: `You are a voice assistant for GitHub repository management. 
        Help users manage their GitHub repositories through voice commands.
        Convert natural language requests into appropriate function calls.
        Always confirm what actions you're taking and provide helpful feedback.`,
        voice: 'alloy',
        input_audio_format: 'pcm16',
        output_audio_format: 'pcm16',
        tools: this.getToolDefinitions()
      }
    };

    this.openaiWs.send(JSON.stringify(sessionUpdate));
  }

  getToolDefinitions() {
    return [
      {
        type: 'function',
        name: 'create_github_issue',
        description: 'Create a new GitHub issue',
        parameters: {
          type: 'object',
          properties: {
            title: {
              type: 'string',
              description: 'The title of the issue'
            },
            body: {
              type: 'string',
              description: 'The body/description of the issue'
            },
            labels: {
              type: 'array',
              items: { type: 'string' },
              description: 'Labels to add to the issue'
            },
            assignees: {
              type: 'array',
              items: { type: 'string' },
              description: 'GitHub usernames to assign to the issue'
            }
          },
          required: ['title']
        }
      },
      {
        type: 'function',
        name: 'update_github_issue',
        description: 'Update an existing GitHub issue',
        parameters: {
          type: 'object',
          properties: {
            number: {
              type: 'number',
              description: 'The issue number to update'
            },
            title: {
              type: 'string',
              description: 'New title for the issue'
            },
            body: {
              type: 'string',
              description: 'New body for the issue'
            },
            state: {
              type: 'string',
              enum: ['open', 'closed'],
              description: 'New state for the issue'
            }
          },
          required: ['number']
        }
      },
      {
        type: 'function',
        name: 'close_github_issue',
        description: 'Close a GitHub issue',
        parameters: {
          type: 'object',
          properties: {
            number: {
              type: 'number',
              description: 'The issue number to close'
            }
          },
          required: ['number']
        }
      },
      {
        type: 'function',
        name: 'list_issues',
        description: 'List GitHub issues for the current repository',
        parameters: {
          type: 'object',
          properties: {
            state: {
              type: 'string',
              enum: ['open', 'closed', 'all'],
              description: 'Filter issues by state'
            },
            limit: {
              type: 'number',
              description: 'Maximum number of issues to return'
            }
          }
        }
      },
      {
        type: 'function',
        name: 'get_repository_info',
        description: 'Get information about the current repository',
        parameters: {
          type: 'object',
          properties: {}
        }
      },
      {
        type: 'function',
        name: 'list_commits',
        description: 'List recent commits for the repository',
        parameters: {
          type: 'object',
          properties: {
            limit: {
              type: 'number',
              description: 'Number of commits to return'
            },
            branch: {
              type: 'string',
              description: 'Branch to get commits from'
            }
          }
        }
      },
      {
        type: 'function',
        name: 'create_pull_request',
        description: 'Create a new pull request',
        parameters: {
          type: 'object',
          properties: {
            title: {
              type: 'string',
              description: 'The title of the pull request'
            },
            body: {
              type: 'string',
              description: 'The body/description of the pull request'
            },
            head: {
              type: 'string',
              description: 'The branch to merge from'
            },
            base: {
              type: 'string',
              description: 'The branch to merge into (default: main/master)'
            }
          },
          required: ['title', 'head']
        }
      },
      {
        type: 'function',
        name: 'list_pull_requests',
        description: 'List pull requests for the repository',
        parameters: {
          type: 'object',
          properties: {
            state: {
              type: 'string',
              enum: ['open', 'closed', 'all'],
              description: 'Filter pull requests by state'
            },
            limit: {
              type: 'number',
              description: 'Maximum number of pull requests to return'
            }
          }
        }
      },
      {
        type: 'function',
        name: 'get_implementation_advice',
        description: 'CALL THIS FUNCTION when user asks implementation questions starting with: "How should I", "How do I", "How can I", "What\'s the best way to", "I need help implementing", "How to implement", "Help me implement", or any variation asking for implementation guidance. This function gets expert advice from Gemini Flash AI for development questions and coding challenges.',
        parameters: {
          type: 'object',
          properties: {
            question: {
              type: 'string',
              description: 'The exact implementation question the user asked, including any typos or informal language'
            },
            context: {
              type: 'string',
              description: 'Additional context about their specific technology stack, requirements, or constraints'
            }
          },
          required: ['question']
        }
      }
    ];
  }

  async handleAudioMessage(data) {
    this.updateActivity();
    
    if (!this.openaiWs || this.openaiWs.readyState !== WebSocket.OPEN) {
      console.warn('OpenAI WebSocket not connected, ignoring audio');
      return;
    }

    try {
      let audioData;
      
      if (Buffer.isBuffer(data)) {
        audioData = data.toString('base64');
      } else if (typeof data === 'string') {
        audioData = data;
      } else if (data.audio_data) {
        audioData = data.audio_data;
      } else {
        console.warn('Invalid audio data format');
        return;
      }

      console.log(`Received audio data: ${audioData.length} chars`);
      
      const audioAppend = {
        type: 'input_audio_buffer.append',
        audio: audioData
      };

      this.openaiWs.send(JSON.stringify(audioAppend));
    } catch (error) {
      console.error('Failed to process audio:', error);
      this.sendStatusMessage('error', 'Failed to process audio', this.currentProject);
    }
  }

  async handleStartRecording() {
    this.isRecording = true;
    this.sendStatusMessage('connecting', 'Connecting to voice assistant...', this.currentProject);
    
    try {
      await this.initializeOpenAIRealtime();
      this.sendStatusMessage('recording', 'Recording started - speak now', this.currentProject);
    } catch (error) {
      console.error('Failed to start recording:', error);
      this.sendStatusMessage('error', 'Failed to connect to voice assistant', this.currentProject);
      this.isRecording = false;
    }
  }

  async handleStopRecording() {
    this.isRecording = false;
    this.sendStatusMessage('processing', 'Processing voice command...', this.currentProject);
    
    if (this.openaiWs && this.openaiWs.readyState === WebSocket.OPEN) {
      try {
        // Commit the audio buffer and request response
        const commitMessage = {
          type: 'input_audio_buffer.commit'
        };
        this.openaiWs.send(JSON.stringify(commitMessage));

        const responseMessage = {
          type: 'response.create',
          response: {
            modalities: ['text', 'audio']
          }
        };
        this.openaiWs.send(JSON.stringify(responseMessage));
      } catch (error) {
        console.error('Failed to stop recording:', error);
        this.sendStatusMessage('error', 'Failed to process recording', this.currentProject);
      }
    }
  }

  async handleSelectProject(data) {
    const projectName = typeof data === 'string' ? data : data.project;
    
    if (!projectName) {
      this.sendStatusMessage('error', 'Project name required', '');
      return;
    }

    try {
      await this.githubManager.selectProject(projectName);
      this.currentProject = projectName;
      this.sendStatusMessage('project_selected', `Selected repository: ${projectName}`, projectName);
      
      await this.updateOpenAIContext();
    } catch (error) {
      console.error('Failed to select repository:', error);
      this.sendStatusMessage('error', `Failed to select repository: ${error.message}`, '');
    }
  }

  handleOpenAIMessage(message) {
    switch (message.type) {
      case 'response.audio.delta':
        this.handleAudioDelta(message);
        break;
      case 'response.audio.done':
        console.log('OpenAI audio response complete');
        break;
      case 'response.audio_transcript.delta':
        this.handleAudioTranscriptDelta(message);
        break;
      case 'response.audio_transcript.done':
        this.handleAudioTranscriptDone(message);
        break;
      case 'conversation.item.input_audio_transcription.completed':
        this.handleTranscription(message);
        break;
      case 'conversation.item.created':
        console.log('OpenAI conversation item created');
        break;
      case 'input_audio_buffer.speech_started':
        console.log('OpenAI detected speech start');
        this.sendStatusMessage('processing', 'Speech detected...', this.currentProject);
        break;
      case 'input_audio_buffer.committed':
        console.log('OpenAI audio buffer committed');
        break;
      case 'input_audio_buffer.speech_stopped':
        console.log('OpenAI detected speech stop');
        this.sendStatusMessage('processing', 'Processing your request...', this.currentProject);
        break;
      case 'response.created':
        console.log('OpenAI response created');
        break;
      case 'response.done':
        console.log('OpenAI response completed');
        this.sendStatusMessage('connected', 'Ready for next command', this.currentProject);
        break;
      case 'response.output_item.added':
        console.log('OpenAI output item added');
        break;
      case 'response.output_item.done':
        console.log('OpenAI output item complete');
        break;
      case 'response.content_part.added':
        console.log('OpenAI content part added');
        break;
      case 'response.content_part.done':
        console.log('OpenAI content part complete');
        break;
      case 'response.function_call_arguments.delta':
        // Function call arguments being built - we'll wait for done
        break;
      case 'response.function_call_arguments.done':
        this.handleFunctionCall(message);
        break;
      case 'rate_limits.updated':
        console.log('OpenAI rate limits updated');
        break;
      case 'error':
        this.handleOpenAIError(message);
        break;
      case 'session.created':
        console.log('OpenAI session created');
        break;
      case 'session.updated':
        console.log('OpenAI session updated');
        break;
      default:
        console.log('Unhandled OpenAI message type:', message.type);
    }
  }

  handleAudioDelta(message) {
    if (message.delta) {
      this.sendMessage({
        type: 'audio_response',
        data: {
          audio_data: message.delta
        }
      });
    }
  }

  handleTranscription(message) {
    if (message.transcript) {
      this.sendMessage({
        type: 'transcription',
        data: {
          text: message.transcript
        }
      });
    }
  }

  async handleFunctionCall(message) {
    const callId = message.call_id;
    const functionName = message.name;
    const args = JSON.parse(message.arguments || '{}');

    // Prevent duplicate function calls
    if (this.processedCallIds.has(callId)) {
      console.log(`Skipping duplicate function call: ${callId}`);
      return;
    }
    this.processedCallIds.add(callId);

    const functionType = functionName === 'get_implementation_advice' ? 'Gemini' : 'GitHub';
    console.log(`Executing ${functionType} function: ${functionName} with args:`, args);
    this.sendStatusMessage('executing', `Executing ${functionType}: ${functionName}`, this.currentProject);

    try {
      const result = await this.githubManager.executeFunction(
        this.currentProject,
        functionName,
        args
      );

      if (result.success) {
        this.sendStatusMessage('completed', `Completed: ${functionName}`, this.currentProject);
        
        // Handle Gemini functions differently - bypass OpenAI entirely
        if (functionName === 'get_implementation_advice') {
          console.log('Sending Gemini advice directly to frontend (bypassing OpenAI)');
          
          this.sendMessage({
            type: 'gemini_advice',
            data: {
              question: result.data?.question || 'Implementation question',
              advice: result.data?.advice || result.message,
              repository: result.data?.repository || this.currentProject
            }
          });
          
          // Send a simple confirmation to OpenAI without requesting a response
          const functionResult = {
            type: 'conversation.item.create',
            item: {
              type: 'function_call_output',
              call_id: callId,
              output: JSON.stringify({ status: 'completed', message: 'Implementation advice provided' })
            }
          };
          
          if (this.openaiWs && this.openaiWs.readyState === WebSocket.OPEN) {
            this.openaiWs.send(JSON.stringify(functionResult));
            // Deliberately NOT sending response.create to prevent OpenAI commentary
          }
        } else {
          // Normal GitHub function handling - keep OpenAI in the loop
          const functionResultData = {
            function: functionName,
            result: result.data || result.message
          };
          
          console.log('Sending function result to frontend:', functionResultData);
          
          this.sendMessage({
            type: 'function_result',
            data: functionResultData
          });
          
          // Send function result back to OpenAI and request response
          const functionResult = {
            type: 'conversation.item.create',
            item: {
              type: 'function_call_output',
              call_id: callId,
              output: JSON.stringify(result.data || result.message)
            }
          };
          
          if (this.openaiWs && this.openaiWs.readyState === WebSocket.OPEN) {
            this.openaiWs.send(JSON.stringify(functionResult));
            
            // Request response after function result
            const responseMessage = {
              type: 'response.create',
              response: {
                modalities: ['text', 'audio']
              }
            };
            this.openaiWs.send(JSON.stringify(responseMessage));
          }
        }
      } else {
        this.sendStatusMessage('error', `Failed: ${result.message}`, this.currentProject);
        
        // Send error back to OpenAI
        const functionResult = {
          type: 'conversation.item.create',
          item: {
            type: 'function_call_output',
            call_id: callId,
            output: `Error: ${result.message}`
          }
        };
        
        if (this.openaiWs && this.openaiWs.readyState === WebSocket.OPEN) {
          this.openaiWs.send(JSON.stringify(functionResult));
        }
      }
    } catch (error) {
      console.error('Function execution error:', error);
      this.sendStatusMessage('error', `Failed to execute ${functionName}: ${error.message}`, this.currentProject);
    }
  }

  handleAudioTranscriptDelta(message) {
    if (message.delta) {
      console.log('Audio transcript delta:', message.delta);
      this.sendMessage({
        type: 'transcription',
        data: {
          text: message.delta
        }
      });
    }
  }

  handleAudioTranscriptDone(message) {
    if (message.transcript) {
      console.log('Audio transcript complete:', message.transcript);
      this.sendMessage({
        type: 'transcription',
        data: {
          text: message.transcript
        }
      });
    }
  }

  handleOpenAIError(message) {
    console.error('OpenAI error:', message.error);
    this.sendStatusMessage('error', message.error.message || 'OpenAI error', this.currentProject);
  }

  async updateOpenAIContext() {
    if (!this.currentProject || !this.openaiWs || this.openaiWs.readyState !== WebSocket.OPEN) {
      return;
    }

    try {
      const status = await this.githubManager.getProjectStatus(this.currentProject);
      
      const contextUpdate = {
        type: 'session.update',
        session: {
          instructions: `
You are controlling GitHub for repository: ${this.currentProject}
Repository: ${status.fullName}
URL: ${status.url}
Available commands: create_github_issue, update_github_issue, close_github_issue, list_issues, get_repository_info, list_commits, create_pull_request, list_pull_requests, get_implementation_advice

FUNCTION PRIORITY RULES:
1. ALWAYS use get_implementation_advice for ANY question asking HOW to implement, build, or do something
2. Only use create_github_issue when user explicitly wants to CREATE an issue, not when asking for implementation help
3. Implementation questions should NEVER create issues - they should get advice first

IMPORTANT: When get_implementation_advice returns results, you must ONLY present the Gemini advice exactly as provided. DO NOT add your own analysis, explanations, or additional recommendations. Act as a presenter of Gemini's response, not as a competing AI assistant. Simply relay the implementation advice that Gemini provided.

When the user gives voice commands, convert them to appropriate function calls.

Examples:
- "Create an issue for adding user authentication" -> create_github_issue
- "Update issue 5 to mark it as completed" -> update_github_issue  
- "Close issue 3" -> close_github_issue
- "Show me the open issues" -> list_issues
- "Get repository information" -> get_repository_info
- "List recent commits" -> list_commits
- "Create a pull request from feature branch to main" -> create_pull_request

IMPLEMENTATION ADVICE EXAMPLES (ALWAYS use get_implementation_advice):
- "How should I implement user authentication?" -> get_implementation_advice
- "How do I add caching to this API?" -> get_implementation_advice
- "How should I implement authentication for my React + Vue website?" -> get_implementation_advice
- "How should I implement authentication fro my React + Vue website?" -> get_implementation_advice (with typo)
- "What's the best way to add user login?" -> get_implementation_advice
- "How can I implement real-time notifications?" -> get_implementation_advice
- "I need help implementing a payment system" -> get_implementation_advice
- "How to implement file uploads?" -> get_implementation_advice
- "Help me implement search functionality" -> get_implementation_advice
- "How do I build a chat feature?" -> get_implementation_advice
- "What's the best approach for user roles?" -> get_implementation_advice
- "How can I add database caching?" -> get_implementation_advice
- "I want to implement OAuth, how should I do it?" -> get_implementation_advice
- "How should I structure my API endpoints?" -> get_implementation_advice

Always be helpful and confirm what actions you're taking.
          `
        }
      };

      this.openaiWs.send(JSON.stringify(contextUpdate));
    } catch (error) {
      console.error('Failed to update OpenAI context:', error);
    }
  }

  sendMessage(message) {
    if (this.closed) return;
    
    try {
      this.socket.emit(message.type, message.data);
    } catch (error) {
      console.error('Failed to send socket message:', error);
    }
  }

  sendStatusMessage(status, message, project) {
    this.sendMessage({
      type: 'status',
      data: {
        type: 'status',
        status,
        message,
        project
      }
    });
  }

  async handleTestFunction(data) {
    const { project, function: functionName, args } = data;
    
    console.log(`Developer mode: Testing function ${functionName} with args:`, args);
    
    try {
      if (project && project !== this.currentProject) {
        await this.handleSelectProject(project);
      }

      const functionType = functionName === 'get_implementation_advice' ? 'Gemini' : 'GitHub';
      this.sendStatusMessage('executing', `Testing ${functionType}: ${functionName}`, this.currentProject);

      const result = await this.githubManager.executeFunction(
        this.currentProject,
        functionName,
        args
      );

      console.log(`Developer mode: Function ${functionName} result:`, result);

      const functionResultData = {
        function: functionName,
        result: result.data || result.message,
        success: result.success
      };

      this.sendMessage({
        type: 'function_result',
        data: functionResultData
      });

      this.sendStatusMessage(
        result.success ? 'completed' : 'error',
        `Test ${result.success ? 'completed' : 'failed'}: ${functionName}`,
        this.currentProject
      );

    } catch (error) {
      console.error(`Developer mode: Function ${functionName} error:`, error);
      this.sendStatusMessage('error', `Test failed: ${functionName} - ${error.message}`, this.currentProject);
    }
  }

  updateActivity() {
    this.lastActivity = new Date();
  }

  close() {
    if (this.closed) return;
    
    this.closed = true;
    this.isRecording = false;
    
    if (this.openaiWs) {
      try {
        this.openaiWs.close();
      } catch (error) {
        console.error('Error closing OpenAI WebSocket:', error);
      }
    }
  }
}