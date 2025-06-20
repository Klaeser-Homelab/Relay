const API_BASE = '/api';

export interface Session {
  id: string;
  title: string;
  repository_name: string;
  repository_full_name: string;
  created_at: string;
  updated_at: string;
  last_accessed_at: string;
  active_issue_number?: number;
  active_issue_title?: string;
  active_issue_url?: string;
  quiet_mode: boolean;
}

export interface SessionMessage {
  id: string;
  chat_id: string;
  type: string;
  content: string;
  metadata?: any;
  created_at: string;
}

export interface SessionSnapshot {
  id: string;
  chat_id: string;
  transcriptions: any[];
  function_results: any[];
  claude_streaming_texts: any[];
  claude_todo_writes: any[];
  repository_issues: any[];
  created_at: string;
}

export interface SessionData {
  chat: Session;
  snapshot?: SessionSnapshot;
  messages: SessionMessage[];
}

class SessionApi {
  async createSession(repositoryName: string, repositoryFullName: string, title?: string): Promise<Session> {
    const response = await fetch(`${API_BASE}/sessions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        repositoryName,
        repositoryFullName,
        title
      }),
    });

    const result = await response.json();
    if (!result.success) {
      throw new Error(result.error || 'Failed to create session');
    }

    return result.session;
  }

  async listSessions(repositoryFullName?: string): Promise<Session[]> {
    const params = repositoryFullName ? `?repository=${encodeURIComponent(repositoryFullName)}` : '';
    const response = await fetch(`${API_BASE}/sessions${params}`);

    const result = await response.json();
    if (!result.success) {
      throw new Error(result.error || 'Failed to list sessions');
    }

    return result.sessions;
  }

  async getSession(sessionId: string): Promise<Session> {
    const response = await fetch(`${API_BASE}/sessions/${sessionId}`);

    const result = await response.json();
    if (!result.success) {
      throw new Error(result.error || 'Failed to get session');
    }

    return result.session;
  }

  async updateSession(sessionId: string, updates: Partial<Session>): Promise<Session> {
    const response = await fetch(`${API_BASE}/sessions/${sessionId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updates),
    });

    const result = await response.json();
    if (!result.success) {
      throw new Error(result.error || 'Failed to update session');
    }

    return result.session;
  }

  async resumeSession(sessionId: string): Promise<SessionData> {
    const response = await fetch(`${API_BASE}/sessions/${sessionId}/resume`, {
      method: 'POST',
    });

    const result = await response.json();
    if (!result.success) {
      throw new Error(result.error || 'Failed to resume session');
    }

    return {
      chat: result.chat,
      snapshot: result.snapshot,
      messages: result.messages || []
    };
  }

  async deleteSession(sessionId: string): Promise<boolean> {
    const response = await fetch(`${API_BASE}/sessions/${sessionId}`, {
      method: 'DELETE',
    });

    const result = await response.json();
    return result.success;
  }
}

export const sessionApi = new SessionApi();