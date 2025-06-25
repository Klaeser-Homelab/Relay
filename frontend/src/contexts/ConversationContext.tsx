import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'
import { api } from '../config/api'

interface Message {
  id: string
  prompt: string
  response: string | null
  model_name: string
  input_tokens: number
  output_tokens: number
  total_cost: number
  processing_time: number
  timestamp: string
  success: boolean
  error_message?: string
}

interface ConversationData {
  id: string
  title: string
  project_name?: string
  repository_name?: string
  repository_full_name?: string
  created_at: string
  updated_at: string
  messages: Message[]
}

interface ConversationContextType {
  // Current conversation ID
  conversationId: string | null
  
  // Conversation data
  conversationData: ConversationData | null
  messages: Message[]
  
  // State
  loading: boolean
  error: string | null
  isStreaming: boolean
  
  // Actions
  setConversationId: (id: string | null) => void
  createNewConversation: (projectName?: string) => Promise<string>
  fetchConversation: (id: string) => Promise<void>
  loadConversation: (id: string) => Promise<void>
  addMessage: (message: Message) => void
  updateMessage: (id: string, updates: Partial<Message>) => void
  clearMessages: () => void
  setIsStreaming: (streaming: boolean) => void
}

const ConversationContext = createContext<ConversationContextType | undefined>(undefined)

export function useConversation(): ConversationContextType {
  const context = useContext(ConversationContext)
  if (!context) {
    throw new Error('useConversation must be used within a ConversationProvider')
  }
  return context
}

interface ConversationProviderProps {
  children: ReactNode
}

export function ConversationProvider({ children }: ConversationProviderProps) {
  const [conversationId, setConversationId] = useState<string | null>(null)
  const [conversationData, setConversationData] = useState<ConversationData | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isStreaming, setIsStreaming] = useState(false)

  // Load conversation ID from localStorage on mount
  useEffect(() => {
    const savedConversationId = localStorage.getItem('currentConversationId')
    if (savedConversationId) {
      setConversationId(savedConversationId)
    }
  }, [])

  // Save conversation ID to localStorage whenever it changes
  useEffect(() => {
    if (conversationId) {
      localStorage.setItem('currentConversationId', conversationId)
    } else {
      localStorage.removeItem('currentConversationId')
    }
  }, [conversationId])

  // Create a new conversation
  const createNewConversation = useCallback(async (projectName?: string): Promise<string> => {
    try {
      setLoading(true)
      setError(null)
      
      const payload: any = {
        title: projectName 
          ? `${projectName} - ${new Date().toLocaleString()}`
          : `Conversation ${new Date().toLocaleString()}`
      }
      
      if (projectName) {
        payload.project_name = projectName
      }
      
      const response = await api.post('/conversations/', payload)
      
      if (response.status !== 200) throw new Error('Failed to create conversation')
      
      const data = await response.data
      const newId = data.id
      setConversationId(newId)
      console.log('Created new conversation:', newId)
      return newId
    } catch (err) {
      console.error('Error creating conversation:', err)
      setError(err instanceof Error ? err.message : 'Failed to create conversation')
      throw err
    } finally {
      setLoading(false)
    }
  }, [])

  // Fetch conversation details
  const fetchConversation = useCallback(async (id: string): Promise<void> => {
    try {
      setLoading(true)
      setError(null)
      
      const response = await api.get(`/conversations/${id}`)
      if (response.status !== 200) throw new Error('Failed to fetch conversation')
      
      const data = await response.data
      console.log('Fetched conversation:', id, data)
    } catch (err) {
      console.error('Error fetching conversation:', err)
      setError(err instanceof Error ? err.message : 'Failed to fetch conversation')
      throw err
    } finally {
      setLoading(false)
    }
  }, [])

  // Load full conversation data including messages
  const loadConversation = useCallback(async (id: string): Promise<void> => {
    try {
      setLoading(true)
      setError(null)
      
      // Fetch conversation with messages included (default behavior)
      const response = await api.get(`/conversations/${id}?include_messages=true`)
      
      if (response.status !== 200) throw new Error('Failed to fetch conversation')
      
      const conversationData = await response.data
      
      // Extract messages from the conversation data
      const messageList = conversationData.messages || []
      
      const fullConversationData: ConversationData = {
        ...conversationData,
        messages: messageList
      }
      
      setConversationData(fullConversationData)
      setMessages(messageList)
      setConversationId(id)
      
      console.log('Loaded conversation:', fullConversationData)
    } catch (err) {
      console.error('Error loading conversation:', err)
      setError(err instanceof Error ? err.message : 'Failed to load conversation')
      throw err
    } finally {
      setLoading(false)
    }
  }, [])

  // Add a message to the conversation
  const addMessage = useCallback((message: Message) => {
    setMessages(prev => [...prev, message])
    
    // Update conversation data if it exists
    if (conversationData) {
      setConversationData(prev => prev ? {
        ...prev,
        messages: [...prev.messages, message]
      } : null)
    }
  }, [conversationData])

  // Update an existing message
  const updateMessage = useCallback((id: string, updates: Partial<Message>) => {
    setMessages(prev => prev.map(msg => 
      msg.id === id ? { ...msg, ...updates } : msg
    ))
    
    // Update conversation data if it exists
    if (conversationData) {
      setConversationData(prev => prev ? {
        ...prev,
        messages: prev.messages.map(msg => 
          msg.id === id ? { ...msg, ...updates } : msg
        )
      } : null)
    }
  }, [conversationData])

  // Clear all messages
  const clearMessages = useCallback(() => {
    setMessages([])
    if (conversationData) {
      setConversationData(prev => prev ? {
        ...prev,
        messages: []
      } : null)
    }
  }, [conversationData])

  const value: ConversationContextType = {
    // Current conversation ID
    conversationId,
    
    // Conversation data
    conversationData,
    messages,
    
    // State
    loading,
    error,
    isStreaming,
    
    // Actions
    setConversationId,
    createNewConversation,
    fetchConversation,
    loadConversation,
    addMessage,
    updateMessage,
    clearMessages,
    setIsStreaming
  }

  return (
    <ConversationContext.Provider value={value}>
      {children}
    </ConversationContext.Provider>
  )
}