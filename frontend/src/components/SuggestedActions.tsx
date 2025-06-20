import React, { useState, useRef, useEffect } from 'react';
import { FileText, Play, X, Loader2, Mic, MicOff, Wifi, MoreHorizontal, Terminal as TerminalIcon } from 'lucide-react';

export interface SuggestedAction {
  id: string;
  label: string;
  icon: React.ReactNode;
  variant: 'primary' | 'secondary' | 'danger' | 'recording';
  action: () => void | Promise<void>;
  loading?: boolean;
  likelihood: number; // 0 to 1 scale for ordering
  size?: 'normal' | 'large'; // For microphone button
}

interface SuggestedActionsProps {
  visible: boolean;
  onUpdatePlan: () => Promise<void>;
  onImplement: () => void;
  onMicrophoneClick: () => void;
  onOpenTerminal: () => void;
  isUpdatingPlan?: boolean;
  connected: boolean;
  isRecording: boolean;
  audioLevel?: number;
  hasActiveIssue?: boolean;
  selectedProject?: any;
}

export const SuggestedActions: React.FC<SuggestedActionsProps> = ({
  visible,
  onUpdatePlan,
  onImplement,
  onMicrophoneClick,
  onOpenTerminal,
  isUpdatingPlan = false,
  connected,
  isRecording,
  audioLevel = 0,
  hasActiveIssue = false,
  selectedProject
}) => {
  const [showOtherPopup, setShowOtherPopup] = useState(false);
  const otherButtonRef = useRef<HTMLButtonElement>(null);
  const popupRef = useRef<HTMLDivElement>(null);

  // Close popup when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (popupRef.current && !popupRef.current.contains(event.target as Node) &&
          otherButtonRef.current && !otherButtonRef.current.contains(event.target as Node)) {
        setShowOtherPopup(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);
  const getMicrophoneIcon = () => {
    if (!connected) {
      return <Wifi className="w-4 h-4" />;
    }
    return isRecording ? <MicOff className="w-4 h-4" /> : <Mic className="w-4 h-4" />;
  };

  const getMicrophoneLabel = () => {
    if (!connected) return 'Connect';
    return isRecording ? 'Stop' : 'Record';
  };

  const handleOtherClick = () => {
    setShowOtherPopup(!showOtherPopup);
  };

  const handleTerminalClick = () => {
    onOpenTerminal();
    setShowOtherPopup(false);
  };

  const actions: SuggestedAction[] = [
    {
      id: 'other',
      label: 'Other',
      icon: <MoreHorizontal className="w-4 h-4" />,
      variant: 'secondary',
      action: handleOtherClick,
      likelihood: 0.1
    },
    // Only show Update Plan when there's an active issue
    ...(hasActiveIssue ? [{
      id: 'update-plan',
      label: 'Update Plan',
      icon: isUpdatingPlan ? <Loader2 className="w-4 h-4 animate-spin" /> : <FileText className="w-4 h-4" />,
      variant: 'secondary' as const,
      action: onUpdatePlan,
      loading: isUpdatingPlan,
      likelihood: 0.8
    }] : []),
    // Only show Implement when there's an active issue
    ...(hasActiveIssue ? [{
      id: 'implement',
      label: 'Implement',
      icon: <Play className="w-4 h-4" />,
      variant: 'primary' as const,
      action: onImplement,
      likelihood: 0.5
    }] : []),
    {
      id: 'microphone',
      label: getMicrophoneLabel(),
      icon: getMicrophoneIcon(),
      variant: 'recording',
      action: onMicrophoneClick,
      likelihood: 1.0
    }
  ];

  // Sort actions by likelihood (lowest first, so highest appears on the right)
  const sortedActions = actions.sort((a, b) => a.likelihood - b.likelihood);

  if (!visible) return null;

  const getButtonStyles = (variant: string, loading: boolean) => {
    const baseStyles = "flex items-center justify-center space-x-2 rounded-lg font-medium transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed px-6 py-3 text-sm";
    
    switch (variant) {
      case 'recording':
        if (!connected) {
          return `${baseStyles} bg-blue-500 border-2 border-blue-600 text-white hover:bg-blue-600 shadow-lg`;
        }
        return `${baseStyles} ${
          isRecording 
            ? 'bg-red-500 border-2 border-red-600 text-white shadow-lg transform scale-105' 
            : 'bg-gray-700 border-2 border-gray-600 text-gray-300 hover:bg-gray-600 hover:border-gray-500'
        }`;
      case 'primary':
        return `${baseStyles} bg-blue-600 text-white hover:bg-blue-700 disabled:hover:bg-blue-600`;
      case 'danger':
        return `${baseStyles} bg-red-600 text-white hover:bg-red-700 disabled:hover:bg-red-600`;
      case 'secondary':
      default:
        return `${baseStyles} bg-gray-700 text-gray-200 hover:bg-gray-600 border border-gray-600 disabled:hover:bg-gray-700`;
    }
  };

  return (
    <div className="fixed bottom-6 left-0 right-0 z-30 px-6">
      <div className="flex items-center justify-center space-x-6 bg-gray-800 px-8 py-4 rounded-xl border border-gray-600 shadow-lg backdrop-blur-sm max-w-none w-full">
        {sortedActions.map((action) => (
          <div key={action.id} className="relative">
            <button
              ref={action.id === 'other' ? otherButtonRef : undefined}
              onClick={action.action}
              disabled={action.loading}
              className={getButtonStyles(action.variant, action.loading || false)}
              title={action.label}
              style={
                action.id === 'microphone' && isRecording && audioLevel > 0
                  ? {
                      transform: `scale(${1 + audioLevel * 0.2})`,
                      boxShadow: `0 0 ${20 + audioLevel * 30}px rgba(239, 68, 68, 0.6)`
                    }
                  : undefined
              }
            >
              {action.icon}
              <span>{action.label}</span>
            </button>

            {/* Other Actions Popup */}
            {action.id === 'other' && showOtherPopup && (
              <div
                ref={popupRef}
                className="absolute bottom-full mb-2 left-1/2 transform -translate-x-1/2 bg-gray-800 border border-gray-600 rounded-lg shadow-xl z-50 min-w-48"
              >
                <div className="py-2">
                  {/* Terminal Option - Only show for cloned repos */}
                  {selectedProject && selectedProject.isCloned && (
                    <button
                      onClick={handleTerminalClick}
                      className="w-full px-4 py-2 text-left text-gray-200 hover:bg-gray-700 flex items-center space-x-2 transition-colors"
                    >
                      <TerminalIcon className="w-4 h-4" />
                      <span>Terminal</span>
                    </button>
                  )}
                  
                  {/* Message when no options available */}
                  {(!selectedProject || !selectedProject.isCloned) && (
                    <div className="px-4 py-2 text-gray-400 text-sm">
                      No additional actions available
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};