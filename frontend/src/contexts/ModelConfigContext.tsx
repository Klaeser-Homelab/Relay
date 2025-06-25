import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { api } from '../config/api'

interface Model {
  name: string
  provider: string
  is_active: boolean
}

interface ModelConfigContextType {
  // Current configured models
  triageModel: Model | null
  planningModel: Model | null
  
  // All available models
  models: Model[]
  
  // State
  error: string | null
  
  // Actions
  fetchModelConfigs: () => Promise<void>
  updateModelConfig: (role: string, modelId: string) => Promise<void>
  getModelByRole: (role: string) => Model | null
  
  // Default model IDs (for use in API calls)
  defaultTriageModelId: string | null
  defaultPlanningModelId: string | null
}

const ModelConfigContext = createContext<ModelConfigContextType | undefined>(undefined)

export function useModelConfig(): ModelConfigContextType {
  const context = useContext(ModelConfigContext)
  if (!context) {
    throw new Error('useModelConfig must be used within a ModelConfigProvider')
  }
  return context
}

interface ModelConfigProviderProps {
  children: ReactNode
}

export function ModelConfigProvider({ children }: ModelConfigProviderProps) {
  const [triageModel, setTriageModel] = useState<Model | null>(null)
  const [planningModel, setPlanningModel] = useState<Model | null>(null)
  const [models, setModels] = useState<Model[]>([])
  const [error, setError] = useState<string | null>(null)
  
  // Default model names derived from current configurations
  const defaultTriageModelId = triageModel?.name || null
  const defaultPlanningModelId = planningModel?.name || null

  // Fetch all available models
  const fetchAllModels = async (): Promise<void> => {
    try {
      const response = await api.get('/models/')
      if (response.status !== 200) throw new Error('Failed to fetch models')
      
      const data = await response.data
      if (data) {
        setModels(data.models || [])
      }
    } catch (err) {
      console.error('Error fetching models:', err)
      setError(err instanceof Error ? err.message : 'Failed to fetch models')
    }
  }

  // Fetch model configurations from the model-configs endpoint
  const fetchModelConfigs = async (): Promise<void> => {
    try {
      setError(null)
      
      const response = await api.get('/model-configs/')
      if (response.status !== 200) throw new Error('Failed to fetch model configurations')
      
      const data = await response.data
      if (data.success && data.data) {
        // Set the configured models based on the response
        const configs = data.data
        
        if (configs.triage) {
          setTriageModel(configs.triage)
          console.log('Loaded triage model:', configs.triage.name)
        }
        
        if (configs.planning) {
          setPlanningModel(configs.planning)
          console.log('Loaded planning model:', configs.planning.name)
        }
        
        console.log('Model configurations loaded successfully')
      }
    } catch (err) {
      console.error('Error fetching model configurations:', err)
      setError(err instanceof Error ? err.message : 'Failed to fetch model configurations')
    }
  }

  // Update model configuration for a specific role
  const updateModelConfig = async (role: string, modelId: string): Promise<void> => {
    try {
      const response = await api.put(`/model-configs/${role}?model_id=${modelId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        }
      })
      
      if (response.status !== 200) throw new Error('Failed to update model configuration')
      
      const data = await response.data
      if (data.success && data.data.model) {
        // Update the appropriate model state
        if (role === 'triage') {
          setTriageModel(data.data.model)
          console.log('Updated triage model:', data.data.model.name)
        } else if (role === 'planning') {
          setPlanningModel(data.data.model)
          console.log('Updated planning model:', data.data.model.name)
        }
      }
    } catch (err) {
      console.error('Error updating model configuration:', err)
      setError(err instanceof Error ? err.message : 'Failed to update model configuration')
      throw err
    }
  }

  // Get model by role
  const getModelByRole = (role: string): Model | null => {
    switch (role) {
      case 'triage':
        return triageModel
      case 'planning':
        return planningModel
      default:
        return null
    }
  }


  // Initialize data on mount
  useEffect(() => {
    const initializeData = async () => {
      await Promise.all([
        fetchAllModels(),
        fetchModelConfigs()
      ])
    }
    
    initializeData()
  }, [])

  const value: ModelConfigContextType = {
    // Current configured models
    triageModel,
    planningModel,
    
    // All available models
    models,

    // State
    error,

    // Actions
    fetchModelConfigs,
    updateModelConfig,
    getModelByRole,
    
    // Default model IDs for API usage
    defaultTriageModelId,
    defaultPlanningModelId
  }

  return (
    <ModelConfigContext.Provider value={value}>
      {children}
    </ModelConfigContext.Provider>
  )
}