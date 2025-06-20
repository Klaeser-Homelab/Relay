import { Chat, Message, ChatSnapshot } from './database.js';
import { v4 as uuidv4 } from 'uuid';

export class SessionManager {
  constructor() {
    this.activeSessions = new Map(); // Map of sessionId to chatId
  }

  /**
   * Create a new chat session
   */
  async createSession(repositoryName, repositoryFullName, title = null) {
    try {
      const chatTitle = title || `${repositoryName} - ${new Date().toLocaleDateString()}`;
      
      const chat = await Chat.create({
        id: uuidv4(),
        title: chatTitle,
        repository_name: repositoryName,
        repository_full_name: repositoryFullName,
        last_accessed_at: new Date()
      });

      console.log(`Created new chat session: ${chat.id} for ${repositoryFullName}`);
      return chat;
    } catch (error) {
      console.error('Failed to create chat session:', error);
      throw error;
    }
  }

  /**
   * Get a session by ID with all its messages
   */
  async getSession(chatId, includeMessages = true) {
    try {
      const options = includeMessages ? {
        include: [{
          model: Message,
          order: [['created_at', 'ASC']]
        }]
      } : {};

      const chat = await Chat.findByPk(chatId, options);
      
      if (chat) {
        // Update last accessed time
        await chat.update({ last_accessed_at: new Date() });
      }

      return chat;
    } catch (error) {
      console.error('Failed to get session:', error);
      throw error;
    }
  }

  /**
   * List sessions for a repository or all sessions
   */
  async listSessions(repositoryFullName = null, limit = 20) {
    try {
      const where = repositoryFullName ? { repository_full_name: repositoryFullName } : {};
      
      const sessions = await Chat.findAll({
        where,
        order: [['last_accessed_at', 'DESC']],
        limit
      });

      return sessions;
    } catch (error) {
      console.error('Failed to list sessions:', error);
      throw error;
    }
  }

  /**
   * Update session details
   */
  async updateSession(chatId, updates) {
    try {
      const chat = await Chat.findByPk(chatId);
      if (!chat) {
        throw new Error(`Chat session ${chatId} not found`);
      }

      await chat.update({
        ...updates,
        last_accessed_at: new Date()
      });

      return chat;
    } catch (error) {
      console.error('Failed to update session:', error);
      throw error;
    }
  }

  /**
   * Add a message to a session
   */
  async addMessage(chatId, type, content, metadata = null) {
    try {
      const message = await Message.create({
        id: uuidv4(),
        chat_id: chatId,
        type,
        content,
        metadata
      });

      // Update chat's last accessed time
      await Chat.update(
        { last_accessed_at: new Date() },
        { where: { id: chatId } }
      );

      console.log(`Added ${type} message to chat ${chatId}`);
      return message;
    } catch (error) {
      console.error('Failed to add message:', error);
      throw error;
    }
  }

  /**
   * Create a snapshot of the current conversation state
   */
  async createSnapshot(chatId, conversationState) {
    try {
      const snapshot = await ChatSnapshot.create({
        id: uuidv4(),
        chat_id: chatId,
        transcriptions: conversationState.transcriptions || [],
        function_results: conversationState.functionResults || [],
        claude_streaming_texts: conversationState.claudeStreamingTexts || [],
        claude_todo_writes: conversationState.claudeTodoWrites || [],
        repository_issues: conversationState.repositoryIssues || [],
        claude_session_id: conversationState.claudeSessionId || null
      });

      console.log(`Created snapshot for chat ${chatId}`);
      return snapshot;
    } catch (error) {
      console.error('Failed to create snapshot:', error);
      throw error;
    }
  }

  /**
   * Get the latest snapshot for a session
   */
  async getLatestSnapshot(chatId) {
    try {
      const snapshot = await ChatSnapshot.findOne({
        where: { chat_id: chatId },
        order: [['created_at', 'DESC']]
      });

      return snapshot;
    } catch (error) {
      console.error('Failed to get latest snapshot:', error);
      throw error;
    }
  }

  /**
   * Resume a session - returns the chat and its latest state
   */
  async resumeSession(chatId) {
    try {
      const chat = await this.getSession(chatId, true);
      if (!chat) {
        throw new Error(`Chat session ${chatId} not found`);
      }

      const snapshot = await this.getLatestSnapshot(chatId);
      
      return {
        chat,
        snapshot,
        messages: chat.Messages || []
      };
    } catch (error) {
      console.error('Failed to resume session:', error);
      throw error;
    }
  }

  /**
   * Delete a session and all its data
   */
  async deleteSession(chatId) {
    try {
      const result = await Chat.destroy({
        where: { id: chatId }
      });

      return result > 0;
    } catch (error) {
      console.error('Failed to delete session:', error);
      throw error;
    }
  }

  /**
   * Associate a WebSocket session with a chat
   */
  setActiveSession(socketSessionId, chatId) {
    this.activeSessions.set(socketSessionId, chatId);
  }

  /**
   * Get the chat ID for a WebSocket session
   */
  getActiveSession(socketSessionId) {
    return this.activeSessions.get(socketSessionId);
  }

  /**
   * Remove a WebSocket session association
   */
  removeActiveSession(socketSessionId) {
    this.activeSessions.delete(socketSessionId);
  }
}

// Export a singleton instance
export const sessionManager = new SessionManager();