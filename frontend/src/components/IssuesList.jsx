import React, { useEffect } from 'react';
import { useConversation } from '../contexts/ConversationContext';

const IssuesList = () => {
  const { 
    repositoryIssues, 
    selectedIssues, 
    issuesLoading, 
    error,
    fetchRepositoryIssues,
    selectIssue,
    conversationData
  } = useConversation();

  // Auto-fetch issues for the conversation's repository
  useEffect(() => {
    if (conversationData?.repository_name) {
      fetchRepositoryIssues(conversationData.repository_name);
    }
  }, [fetchRepositoryIssues, conversationData?.repository_name]);

  if (issuesLoading) {
    return (
      <div className="text-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
        <p className="text-gray-400 mt-2">Loading repository issues...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <p className="text-red-400">Failed to load issues: {error}</p>
      </div>
    );
  }

  if (!conversationData?.repository_name) {
    return (
      <div className="text-center py-8">
        <p className="text-gray-400">No repository associated with this conversation.</p>
      </div>
    );
  }

  if (repositoryIssues.length === 0) {
    return (
      <div className="text-center py-8">
        <p className="text-gray-400">No open issues found in {conversationData.repository_name}.</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="text-center">
        <h3 className="text-xl text-gray-100 mb-2">What would you like to work on?</h3>
        <p className="text-gray-400 text-sm">Select an issue to get started</p>
      </div>
      
      <div className="space-y-3">
        {repositoryIssues.map((issue) => {
          const isSelected = selectedIssues.find(i => i.number === issue.number);
          
          return (
            <button
              key={issue.number}
              onClick={() => selectIssue(issue)}
              disabled={!!isSelected}
              className={`w-full text-left px-4 py-3 rounded-lg text-sm border transition-colors ${
                isSelected
                  ? 'bg-blue-900 border-blue-600 text-blue-200 cursor-not-allowed'
                  : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border-gray-600'
              }`}
            >
              <span className="font-mono text-blue-400">#{issue.number}</span>{' '}
              <span className="truncate">{issue.title}</span>
              {issue.labels.length > 0 && (
                <div className="mt-2">
                  {issue.labels.slice(0, 3).map((label) => (
                    <span 
                      key={label} 
                      className="inline-block bg-gray-700 text-gray-300 text-xs px-2 py-1 rounded mr-2"
                    >
                      {label}
                    </span>
                  ))}
                </div>
              )}
              {isSelected && (
                <div className="mt-2 text-xs text-blue-300">
                  ✓ Selected
                </div>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
};

export default IssuesList;