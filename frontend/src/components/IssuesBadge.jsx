import React from 'react';
import { useConversation } from '../contexts/ConversationContext';

const IssuesBadge = () => {
  const { selectedIssues, removeIssue } = useConversation();

  if (selectedIssues.length === 0) {
    return null;
  }

  return (
    <div className="flex flex-wrap gap-2 mb-4">
      {selectedIssues.map((issue) => (
        <span 
          key={issue.number} 
          className="inline-flex items-center px-3 py-1 rounded-full text-sm bg-blue-600 text-white"
        >
          <span className="font-mono mr-1">#{issue.number}:</span>
          <span className="truncate max-w-xs">{issue.title}</span>
          <button
            onClick={() => removeIssue(issue.number)}
            className="ml-2 text-blue-200 hover:text-white transition-colors"
            title="Remove issue"
          >
            ×
          </button>
        </span>
      ))}
    </div>
  );
};

export default IssuesBadge;