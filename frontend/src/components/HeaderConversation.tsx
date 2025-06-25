import { Menu } from 'lucide-react'
import { useNavigate, useParams } from 'react-router-dom'
import React from 'react'
import { useConversation } from '../contexts/ConversationContext'

export const HeaderConversation = (props: { conversationId: string }) => {
  const navigate = useNavigate()
  const { conversationId } = props
  const { conversationData } = useConversation()
  
  const projectName = conversationData?.project_name || conversationData?.title?.split(' - ')[0] || 'Unknown Project'

  return (
<header className="p-4 bg-gray-900 border-b border-gray-700">
        <div className="flex justify-between items-center">
          <div className="flex items-center space-x-4">
            <button
              onClick={() => navigate('/')}
              className="text-gray-400 hover:text-gray-200 hover:bg-gray-700 rounded-md transition-colors p-2"
              title="Back to Projects"
            >
              <Menu className="w-5 h-5" />
            </button>
            
            <div>
              <h1 className="text-2xl font-bold text-white">
                {projectName}
              </h1>
              <p className="text-sm text-gray-400">Conversation ID: {conversationId}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
          <button
                onClick={() => navigate('/')}
                className="px-3 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 flex items-center gap-1"
                title="Start new conversation"
              >
                <span className="text-lg">+</span>
                New
              </button>
          </div>

        </div>
      </header>
    )
}
