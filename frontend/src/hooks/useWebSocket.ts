import { useState, useEffect, useRef, useCallback } from 'react';
import { io, Socket } from 'socket.io-client';
import type { VoiceSessionStatus, TranscriptionData, AudioResponseData, FunctionResultData, GeminiAdviceData, ClaudePlanRequestData, ClaudePlanResponseData, ClaudeStreamingTextData, ClaudeTodoWriteData } from '../types/api';

export interface WebSocketState {
  socket: Socket | null;
  connected: boolean;
  status: VoiceSessionStatus | null;
  transcriptions: TranscriptionData[];
  functionResults: FunctionResultData[];
  geminiAdvice: GeminiAdviceData[];
  claudePlanRequests: ClaudePlanRequestData[];
  claudePlanResponses: ClaudePlanResponseData[];
  claudeStreamingTexts: ClaudeStreamingTextData[];
  claudeTodoWrites: ClaudeTodoWriteData[];
}

export function useWebSocket() {
  const [state, setState] = useState<WebSocketState>({
    socket: null,
    connected: false,
    status: null,
    transcriptions: [],
    functionResults: [],
    geminiAdvice: [],
    claudePlanRequests: [],
    claudePlanResponses: [],
    claudeStreamingTexts: [],
    claudeTodoWrites: []
  });

  const socketRef = useRef<Socket | null>(null);
  const claudeProcessingStartIndexRef = useRef<number>(-1);

  const connect = useCallback((chatId?: string) => {
    if (socketRef.current?.connected) {
      return;
    }

    const socket = io('http://localhost:8080', {
      transports: ['websocket'],
      autoConnect: true
    });

    socket.on('connect', () => {
      console.log('🔌 [DEBUG] WebSocket connected, initializing voice session...');
      setState(prev => ({ ...prev, connected: true, socket }));
      
      // Initialize voice session on connection with optional chat ID
      socket.emit('voice_session', { chatId });
      console.log('🔌 [DEBUG] voice_session event sent to backend with chatId:', chatId);
    });

    socket.on('disconnect', () => {
      console.log('WebSocket disconnected');
      setState(prev => ({ ...prev, connected: false }));
    });

    socket.on('status', (data: VoiceSessionStatus) => {
      console.log('Status update:', data);
      setState(prev => ({ ...prev, status: data }));
    });

    socket.on('transcription', (data: TranscriptionData) => {
      console.log('📝 [DEBUG] Transcription received in frontend:', data);
      setState(prev => {
        const newTranscriptions = [...prev.transcriptions, data];
        console.log('📝 [DEBUG] Updated transcriptions count:', newTranscriptions.length);
        return { ...prev, transcriptions: newTranscriptions };
      });
    });

    socket.on('audio_response', (data: AudioResponseData) => {
      console.log('Audio response received');
      // Handle audio playback here
      playAudioData(data.audio_data);
    });

    socket.on('function_result', (data: FunctionResultData) => {
      console.log('Function result received in frontend:', data);
      setState(prev => {
        const newResults = [...prev.functionResults, data];
        console.log('Updated function results:', newResults);
        return {
          ...prev,
          functionResults: newResults
        };
      });
    });

    socket.on('gemini_advice', (data: GeminiAdviceData) => {
      console.log('Gemini advice received in frontend:', data);
      setState(prev => {
        const newAdvice = [...prev.geminiAdvice, data];
        console.log('Updated Gemini advice:', newAdvice);
        return {
          ...prev,
          geminiAdvice: newAdvice
        };
      });
    });

    socket.on('claude_plan_request', (data: ClaudePlanRequestData) => {
      console.log('=== FRONTEND: Claude plan request received ===');
      console.log('Request data:', JSON.stringify(data, null, 2));
      setState(prev => {
        const newRequests = [...prev.claudePlanRequests, data];
        console.log('✓ Added to state, total requests:', newRequests.length);
        return {
          ...prev,
          claudePlanRequests: newRequests
        };
      });
    });

    socket.on('claude_plan_response', (data: ClaudePlanResponseData) => {
      console.log('=== FRONTEND: Claude plan response received ===');
      console.log('Response data:', JSON.stringify({ ...data, plan: `${data.plan?.length || 0} characters` }, null, 2));
      setState(prev => {
        const newResponses = [...prev.claudePlanResponses, data];
        console.log('✓ Added plan to state, total responses:', newResponses.length);
        return {
          ...prev,
          claudePlanResponses: newResponses
        };
      });
    });

    socket.on('claude_streaming_text', (data: ClaudeStreamingTextData) => {
      console.log('✓ Claude streaming text received:', data.content.substring(0, 100) + '...');
      setState(prev => {
        // Track the start of Claude processing
        if (claudeProcessingStartIndexRef.current === -1) {
          claudeProcessingStartIndexRef.current = prev.claudeStreamingTexts.length;
          console.log('📍 Claude processing started at index:', claudeProcessingStartIndexRef.current);
        }
        
        const newTexts = [...prev.claudeStreamingTexts, data];
        return {
          ...prev,
          claudeStreamingTexts: newTexts
        };
      });
    });

    socket.on('claude_todowrite', (data: ClaudeTodoWriteData) => {
      console.log('✓ Claude TodoWrite received:', data.todos.length, 'todos');
      setState(prev => {
        const newTodoWrites = [...prev.claudeTodoWrites, data];
        return {
          ...prev,
          claudeTodoWrites: newTodoWrites
        };
      });
    });

    socket.on('session_resumed', (data: any) => {
      console.log('Session resumed:', data);
      // Restore conversation state from the resumed session
      if (data.snapshot) {
        setState(prev => ({
          ...prev,
          transcriptions: data.snapshot.transcriptions || [],
          functionResults: data.snapshot.function_results || [],
          claudeStreamingTexts: data.snapshot.claude_streaming_texts || [],
          claudeTodoWrites: data.snapshot.claude_todo_writes || []
        }));
      }
    });

    socket.on('processing_interrupted', (data: any) => {
      console.log('Processing interrupted:', data);
      
      setState(prev => {
        let updatedClaudeTexts = prev.claudeStreamingTexts;
        let updatedTodoWrites = prev.claudeTodoWrites;
        
        // Remove Claude messages that were added after processing started
        if (claudeProcessingStartIndexRef.current !== -1) {
          console.log('🗑️ Removing Claude messages from index:', claudeProcessingStartIndexRef.current);
          updatedClaudeTexts = prev.claudeStreamingTexts.slice(0, claudeProcessingStartIndexRef.current);
          
          // Also remove any TodoWrites that were added during this session
          // We'll keep TodoWrites that existed before the interruption
          const interruptTime = Date.now() - 60000; // Messages from last minute are likely from this session
          updatedTodoWrites = prev.claudeTodoWrites.filter(todo => {
            const todoTime = new Date(todo.timestamp).getTime();
            return todoTime < interruptTime;
          });
        }
        
        // Add only the interrupt message
        updatedClaudeTexts = [...updatedClaudeTexts, {
          content: 'Interrupted by user',
          timestamp: new Date().toISOString()
        }];
        
        // Reset the processing start index
        claudeProcessingStartIndexRef.current = -1;
        
        return {
          ...prev,
          claudeStreamingTexts: updatedClaudeTexts,
          claudeTodoWrites: updatedTodoWrites
        };
      });
    });

    socket.on('connect_error', (error) => {
      console.error('WebSocket connection error:', error);
    });

    socketRef.current = socket;
  }, []);

  const disconnect = useCallback(() => {
    if (socketRef.current) {
      socketRef.current.disconnect();
      socketRef.current = null;
      setState(prev => ({ ...prev, socket: null, connected: false }));
    }
  }, []);

  const startRecording = useCallback(() => {
    console.log('🎬 [DEBUG] WebSocket startRecording called, connected:', socketRef.current?.connected);
    if (socketRef.current?.connected) {
      console.log('🎬 [DEBUG] Emitting start_recording to backend');
      socketRef.current.emit('start_recording');
    } else {
      console.log('🎬 [DEBUG] Socket not connected, cannot start recording');
    }
  }, []);

  const stopRecording = useCallback(() => {
    console.log('🛑 [DEBUG] WebSocket stopRecording called, connected:', socketRef.current?.connected);
    if (socketRef.current?.connected) {
      console.log('🛑 [DEBUG] Emitting stop_recording to backend');
      socketRef.current.emit('stop_recording');
    } else {
      console.log('🛑 [DEBUG] Socket not connected, cannot stop recording');
    }
  }, []);

  const sendAudio = useCallback((audioData: string) => {
    console.log('📡 [DEBUG] sendAudio called, socket connected:', socketRef.current?.connected, 'audio length:', audioData.length);
    if (socketRef.current?.connected) {
      socketRef.current.emit('audio', { audio_data: audioData });
      console.log('📡 [DEBUG] Audio data sent to backend');
    } else {
      console.log('📡 [DEBUG] Socket not connected, audio not sent');
    }
  }, []);

  const sendTextMessage = useCallback((text: string) => {
    console.log('💬 [DEBUG] sendTextMessage called, socket connected:', socketRef.current?.connected, 'text:', text);
    if (socketRef.current?.connected) {
      socketRef.current.emit('text_message', { text });
      console.log('💬 [DEBUG] Text message sent to backend');
    } else {
      console.log('💬 [DEBUG] Socket not connected, text message not sent');
    }
  }, []);

  const selectProject = useCallback((projectName: string) => {
    if (socketRef.current?.connected) {
      socketRef.current.emit('select_project', { project: projectName });
    }
  }, []);

  const clearTranscriptions = useCallback(() => {
    setState(prev => ({ ...prev, transcriptions: [] }));
  }, []);

  const clearFunctionResults = useCallback(() => {
    setState(prev => ({ ...prev, functionResults: [] }));
  }, []);

  const clearGeminiAdvice = useCallback(() => {
    setState(prev => ({ ...prev, geminiAdvice: [] }));
  }, []);

  const clearClaudePlanRequests = useCallback(() => {
    setState(prev => ({ ...prev, claudePlanRequests: [] }));
  }, []);

  const clearClaudePlanResponses = useCallback(() => {
    setState(prev => ({ ...prev, claudePlanResponses: [] }));
  }, []);

  const clearClaudeStreamingTexts = useCallback(() => {
    setState(prev => ({ ...prev, claudeStreamingTexts: [] }));
  }, []);

  const clearClaudeTodoWrites = useCallback(() => {
    setState(prev => ({ ...prev, claudeTodoWrites: [] }));
  }, []);

  const resetClaudeProcessingIndex = useCallback(() => {
    console.log('🔄 Resetting Claude processing index');
    claudeProcessingStartIndexRef.current = -1;
  }, []);

  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    ...state,
    connect,
    disconnect,
    startRecording,
    stopRecording,
    sendAudio,
    sendTextMessage,
    selectProject,
    clearTranscriptions,
    clearFunctionResults,
    clearGeminiAdvice,
    clearClaudePlanRequests,
    clearClaudePlanResponses,
    clearClaudeStreamingTexts,
    clearClaudeTodoWrites,
    resetClaudeProcessingIndex,
    socket: socketRef.current
  };
}

// Helper function to play audio data
async function playAudioData(base64Audio: string) {
  try {
    // Decode base64 audio data
    const audioData = atob(base64Audio);
    const arrayBuffer = new ArrayBuffer(audioData.length);
    const uint8Array = new Uint8Array(arrayBuffer);
    
    for (let i = 0; i < audioData.length; i++) {
      uint8Array[i] = audioData.charCodeAt(i);
    }

    // Create audio context and play the audio
    const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
    const audioBuffer = await audioContext.decodeAudioData(arrayBuffer);
    
    const source = audioContext.createBufferSource();
    source.buffer = audioBuffer;
    source.connect(audioContext.destination);
    source.start();
  } catch (error) {
    console.error('Failed to play audio:', error);
  }
}