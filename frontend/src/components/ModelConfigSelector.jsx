import { useEffect, useState } from 'react';
import { ModelSelector } from './ModelSelector';
import { useModelConfig } from '../contexts/ModelConfigContext.tsx'


export function ModelConfigSelector() {
    const { 
        routingModel,
        planningModel,
        models
    } = useModelConfig()
    

    // Console log current model configuration once on mount
    useEffect(() => {
      console.log('Current Model Configuration:', {
          routing: routingModel,
          planning: planningModel,
          models: models
      })
  }, [routingModel, planningModel, models])

  return (
    <div className="space-y-4">
      <ModelSelector 
        role="routing" 
        label="routing Model"
      />
      <ModelSelector 
        role="planning" 
        label="Planning Model"
      />
    </div>
  );
}