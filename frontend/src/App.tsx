import { useEffect, useState } from 'react';
import { Star, GitFork, Lock, Globe, HardDrive, Wifi, WifiOff, Mic, MicOff, Menu, X, Square, Plus, Terminal as TerminalIcon } from 'lucide-react';
import { ProjectSelector } from './components/ProjectSelector';
import { VoiceChat } from './components/VoiceChat';
import { StatusDisplay } from './components/StatusDisplay';
import { TranscriptionView } from './components/TranscriptionView';
import { FunctionResults } from './components/FunctionResults';
import { GeminiAdvice } from './components/GeminiAdvice';
import { DeveloperMode } from './components/DeveloperMode';
import { Terminal } from './components/Terminal';
import { ClaudePlan } from './components/ClaudePlan';
import { Settings } from './components/Settings';
import { ConversationHistory } from './components/ConversationHistory';
import { Sidebar } from './components/Sidebar';
import { useGitHubProjects } from './hooks/useGitHubProjects';
import { useWebSocket } from './hooks/useWebSocket';
import { useAudioRecording } from './hooks/useAudioRecording';
import { sessionApi } from './services/sessionApi';

function App() {
  const [developerMode, setDeveloperMode] = useState(false);
  const [showTerminal, setShowTerminal] = useState(false);
  const [terminalWorkingDir, setTerminalWorkingDir] = useState<string | undefined>();
  const [claudePrompt, setClaudePrompt] = useState<string | undefined>();
  const [showSettings, setShowSettings] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [activeIssue, setActiveIssue] = useState<any>(null);
  const [repositoryIssues, setRepositoryIssues] = useState<any[]>([]);
  const [quietModeByRepo, setQuietModeByRepo] = useState<{[key: string]: boolean}>({});
  const [showQuietOverlay, setShowQuietOverlay] = useState(false);
  const [inactivityTimer, setInactivityTimer] = useState<NodeJS.Timeout | null>(null);
  const [recentChats, setRecentChats] = useState<any[]>([]);
  const [currentChatId, setCurrentChatId] = useState<string | null>(null);
  
  const {
    projects,
    selectedProject,
    loading: projectsLoading,
    error: projectsError,
    selectProject,
    clearProject,
    fetchProjects
  } = useGitHubProjects();

  const {
    connected,
    status,
    transcriptions,
    functionResults,
    geminiAdvice,
    claudePlanRequests,
    claudePlanResponses,
    claudeStreamingTexts,
    claudeTodoWrites,
    connect,
    disconnect,
    startRecording,
    stopRecording,
    sendAudio,
    selectProject: selectProjectWS,
    clearTranscriptions,
    clearFunctionResults,
    clearGeminiAdvice,
    clearClaudePlanRequests,
    clearClaudePlanResponses,
    clearClaudeStreamingTexts,
    clearClaudeTodoWrites,
    socket
  } = useWebSocket();

  const {
    isRecording,
    audioLevel,
    startRecording: startAudioRecording,
    stopRecording: stopAudioRecording,
    setOnAudioData
  } = useAudioRecording();

  // Set up audio data callback
  useEffect(() => {
    setOnAudioData((audioData: string) => {
      sendAudio(audioData);
    });
  }, [setOnAudioData, sendAudio]);

  // Handle Claude plan requests (legacy - keeping for backward compatibility)
  useEffect(() => {
    console.log('=== APP: Claude plan requests changed ===');
    console.log('Current requests:', claudePlanRequests.length);
    
    if (claudePlanRequests.length > 0) {
      const latestRequest = claudePlanRequests[claudePlanRequests.length - 1];
      console.log('✓ Processing latest request:', JSON.stringify(latestRequest, null, 2));
      
      // Set terminal working directory, prompt, and show terminal
      console.log('✓ Setting terminal state...');
      setTerminalWorkingDir(latestRequest.workingDirectory);
      setClaudePrompt(latestRequest.prompt);
      setShowTerminal(true);
      console.log('✓ Terminal should now be visible');
      
      // Clear the request after handling
      setTimeout(() => {
        console.log('✓ Clearing Claude plan requests');
        clearClaudePlanRequests();
      }, 1000);
    }
  }, [claudePlanRequests, clearClaudePlanRequests]);

  // Log Claude plan responses
  useEffect(() => {
    console.log('=== APP: Claude plan responses changed ===');
    console.log('Current responses:', claudePlanResponses.length);
    if (claudePlanResponses.length > 0) {
      const latest = claudePlanResponses[claudePlanResponses.length - 1];
      console.log('✓ Latest plan response:', latest.prompt, 'Plan length:', latest.plan.length);
    }
  }, [claudePlanResponses]);

  // Handle quiet mode inactivity timer
  useEffect(() => {
    if (selectedProject && quietModeByRepo[selectedProject.fullName]) {
      resetInactivityTimer();
    } else {
      // Clean up timer when quiet mode is off
      if (inactivityTimer) {
        clearTimeout(inactivityTimer);
        setInactivityTimer(null);
      }
      setShowQuietOverlay(false);
    }
    
    // Cleanup on unmount
    return () => {
      if (inactivityTimer) {
        clearTimeout(inactivityTimer);
      }
    };
  }, [selectedProject, quietModeByRepo]);

  // Add mouse movement listener for quiet mode
  useEffect(() => {
    const handleMouseMove = () => {
      if (selectedProject && quietModeByRepo[selectedProject.fullName]) {
        resetInactivityTimer();
      }
    };

    const handleKeyPress = () => {
      if (selectedProject && quietModeByRepo[selectedProject.fullName]) {
        resetInactivityTimer();
      }
    };

    if (selectedProject && quietModeByRepo[selectedProject.fullName]) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('keydown', handleKeyPress);
      document.addEventListener('click', handleMouseMove);
      
      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('keydown', handleKeyPress);
        document.removeEventListener('click', handleMouseMove);
      };
    }
  }, [selectedProject, quietModeByRepo]);

  const clearAllConversation = () => {
    clearTranscriptions();
    clearFunctionResults();
    clearGeminiAdvice();
    clearClaudeStreamingTexts();
    clearClaudeTodoWrites();
  };

  const handleSelectProject = async (project: any) => {
    const success = await selectProject(project);
    if (success) {
      try {
        // Create new session in database
        const session = await sessionApi.createSession(
          project.name,
          project.fullName,
          `${project.name} - ${new Date().toLocaleDateString()}`
        );
        
        setCurrentChatId(session.id);
        
        // Connect WebSocket with the new session ID
        if (!connected) {
          connect(session.id);
        } else {
          // Disconnect and reconnect with new session
          disconnect();
          setTimeout(() => connect(session.id), 100);
        }
        
        // Update the WebSocket project selection
        if (connected) {
          selectProjectWS(project.name);
        }
        
        // Fetch issues for the selected project
        fetchRepositoryIssues(project.fullName);
        // Clear active issue when switching projects
        setActiveIssue(null);
        
        // Load recent sessions for this project
        const sessions = await sessionApi.listSessions(project.fullName);
        setRecentChats(sessions);
      } catch (error) {
        console.error('Failed to create session:', error);
        // Fall back to non-persistent mode
        if (!connected) {
          connect();
        }
        if (connected) {
          selectProjectWS(project.name);
        }
      }
    }
    return success;
  };

  const handleIssueClick = async (issue: any) => {
    setActiveIssue(issue);
    if (issue) {
      console.log(`Active issue set to #${issue.number}: ${issue.title}`);
      
      // Update the session with the active issue
      if (currentChatId) {
        try {
          await sessionApi.updateSession(currentChatId, {
            active_issue_number: issue.number,
            active_issue_title: issue.title,
            active_issue_url: issue.url || issue.html_url
          });
        } catch (error) {
          console.error('Failed to update session with active issue:', error);
        }
      }
    } else {
      console.log('Active issue cleared');
      
      // Clear the active issue from the session
      if (currentChatId) {
        try {
          await sessionApi.updateSession(currentChatId, {
            active_issue_number: null,
            active_issue_title: null,
            active_issue_url: null
          });
        } catch (error) {
          console.error('Failed to clear active issue from session:', error);
        }
      }
    }
  };

  const handleToggleQuietMode = () => {
    if (selectedProject) {
      const newQuietModeByRepo = {
        ...quietModeByRepo,
        [selectedProject.fullName]: !quietModeByRepo[selectedProject.fullName]
      };
      setQuietModeByRepo(newQuietModeByRepo);
      
      // If turning off quiet mode, hide overlay immediately
      if (quietModeByRepo[selectedProject.fullName]) {
        setShowQuietOverlay(false);
        if (inactivityTimer) {
          clearTimeout(inactivityTimer);
          setInactivityTimer(null);
        }
      }
    }
  };

  const resetInactivityTimer = () => {
    if (!selectedProject || !quietModeByRepo[selectedProject.fullName]) return;
    
    // Clear existing timer
    if (inactivityTimer) {
      clearTimeout(inactivityTimer);
    }
    
    // Hide overlay if showing
    setShowQuietOverlay(false);
    
    // Set new timer for 10 seconds
    const newTimer = setTimeout(() => {
      setShowQuietOverlay(true);
    }, 10000);
    
    setInactivityTimer(newTimer);
  };

  const handleSelectChat = async (chat: any) => {
    try {
      // Resume the session
      const sessionData = await sessionApi.resumeSession(chat.id);
      
      setCurrentChatId(chat.id);
      
      // Select the project associated with the chat
      const project = projects.find(p => p.fullName === chat.repository_full_name);
      if (project) {
        await selectProject(project);
        
        // Reconnect with the resumed session ID
        if (!connected) {
          connect(chat.id);
        } else {
          disconnect();
          setTimeout(() => connect(chat.id), 100);
        }
        
        // Restore active issue if any
        if (sessionData.chat.active_issue_number) {
          setActiveIssue({
            number: sessionData.chat.active_issue_number,
            title: sessionData.chat.active_issue_title,
            url: sessionData.chat.active_issue_url
          });
        }
        
        // Fetch issues for the project
        fetchRepositoryIssues(project.fullName);
      }
    } catch (error) {
      console.error('Failed to resume session:', error);
    }
  };

  const handleNewChat = () => {
    // Save current chat to recents if we have a selected project
    if (selectedProject && currentChatId) {
      const currentChat = {
        id: currentChatId,
        title: `${selectedProject.name} - ${new Date().toLocaleDateString()}`,
        repository_name: selectedProject.name,
        repository_full_name: selectedProject.fullName,
        updated_at: new Date().toISOString(),
        last_accessed_at: new Date().toISOString()
      };
      
      // Add to recents if not already there
      setRecentChats(prev => {
        const exists = prev.some(chat => chat.id === currentChatId);
        if (!exists) {
          return [currentChat, ...prev.slice(0, 19)]; // Keep last 20 chats
        }
        return prev;
      });
    }
    
    // Clear current selection and show full repository selector
    clearProject();
    setCurrentChatId(null);
    setActiveIssue(null);
    clearAllConversation();
  };

  // Load recent sessions on component mount
  useEffect(() => {
    loadRecentSessions();
  }, []);
  
  const loadRecentSessions = async () => {
    try {
      const sessions = await sessionApi.listSessions();
      setRecentChats(sessions);
      
      // Auto-resume last session if available
      if (sessions.length > 0 && !currentChatId) {
        const lastSession = sessions[0];
        await handleSelectChat(lastSession);
      }
    } catch (error) {
      console.error('Failed to load recent sessions:', error);
    }
  };

  const handleStartRecording = () => {
    console.log('🎯 [DEBUG] App handleStartRecording called');
    console.log('🎯 [DEBUG] Socket connected:', connected);
    console.log('🎯 [DEBUG] Socket object:', socket);
    startRecording();
    startAudioRecording();
  };

  const handleStopRecording = () => {
    console.log('🎯 [DEBUG] App handleStopRecording called');
    stopRecording();
    stopAudioRecording();
  };

  const handleMicrophoneClick = async () => {
    if (!connected) {
      // If we have a current chat ID, connect with it
      // Otherwise, connect without a session (will need to select project first)
      connect(currentChatId || undefined);
    } else if (isRecording) {
      handleStopRecording();
    } else {
      handleStartRecording();
    }
  };

  const handleCloneRepository = async (project: any) => {
    try {
      const response = await fetch('/api/repositories/clone', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ repository: project }),
      });

      const result = await response.json();
      
      if (result.success) {
        // Refresh projects to update clone status
        await fetchProjects();
        return true;
      } else {
        console.error('Failed to clone repository:', result.message);
        alert(`Failed to clone repository: ${result.message}`);
        return false;
      }
    } catch (error) {
      console.error('Error cloning repository:', error);
      alert('Error cloning repository. Please try again.');
      return false;
    }
  };

  const fetchRepositoryIssues = async (repositoryName: string) => {
    try {
      // First select the project to ensure context
      const selectResponse = await fetch(`/api/projects/${encodeURIComponent(repositoryName)}/select`, {
        method: 'POST',
      });
      
      if (!selectResponse.ok) {
        console.error('Failed to select project for issues');
        return;
      }

      // Fetch issues using GitHub manager function
      const response = await fetch('/api/github/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          projectName: repositoryName,
          functionName: 'list_issues',
          args: { state: 'open', limit: 10 }
        }),
      });

      const result = await response.json();
      
      if (result.success && result.data && result.data.issues) {
        setRepositoryIssues(result.data.issues);
        console.log(`Loaded ${result.data.issues.length} issues for ${repositoryName}`);
      } else {
        console.error('Failed to fetch issues:', result.message);
        setRepositoryIssues([]);
      }
    } catch (error) {
      console.error('Error fetching repository issues:', error);
      setRepositoryIssues([]);
    }
  };

  const handleSavePlan = async (plan: any) => {
    try {
      if (activeIssue) {
        // Add as comment to active issue
        const response = await fetch('/api/github/execute', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            projectName: selectedProject?.fullName,
            functionName: 'add_issue_comment',
            args: {
              number: activeIssue.number,
              body: `## Claude Generated Plan\n\n**Prompt:** ${plan.prompt}\n\n**Plan:**\n\n${plan.plan}\n\n---\n*Generated by Claude Code at ${new Date().toLocaleString()}*`
            }
          }),
        });

        const result = await response.json();
        
        if (result.success) {
          alert(`Plan added as comment to issue #${activeIssue.number}!\nView at: ${activeIssue.url}`);
        } else {
          console.error('Failed to add comment:', result.message);
          alert(`Failed to add comment: ${result.message}`);
        }
      } else {
        // Create new issue as before
        const title = `Plan: ${plan.prompt}`;
        
        const response = await fetch('/api/plans/save-as-issue', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            plan: plan.plan,
            title: title,
            repository: plan.repository || selectedProject?.fullName
          }),
        });

        const result = await response.json();
        
        if (result.success) {
          alert(`Plan saved as GitHub issue #${result.issueNumber}!\nView at: ${result.issueUrl}`);
          // Refresh issues to include the new one
          if (selectedProject) {
            fetchRepositoryIssues(selectedProject.fullName);
          }
        } else {
          console.error('Failed to save plan:', result.error);
          alert(`Failed to save plan: ${result.error}`);
        }
      }
    } catch (error) {
      console.error('Error saving plan:', error);
      alert('Error saving plan. Please try again.');
    }
  };

  return (
    <div className="min-h-screen bg-gray-800">
      {/* Sidebar */}
      <Sidebar 
        developerMode={developerMode}
        onToggleDeveloperMode={() => setDeveloperMode(!developerMode)}
        onOpenSettings={() => setShowSettings(true)}
        isOpen={sidebarOpen}
        onToggle={() => setSidebarOpen(!sidebarOpen)}
        selectedProject={selectedProject}
        repositoryCount={projects.length}
        quietMode={selectedProject ? quietModeByRepo[selectedProject.fullName] || false : false}
        onToggleQuietMode={handleToggleQuietMode}
        recentChats={recentChats}
        currentChatId={currentChatId}
        onSelectChat={handleSelectChat}
        onNewChat={handleNewChat}
      />

      {/* Header - Fixed at top */}
      <header className="fixed top-0 left-0 right-0 z-40 p-4 bg-gray-900 border-b border-gray-700">
          <div className="">
            <div className="flex justify-between items-center">
            <div className="flex items-center space-x-4">
                {/* Menu Button */}
                <button
                  onClick={() => setSidebarOpen(!sidebarOpen)}
                  className=" text-gray-400 hover:text-gray-200 hover:bg-gray-700 rounded-md transition-colors"
                  title="Menu"
                >
                  <Menu className="w-5 h-5" />
                </button>

                <div className="flex-1">
                  {selectedProject ? (
                    <div>
                      <h1 className="text-2xl font-bold text-white">
                        {selectedProject.fullName}
                      </h1>
                    </div>
                  ) : (
                    <div>
                      <h1 className="text-2xl font-bold text-white">
                        What next?
                      </h1>
                    </div>
                  )}
                </div>
              </div>
              <div className="flex items-center space-x-4">
                {/* Terminal Button - Only show for cloned repos */}
                {selectedProject && selectedProject.isCloned && (
                  <button
                    onClick={() => setShowTerminal(true)}
                    className="p-2 text-gray-400 hover:text-gray-200 rounded-md hover:bg-gray-700 transition-colors"
                    title="Open terminal for repository"
                  >
                    <TerminalIcon className="w-5 h-5" />
                  </button>
                )}
                
                {/* Plus Button - New Chat/Repository - Only show when repo selected */}
                {selectedProject && (
                  <button
                    onClick={handleNewChat}
                    className=" text-gray-400 hover:text-gray-200 rounded-md hover:bg-gray-700 transition-colors"
                    title="New chat with repository"
                  >
                    <Plus className="w-5 h-5" />
                  </button>
                )}


                {/* Repository count and change button moved to sidebar */}
              </div>
            </div>
          </div>
        </header>

        {/* Main Content - Account for fixed header */}
        <main className="pt-20 bg-gray-900 relative flex flex-col min-h-screen">
          
          {/* Active Issue Display - Top Right */}
          {activeIssue && selectedProject && (
            <div className="absolute top-4 right-4 z-10">
              <div className="flex items-center space-x-2 bg-gray-800 px-3 py-2 rounded-lg border border-gray-600 shadow-lg">
                <span className="text-gray-300 text-sm">Working on:</span>
                <span className="font-mono text-blue-400 text-sm">#{activeIssue.number}</span>
                <span className="text-white text-sm truncate max-w-48">{activeIssue.title}</span>
                {activeIssue.labels.length > 0 && (
                  <span className="ml-2">
                    {activeIssue.labels.slice(0, 2).map((label: string) => (
                      <span key={label} className="inline-block bg-gray-700 text-gray-300 text-xs px-1 rounded mr-1">
                        {label}
                      </span>
                    ))}
                  </span>
                )}
                <button
                  onClick={() => setActiveIssue(null)}
                  className="p-1 text-gray-400 hover:text-gray-200 rounded hover:bg-gray-700 transition-colors"
                  title="Clear active issue"
                >
                  <X className="w-3 h-3" />
                </button>
              </div>
            </div>
          )}
          
          {selectedProject ? (
            <>
              {/* Conversation Section - Scrollable with fixed height */}
              <div className="flex-1 px-4 sm:px-6 lg:px-8 py-4 overflow-y-auto" style={{height: 'calc(100vh - 240px)'}}>
                <ConversationHistory
                  transcriptions={transcriptions}
                  functionResults={functionResults}
                  claudeStreamingTexts={claudeStreamingTexts}
                  claudeTodoWrites={claudeTodoWrites}
                  isRecording={isRecording}
                  status={status}
                  selectedProject={selectedProject}
                  connected={connected}
                  audioLevel={audioLevel}
                  repositoryIssues={repositoryIssues}
                  activeIssue={activeIssue}
                  onClear={clearAllConversation}
                  onConnect={connect}
                  onStartRecording={handleStartRecording}
                  onStopRecording={handleStopRecording}
                  onIssueClick={handleIssueClick}
                />
              </div>

              {/* Recording Button - Absolutely positioned at bottom */}
              <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 z-30">
                <div className="flex flex-col items-center space-y-4">
                  <button
                    onClick={handleMicrophoneClick}
                    className={`w-32 h-32 rounded-full border-4 transition-all duration-200 flex items-center justify-center ${
                      !connected
                        ? 'bg-blue-500 border-blue-600 text-white hover:bg-blue-600 shadow-lg'
                        : isRecording
                        ? 'bg-red-500 border-red-600 text-white shadow-lg transform scale-110'
                        : 'bg-gray-700 border-gray-600 text-gray-300 hover:bg-gray-600 hover:border-gray-500'
                    }`}
                    title={
                      !connected 
                        ? 'Connect to Voice Assistant' 
                        : isRecording 
                        ? 'Stop Recording' 
                        : 'Start Recording'
                    }
                  >
                    {!connected ? (
                      <Mic className="w-14 h-14" />
                    ) : isRecording ? (
                      <Square className="w-10 h-10" />
                    ) : (
                      <Mic className="w-14 h-14" />
                    )}

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
                  </button>
                  
                  <div className="text-center">
                    <p className="text-gray-400 text-sm">
                      {!connected 
                        ? 'Click to start recording' 
                        : isRecording 
                        ? 'Listening...' 
                        : 'Click to start recording'
                      }
                    </p>
                  </div>
                </div>
              </div>
            </>
          ) : (
            // Show project selector when no project is selected
            <div className="flex-1 px-4 sm:px-6 lg:px-8 py-8">
              <div className="max-w-7xl mx-auto">
                <ProjectSelector
                  projects={projects}
                  selectedProject={selectedProject}
                  loading={projectsLoading}
                  error={projectsError}
                  onSelectProject={handleSelectProject}
                  onRefresh={fetchProjects}
                  onCloneRepository={handleCloneRepository}
                />
              </div>
            </div>
          )}

        </main>

        {/* Developer/Debug Section - Only show if developer mode or there's additional data */}
        {(developerMode || geminiAdvice.length > 0 || claudePlanResponses.length > 0) && selectedProject && (
          <div className="bg-gray-900 px-4 py-6">
            <div className="max-w-7xl mx-auto">
              {/* Developer Mode Status */}
              {developerMode && (
                <div className="flex items-center space-x-4 bg-gray-800 px-4 py-3 rounded-lg border border-gray-600 mb-6">
                  {/* Connection Status */}
                  <div className="flex items-center space-x-2">
                    {connected ? (
                      <div className="flex items-center space-x-1 text-green-600">
                        <Wifi className="w-4 h-4" />
                        <span className="text-sm">Connected</span>
                      </div>
                    ) : (
                      <div className="flex items-center space-x-1 text-red-600">
                        <WifiOff className="w-4 h-4" />
                        <span className="text-sm">Disconnected</span>
                      </div>
                    )}
                  </div>
                  
                  {/* Recording Status */}
                  <div className="flex items-center space-x-2">
                    {isRecording ? (
                      <div className="flex items-center space-x-1 text-red-600">
                        <Mic className="w-4 h-4" />
                        <span className="text-sm">Recording</span>
                      </div>
                    ) : (
                      <div className="flex items-center space-x-1 text-gray-400">
                        <MicOff className="w-4 h-4" />
                        <span className="text-sm">Not Recording</span>
                      </div>
                    )}
                  </div>
                </div>
              )}

              <div className={`grid gap-8 ${geminiAdvice.length > 0 || claudePlanResponses.length > 0 ? 'grid-cols-1 xl:grid-cols-3' : 'grid-cols-1 lg:grid-cols-2'}`}>
                {/* Developer Mode */}
                {developerMode && (
                  <DeveloperMode 
                    selectedProject={selectedProject?.name || null} 
                    socket={socket}
                    connected={connected}
                  />
                )}

                {/* Claude Plans - Show if there are plans */}
                {claudePlanResponses.length > 0 && (
                  <ClaudePlan
                    plans={claudePlanResponses}
                    onClear={clearClaudePlanResponses}
                    onSavePlan={handleSavePlan}
                  />
                )}

                {/* Transcriptions - Developer Debug */}
                {developerMode && (
                  <TranscriptionView
                    transcriptions={transcriptions}
                    onClear={clearTranscriptions}
                  />
                )}

                {/* Function Results - Developer Debug */}
                {developerMode && (
                  <FunctionResults
                    results={functionResults}
                    onClear={clearFunctionResults}
                  />
                )}

                {/* Gemini Advice - Only show if there's advice */}
                {geminiAdvice.length > 0 && (
                  <GeminiAdvice
                    advice={geminiAdvice}
                    onClear={clearGeminiAdvice}
                  />
                )}
              </div>
            </div>
          </div>
        )}

      {/* Terminal Modal */}
      {showTerminal && (
        <Terminal
          localPath={terminalWorkingDir || (selectedProject?.localPath || `/Users/reed/Code/${selectedProject?.name}`)}
          claudePrompt={claudePrompt}
          onClose={() => {
            setShowTerminal(false);
            setClaudePrompt(undefined);
            setTerminalWorkingDir(undefined);
          }}
        />
      )}


      {/* Settings Modal */}
      <Settings
        isOpen={showSettings}
        onClose={() => setShowSettings(false)}
        onConfigUpdated={fetchProjects}
      />

      {/* Quiet Mode Overlay */}
      {showQuietOverlay && selectedProject && quietModeByRepo[selectedProject.fullName] && (
        <div className="fixed inset-0 bg-black bg-opacity-95 z-50 flex items-center justify-center">
          {/* Subtle message */}
          <div className="text-center">
            <div className="text-gray-400 text-lg mb-4">Quiet Mode</div>
            <div className="text-gray-500 text-sm">Move mouse to wake</div>
          </div>
          
          {/* Preserve microphone button access */}
          <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 z-60">
            <button
              onClick={handleMicrophoneClick}
              className={`w-32 h-32 rounded-full border-4 transition-all duration-200 flex items-center justify-center ${
                !connected
                  ? 'bg-blue-500 border-blue-600 text-white hover:bg-blue-600 shadow-lg'
                  : isRecording
                  ? 'bg-red-500 border-red-600 text-white shadow-lg transform scale-110'
                  : 'bg-gray-700 border-gray-600 text-gray-300 hover:bg-gray-600 hover:border-gray-500'
              }`}
              title={
                !connected 
                  ? 'Connect to Voice Assistant' 
                  : isRecording 
                  ? 'Stop Recording' 
                  : 'Start Recording'
              }
            >
              {!connected ? (
                <Mic className="w-14 h-14" />
              ) : isRecording ? (
                <Square className="w-10 h-10" />
              ) : (
                <Mic className="w-14 h-14" />
              )}

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
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;