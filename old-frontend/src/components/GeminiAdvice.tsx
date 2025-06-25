import { useState } from 'react';
import type { GeminiAdviceData } from '../types/api';

interface GeminiAdviceProps {
  advice: GeminiAdviceData[];
  onClear: () => void;
}

export function GeminiAdvice({ advice, onClear }: GeminiAdviceProps) {
  const [expandedIndex, setExpandedIndex] = useState<number | null>(null);

  const toggleExpanded = (index: number) => {
    setExpandedIndex(expandedIndex === index ? null : index);
  };

  const formatAdvice = (text: string) => {
    // Simple formatting for code blocks and sections
    return text
      .replace(/```(\w+)?\n([\s\S]*?)```/g, '<pre class="bg-gray-800 text-green-400 p-3 rounded text-sm overflow-x-auto my-2"><code>$2</code></pre>')
      .replace(/\*\*(.*?)\*\*/g, '<strong class="font-semibold">$1</strong>')
      .replace(/\*(.*?)\*/g, '<em class="italic">$1</em>')
      .replace(/\n\n/g, '</p><p class="mb-2">')
      .replace(/\n/g, '<br/>');
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200">
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-purple-500 rounded-full"></div>
            <h3 className="text-lg font-semibold text-gray-900">
              Implementation Advice
            </h3>
            <span className="px-2 py-1 text-xs font-medium bg-purple-100 text-purple-800 rounded-full">
              Powered by Gemini Flash
            </span>
          </div>
          {advice.length > 0 && (
            <button
              onClick={onClear}
              className="text-sm text-gray-500 hover:text-gray-700 transition-colors"
            >
              Clear
            </button>
          )}
        </div>
      </div>

      <div className="p-6">
        {advice.length === 0 ? (
          <div className="text-center py-8">
            <div className="w-12 h-12 mx-auto mb-4 bg-purple-100 rounded-full flex items-center justify-center">
              <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
              </svg>
            </div>
            <p className="text-gray-500">
              Ask implementation questions like "How should I implement authentication?" to get expert advice from Gemini Flash.
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            {advice.map((item, index) => (
              <div
                key={index}
                className="border border-purple-200 rounded-lg overflow-hidden"
              >
                <div 
                  className="px-4 py-3 bg-purple-50 cursor-pointer hover:bg-purple-100 transition-colors"
                  onClick={() => toggleExpanded(index)}
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-purple-900 mb-1">
                        {item.question}
                      </h4>
                      <p className="text-sm text-purple-600">
                        Repository: {item.repository}
                      </p>
                    </div>
                    <svg 
                      className={`w-5 h-5 text-purple-600 transition-transform ${
                        expandedIndex === index ? 'rotate-180' : ''
                      }`} 
                      fill="none" 
                      stroke="currentColor" 
                      viewBox="0 0 24 24"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                    </svg>
                  </div>
                </div>
                
                {expandedIndex === index && (
                  <div className="p-4 bg-white border-t border-purple-200">
                    <div 
                      className="prose prose-sm max-w-none text-gray-700 leading-relaxed"
                      dangerouslySetInnerHTML={{ 
                        __html: `<p class="mb-2">${formatAdvice(item.advice)}</p>` 
                      }}
                    />
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}