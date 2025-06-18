import { useState, useRef, useCallback } from 'react';

export interface AudioRecordingState {
  isRecording: boolean;
  isSupported: boolean;
  audioLevel: number;
}

export function useAudioRecording() {
  const [state, setState] = useState<AudioRecordingState>({
    isRecording: false,
    isSupported: typeof navigator !== 'undefined' && !!navigator.mediaDevices?.getUserMedia,
    audioLevel: 0
  });

  const audioContextRef = useRef<AudioContext | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const animationFrameRef = useRef<number | null>(null);
  const processorRef = useRef<ScriptProcessorNode | null>(null);

  const onAudioData = useRef<((audioData: string) => void) | null>(null);

  const setOnAudioData = useCallback((callback: (audioData: string) => void) => {
    onAudioData.current = callback;
  }, []);

  const startRecording = useCallback(async () => {
    if (!state.isSupported || state.isRecording) {
      return false;
    }

    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          sampleRate: 24000,
          channelCount: 1,
          echoCancellation: true,
          noiseSuppression: true
        }
      });

      streamRef.current = stream;

      // Set up audio context for processing
      audioContextRef.current = new (window.AudioContext || (window as any).webkitAudioContext)({
        sampleRate: 24000
      });
      
      const source = audioContextRef.current.createMediaStreamSource(stream);
      
      // Set up analyser for level monitoring
      analyserRef.current = audioContextRef.current.createAnalyser();
      analyserRef.current.fftSize = 256;
      source.connect(analyserRef.current);

      // Set up script processor for PCM16 conversion
      const bufferSize = 4096;
      processorRef.current = audioContextRef.current.createScriptProcessor(bufferSize, 1, 1);
      
      processorRef.current.onaudioprocess = (event) => {
        if (!onAudioData.current) return;
        
        const inputBuffer = event.inputBuffer;
        const inputData = inputBuffer.getChannelData(0);
        
        // Convert float32 to PCM16
        const pcm16Buffer = new Int16Array(inputData.length);
        for (let i = 0; i < inputData.length; i++) {
          // Clamp to [-1, 1] and convert to 16-bit signed integer
          const sample = Math.max(-1, Math.min(1, inputData[i]));
          pcm16Buffer[i] = sample * 0x7FFF;
        }
        
        // Convert to base64
        const uint8Array = new Uint8Array(pcm16Buffer.buffer);
        const base64 = btoa(String.fromCharCode(...uint8Array));
        onAudioData.current(base64);
      };

      source.connect(processorRef.current);
      processorRef.current.connect(audioContextRef.current.destination);

      setState(prev => ({ ...prev, isRecording: true }));
      
      // Start audio level monitoring
      monitorAudioLevel();

      return true;
    } catch (error) {
      console.error('Failed to start recording:', error);
      return false;
    }
  }, [state.isSupported, state.isRecording]);

  const stopRecording = useCallback(() => {
    if (!state.isRecording) {
      return;
    }

    if (processorRef.current) {
      processorRef.current.disconnect();
      processorRef.current = null;
    }

    if (streamRef.current) {
      streamRef.current.getTracks().forEach(track => track.stop());
      streamRef.current = null;
    }

    if (audioContextRef.current) {
      audioContextRef.current.close();
      audioContextRef.current = null;
    }

    if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current);
      animationFrameRef.current = null;
    }

    setState(prev => ({ ...prev, isRecording: false, audioLevel: 0 }));
  }, [state.isRecording]);

  const monitorAudioLevel = useCallback(() => {
    if (!analyserRef.current) {
      return;
    }

    const dataArray = new Uint8Array(analyserRef.current.frequencyBinCount);
    
    const updateLevel = () => {
      if (!analyserRef.current || !state.isRecording) {
        return;
      }

      analyserRef.current.getByteFrequencyData(dataArray);
      
      // Calculate average volume
      const average = dataArray.reduce((sum, value) => sum + value, 0) / dataArray.length;
      const normalizedLevel = Math.min(average / 128, 1); // Normalize to 0-1

      setState(prev => ({ ...prev, audioLevel: normalizedLevel }));
      
      animationFrameRef.current = requestAnimationFrame(updateLevel);
    };

    updateLevel();
  }, [state.isRecording]);

  return {
    ...state,
    startRecording,
    stopRecording,
    setOnAudioData
  };
}