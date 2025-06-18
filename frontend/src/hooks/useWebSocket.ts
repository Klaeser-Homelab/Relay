import { useState, useEffect, useRef, useCallback } from 'react';
import { io, Socket } from 'socket.io-client';
import type { VoiceSessionStatus, TranscriptionData, AudioResponseData, FunctionResultData, GeminiAdviceData } from '../types/api';

export interface WebSocketState {
  socket: Socket | null;
  connected: boolean;
  status: VoiceSessionStatus | null;
  transcriptions: TranscriptionData[];
  functionResults: FunctionResultData[];
  geminiAdvice: GeminiAdviceData[];
}

export function useWebSocket() {
  const [state, setState] = useState<WebSocketState>({
    socket: null,
    connected: false,
    status: null,
    transcriptions: [],
    functionResults: [],
    geminiAdvice: []
  });

  const socketRef = useRef<Socket | null>(null);

  const connect = useCallback(() => {
    if (socketRef.current?.connected) {
      return;
    }

    const socket = io('http://localhost:8080', {
      transports: ['websocket'],
      autoConnect: true
    });

    socket.on('connect', () => {
      console.log('WebSocket connected');
      setState(prev => ({ ...prev, connected: true, socket }));
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
      console.log('Transcription:', data);
      setState(prev => ({
        ...prev,
        transcriptions: [...prev.transcriptions, data]
      }));
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
    if (socketRef.current?.connected) {
      socketRef.current.emit('start_recording');
    }
  }, []);

  const stopRecording = useCallback(() => {
    if (socketRef.current?.connected) {
      socketRef.current.emit('stop_recording');
    }
  }, []);

  const sendAudio = useCallback((audioData: string) => {
    if (socketRef.current?.connected) {
      socketRef.current.emit('audio', { audio_data: audioData });
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
    selectProject,
    clearTranscriptions,
    clearFunctionResults,
    clearGeminiAdvice,
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