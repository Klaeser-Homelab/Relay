import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { ConversationUsage } from './ConversationUsage'
import { ModelConfigSelector } from './ModelConfigSelector'
import { useConversation } from '../contexts/ConversationContext'
import IssuesList from './IssuesList'

const Conversation = ({ conversationId }) => {
    const navigate = useNavigate()
    const { 
        messages: contextMessages, 
        conversationData, 
        loading: contextLoading, 
        loadConversation,
        selectedIssue 
    } = useConversation()
    
    // Auto-load conversation once on mount if conversationId is provided
    useEffect(() => {
        if (conversationId) {
            console.log('Auto-loading conversation:', conversationId)
            loadConversation(conversationId).catch(console.error)
        }
        else {
            throw new Error('No conversationId provided')
        }
    }, [])

    return (
        <div className="space-y-4">
            {conversationId && (
                <div className="text-xs text-gray-500 bg-gray-100 p-2 rounded">
                    Conversation ID: {conversationId}
                </div>
            )}
            <ConversationUsage 
                conversationId={conversationId}
                onViewTotal={() => navigate('/usage')}
            />
            <ModelConfigSelector />
            
            {/* Show issues list when no messages exist and no issues are selected */}
            {(!contextMessages || contextMessages.length === 0) && !selectedIssue && (
                <IssuesList />
            )}
            
            {contextMessages && contextMessages.map((conversation, idx) => (
                <div key={conversation.id || idx} className="space-y-2">
                    {/* User prompt */}
                    <div className="flex justify-end">
                        <div className="max-w-2xl p-3 rounded-lg bg-blue-500 text-white">
                            {conversation.prompt}
                        </div>
                    </div>
                    {/* Assistant response or error */}
                    {conversation.response && (
                        <div className="flex justify-start">
                            <div className="max-w-2xl p-3 rounded-lg bg-white text-gray-800 shadow">
                                <div>{conversation.response}</div>
                                {/* Show conversation metadata */}
                                <div className="text-xs text-gray-500 mt-2 border-t pt-2">
                                    Model: {conversation.model_name} | 
                                    Tokens: {conversation.input_tokens + conversation.output_tokens} | 
                                    Time: {conversation.processing_time?.toFixed(2)}s
                                    {conversation.total_cost > 0 && ` | Cost: $${conversation.total_cost.toFixed(6)}`}
                                </div>
                            </div>
                        </div>
                    )}
                    {/* Show error message if chat failed */}
                    {!conversation.success && conversation.error_message && (
                        <div className="flex justify-start">
                            <div className="max-w-2xl p-3 rounded-lg bg-red-100 text-red-800 border border-red-300">
                                <div className="font-medium text-sm">Error:</div>
                                <div>{conversation.error_message}</div>
                                {/* Show conversation metadata for failed chats */}
                                <div className="text-xs text-red-600 mt-2 border-t border-red-200 pt-2">
                                    Model: {conversation.model_name} | 
                                    Time: {conversation.processing_time?.toFixed(2)}s
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            ))}
            
        </div>
    );
};

export default Conversation;