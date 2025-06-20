import { useEffect, useState } from 'react';
import { Mic, Square } from 'lucide-react';
import type { GitHubRepository } from '../types/api';

interface VoiceChatProps {
  selectedProject: GitHubRepository | null;
  connected: boolean;
  isRecording: boolean;
  audioLevel: number;
  onStartRecording: () => void;
  onStopRecording: () => void;
  onConnect: () => void;
  onDisconnect: () => void;
}

export function VoiceChat({
  selectedProject,
  connected,
  isRecording,
  audioLevel,
  onStartRecording,
  onStopRecording,
  onConnect,
  onDisconnect
}: VoiceChatProps) {
  const [isHolding, setIsHolding] = useState(false);
  const [recordingMode, setRecordingMode] = useState<'push-to-talk' | 'toggle'>('push-to-talk');

  const handleMouseDown = () => {
    if (recordingMode === 'push-to-talk' && connected && !isRecording) {
      setIsHolding(true);
      onStartRecording();
    }
  };

  const handleMouseUp = () => {
    if (recordingMode === 'push-to-talk' && isHolding) {
      setIsHolding(false);
      onStopRecording();
    }
  };

  const handleClick = () => {
    if (recordingMode === 'toggle') {
      if (isRecording) {
        onStopRecording();
      } else {
        onStartRecording();
      }
    }
  };

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.code === 'Space' && e.target === document.body) {
        e.preventDefault();
        if (recordingMode === 'push-to-talk' && connected && !isRecording && !isHolding) {
          setIsHolding(true);
          onStartRecording();
        }
      }
    };

    const handleKeyUp = (e: KeyboardEvent) => {
      if (e.code === 'Space' && isHolding) {
        e.preventDefault();
        setIsHolding(false);
        onStopRecording();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    document.addEventListener('keyup', handleKeyUp);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.removeEventListener('keyup', handleKeyUp);
    };
  }, [recordingMode, connected, isRecording, isHolding, onStartRecording, onStopRecording]);

  if (!selectedProject) {
    return (
      <div className="card">
        <div className="text-center py-12">
          <Mic className="w-16 h-16 mx-auto text-gray-300 mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            Select a Repository
          </h3>
          <p className="text-gray-600">
            Choose a GitHub repository to start voice chatting about your project.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <div className="text-center">
        <h3 className="text-lg font-semibold text-gray-900 mb-2">
          Voice Chat
        </h3>
        <div className="text-sm text-gray-600 mb-6">
          {selectedProject.fullName}
        </div>

        {/* Connection Controls */}
        <div className="mb-6">
          {connected ? (
            <button
              onClick={onDisconnect}
              className="btn-danger"
            >
              Disconnect
            </button>
          ) : (
            <button
              onClick={onConnect}
              className="btn-primary"
            >
              Connect to Voice Assistant
            </button>
          )}
        </div>

        {connected && (
          <>
            {/* Recording Mode Toggle */}
            <div className="mb-6">
              <div className="text-sm text-gray-600 mb-2">Recording Mode</div>
              <div className="flex items-center justify-center space-x-2">
                <button
                  onClick={() => setRecordingMode('push-to-talk')}
                  className={`px-3 py-1 rounded-full text-sm ${
                    recordingMode === 'push-to-talk'
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  Push to Talk
                </button>
                <button
                  onClick={() => setRecordingMode('toggle')}
                  className={`px-3 py-1 rounded-full text-sm ${
                    recordingMode === 'toggle'
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  Click to Toggle
                </button>
              </div>
            </div>

            {/* Voice Recording Button */}
            <div className="mb-6">
              <div className="relative inline-block">
                <button
                  className={`w-24 h-24 rounded-full border-4 transition-all duration-200 flex items-center justify-center ${
                    isRecording
                      ? 'bg-red-500 border-red-600 text-white shadow-lg transform scale-110'
                      : 'bg-white border-gray-300 text-gray-600 hover:border-gray-400 hover:shadow-md'
                  }`}
                  onMouseDown={handleMouseDown}
                  onMouseUp={handleMouseUp}
                  onMouseLeave={handleMouseUp}
                  onClick={handleClick}
                  disabled={!connected}
                >
                  {isRecording ? (
                    <Square className="w-8 h-8" />
                  ) : (
                    <Mic className="w-8 h-8" />
                  )}
                </button>

                {/* Audio Level Ring */}
                {isRecording && audioLevel > 0 && (
                  <div
                    className="absolute inset-0 rounded-full border-4 border-red-400 animate-pulse"
                    style={{
                      transform: `scale(${1 + audioLevel * 0.3})`,
                      opacity: 0.6
                    }}
                  />
                )}
              </div>

              <div className="mt-4 text-sm text-gray-600">
                {isRecording ? (
                  <span className="text-red-600 font-medium">
                    🔴 Recording... {recordingMode === 'push-to-talk' ? 'Release to stop' : 'Click to stop'}
                  </span>
                ) : (
                  <span>
                    {recordingMode === 'push-to-talk' 
                      ? 'Hold to record (or press Space)' 
                      : 'Click to start recording'
                    }
                  </span>
                )}
              </div>
            </div>

          </>
        )}
      </div>
    </div>
  );
}