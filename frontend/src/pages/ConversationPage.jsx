import { useParams, useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useConversation } from '../contexts/ConversationContext'
import Conversation from '../components/Conversation'
import { HeaderConversation } from '../components/HeaderConversation'
import Actions from '../components/Actions'

export function ConversationPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { setConversationId } = useConversation()

  useEffect(() => {
    if (id) {
      setConversationId(id)
    }
  }, [id, setConversationId])

  return (
    <div className="min-h-screen bg-gray-800 flex flex-col">
      {/* Simple header */}
      <HeaderConversation conversationId={id} />

      {/* Main conversation area */}
      <main className="flex-1 flex flex-col">
        <div className="flex-1 px-4 sm:px-6 lg:px-8 py-4 overflow-y-auto">
          <Conversation
            conversationId={id}
          />
        </div>

        {/* Actions component for message input */}
        <Actions conversationId={id} />
      </main>
    </div>
  )
}