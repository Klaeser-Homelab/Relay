import { useEffect, useState } from 'react';
import { ProjectSelector } from './components/ProjectSelector';
import { VoiceChat } from './components/VoiceChat';
import { StatusDisplay } from './components/StatusDisplay';
import { TranscriptionView } from './components/TranscriptionView';
import { FunctionResults } from './components/FunctionResults';
import { DeveloperMode } from './components/DeveloperMode';
import { useGitHubProjects } from './hooks/useGitHubProjects';
import { useWebSocket } from './hooks/useWebSocket';
import { useAudioRecording } from './hooks/useAudioRecording';

function App() {
  const [developerMode, setDeveloperMode] = useState(false);
  
  const {
    projects,
    selectedProject,
    loading: projectsLoading,
    error: projectsError,
    selectProject,
    fetchProjects
  } = useGitHubProjects();

  const {
    connected,
    status,
    transcriptions,
    functionResults,
    connect,
    disconnect,
    startRecording,
    stopRecording,
    sendAudio,
    selectProject: selectProjectWS,
    clearTranscriptions,
    clearFunctionResults,
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

  const handleSelectProject = async (project: any) => {
    const success = await selectProject(project);
    if (success && connected) {
      selectProjectWS(project.name);
    }
    return success;
  };

  const handleStartRecording = () => {
    startRecording();
    startAudioRecording();
  };

  const handleStopRecording = () => {
    stopRecording();
    stopAudioRecording();
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">
                Relay Voice
              </h1>
              <p className="text-gray-600">
                Voice-controlled GitHub repository management
              </p>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={() => setDeveloperMode(!developerMode)}
                className={`px-3 py-1 text-xs font-medium rounded-full transition-colors ${
                  developerMode
                    ? 'bg-orange-100 text-orange-800 border border-orange-200'
                    : 'bg-gray-100 text-gray-600 border border-gray-200 hover:bg-gray-200'
                }`}
              >
                {developerMode ? '🔧 Dev Mode ON' : '🔧 Dev Mode'}
              </button>
              <div className="text-sm text-gray-500">
                {projects.length} repositories
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left Column - Project Selection */}
          <div className="lg:col-span-2">
            <ProjectSelector
              projects={projects}
              selectedProject={selectedProject}
              loading={projectsLoading}
              error={projectsError}
              onSelectProject={handleSelectProject}
              onRefresh={fetchProjects}
            />
          </div>

          {/* Right Column - Voice Interface */}
          <div className="space-y-6">
            {/* Voice Chat */}
            <VoiceChat
              selectedProject={selectedProject}
              connected={connected}
              isRecording={isRecording}
              audioLevel={audioLevel}
              onStartRecording={handleStartRecording}
              onStopRecording={handleStopRecording}
              onConnect={connect}
              onDisconnect={disconnect}
            />

            {/* Status Display */}
            <StatusDisplay
              connected={connected}
              status={status}
              audioLevel={audioLevel}
              isRecording={isRecording}
            />
          </div>
        </div>

        {/* Developer Mode */}
        {developerMode && (
          <div className="mt-8">
            <DeveloperMode 
              selectedProject={selectedProject?.name || null} 
              socket={socket}
              connected={connected}
            />
          </div>
        )}

        {/* Bottom Section - Transcriptions and Results */}
        <div className="mt-8 grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Transcriptions */}
          <TranscriptionView
            transcriptions={transcriptions}
            onClear={clearTranscriptions}
          />

          {/* Function Results */}
          <FunctionResults
            results={functionResults}
            onClear={clearFunctionResults}
          />
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-200 mt-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center text-gray-500 text-sm">
            <p>
              Relay Voice - Powered by OpenAI Realtime API and GitHub
            </p>
            <div className="mt-2 flex items-center justify-center space-x-4">
              <span className={`w-2 h-2 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'}`}></span>
              <span>
                {connected ? 'Connected to voice server' : 'Disconnected from voice server'}
              </span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;