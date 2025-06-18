import { useEffect, useRef } from 'react';
import { MessageSquare, Trash2 } from 'lucide-react';
import type { TranscriptionData } from '../types/api';

interface TranscriptionViewProps {
  transcriptions: TranscriptionData[];
  onClear: () => void;
}

export function TranscriptionView({ transcriptions, onClear }: TranscriptionViewProps) {
  const scrollRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new transcriptions arrive
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [transcriptions]);

  if (transcriptions.length === 0) {
    return (
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900 flex items-center">
            <MessageSquare className="w-5 h-5 mr-2" />
            Voice Transcription
          </h3>
        </div>
        <div className="text-center py-8 text-gray-500">
          <MessageSquare className="w-12 h-12 mx-auto mb-3 text-gray-300" />
          <p>Start recording to see your voice transcribed here</p>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900 flex items-center">
          <MessageSquare className="w-5 h-5 mr-2" />
          Voice Transcription
        </h3>
        <button
          onClick={onClear}
          className="btn-secondary text-sm flex items-center"
        >
          <Trash2 className="w-4 h-4 mr-1" />
          Clear
        </button>
      </div>

      <div 
        ref={scrollRef}
        className="space-y-3 max-h-64 overflow-y-auto bg-gray-50 rounded-lg p-4"
      >
        {transcriptions.map((transcription, index) => (
          <TranscriptionItem
            key={index}
            transcription={transcription}
            timestamp={new Date()}
          />
        ))}
      </div>
    </div>
  );
}

interface TranscriptionItemProps {
  transcription: TranscriptionData;
  timestamp: Date;
}

function TranscriptionItem({ transcription, timestamp }: TranscriptionItemProps) {
  const formatTime = (date: Date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  return (
    <div className="bg-white rounded-lg p-3 shadow-sm border border-gray-200">
      <div className="flex items-start justify-between mb-2">
        <div className="text-xs text-gray-500">
          {formatTime(timestamp)}
        </div>
      </div>
      <div className="text-gray-900">
        {transcription.text}
      </div>
    </div>
  );
}