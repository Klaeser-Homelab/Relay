import { useModelConfig } from '../contexts/ModelConfigContext.tsx';

export function ModelSelector({ role, label }) {
  const { 
    getModelByRole,
    models, 
    updateModelConfig, 
    error 
  } = useModelConfig();

  const currentModel = getModelByRole(role);

  const handleModelChange = async (modelId) => {
    try {
      await updateModelConfig(role, modelId);
    } catch (err) {
      console.error(`Failed to update ${role} model:`, err);
    }
  };


  if (error) {
    return (
      <div className="bg-gray-800 border border-red-600 rounded-lg p-4 mb-4">
        <div className="flex items-center justify-between">
          <span className="text-white text-sm font-medium">{label || `${role} Model`}</span>
          <span className="text-red-400 text-sm">Error: {error}</span>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 mb-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <span className="text-white text-sm font-medium">{label || `${role} Model`}</span>
          {currentModel && (
            <span className="text-gray-400 text-xs">({currentModel.provider})</span>
          )}
        </div>
        <div className="flex items-center space-x-2">
          <select
            value={currentModel?.name || ''}
            onChange={(e) => handleModelChange(e.target.value)}
            className="bg-gray-700 border border-gray-600 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={models.length === 0}
          >
            {!currentModel && (
              <option value="">Select a model...</option>
            )}
            {models.map((model) => (
              <option key={model.name} value={model.name}>
                {model.name}
              </option>
            ))}
          </select>
        </div>
      </div>
    </div>
  );
}