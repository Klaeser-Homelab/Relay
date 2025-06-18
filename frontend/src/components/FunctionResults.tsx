import React from 'react';
import { Terminal, ExternalLink, Trash2, CheckCircle, XCircle, AlertCircle } from 'lucide-react';
import type { FunctionResultData } from '../types/api';

interface FunctionResultsProps {
  results: FunctionResultData[];
  onClear: () => void;
}

export function FunctionResults({ results, onClear }: FunctionResultsProps) {
  if (results.length === 0) {
    return (
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900 flex items-center">
            <Terminal className="w-5 h-5 mr-2" />
            Function Results
          </h3>
        </div>
        <div className="text-center py-8 text-gray-500">
          <Terminal className="w-12 h-12 mx-auto mb-3 text-gray-300" />
          <p>Function execution results will appear here</p>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900 flex items-center">
          <Terminal className="w-5 h-5 mr-2" />
          Function Results
        </h3>
        <button
          onClick={onClear}
          className="btn-secondary text-sm flex items-center"
        >
          <Trash2 className="w-4 h-4 mr-1" />
          Clear
        </button>
      </div>

      <div className="space-y-4 max-h-80 overflow-y-auto">
        {results.map((result, index) => (
          <FunctionResultItem
            key={index}
            result={result}
            timestamp={new Date()}
          />
        ))}
      </div>
    </div>
  );
}

interface FunctionResultItemProps {
  result: FunctionResultData;
  timestamp: Date;
}

function FunctionResultItem({ result, timestamp }: FunctionResultItemProps) {
  const formatTime = (date: Date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  const getFunctionIcon = (functionName: string) => {
    switch (functionName) {
      case 'create_github_issue':
      case 'update_github_issue':
      case 'close_github_issue':
        return '🐛';
      case 'list_issues':
        return '📋';
      case 'get_repository_info':
        return '📊';
      case 'list_commits':
        return '📝';
      case 'create_pull_request':
      case 'list_pull_requests':
        return '🔀';
      default:
        return '⚡';
    }
  };

  const getStatusIcon = (result: any) => {
    if (result?.success === true) {
      return <CheckCircle className="w-4 h-4 text-green-500" />;
    } else if (result?.success === false) {
      return <XCircle className="w-4 h-4 text-red-500" />;
    } else {
      return <AlertCircle className="w-4 h-4 text-blue-500" />;
    }
  };

  const renderResultData = (data: any) => {
    if (typeof data === 'string') {
      return <div className="text-gray-700">{data}</div>;
    }

    if (typeof data === 'object' && data !== null) {
      // Handle specific GitHub data types
      if (data.issues) {
        return (
          <div className="space-y-2">
            <div className="text-sm font-medium text-gray-700">
              {data.issues.length} issue(s) found:
            </div>
            {data.issues.slice(0, 3).map((issue: any, idx: number) => (
              <div key={idx} className="bg-gray-50 rounded p-2 text-sm">
                <div className="flex items-center justify-between">
                  <span className="font-medium">#{issue.number} {issue.title}</span>
                  <span className={`px-2 py-1 rounded-full text-xs ${
                    issue.state === 'open' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                  }`}>
                    {issue.state}
                  </span>
                </div>
                {issue.url && (
                  <a 
                    href={issue.url} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="text-blue-600 hover:text-blue-800 text-xs flex items-center mt-1"
                  >
                    View on GitHub <ExternalLink className="w-3 h-3 ml-1" />
                  </a>
                )}
              </div>
            ))}
            {data.issues.length > 3 && (
              <div className="text-xs text-gray-500">
                ... and {data.issues.length - 3} more
              </div>
            )}
          </div>
        );
      }

      if (data.commits) {
        return (
          <div className="space-y-2">
            <div className="text-sm font-medium text-gray-700">
              {data.commits.length} commit(s):
            </div>
            {data.commits.slice(0, 3).map((commit: any, idx: number) => (
              <div key={idx} className="bg-gray-50 rounded p-2 text-sm">
                <div className="font-mono text-xs text-gray-600">{commit.sha}</div>
                <div className="text-gray-800">{commit.message}</div>
                <div className="text-xs text-gray-500 mt-1">
                  by {commit.author} on {new Date(commit.date).toLocaleDateString()}
                </div>
                {commit.url && (
                  <a 
                    href={commit.url} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="text-blue-600 hover:text-blue-800 text-xs flex items-center mt-1"
                  >
                    View commit <ExternalLink className="w-3 h-3 ml-1" />
                  </a>
                )}
              </div>
            ))}
          </div>
        );
      }

      if (data.number && data.title) {
        // Single issue or PR
        return (
          <div className="bg-gray-50 rounded p-3">
            <div className="flex items-center justify-between mb-2">
              <span className="font-medium">#{data.number} {data.title}</span>
              {data.state && (
                <span className={`px-2 py-1 rounded-full text-xs ${
                  data.state === 'open' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                }`}>
                  {data.state}
                </span>
              )}
            </div>
            {data.url && (
              <a 
                href={data.url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="text-blue-600 hover:text-blue-800 text-sm flex items-center"
              >
                View on GitHub <ExternalLink className="w-4 h-4 ml-1" />
              </a>
            )}
          </div>
        );
      }

      // Generic object display
      return (
        <pre className="bg-gray-50 rounded p-3 text-xs overflow-x-auto">
          {JSON.stringify(data, null, 2)}
        </pre>
      );
    }

    return <div className="text-gray-500 italic">No data</div>;
  };

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4">
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center space-x-2">
          <span className="text-lg">{getFunctionIcon(result.function)}</span>
          <div>
            <div className="font-medium text-gray-900 flex items-center">
              {result.function.replace(/_/g, ' ')}
              {getStatusIcon(result.result)}
            </div>
            <div className="text-xs text-gray-500">
              {formatTime(timestamp)}
            </div>
          </div>
        </div>
      </div>

      <div className="mt-3">
        {renderResultData(result.result)}
      </div>
    </div>
  );
}