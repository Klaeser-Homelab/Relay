import { useEffect, useState } from 'react';
import { ModelSelector } from './ModelSelector';
import { useModelConfig } from '../contexts/ModelConfigContext.tsx'


export function ModelConfigSelector() {
    const { 
        triageModel,
        planningModel,
        models
    } = useModelConfig()
    

    // Console log current model configuration once on mount
    useEffect(() => {
      console.log('Current Model Configuration:', {
          triage: triageModel,
          planning: planningModel,
          models: models
      })
  }, [triageModel, planningModel, models])

  return (
    <div className="space-y-4">
      <ModelSelector 
        role="triage" 
        label="Triage Model"
      />
      <ModelSelector 
        role="planning" 
        label="Planning Model"
      />
    </div>
  );
}