import React, { useState } from 'react';
import { useConversation } from '../contexts/ConversationContext';
import { useActions } from '../contexts/ActionsContext';
import { useModelConfig } from '../contexts/ModelConfigContext.tsx';
import { api } from '../config/api';

const Actions = ({ conversationId }) => {
    const [inputMessage, setInputMessage] = useState('');
    const { isStreaming, addMessage, updateMessage, setIsStreaming, repository } = useConversation();
    const { setCurrentAction } = useActions();
    const { triageModel, planningModel } = useModelConfig();

    const handleKeyPress = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    };

    const sendMessage = async () => {
        if (!inputMessage.trim() || isStreaming || !conversationId) return;

        const prompt = inputMessage.trim();
        setCurrentAction({ type: 'send_text', message: prompt });
        setInputMessage('');
        setIsStreaming(true);
        
        // Create and immediately display user message
        const userMessageId = `temp_${Date.now()}`;
        const userMessage = {
            id: userMessageId,
            conversation_id: conversationId,
            prompt: prompt,
            response: null,
            success: true,
            timestamp: new Date().toISOString(),
            model_name: 'user',
            input_tokens: 0,
            output_tokens: 0,
            total_cost: 0,
            processing_time: 0
        };
        
        addMessage(userMessage);
        
        try {
            const response = await api.post(`/conversations/${conversationId}/run`, {
                prompt: prompt,
                triage_model: triageModel?.name,
                planning_model: planningModel?.name,
                repository_name: repository
            });
            
            // The response is the complete chat object with AI response
            const chat = response.data;
            
            // Remove the temporary user message and add the complete one
            updateMessage(userMessageId, chat);
            
        } catch (error) {
            console.error('Error sending message:', error);
            
            // Update the temporary user message with error details
            updateMessage(userMessageId, {
                error_message: error.response?.data?.detail || error.message || 'Failed to send message',
                success: false,
                model_name: 'error'
            });
        } finally {
            setIsStreaming(false);
        }
    };
    
    return (
        <div className="p-4 bg-gray-800 border-t border-gray-700">
            <div className="flex gap-2">
                <textarea
                    value={inputMessage}
                    onChange={(e) => setInputMessage(e.target.value)}
                    onKeyDown={handleKeyPress}
                    placeholder="Type your message... (Enter to send, Shift+Enter for new line)"
                    className="flex-1 p-3 bg-gray-700 text-white border border-gray-600 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 placeholder-gray-400"
                    rows="2"
                    disabled={isStreaming}
                />
                <button
                    onClick={sendMessage}
                    disabled={isStreaming || !inputMessage.trim() || !conversationId}
                    className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed transition-colors"
                >
                    {isStreaming ? 'Thinking...' : 'Send'}
                </button>
            </div>
        </div>
    );
};

export default Actions;