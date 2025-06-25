import asyncio
from app.core.database import async_session_maker
from app.services.model_config_service import ModelConfigService

async def test_model_configs():
    async with async_session_maker() as db:
        configs = await ModelConfigService.get_models_by_roles(db)
        print('Model configs by role:', configs)
        
        # Test individual role lookup
        triage_config = await ModelConfigService.get_model_config_by_role(db, 'triage')
        planning_config = await ModelConfigService.get_model_config_by_role(db, 'planning')
        
        print('Triage config:', triage_config.model_role if triage_config else None, 
              triage_config.model_ref.name if triage_config and triage_config.model_ref else None)
        print('Planning config:', planning_config.model_role if planning_config else None,
              planning_config.model_ref.name if planning_config and planning_config.model_ref else None)

if __name__ == "__main__":
    asyncio.run(test_model_configs())