import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { CompareMode } from '../components/CompareMode'
import { UsagePage } from '../components/UsagePage'
import { useConversation } from '../contexts/ConversationContext'
import { api } from '../config/api'

export function ComparePage() {
  const [conversations, setConversations] = useState({ left: [], right: [] })
  const [inputMessage, setInputMessage] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [showUsagePage, setShowUsagePage] = useState(false)
  const [conversationId, setConversationId] = useState(null)
  const navigate = useNavigate()

  // Helper function to create a new conversation
  const createNewConversation = async (routingModel = 'gpt-4.1-nano') => {
    try {
      const response = await api.post('/conversations/', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: `Conversation ${new Date().toLocaleString()}`,
          default_model: routingModel
        })
      })
      
      if (!response.ok) throw new Error('Failed to create conversation')
      
      const conversation = await response.json()
      setConversationId(conversation.id)
      console.log('Created new conversation:', conversation.id)
      return conversation.id
    } catch (error) {
      console.error('Error creating conversation:', error)
      // Generate a fallback ID if API fails
      const fallbackId = `local_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
      setConversationId(fallbackId)
      console.log('Using fallback conversation ID:', fallbackId)
      return fallbackId
    }
  }

  // Console log the current conversation's conversation list when it changes
  useEffect(() => {
    console.log('=== COMPARE MODE CONVERSATION LIST UPDATED ===');
    console.log('Left conversations:', conversations.left);
    console.log('Right conversations:', conversations.right);
    console.log('Total conversations in compare mode:', Math.max(conversations.left.length, conversations.right.length));
    console.log('===========================================');
  }, [conversations]);

  // Auto-create conversation when page loads
  useEffect(() => {
    createNewConversation()
  }, [])

  const sendMessage = async (leftModel, rightModel) => {
    if (!inputMessage.trim()) return

    const prompt = inputMessage
    setInputMessage('')
    setIsLoading(true)
    
    try {
      // Send to both models simultaneously
      const [leftResponse, rightResponse] = await Promise.all([
        api.post('/agent/run', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ 
            prompt, 
            routing_model: leftModel,
            weather_model: leftModel,
            conversation_id: conversationId 
          })
        }),
        api.post('/agent/run', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ 
            prompt, 
            routing_model: rightModel,
            weather_model: rightModel,
            conversation_id: conversationId 
          })
        })
      ])

      const [leftData, rightData] = await Promise.all([
        leftResponse.json(),
        rightResponse.json()
      ])

      // Use the complete Chat objects returned from the backend
      const leftConversation = {
        id: leftData.id,
        prompt: leftData.prompt,
        response: leftData.response || 'No response',
        model_id: leftData.model_id,
        input_tokens: leftData.input_tokens,
        output_tokens: leftData.output_tokens,
        total_cost: leftData.total_cost,
        processing_time: leftData.processing_time,
        success: leftData.success,
        timestamp: leftData.timestamp
      }

      const rightConversation = {
        id: rightData.id,
        prompt: rightData.prompt,
        response: rightData.response || 'No response',
        model_id: rightData.model_id,
        input_tokens: rightData.input_tokens,
        output_tokens: rightData.output_tokens,
        total_cost: rightData.total_cost,
        processing_time: rightData.processing_time,
        success: rightData.success,
        timestamp: rightData.timestamp
      }

      setConversations(prev => ({
        left: [...prev.left, leftConversation],
        right: [...prev.right, rightConversation]
      }))
    } catch (error) {
      console.error('Error:', error)
      const errorConversation = {
        id: Date.now().toString(),
        prompt,
        response: 'Error: Failed to get response',
        model_id: leftModel || rightModel,
        input_tokens: 0,
        output_tokens: 0,
        total_cost: 0,
        processing_time: 0,
        success: false,
        timestamp: new Date().toISOString()
      }
      setConversations(prev => ({
        left: [...prev.left, errorConversation],
        right: [...prev.right, errorConversation]
      }))
    } finally {
      setIsLoading(false)
    }
  }

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      // Note: This will be called from CompareMode component with models
    }
  }

  const handleStartNewConversation = async () => {
    setConversations({ left: [], right: [] })
    setInputMessage('')
    await createNewConversation()
    console.log('Started new conversation in compare mode:', conversationId)
  }

  if (showUsagePage) {
    return <UsagePage onBack={() => setShowUsagePage(false)} />
  }

  return (
    <div className="flex flex-col h-screen bg-gray-400">
      <div className="p-4 bg-gray-800 border-b border-gray-700">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-white text-xl font-bold">Compare Mode</h1>
            {conversationId && (
              <p className="text-gray-400 text-xs">ID: {conversationId.slice(0, 20)}...</p>
            )}
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={handleStartNewConversation}
              className="px-3 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 flex items-center gap-1"
              title="Start new conversation"
            >
              <span className="text-lg">+</span>
              New
            </button>
            <button
              onClick={() => navigate('/')}
              className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
            >
              Switch to Single Mode
            </button>
          </div>
        </div>
      </div>
      
      <CompareMode
        exchanges={conversations}
        isLoading={isLoading}
        inputMessage={inputMessage}
        setInputMessage={setInputMessage}
        sendMessage={sendMessage}
        handleKeyPress={handleKeyPress}
        onViewTotal={() => setShowUsagePage(true)}
      />
    </div>
  )
}