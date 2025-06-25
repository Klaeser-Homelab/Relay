from decimal import Decimal
from typing import Dict


class Settings:
    """Application settings and configuration"""
    PROJECT_NAME: str = "Agent API Example"
    VERSION: str = "1.0.0"
    
    # Model pricing per 1M tokens (in USD) - Current as of December 2024
    MODEL_PRICING: Dict[str, Dict[str, Decimal]] = {
        "gpt-4o": {
            "input": Decimal("2.50"),      # $2.50 per 1M input tokens
            "output": Decimal("10.00"),    # $10.00 per 1M output tokens
        },
        "gpt-4o-2024-11-20": {
            "input": Decimal("2.50"),      # $2.50 per 1M input tokens
            "output": Decimal("10.00"),    # $10.00 per 1M output tokens
        },
        "gpt-4o-2024-08-06": {
            "input": Decimal("2.50"),      # $2.50 per 1M input tokens
            "output": Decimal("10.00"),    # $10.00 per 1M output tokens
        },
        "gpt-4o-mini": {
            "input": Decimal("0.15"),      # $0.15 per 1M input tokens
            "output": Decimal("0.60"),     # $0.60 per 1M output tokens
        },
        "gpt-4o-mini-2024-07-18": {
            "input": Decimal("0.15"),      # $0.15 per 1M input tokens
            "output": Decimal("0.60"),     # $0.60 per 1M output tokens
        },
        "gpt-4-turbo": {
            "input": Decimal("10.00"),     # $10.00 per 1M input tokens
            "output": Decimal("30.00"),    # $30.00 per 1M output tokens
        },
        "gpt-4-turbo-preview": {
            "input": Decimal("10.00"),     # $10.00 per 1M input tokens
            "output": Decimal("30.00"),    # $30.00 per 1M output tokens
        },
        "gpt-4": {
            "input": Decimal("30.00"),     # $30.00 per 1M input tokens
            "output": Decimal("60.00"),    # $60.00 per 1M output tokens
        },
        "gpt-3.5-turbo": {
            "input": Decimal("0.50"),      # $0.50 per 1M input tokens
            "output": Decimal("1.50"),     # $1.50 per 1M output tokens
        },
        "gpt-3.5-turbo-16k": {
            "input": Decimal("3.00"),      # $3.00 per 1M input tokens
            "output": Decimal("4.00"),     # $4.00 per 1M output tokens
        },
        # Map common variations
        "gpt-4.1-mini": {  # This seems to be an alias for gpt-4o-mini
            "input": Decimal("0.40"),      # $0.15 per 1M input tokens
            "output": Decimal("1.60"),     # $0.60 per 1M output tokens
        },
        "gpt-4.1-nano": {  # This seems to be an alias for gpt-4o-mini
            "input": Decimal("0.10"),      # $0.15 per 1M input tokens
            "output": Decimal("0.40"),     # $0.60 per 1M output tokens
        },
        "o4-mini": {  # OpenAI o4-mini model
            "input": Decimal("0.05"),      # $0.05 per 1M input tokens
            "output": Decimal("0.20"),     # $0.20 per 1M output tokens
        },
        # Default pricing for unknown models
        "default": {
            "input": Decimal("1.00"),      # $1.00 per 1M input tokens
            "output": Decimal("2.00"),     # $2.00 per 1M output tokens
        }
    }
    
    def get_model_pricing(self, model: str) -> Dict[str, Decimal]:
        """Get pricing for a specific model, returns default if not found"""
        # Try exact match first
        if model in self.MODEL_PRICING:
            return self.MODEL_PRICING[model]
            
        # Handle model variants by checking prefixes
        for key in self.MODEL_PRICING.keys():
            if model.startswith(key) or key.startswith(model):
                return self.MODEL_PRICING[key]
                
        return self.MODEL_PRICING["default"]
    
    def get_price_per_token(self, model: str) -> Dict[str, Decimal]:
        """Get price per single token (not per million)"""
        pricing = self.get_model_pricing(model)
        return {
            "input": pricing["input"] / Decimal("1000000"),
            "output": pricing["output"] / Decimal("1000000")
        }


settings = Settings()