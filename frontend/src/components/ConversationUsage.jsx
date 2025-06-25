import { useState } from 'react';

export function ConversationUsage({ conversationStats, onViewTotal }) {
  if (!conversationStats) {
    return (
      <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 mb-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center space-x-2">
            <span className="text-white text-sm font-medium">This Conversation</span>
          </div>
          <button
            onClick={onViewTotal}
            className="text-blue-400 hover:text-blue-300 text-xs px-2 py-1 rounded hover:bg-gray-700 transition-colors"
          >
            View Total Usage
          </button>
        </div>
        
        <div className="text-gray-400 text-sm">No messages in this conversation yet</div>
      </div>
    );
  }

  const formatCost = (cost) => {
    return `$${Number(cost).toFixed(6)}`;
  };

  const formatTime = (seconds) => {
    return `${Number(seconds).toFixed(2)}s`;
  };

  const formatNumber = (num) => {
    return Number(num).toLocaleString();
  };

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 mb-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center space-x-2">
          <span className="text-white text-sm font-medium">This Conversation</span>
        </div>
        <button
          onClick={onViewTotal}
          className="text-blue-400 hover:text-blue-300 text-xs px-2 py-1 rounded hover:bg-gray-700 transition-colors"
        >
          View Total Usage
        </button>
      </div>
      
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-gray-700 rounded-lg p-3">
          <div className="flex items-center space-x-2 mb-1">
            <span className="text-gray-300 text-xs">Messages</span>
          </div>
          <div className="text-white text-lg font-semibold">
            {formatNumber(conversationStats.messages)}
          </div>
          <div className="text-gray-400 text-xs">
            Total exchanges
          </div>
        </div>

        <div className="bg-gray-700 rounded-lg p-3">
          <div className="flex items-center space-x-2 mb-1">
            <span className="text-gray-300 text-xs">Tokens</span>
          </div>
          <div className="text-white text-lg font-semibold">
            {formatNumber(conversationStats.totalTokens)}
          </div>
          <div className="text-gray-400 text-xs">
            {formatNumber(conversationStats.inputTokens)} in / {formatNumber(conversationStats.outputTokens)} out
          </div>
        </div>

        <div className="bg-gray-700 rounded-lg p-3">
          <div className="flex items-center space-x-2 mb-1">
            <span className="text-gray-300 text-xs">Cost</span>
          </div>
          <div className="text-white text-lg font-semibold">
            {formatCost(conversationStats.totalCost)}
          </div>
          <div className="text-gray-400 text-xs">
            This session
          </div>
        </div>

        <div className="bg-gray-700 rounded-lg p-3">
          <div className="flex items-center space-x-2 mb-1">
            <span className="text-gray-300 text-xs">Avg Time</span>
          </div>
          <div className="text-white text-lg font-semibold">
            {formatTime(conversationStats.avgTime)}
          </div>
          <div className="text-gray-400 text-xs">
            Per message
          </div>
        </div>
      </div>

      {conversationStats.models && Object.keys(conversationStats.models).length > 1 && (
        <div className="mt-3 pt-3 border-t border-gray-600">
          <div className="text-gray-300 text-xs mb-2">Model breakdown:</div>
          <div className="flex flex-wrap gap-2">
            {Object.entries(conversationStats.models).map(([model, count]) => (
              <span key={model} className="bg-gray-600 text-gray-200 text-xs px-2 py-1 rounded">
                {model}: {count} msgs
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}