import React from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { Settings, Code, X, Moon, Plus, MessageSquare, Archive, GitCompare } from 'lucide-react'
import { useSidebar } from '../contexts/SidebarContext'
import { useConversation } from '../contexts/ConversationContext'

interface Conversation {
  id: string
  title: string
  repository_name?: string
  repository_full_name?: string
  updated_at: string
  last_accessed_at?: string
}

export function Sidebar() {
  const navigate = useNavigate()
  const location = useLocation()
  const {
    isOpen,
    closeSidebar,
    developerMode,
    quietMode,
    toggleDeveloperMode,
    toggleQuietMode,
    recentConversations,
    loadingConversations,
    openSettings,
    selectedProject
  } = useSidebar()
  
  const { conversationId, createNewConversation } = useConversation()
  
  const isComparePage = location.pathname === '/compare'
  
  if (!isOpen) {
    return null
  }
  
  const handleSelectConversation = (conversation: Conversation) => {
    closeSidebar()
    navigate(`/conversation/${conversation.id}`)
  }
  
  const handleNewConversation = async () => {
    try {
      const newId = await createNewConversation()
      closeSidebar()
      navigate(`/conversation/${newId}`)
    } catch (error) {
      console.error('Error creating new conversation:', error)
    }
  }
  
  const handleShowAllConversations = () => {
    closeSidebar()
    // Navigate to all conversations page when implemented
    console.log('Show all conversations')
  }
  
  const handleToggleMode = () => {
    closeSidebar()
    if (isComparePage) {
      navigate('/')
    } else {
      navigate('/compare')
    }
  }
  
  return (
    <>
      {/* Backdrop */}
      <div 
        className="fixed inset-0 bg-black bg-opacity-50 z-40" 
        onClick={closeSidebar}
      />
      
      {/* Sidebar */}
      <div className="fixed top-0 left-0 h-full bg-gray-950 text-gray-200 w-64 flex flex-col border-r border-gray-700 z-50 shadow-xl">
        {/* Header */}
        <div className="p-4 border-b border-gray-700">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className="grid grid-cols-2 gap-1 w-4 h-4">
                <div className="w-1.5 h-1.5 bg-white rounded-full"></div>
                <div className="w-1.5 h-1.5 bg-green-500 rounded-full"></div>
                <div className="w-1.5 h-1.5 bg-orange-500 rounded-full"></div>
                <div className="w-1.5 h-1.5 bg-blue-500 rounded-full"></div>
              </div>
              <h1 className="text-2xl font-bold text-white">Relay</h1>
            </div>
            <button
              onClick={closeSidebar}
              className="p-1 text-gray-400 hover:text-gray-200 rounded hover:bg-gray-800 transition-colors"
              title="Close sidebar"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>
        
        {/* Recents Section */}
        <div className="flex-1 p-4 overflow-y-auto">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-sm font-medium text-gray-400 uppercase tracking-wide">Recent Conversations</h3>
            <button
              onClick={handleNewConversation}
              className="p-1 text-gray-400 hover:text-gray-200 rounded hover:bg-gray-800 transition-colors"
              title="New conversation"
            >
              <Plus className="w-4 h-4" />
            </button>
          </div>
          
          {/* Recent Conversations List */}
          <div className="space-y-1 mb-3">
            {loadingConversations ? (
              <div className="text-center py-8 text-gray-500">
                <p className="text-sm">Loading conversations...</p>
              </div>
            ) : recentConversations.length > 0 ? (
              recentConversations.slice(0, 10).map((conversation) => (
                <button
                  key={conversation.id}
                  onClick={() => handleSelectConversation(conversation)}
                  className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors group ${
                    conversationId === conversation.id
                      ? 'bg-gray-800 text-white'
                      : 'text-gray-300 hover:text-white hover:bg-gray-800'
                  }`}
                >
                  <div className="flex items-start space-x-2">
                    <MessageSquare className="w-4 h-4 mt-0.5 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="font-medium truncate">{conversation.title}</div>
                      {conversation.repository_name && (
                        <div className="text-xs text-gray-500 truncate">{conversation.repository_name}</div>
                      )}
                    </div>
                  </div>
                </button>
              ))
            ) : (
              <div className="text-center py-8 text-gray-500">
                <MessageSquare className="w-8 h-8 mx-auto mb-2 opacity-50" />
                <p className="text-xs">No recent conversations</p>
                <p className="text-xs">Start a new conversation</p>
              </div>
            )}
          </div>
          
          {/* All Conversations Button */}
          {recentConversations.length > 0 && (
            <button
              onClick={handleShowAllConversations}
              className="w-full flex items-center space-x-2 px-3 py-2 rounded-md text-sm text-gray-400 hover:text-white hover:bg-gray-800 transition-colors"
            >
              <Archive className="w-4 h-4 flex-shrink-0" />
              <span>All conversations</span>
              {recentConversations.length > 10 && (
                <span className="text-xs text-gray-500">({recentConversations.length})</span>
              )}
            </button>
          )}
        </div>
        
        {/* Navigation */}
        <div className="px-4 pb-4">
          <nav className="space-y-2">
            <button
              onClick={handleToggleMode}
              className={`w-full flex items-center space-x-3 px-3 py-2 rounded-md transition-colors ${
                isComparePage
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-300 hover:text-white hover:bg-gray-800'
              }`}
            >
              <GitCompare className="w-5 h-5 flex-shrink-0" />
              <span>{isComparePage ? 'Single Mode' : 'Compare Mode'}</span>
            </button>
            
            <button
              onClick={toggleDeveloperMode}
              className={`w-full flex items-center space-x-3 px-3 py-2 rounded-md transition-colors ${
                developerMode
                  ? 'bg-orange-600 text-white'
                  : 'text-gray-300 hover:text-white hover:bg-gray-800'
              }`}
            >
              <Code className="w-5 h-5 flex-shrink-0" />
              <span>Dev Mode</span>
            </button>
            
            {/* Only show quiet mode when a project is selected */}
            {selectedProject && (
              <button
                onClick={toggleQuietMode}
                className={`w-full flex items-center space-x-3 px-3 py-2 rounded-md transition-colors ${
                  quietMode
                    ? 'bg-purple-600 text-white'
                    : 'text-gray-300 hover:text-white hover:bg-gray-800'
                }`}
              >
                <Moon className="w-5 h-5 flex-shrink-0" />
                <span>Quiet Mode</span>
              </button>
            )}
          </nav>
        </div>
        
        {/* Footer */}
        <div className="p-4 border-t border-gray-700">
          <button
            onClick={openSettings}
            className="w-full flex items-center space-x-3 px-3 py-2 rounded-md text-gray-300 hover:text-white hover:bg-gray-800 transition-colors"
          >
            <Settings className="w-5 h-5 flex-shrink-0" />
            <span>Settings</span>
          </button>
        </div>
      </div>
    </>
  )
}