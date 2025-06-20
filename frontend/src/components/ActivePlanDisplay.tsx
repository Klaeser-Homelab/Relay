import React, { useState } from 'react';
import { ChevronDown, ChevronUp, Calendar, User, GitBranch } from 'lucide-react';

interface Plan {
  type: string;
  version: string;
  timestamp: string;
  text: string;
  metadata: {
    author: string;
    issue_number: number;
    repository: string;
  };
  comment_id?: string;
  comment_created_at?: string;
  comment_author?: string;
}

interface ActivePlanDisplayProps {
  plan: Plan | null;
  issueNumber: number;
}

export const ActivePlanDisplay: React.FC<ActivePlanDisplayProps> = ({ plan, issueNumber }) => {
  const [isExpanded, setIsExpanded] = useState(false);

  if (!plan) {
    return (
      <div className="bg-gray-800 border border-gray-600 rounded-lg p-4 mb-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <GitBranch className="w-4 h-4 text-gray-400" />
            <span className="text-gray-300 text-sm">No plan available</span>
          </div>
          <span className="text-xs text-gray-500">Issue #{issueNumber}</span>
        </div>
        <p className="text-gray-400 text-xs mt-2">
          Click "Create Plan" to create an implementation plan for this issue.
        </p>
      </div>
    );
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const renderMarkdown = (text: string) => {
    // Simple markdown rendering for plans
    return text
      .replace(/^# (.*$)/gim, '<h1 class="text-lg font-bold text-white mb-3">$1</h1>')
      .replace(/^## (.*$)/gim, '<h2 class="text-base font-semibold text-white mb-2 mt-4">$1</h2>')
      .replace(/^### (.*$)/gim, '<h3 class="text-sm font-medium text-gray-200 mb-2 mt-3">$1</h3>')
      .replace(/^\*\*(.*?)\*\*/gim, '<strong class="font-semibold text-white">$1</strong>')
      .replace(/^\* (.*$)/gim, '<li class="text-gray-300 text-sm ml-4">• $1</li>')
      .replace(/^(\d+)\. (.*$)/gim, '<li class="text-gray-300 text-sm ml-4">$1. $2</li>')
      .replace(/^- (.*$)/gim, '<li class="text-gray-300 text-sm ml-6">- $1</li>')
      .replace(/\n/g, '<br />');
  };

  return (
    <div className="bg-gray-800 border border-gray-600 rounded-lg mb-4 overflow-hidden">
      {/* Header */}
      <div 
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-750 transition-colors"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center space-x-3">
          <GitBranch className="w-4 h-4 text-blue-400" />
          <div>
            <span className="text-white font-medium">Current Plan</span>
            <div className="flex items-center space-x-3 mt-1">
              <div className="flex items-center space-x-1">
                <Calendar className="w-3 h-3 text-gray-400" />
                <span className="text-xs text-gray-400">
                  {formatDate(plan.timestamp)}
                </span>
              </div>
              <div className="flex items-center space-x-1">
                <User className="w-3 h-3 text-gray-400" />
                <span className="text-xs text-gray-400">
                  {plan.comment_author || plan.metadata.author}
                </span>
              </div>
              <span className="text-xs text-gray-500">
                v{plan.version}
              </span>
            </div>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <span className="text-xs text-gray-500">Issue #{issueNumber}</span>
          {isExpanded ? (
            <ChevronUp className="w-4 h-4 text-gray-400" />
          ) : (
            <ChevronDown className="w-4 h-4 text-gray-400" />
          )}
        </div>
      </div>

      {/* Plan Content */}
      {isExpanded && (
        <div className="border-t border-gray-600 p-4">
          <div 
            className="prose prose-sm prose-invert max-w-none"
            dangerouslySetInnerHTML={{ 
              __html: renderMarkdown(plan.text) 
            }}
          />
          
          {/* Metadata */}
          <div className="mt-4 pt-3 border-t border-gray-700">
            <div className="flex flex-wrap gap-4 text-xs text-gray-500">
              <span>Plan ID: {plan.comment_id || 'N/A'}</span>
              <span>Repository: {plan.metadata.repository}</span>
              {plan.comment_created_at && (
                <span>Created: {formatDate(plan.comment_created_at)}</span>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};