import React from 'react';
import { Wifi, WifiOff, Mic, MicOff, Loader } from 'lucide-react';
import type { VoiceSessionStatus } from '../types/api';

interface StatusDisplayProps {
  connected: boolean;
  status: VoiceSessionStatus | null;
  audioLevel: number;
  isRecording: boolean;
}

export function StatusDisplay({ connected, status, audioLevel, isRecording }: StatusDisplayProps) {
  const getConnectionStatus = () => {
    if (!connected) {
      return {
        icon: <WifiOff className="w-4 h-4" />,
        text: 'Disconnected',
        className: 'status-disconnected'
      };
    }

    if (status?.status === 'connecting') {
      return {
        icon: <Loader className="w-4 h-4 animate-spin" />,
        text: 'Connecting...',
        className: 'status-connecting'
      };
    }

    return {
      icon: <Wifi className="w-4 h-4" />,
      text: 'Connected',
      className: 'status-connected'
    };
  };

  const getRecordingStatus = () => {
    if (!isRecording) {
      return {
        icon: <MicOff className="w-4 h-4" />,
        text: 'Not Recording',
        className: 'text-gray-500'
      };
    }

    return {
      icon: <Mic className="w-4 h-4" />,
      text: 'Recording',
      className: 'text-red-500'
    };
  };

  const connectionStatus = getConnectionStatus();
  const recordingStatus = getRecordingStatus();

  return (
    <div className="card">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">Connection Status</h3>
      
      <div className="space-y-4">
        {/* Connection Status */}
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <span className={`status-indicator ${connectionStatus.className}`}></span>
            <div className="flex items-center space-x-2">
              {connectionStatus.icon}
              <span className="font-medium">{connectionStatus.text}</span>
            </div>
          </div>
        </div>

        {/* Recording Status */}
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <span className={`status-indicator ${isRecording ? 'status-recording' : 'bg-gray-300'}`}></span>
            <div className={`flex items-center space-x-2 ${recordingStatus.className}`}>
              {recordingStatus.icon}
              <span className="font-medium">{recordingStatus.text}</span>
            </div>
          </div>
        </div>

        {/* Audio Level Indicator */}
        {isRecording && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-gray-600">Audio Level</span>
              <span className="text-gray-600">{Math.round(audioLevel * 100)}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className="bg-red-500 h-2 rounded-full transition-all duration-100"
                style={{ width: `${audioLevel * 100}%` }}
              ></div>
            </div>
          </div>
        )}

        {/* Current Status Message */}
        {status && (
          <div className="pt-3 border-t border-gray-200">
            <div className="flex items-start space-x-2">
              <div className={`w-2 h-2 rounded-full mt-2 flex-shrink-0 ${getStatusColor(status.status)}`}></div>
              <div>
                <div className="text-sm font-medium text-gray-900 capitalize">
                  {status.status.replace('_', ' ')}
                </div>
                <div className="text-sm text-gray-600 mt-1">
                  {status.message}
                </div>
                {status.project && (
                  <div className="text-xs text-blue-600 mt-1">
                    Project: {status.project}
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'connected':
    case 'completed':
    case 'project_selected':
      return 'bg-green-500';
    case 'connecting':
    case 'processing':
    case 'executing':
      return 'bg-yellow-500';
    case 'recording':
      return 'bg-red-500';
    case 'error':
      return 'bg-red-600';
    default:
      return 'bg-gray-400';
  }
}