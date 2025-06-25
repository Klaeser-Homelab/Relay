import { useState } from 'react';
import { ConversationUsage } from './ConversationUsage';
import { ModelSelector } from './ModelSelector';

export function CompareMode({ 
  exchanges, 
  isLoading, 
  inputMessage, 
  setInputMessage, 
  sendMessage, 
  handleKeyPress,
  onViewTotal 
}) {
  const [leftModel, setLeftModel] = useState('gpt-4.1-mini');
  const [rightModel, setRightModel] = useState('gpt-4.1-nano');

  const leftConversations = exchanges.left || [];
  const rightConversations = exchanges.right || [];

  const ConversationChain = ({ conversations, model, side }) => (
    <div className="flex-1 flex flex-col h-full">
      <div className="mb-4">
        <h3 className="text-white text-lg font-semibold mb-2">{side} Model</h3>
        <ModelSelector 
          selectedModel={model} 
          onModelChange={side === 'Left' ? setLeftModel : setRightModel}
        />
        <ConversationUsage 
          conversationStats={null} // TODO: Calculate side-specific stats
          onViewTotal={onViewTotal}
        />
      </div>
      
      <div className="flex-1 overflow-y-auto space-y-4 pr-2">
        {conversations.map((conversation, idx) => (
          <div key={conversation.id || idx} className="space-y-2">
            {/* User prompt */}
            <div className="flex justify-end">
              <div className="max-w-2xl p-3 rounded-lg bg-blue-500 text-white">
                {conversation.prompt}
              </div>
            </div>
            {/* Assistant response */}
            {conversation.response && (
              <div className="flex justify-start">
                <div className={`max-w-2xl p-3 rounded-lg shadow ${
                  side === 'Left' 
                    ? 'bg-green-100 text-gray-800' 
                    : 'bg-purple-100 text-gray-800'
                }`}>
                  <div>{conversation.response}</div>
                  {/* Show conversation metadata */}
                  <div className="text-xs text-gray-600 mt-2 border-t pt-2">
                    Model: {conversation.model_id} | 
                    Tokens: {conversation.input_tokens + conversation.output_tokens} | 
                    Time: {conversation.processing_time?.toFixed(2)}s
                    {conversation.total_cost > 0 && ` | Cost: $${conversation.total_cost.toFixed(6)}`}
                  </div>
                </div>
              </div>
            )}
          </div>
        ))}
        {isLoading && (
          <div className="flex justify-start">
            <div className={`p-3 rounded-lg shadow ${
              side === 'Left' 
                ? 'bg-green-100 text-gray-800' 
                : 'bg-purple-100 text-gray-800'
            }`}>
              Thinking...
            </div>
          </div>
        )}
      </div>
    </div>
  );

  return (
    <div className="flex flex-col h-screen bg-gray-400">
      <div className="flex-1 flex gap-4 p-4">
        <ConversationChain 
          conversations={leftConversations} 
          model={leftModel} 
          side="Left"
        />
        
        {/* Divider */}
        <div className="w-px bg-gray-600"></div>
        
        <ConversationChain 
          conversations={rightConversations} 
          model={rightModel} 
          side="Right"
        />
      </div>
      
      <div className="p-4 bg-white border-t">
        <div className="flex gap-2">
          <textarea
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Type your message (will be sent to both models)..."
            className="flex-1 p-2 border border-gray-300 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
            rows="2"
          />
          <button
            onClick={() => sendMessage(leftModel, rightModel)}
            disabled={isLoading}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            Send to Both
          </button>
        </div>
      </div>
    </div>
  );
}