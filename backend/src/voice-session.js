import OpenAI from 'openai/index.mjs';
import WebSocket from 'ws';
import { v4 as uuidv4 } from 'uuid';

export class VoiceSession {
  constructor(sessionId, socket, openaiAPIKey, githubManager, sessionManager, chatId = null) {
    this.sessionId = sessionId;
    this.socket = socket;
    this.openaiAPIKey = openaiAPIKey;
    this.githubManager = githubManager;
    this.sessionManager = sessionManager;
    this.chatId = chatId;
    
    this.currentProject = null;
    this.isRecording = false;
    this.lastActivity = new Date();
    
    this.openaiWs = null;
    this.openaiClient = new OpenAI({ apiKey: openaiAPIKey });
    this.closed = false;
    this.processedCallIds = new Set(); // Track processed function calls to prevent duplicates
    this.currentTranscript = ''; // Accumulate streaming transcription
    this.recentFunctionCalls = new Map(); // Track recent function calls by content hash
    this.conversationState = {
      transcriptions: [],
      functionResults: [],
      claudeStreamingTexts: [],
      claudeTodoWrites: [],
      repositoryIssues: [],
      claudeSessionId: null // Track Claude session for conversation continuation
    };
    this.snapshotInterval = null;
    
    this.setupSocketHandlers();
    this.startSnapshotInterval();
  }
  
  startSnapshotInterval() {
    // Create snapshots every 5 minutes if we have a chat ID
    if (this.chatId && this.sessionManager) {
      this.snapshotInterval = setInterval(() => {
        this.createSnapshot();
      }, 5 * 60 * 1000); // 5 minutes
    }
  }
  
  async createSnapshot() {
    if (!this.chatId || !this.sessionManager) return;
    
    try {
      await this.sessionManager.createSnapshot(this.chatId, this.conversationState);
      console.log(`Created snapshot for chat ${this.chatId}`);
    } catch (error) {
      console.error('Failed to create snapshot:', error);
    }
  }

  setupSocketHandlers() {
    console.log('🔌 [DEBUG] Setting up socket handlers for session:', this.sessionId);
    this.socket.on('audio', (data) => {
      console.log('🎵 [DEBUG] Audio event received');
      this.handleAudioMessage(data);
    });
    this.socket.on('start_recording', () => {
      console.log('▶️ [DEBUG] start_recording event received');
      this.handleStartRecording();
    });
    this.socket.on('stop_recording', () => {
      console.log('⏹️ [DEBUG] stop_recording event received');
      this.handleStopRecording();
    });
    this.socket.on('select_project', (data) => {
      console.log('📁 [DEBUG] select_project event received:', data);
      this.handleSelectProject(data);
    });
    this.socket.on('test_function', (data) => {
      console.log('🧪 [DEBUG] test_function event received:', data);
      this.handleTestFunction(data);
    });
    this.socket.on('disconnect', () => {
      console.log('🔌 [DEBUG] Socket disconnect event received');
      this.close();
    });
  }

  async start() {
    try {
      // If we have a chat ID, resume the session
      if (this.chatId && this.sessionManager) {
        try {
          const sessionData = await this.sessionManager.resumeSession(this.chatId);
          console.log(`Resumed chat session ${this.chatId}`);
          
          // Send the session data to the frontend
          this.sendMessage({
            type: 'session_resumed',
            data: {
              chatId: this.chatId,
              chat: sessionData.chat,
              snapshot: sessionData.snapshot,
              messages: sessionData.messages
            }
          });
        } catch (error) {
          console.error('Failed to resume session:', error);
        }
      }
      
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
        input_audio_transcription: {
          model: 'whisper-1'
        },
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
      // COMMENTED OUT: Gemini implementation advice - keeping for potential future use
      // {
      //   type: 'function',
      //   name: 'get_implementation_advice',
      //   description: 'CALL THIS FUNCTION when user asks implementation questions starting with: "How should I", "How do I", "How can I", "What\'s the best way to", "I need help implementing", "How to implement", "Help me implement", or any variation asking for implementation guidance. This function gets expert advice from Gemini Flash AI for development questions and coding challenges.',
      //   parameters: {
      //     type: 'object',
      //     properties: {
      //       question: {
      //         type: 'string',
      //         description: 'The exact implementation question the user asked, including any typos or informal language'
      //       },
      //       context: {
      //         type: 'string',
      //         description: 'Additional context about their specific technology stack, requirements, or constraints'
      //       }
      //     },
      //     required: ['question']
      //   }
      // },
      {
        type: 'function',
        name: 'ask_claude',
        description: 'CALL THIS FUNCTION when user asks ANY question, seeks advice, or wants help with implementation, planning, or coding. Routes all questions to Claude for expert assistance.',
        parameters: {
          type: 'object',
          properties: {
            prompt: {
              type: 'string',
              description: 'The planning prompt to send to Claude'
            },
            workingDirectory: {
              type: 'string',
              description: 'The working directory path for the terminal session (optional, defaults to repository path)'
            }
          },
          required: ['prompt']
        }
      }
    ];
  }

  async handleAudioMessage(data) {
    console.log('🎙️ [DEBUG] handleAudioMessage called, data type:', typeof data);
    this.updateActivity();
    
    if (!this.openaiWs || this.openaiWs.readyState !== WebSocket.OPEN) {
      console.warn('🎙️ [DEBUG] OpenAI WebSocket not connected, ignoring audio');
      return;
    }

    try {
      let audioData;
      
      if (Buffer.isBuffer(data)) {
        audioData = data.toString('base64');
        console.log('🎙️ [DEBUG] Converted Buffer to base64, length:', audioData.length);
      } else if (typeof data === 'string') {
        audioData = data;
        console.log('🎙️ [DEBUG] Using string data directly, length:', audioData.length);
      } else if (data.audio_data) {
        audioData = data.audio_data;
        console.log('🎙️ [DEBUG] Extracted audio_data, length:', audioData.length);
      } else {
        console.warn('🎙️ [DEBUG] Invalid audio data format:', data);
        return;
      }

      console.log(`🎙️ [DEBUG] Sending audio to OpenAI: ${audioData.length} chars`);
      
      const audioAppend = {
        type: 'input_audio_buffer.append',
        audio: audioData
      };

      this.openaiWs.send(JSON.stringify(audioAppend));
      console.log('🎙️ [DEBUG] Audio sent to OpenAI successfully');
    } catch (error) {
      console.error('🎙️ [DEBUG] Failed to process audio:', error);
      this.sendStatusMessage('error', 'Failed to process audio', this.currentProject);
    }
  }

  async handleStartRecording() {
    console.log('🎬 [DEBUG] handleStartRecording called');
    this.isRecording = true;
    this.sendStatusMessage('connecting', 'Connecting to voice assistant...', this.currentProject);
    
    try {
      console.log('🎬 [DEBUG] Initializing OpenAI Realtime...');
      await this.initializeOpenAIRealtime();
      console.log('🎬 [DEBUG] OpenAI Realtime initialized successfully');
      this.sendStatusMessage('recording', 'Recording started - speak now', this.currentProject);
    } catch (error) {
      console.error('🎬 [DEBUG] Failed to start recording:', error);
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
    console.log('🤖 [DEBUG] OpenAI message received, type:', message.type);
    switch (message.type) {
      case 'response.audio.delta':
        this.handleAudioDelta(message);
        break;
      case 'response.audio.done':
        console.log('🤖 [DEBUG] OpenAI audio response complete');
        break;
      case 'response.audio_transcript.delta':
        console.log('🤖 [DEBUG] Audio transcript delta received (ignoring - this is OpenAI response)');
        // Don't handle OpenAI's response transcription - we only want user input transcriptions
        break;
      case 'response.audio_transcript.done':
        console.log('🤖 [DEBUG] Audio transcript done received (ignoring - this is OpenAI response)');
        // Don't handle OpenAI's response transcription - we only want user input transcriptions
        break;
      case 'conversation.item.input_audio_transcription.completed':
        console.log('🤖 [DEBUG] Input audio transcription completed');
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
    console.log('📝 [DEBUG] handleTranscription called:', message);
    if (message.transcript) {
      console.log('📝 [DEBUG] Final transcription received, sending to frontend:', message.transcript);
      this.sendMessage({
        type: 'transcription',
        data: {
          text: message.transcript
        }
      });
      
      // Add to conversation state
      this.conversationState.transcriptions.push({
        text: message.transcript,
        timestamp: new Date().toISOString()
      });
      
      // Save transcription to database if we have a chat ID
      if (this.chatId && this.sessionManager) {
        this.sessionManager.addMessage(
          this.chatId,
          'user',
          message.transcript,
          { source: 'voice' }
        ).catch(error => {
          console.error('Failed to save transcription to database:', error);
        });
      }
    } else {
      console.log('📝 [DEBUG] No transcript in message');
    }
  }

  async handleFunctionCall(message) {
    const callId = message.call_id;
    const functionName = message.name;
    const args = JSON.parse(message.arguments || '{}');

    // Prevent duplicate function calls by call_id
    if (this.processedCallIds.has(callId)) {
      console.log(`Skipping duplicate function call by call_id: ${callId}`);
      return;
    }
    this.processedCallIds.add(callId);

    // Additional deduplication by function content hash (to catch different call_ids with same content)
    const contentHash = `${functionName}:${JSON.stringify(args)}`;
    const now = Date.now();
    const recentCallTime = this.recentFunctionCalls.get(contentHash);
    
    // If same function with same args was called within last 5 seconds, skip it
    if (recentCallTime && (now - recentCallTime) < 5000) {
      console.log(`Skipping duplicate function call by content: ${functionName} (${contentHash})`);
      return;
    }
    this.recentFunctionCalls.set(contentHash, now);

    // Clean up old entries (older than 30 seconds)
    for (const [hash, timestamp] of this.recentFunctionCalls.entries()) {
      if (now - timestamp > 30000) {
        this.recentFunctionCalls.delete(hash);
      }
    }

    // Determine function type for logging
    let functionType = 'GitHub';
    if (functionName === 'ask_claude') {
      functionType = 'Claude';
    }
    console.log(`Executing ${functionType} function: ${functionName} with args:`, args);
    this.sendStatusMessage('executing', `Executing ${functionType}: ${functionName}`, this.currentProject);

    // Send immediate "Asking Claude..." feedback for Claude functions
    if (functionName === 'ask_claude') {
      console.log('✓ Sending immediate "Asking Claude..." feedback');
      this.sendMessage({
        type: 'function_result',
        data: {
          function: functionName,
          result: 'Asking Claude...'
        }
      });
    }

    try {
      // Create streaming callback for Claude functions
      const streamCallback = (functionName === 'ask_claude') ? 
        (streamMessage) => this.handleClaudeStreamMessage(streamMessage) : null;

      // Add Claude session ID for continuation if this is a Claude function
      const enhancedArgs = functionName === 'ask_claude' ? 
        { ...args, claudeSessionId: this.conversationState.claudeSessionId } : args;

      const result = await this.githubManager.executeFunction(
        this.currentProject,
        functionName,
        enhancedArgs,
        streamCallback
      );

      if (result.success) {
        this.sendStatusMessage('completed', `Completed: ${functionName}`, this.currentProject);
        
        // COMMENTED OUT: Gemini handling logic - keeping for potential future use
        // if (functionName === 'get_implementation_advice') {
        //   console.log('Sending Gemini advice directly to frontend (bypassing OpenAI)');
        //   
        //   this.sendMessage({
        //     type: 'gemini_advice',
        //     data: {
        //       question: result.data?.question || 'Implementation question',
        //       advice: result.data?.advice || result.message,
        //       repository: result.data?.repository || this.currentProject
        //     }
        //   });
        //   
        //   // Send a simple confirmation to OpenAI without requesting a response
        //   const functionResult = {
        //     type: 'conversation.item.create',
        //     item: {
        //       type: 'function_call_output',
        //       call_id: callId,
        //       output: JSON.stringify({ status: 'completed', message: 'Implementation advice provided' })
        //     }
        //   };
        //   
        //   if (this.openaiWs && this.openaiWs.readyState === WebSocket.OPEN) {
        //     this.openaiWs.send(JSON.stringify(functionResult));
        //     // Deliberately NOT sending response.create to prevent OpenAI commentary
        //   }
        // } else 
        if (functionName === 'ask_claude') {
          // Handle Claude with SDK - streaming already handled, no need for final response
          console.log('=== VOICE SESSION: HANDLING CLAUDE SDK RESPONSE ===');
          console.log('Function result from GitHub manager:', JSON.stringify(result, null, 2));
          
          // Store Claude session ID for conversation continuation
          if (result.data?.sessionId) {
            this.conversationState.claudeSessionId = result.data.sessionId;
            console.log(`✓ Stored Claude session ID for continuation: ${result.data.sessionId}`);
          }
          
          // Note: function_result was already sent immediately when function started
          // Individual sections and todos were streamed progressively
          // No need to send the complete plan again as it would duplicate content
          
          console.log('✓ Claude response streaming completed, all content already sent');
          
          // Send simple confirmation to OpenAI (bypass response generation)
          const functionResult = {
            type: 'conversation.item.create',
            item: {
              type: 'function_call_output',
              call_id: callId,
              output: JSON.stringify({ 
                status: 'completed', 
                message: 'Claude response generated successfully' 
              })
            }
          };
          
          console.log('✓ Sending confirmation to OpenAI');
          
          if (this.openaiWs && this.openaiWs.readyState === WebSocket.OPEN) {
            this.openaiWs.send(JSON.stringify(functionResult));
            console.log('✓ OpenAI confirmation sent');
            // Deliberately NOT sending response.create to prevent OpenAI commentary
          } else {
            console.log('❌ OpenAI WebSocket not available');
          }
          
          console.log('=== VOICE SESSION: CLAUDE SDK PLAN COMPLETE ===');
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
          
          // Add to conversation state
          this.conversationState.functionResults.push(functionResultData);
          
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
      console.log('📝 [DEBUG] Audio transcript delta (not sending to frontend):', message.delta);
      // Accumulate the streaming transcript but don't send individual deltas
      this.currentTranscript += message.delta;
    }
  }

  handleAudioTranscriptDone(message) {
    if (message.transcript) {
      console.log('📝 [DEBUG] Audio transcript complete, sending to frontend:', message.transcript);
      this.sendMessage({
        type: 'transcription',
        data: {
          text: message.transcript
        }
      });
      // Reset accumulated transcript
      this.currentTranscript = '';
    }
  }

  handleClaudeStreamMessage(streamMessage) {
    console.log('✓ Handling Claude stream message:', streamMessage.type);
    
    if (streamMessage.type === 'claude_text') {
      // Send streaming text response as a Claude message
      this.sendMessage({
        type: 'claude_streaming_text',
        data: {
          content: streamMessage.content,
          timestamp: streamMessage.timestamp
        }
      });
      
      // Add to conversation state
      this.conversationState.claudeStreamingTexts.push({
        content: streamMessage.content,
        timestamp: streamMessage.timestamp
      });
      
      // Save Claude response to database
      if (this.chatId && this.sessionManager) {
        this.sessionManager.addMessage(
          this.chatId,
          'claude',
          streamMessage.content,
          { streaming: true, timestamp: streamMessage.timestamp }
        ).catch(error => {
          console.error('Failed to save Claude response to database:', error);
        });
      }
    } else if (streamMessage.type === 'claude_todowrite') {
      // Send TodoWrite tool call as a structured message
      this.sendMessage({
        type: 'claude_todowrite',
        data: {
          todos: streamMessage.content.todos,
          timestamp: streamMessage.timestamp
        }
      });
      
      // Add to conversation state
      this.conversationState.claudeTodoWrites.push({
        todos: streamMessage.content.todos,
        timestamp: streamMessage.timestamp
      });
      
      // Save TodoWrite to database
      if (this.chatId && this.sessionManager) {
        this.sessionManager.addMessage(
          this.chatId,
          'claude',
          'TodoWrite',
          { todos: streamMessage.content.todos, timestamp: streamMessage.timestamp }
        ).catch(error => {
          console.error('Failed to save Claude TodoWrite to database:', error);
        });
      }
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
          tools: this.getToolDefinitions(),
          instructions: `
You are controlling GitHub for repository: ${this.currentProject}
Repository: ${status.fullName}
URL: ${status.url}
Available commands: create_github_issue, update_github_issue, close_github_issue, list_issues, get_repository_info, list_commits, create_pull_request, list_pull_requests, ask_claude

FUNCTION PRIORITY RULES:
1. ALWAYS use ask_claude for ANY question, advice request, implementation help, planning, or general queries
2. Only use GitHub commands when user explicitly wants to perform GitHub actions (create issue, update issue, etc.)
3. Default behavior: Route questions to Claude

IMPORTANT: When ask_claude returns results, you must ONLY present Claude's response exactly as provided. DO NOT add your own analysis, explanations, or additional recommendations. Act as a presenter of Claude's response, not as a competing AI assistant. Simply relay the advice that Claude provided.

When the user gives voice commands, convert them to appropriate function calls.

Examples:
- "Ask claude to make a plan" -> ask_claude
- "How should I implement user authentication?" -> ask_claude
- "What's the best way to add user login?" -> ask_claude
- "How do I add caching to this API?" -> ask_claude
- "Help me implement search functionality" -> ask_claude
- "Why not use Firebase auth instead?" -> ask_claude
- "What are the pros and cons of different approaches?" -> ask_claude
- "Create an issue for adding user authentication" -> create_github_issue
- "Update issue 5 to mark it as completed" -> update_github_issue  
- "Close issue 3" -> close_github_issue
- "Show me the open issues" -> list_issues
- "Get repository information" -> get_repository_info
- "List recent commits" -> list_commits
- "Create a pull request from feature branch to main" -> create_pull_request

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

      let functionType = 'GitHub';
      // COMMENTED OUT: Gemini support
      // if (functionName === 'get_implementation_advice') {
      //   functionType = 'Gemini';
      // } else 
      if (functionName === 'ask_claude') {
        functionType = 'Claude Code SDK';
      }
      
      this.sendStatusMessage('executing', `Testing ${functionType}: ${functionName}`, this.currentProject);

      const result = await this.githubManager.executeFunction(
        this.currentProject,
        functionName,
        args
      );

      console.log(`Developer mode: Function ${functionName} result:`, result);

      // Handle Claude responses specially
      if (functionName === 'ask_claude' && result.success) {
        console.log('✓ Sending Claude plan response to frontend from test function');
        
        const messageData = {
          prompt: result.data?.prompt,
          workingDirectory: result.data?.workingDirectory,
          repository: result.data?.repository,
          plan: result.data?.plan,
          timestamp: result.data?.timestamp
        };
        
        this.sendMessage({
          type: 'claude_plan_response',
          data: messageData
        });
        
        // Also send as function result for Developer Mode display
        const functionResultData = {
          function: functionName,
          result: {
            prompt: result.data?.prompt,
            planLength: result.data?.plan?.length || 0,
            repository: result.data?.repository,
            workingDirectory: result.data?.workingDirectory
          },
          success: result.success
        };

        this.sendMessage({
          type: 'function_result',
          data: functionResultData
        });
      } else {
        // Normal function result handling
        const functionResultData = {
          function: functionName,
          result: result.data || result.message,
          success: result.success
        };

        this.sendMessage({
          type: 'function_result',
          data: functionResultData
        });
      }

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
    
    // Clear snapshot interval
    if (this.snapshotInterval) {
      clearInterval(this.snapshotInterval);
      this.snapshotInterval = null;
    }
    
    // Create final snapshot before closing
    if (this.chatId && this.sessionManager) {
      this.createSnapshot();
    }
    
    if (this.openaiWs) {
      try {
        this.openaiWs.close();
      } catch (error) {
        console.error('Error closing OpenAI WebSocket:', error);
      }
    }
  }
}