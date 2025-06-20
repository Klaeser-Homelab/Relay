import { useState } from 'react';

interface DeveloperModeProps {
  selectedProject: string | null;
  socket: any;
  connected: boolean;
}

const AVAILABLE_FUNCTIONS = [
  {
    name: 'ask_claude_to_make_plan',
    description: 'Ask Claude to make a plan (opens terminal with claude -p)',
    parameters: [
      { name: 'prompt', type: 'string', required: true, placeholder: 'Describe what you want Claude to plan for' },
      { name: 'workingDirectory', type: 'string', required: false, placeholder: 'Working directory (optional)' }
    ]
  },
  {
    name: 'get_implementation_advice',
    description: 'Get implementation advice from Gemini Flash',
    parameters: [
      { name: 'question', type: 'string', required: true, placeholder: 'How should I implement user authentication?' },
      { name: 'context', type: 'string', required: false, placeholder: 'Additional context (optional)' }
    ]
  },
  {
    name: 'create_github_issue',
    description: 'Create a new GitHub issue',
    parameters: [
      { name: 'title', type: 'string', required: true, placeholder: 'Issue title' },
      { name: 'body', type: 'string', required: false, placeholder: 'Issue description' },
      { name: 'labels', type: 'array', required: false, placeholder: 'bug,enhancement (comma-separated)' }
    ]
  },
  {
    name: 'list_issues',
    description: 'List GitHub issues',
    parameters: [
      { name: 'state', type: 'select', required: false, options: ['open', 'closed', 'all'], placeholder: 'Issue state' },
      { name: 'limit', type: 'number', required: false, placeholder: '10' }
    ]
  },
  {
    name: 'get_repository_info',
    description: 'Get repository information',
    parameters: []
  },
  {
    name: 'list_commits',
    description: 'List recent commits',
    parameters: [
      { name: 'limit', type: 'number', required: false, placeholder: '10' },
      { name: 'branch', type: 'string', required: false, placeholder: 'Branch name' }
    ]
  }
];

export function DeveloperMode({ selectedProject, socket, connected }: DeveloperModeProps) {
  const [selectedFunction, setSelectedFunction] = useState<string>('');
  const [parameters, setParameters] = useState<Record<string, string>>({});
  const [response, setResponse] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const currentFunction = AVAILABLE_FUNCTIONS.find(f => f.name === selectedFunction);

  const handleParameterChange = (paramName: string, value: string) => {
    setParameters(prev => ({
      ...prev,
      [paramName]: value
    }));
  };

  const executeFunction = async () => {
    if (!selectedFunction || !selectedProject || !socket) {
      return;
    }

    setLoading(true);
    setResponse(null);

    try {
      // Convert parameters to appropriate types
      const processedParams: Record<string, any> = {};
      
      currentFunction?.parameters.forEach(param => {
        const value = parameters[param.name];
        if (value) {
          if (param.type === 'number') {
            processedParams[param.name] = parseInt(value);
          } else if (param.type === 'array') {
            processedParams[param.name] = value.split(',').map(s => s.trim()).filter(s => s);
          } else {
            processedParams[param.name] = value;
          }
        }
      });

      // Listen for function result
      const handleFunctionResult = (data: any) => {
        if (data.function === selectedFunction) {
          setResponse(data.result);
          setLoading(false);
          socket.off('function_result', handleFunctionResult);
        }
      };

      socket.on('function_result', handleFunctionResult);

      // Simulate function call by emitting a test event
      socket.emit('test_function', {
        project: selectedProject,
        function: selectedFunction,
        args: processedParams
      });

      // Timeout after 30 seconds
      setTimeout(() => {
        setLoading(false);
        socket.off('function_result', handleFunctionResult);
      }, 30000);

    } catch (error) {
      console.error('Failed to execute function:', error);
      setLoading(false);
    }
  };

  if (!selectedProject) {
    return (
      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
        <p className="text-yellow-800">Please select a project first to use developer mode.</p>
      </div>
    );
  }

  return (
    <div className="bg-gray-50 border border-gray-200 rounded-lg p-6">
      <div className="flex items-center gap-2 mb-4">
        <div className="w-3 h-3 bg-orange-500 rounded-full"></div>
        <h3 className="text-lg font-semibold text-gray-900">Function Tester</h3>
        <span className="text-sm text-gray-500">Project: {selectedProject}</span>
      </div>

      <div className="space-y-4">
        {/* Function Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Select Function
          </label>
          <select
            value={selectedFunction}
            onChange={(e) => {
              setSelectedFunction(e.target.value);
              setParameters({});
              setResponse(null);
            }}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="">Choose a function...</option>
            {AVAILABLE_FUNCTIONS.map(func => (
              <option key={func.name} value={func.name}>
                {func.name} - {func.description}
              </option>
            ))}
          </select>
        </div>

        {/* Parameters */}
        {currentFunction && currentFunction.parameters.length > 0 && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Parameters
            </label>
            <div className="space-y-3">
              {currentFunction.parameters.map(param => (
                <div key={param.name}>
                  <label className="block text-xs font-medium text-gray-600 mb-1">
                    {param.name} {param.required && <span className="text-red-500">*</span>}
                  </label>
                  {param.type === 'select' ? (
                    <select
                      value={parameters[param.name] || ''}
                      onChange={(e) => handleParameterChange(param.name, e.target.value)}
                      className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
                    >
                      <option value="">{param.placeholder}</option>
                      {param.options?.map(option => (
                        <option key={option} value={option}>{option}</option>
                      ))}
                    </select>
                  ) : (
                    <input
                      type={param.type === 'number' ? 'number' : 'text'}
                      value={parameters[param.name] || ''}
                      onChange={(e) => handleParameterChange(param.name, e.target.value)}
                      placeholder={param.placeholder}
                      className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Execute Button */}
        <button
          onClick={executeFunction}
          disabled={!selectedFunction || !connected || loading}
          className={`w-full py-2 px-4 rounded-md font-medium transition-colors ${
            !selectedFunction || !connected || loading
              ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
              : 'bg-blue-600 hover:bg-blue-700 text-white'
          }`}
        >
          {loading ? 'Executing...' : 'Execute Function'}
        </button>

        {/* Response */}
        {response && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Response
            </label>
            <pre className="bg-gray-900 text-green-400 p-4 rounded-md text-sm overflow-auto max-h-96">
              {JSON.stringify(response, null, 2)}
            </pre>
          </div>
        )}

        {/* Connection Status */}
        <div className="text-xs text-gray-500">
          Status: {connected ? '🟢 Connected' : '🔴 Disconnected'}
        </div>
      </div>
    </div>
  );
}