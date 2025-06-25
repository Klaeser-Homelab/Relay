import { useState } from 'react';
import { MessageSquare, ChevronDown, ChevronUp, X, Mic, Square, Copy, Check } from 'lucide-react';
import type { TranscriptionData, FunctionResultData, ClaudePlanResponseData, ClaudeStreamingTextData, ClaudeTodoWriteData } from '../types/api';
import { TuiMarkdown } from './TuiMarkdown';
import { TodoWriteRenderer } from './TodoWriteRenderer';

interface ConversationHistoryProps {
  transcriptions: TranscriptionData[];
  functionResults: FunctionResultData[];
  claudeStreamingTexts: ClaudeStreamingTextData[];
  claudeTodoWrites: ClaudeTodoWriteData[];
  isRecording: boolean;
  status: any;
  selectedProject: any;
  connected: boolean;
  audioLevel: number;
  repositoryIssues: any[];
  activeIssue: any;
  onClear: () => void;
  onConnect: () => void;
  onStartRecording: () => void;
  onStopRecording: () => void;
  onIssueClick: (issue: any) => void;
}

interface ConversationMessage {
  id: string;
  type: 'user' | 'openai' | 'claude' | 'claude_streaming' | 'claude_todowrite';
  content: string;
  data?: any;
  timestamp?: string;
}

export function ConversationHistory({ 
  transcriptions, 
  functionResults, 
  claudeStreamingTexts,
  claudeTodoWrites,
  isRecording,
  status,
  selectedProject,
  connected,
  audioLevel,
  repositoryIssues,
  activeIssue,
  onClear,
  onConnect,
  onStartRecording,
  onStopRecording,
  onIssueClick
}: ConversationHistoryProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const [isCopying, setIsCopying] = useState(false);
  const [copySuccess, setCopySuccess] = useState(false);

  // Debug logging
  console.log('🔍 ConversationHistory Debug:');
  console.log('  - Transcriptions:', transcriptions.length, transcriptions);
  console.log('  - Function Results:', functionResults.length, functionResults);
  console.log('  - Claude Streaming Texts:', claudeStreamingTexts.length, claudeStreamingTexts);
  console.log('  - Claude TodoWrites:', claudeTodoWrites.length, claudeTodoWrites);
  console.log('  - Status:', status);
  console.log('  - Is Recording:', isRecording);

  // Combine all messages into a chronological conversation
  const messages: ConversationMessage[] = [];

  // Add welcome message and issues when a project is selected
  if (selectedProject) {
    messages.push({
      id: 'welcome-message',
      type: 'openai',
      content: activeIssue ? 'Let\'s get started!' : 'What should we work on?',
      timestamp: new Date(Date.now() - 1000000).toISOString() // Make it appear first
    });

    // Add issues as clickable items (only when no active issue)
    if (repositoryIssues.length > 0 && !activeIssue) {
      messages.push({
        id: 'issues-list',
        type: 'openai',
        content: 'Open Issues:',
        data: { issues: repositoryIssues, activeIssue },
        timestamp: new Date(Date.now() - 999000).toISOString()
      });
    }
  }

  // Add transcriptions as user messages (filter out very short ones which are streaming fragments)
  transcriptions.forEach((transcription, index) => {
    // Only add transcriptions that are likely complete sentences (more than 3 words)
    const words = transcription.text.trim().split(/\s+/);
    if (words.length > 3 || transcription.text.endsWith('.') || transcription.text.endsWith('?') || transcription.text.endsWith('!')) {
      messages.push({
        id: `transcription-${index}`,
        type: 'user',
        content: transcription.text,
      });
    }
  });

  // Add function results as OpenAI responses
  functionResults.forEach((result, index) => {
    if (result.function === 'ask_claude_to_make_plan') {
      messages.push({
        id: `function-${index}`,
        type: 'openai',
        content: `Asking Claude...`,
        data: result
      });
    }
  });

  // Add Claude streaming texts
  claudeStreamingTexts.forEach((text, index) => {
    messages.push({
      id: `claude-streaming-${index}`,
      type: 'claude_streaming',
      content: text.content,
      timestamp: text.timestamp
    });
  });

  // Add Claude TodoWrite calls - merge all todos into a single latest state
  if (claudeTodoWrites.length > 0) {
    // Merge all todos by ID to get the latest state
    const allTodos = new Map();
    claudeTodoWrites.forEach(todoWrite => {
      todoWrite.todos.forEach(todo => {
        allTodos.set(todo.id, todo);
      });
    });

    const latestTodoWrite = claudeTodoWrites[claudeTodoWrites.length - 1];
    messages.push({
      id: 'claude-todowrite-merged',
      type: 'claude_todowrite',
      content: 'TodoWrite',
      data: {
        todos: Array.from(allTodos.values()),
        timestamp: latestTodoWrite.timestamp
      },
      timestamp: latestTodoWrite.timestamp
    });
  }

  // Skip Claude plan responses since we're streaming progressively
  // The content is already shown via claude_streaming messages

  // Messages are already in chronological order from how they were added
  // No need to sort - this preserves the correct conversation flow

  // Add current status if recording or processing
  if (isRecording && status?.status === 'recording') {
    messages.push({
      id: 'current-recording',
      type: 'openai',
      content: 'Relay Listening...',
      timestamp: new Date().toISOString()
    });
  } else if (status?.status === 'processing') {
    messages.push({
      id: 'current-processing',
      type: 'openai',
      content: 'Processing...',
      timestamp: new Date().toISOString()
    });
  }

  const exportConversation = async () => {
    setIsCopying(true);
    
    try {
      const conversationData = {
        conversation: messages.map((message, index) => {
          let sender = 'system';
          let content = message.content;
          
          // Determine sender based on message type
          if (message.type === 'user') {
            sender = 'user';
          } else if (message.type === 'claude' || message.type === 'claude_streaming') {
            sender = 'claude';
          } else if (message.type === 'claude_todowrite') {
            sender = 'claude';
            // Format todo content
            if (message.data?.todos) {
              content = `TodoWrite: ${message.data.todos.map(todo => 
                `[${todo.status === 'completed' ? 'x' : todo.status === 'in_progress' ? '.' : ' '}] ${todo.content} (${todo.priority})`
              ).join('\n')}`;
            }
          } else if (message.type === 'openai') {
            sender = 'system';
          }
          
          return {
            sequence: index + 1,
            sender,
            content,
            timestamp: message.timestamp || new Date().toISOString(),
            type: message.type
          };
        }),
        metadata: {
          project: selectedProject?.fullName || 'Unknown',
          active_issue: activeIssue ? `#${activeIssue.number}: ${activeIssue.title}` : null,
          exported_at: new Date().toISOString(),
          total_messages: messages.length,
          export_source: 'Relay Conversation History'
        }
      };
      
      const jsonString = JSON.stringify(conversationData, null, 2);
      await navigator.clipboard.writeText(jsonString);
      
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    } catch (error) {
      console.error('Failed to copy conversation:', error);
    } finally {
      setIsCopying(false);
    }
  };

  // Always show the conversation when a project is selected
  if (!selectedProject) {
    return null;
  }


  return (
    <div className="relative h-full">
      {/* Copy Button - Fixed position */}
      {messages.length > 0 && (
        <div className="absolute top-4 right-4 z-10">
          <button
            onClick={exportConversation}
            disabled={isCopying}
            className={`p-2 rounded-lg transition-all duration-200 ${
              copySuccess 
                ? 'bg-green-600 text-white' 
                : 'bg-gray-700 hover:bg-gray-600 text-gray-300 hover:text-white'
            } border border-gray-600 shadow-lg`}
            title="Copy conversation as JSON"
          >
            {isCopying ? (
              <div className="w-4 h-4 border-2 border-gray-400 border-t-transparent rounded-full animate-spin" />
            ) : copySuccess ? (
              <Check className="w-4 h-4" />
            ) : (
              <Copy className="w-4 h-4" />
            )}
          </button>
        </div>
      )}
      
      <div className="h-full overflow-y-auto bg-gray-900 p-6 pr-16 rounded-md font-mono text-sm leading-relaxed scrollbar-hide">
        {messages.map((message) => (
          <div key={message.id} className="mb-4">
            <div className="flex items-start space-x-3">
              <span className={
                message.type === 'user' 
                  ? 'text-gray-400 flex-shrink-0'
                  : message.type === 'openai'
                  ? 'text-gray-200 flex-shrink-0'
                  : 'text-orange-400 flex-shrink-0'
              }>
                {message.type === 'user' && '> '}
                {message.type === 'openai' && '● '}
                {(message.type === 'claude' || message.type === 'claude_streaming' || message.type === 'claude_todowrite') && '● '}
              </span>
              <div className="flex-1">
                {message.type === 'claude' ? (
                  <TuiMarkdown content={message.content} />
                ) : message.type === 'claude_streaming' ? (
                  <TuiMarkdown content={message.content} />
                ) : message.type === 'claude_todowrite' ? (
                  <TodoWriteRenderer data={message.data} />
                ) : message.id === 'issues-list' && message.data?.issues ? (
                  <div>
                    <span className="text-gray-100 block mb-2">{message.content}</span>
                    <div className="space-y-3">
                      {message.data.issues.map((issue: any) => (
                        <button
                          key={issue.number}
                          onClick={() => onIssueClick(issue)}
                          className="w-full text-left px-4 py-3 rounded-lg text-sm bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-600 transition-colors"
                        >
                          <span className="font-mono text-blue-400">#{issue.number}</span>{' '}
                          <span className="truncate">{issue.title}</span>
                          {issue.labels.length > 0 && (
                            <span className="ml-2">
                              {issue.labels.slice(0, 3).map((label: string) => (
                                <span key={label} className="inline-block bg-gray-700 text-gray-300 text-xs px-2 py-1 rounded mr-2">
                                  {label}
                                </span>
                              ))}
                            </span>
                          )}
                        </button>
                      ))}
                    </div>
                  </div>
                ) : (
                  <span className={
                    message.type === 'user' 
                      ? 'text-gray-400'
                      : 'text-gray-100'
                  }>
                    {message.content}
                  </span>
                )}
              </div>
            </div>
          </div>
        ))}

        {messages.length === 0 && (
          <div className="text-center py-8 text-gray-400">
            <MessageSquare className="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p className="text-sm">No conversation yet</p>
            <p className="text-xs">Start recording to begin a conversation</p>
          </div>
        )}
      </div>

    </div>
  );
}