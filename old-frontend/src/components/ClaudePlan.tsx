import { X, Clock, Folder, MessageSquare, Save } from 'lucide-react';
import type { ClaudePlanResponseData } from '../types/api';

interface ClaudePlanProps {
  plans: ClaudePlanResponseData[];
  onClear: () => void;
  onSavePlan?: (plan: ClaudePlanResponseData) => void;
}

export function ClaudePlan({ plans, onClear, onSavePlan }: ClaudePlanProps) {
  if (plans.length === 0) {
    return null;
  }

  return (
    <div className="card">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-2">
          <MessageSquare className="w-5 h-5 text-blue-600" />
          <h3 className="text-lg font-semibold text-white">Claude Plans</h3>
          <span className="bg-blue-900 text-blue-200 text-xs font-medium px-2.5 py-0.5 rounded-full">
            {plans.length}
          </span>
        </div>
        <button
          onClick={onClear}
          className="p-1 text-gray-400 hover:text-gray-200 rounded-md hover:bg-gray-700"
          title="Clear plans"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      <div className="space-y-4 max-h-96 overflow-y-auto">
        {plans.map((plan, index) => (
          <div
            key={`${plan.timestamp}-${index}`}
            className="border border-gray-600 rounded-lg p-4 bg-gray-800"
          >
            {/* Plan Header */}
            <div className="flex items-start justify-between mb-3">
              <div className="flex-1 min-w-0">
                <div className="flex items-center space-x-2 mb-1">
                  <MessageSquare className="w-4 h-4 text-blue-500 flex-shrink-0" />
                  <h4 className="text-sm font-medium text-white truncate">
                    {plan.prompt}
                  </h4>
                </div>
                <div className="flex items-center space-x-4 text-xs text-gray-400">
                  <div className="flex items-center space-x-1">
                    <Clock className="w-3 h-3" />
                    <span>{new Date(plan.timestamp).toLocaleTimeString()}</span>
                  </div>
                  {plan.repository && (
                    <div className="flex items-center space-x-1">
                      <Folder className="w-3 h-3" />
                      <span className="truncate max-w-32">{plan.repository}</span>
                    </div>
                  )}
                </div>
              </div>
              {/* Save Plan Button */}
              {onSavePlan && (
                <button
                  onClick={() => onSavePlan(plan)}
                  className="ml-2 p-1.5 text-gray-400 hover:text-blue-400 rounded-md hover:bg-gray-700 transition-colors"
                  title="Save this plan as GitHub issue"
                >
                  <Save className="w-4 h-4" />
                </button>
              )}
            </div>

            {/* Plan Content */}
            <div className="bg-gray-900 rounded-md p-3 border border-gray-600">
              <div className="prose prose-sm max-w-none">
                <pre className="whitespace-pre-wrap text-sm text-gray-200 font-mono leading-relaxed">
                  {plan.plan}
                </pre>
              </div>
            </div>

            {/* Working Directory Info */}
            {plan.workingDirectory && (
              <div className="mt-2 text-xs text-gray-400">
                <span className="font-medium">Working Directory:</span> {plan.workingDirectory}
              </div>
            )}
          </div>
        ))}
      </div>

      {plans.length > 3 && (
        <div className="mt-4 text-center">
          <p className="text-sm text-gray-400">
            Showing {plans.length} plan{plans.length !== 1 ? 's' : ''}
          </p>
        </div>
      )}
    </div>
  );
}