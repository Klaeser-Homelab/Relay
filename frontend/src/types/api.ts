export interface GitHubRepository {
  name: string;
  fullName: string;
  path: string;
  description?: string;
  url: string;
  cloneUrl: string;
  defaultBranch: string;
  isPrivate: boolean;
  language?: string;
  lastOpened: string;
  stars: number;
  forks: number;
  isCloned?: boolean;
  localPath?: string;
}

export interface RepositoryStatus {
  name: string;
  fullName: string;
  path: string;
  url: string;
  description?: string;
  defaultBranch: string;
  isPrivate: boolean;
  language?: string;
  stars: number;
  forks: number;
  lastCommit?: {
    sha: string;
    message: string;
    author: string;
    date: string;
  };
  openIssues: number;
  updatedAt: string;
}

export interface VoiceSessionStatus {
  type: 'status';
  status: 'connected' | 'connecting' | 'disconnected' | 'recording' | 'processing' | 'error' | 'project_selected' | 'executing' | 'completed';
  message: string;
  project?: string;
}

export interface TranscriptionData {
  text: string;
}

export interface AudioResponseData {
  audio_data: string; // base64 encoded audio
}

export interface FunctionResultData {
  function: string;
  result: any;
}

export interface GeminiAdviceData {
  question: string;
  advice: string;
  repository: string;
}

export interface WebSocketMessage {
  type: 'status' | 'transcription' | 'audio_response' | 'function_result' | 'gemini_advice';
  data: VoiceSessionStatus | TranscriptionData | AudioResponseData | FunctionResultData | GeminiAdviceData;
}

export interface GitHubIssue {
  number: number;
  title: string;
  state: 'open' | 'closed';
  url: string;
  author: string;
  labels: string[];
  assignees: string[];
  createdAt: string;
  updatedAt: string;
}

export interface GitHubCommit {
  sha: string;
  message: string;
  author: string;
  date: string;
  url: string;
}

export interface GitHubPullRequest {
  number: number;
  title: string;
  state: 'open' | 'closed';
  url: string;
  author: string;
  head: string;
  base: string;
  createdAt: string;
  updatedAt: string;
}

export interface GitConfig {
  baseDirectory: string;
  gitUsername: string;
  hasToken: boolean;
}

export interface RepositoryCloneResult {
  success: boolean;
  message: string;
  localPath?: string;
}

export interface RepositoryInfo {
  success: boolean;
  data?: {
    localPath: string;
    isClean: boolean;
    currentBranch: string;
    lastCommit: string;
    hasUncommittedChanges: boolean;
  };
  message?: string;
}