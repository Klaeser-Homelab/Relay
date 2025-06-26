from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func, desc
from sqlalchemy.orm import selectinload
from app.core.database import Conversation, Chat
from app.models.conversation import ConversationCreate, MessageCreate, ConversationStats
from app.services.model_service import ModelService
from datetime import datetime
from typing import Optional, List, Tuple
from decimal import Decimal
import uuid
import time


class ConversationService:
    """Service for managing conversations and messages"""
    
    @staticmethod
    def generate_conversation_id() -> str:
        """Generate a unique conversation ID"""
        timestamp = int(time.time())
        random_suffix = uuid.uuid4().hex[:8]
        return f"conv_{timestamp}_{random_suffix}"
    
    @staticmethod
    def generate_chat_id() -> str:
        """Generate a unique chat ID"""
        timestamp = int(time.time())
        random_suffix = uuid.uuid4().hex[:8]
        return f"chat_{timestamp}_{random_suffix}"
    
    @staticmethod
    async def create_conversation(
        db: AsyncSession,
        conversation_data: ConversationCreate
    ) -> Conversation:
        """Create a new conversation"""
        conversation = Conversation(
            id=ConversationService.generate_conversation_id(),
            title=conversation_data.title,
            repository_name=conversation_data.repository_name,
            created_at=datetime.utcnow(),
            updated_at=datetime.utcnow()
        )
        
        db.add(conversation)
        await db.commit()
        
        # Explicitly load the conversation with messages relationship
        # to avoid lazy loading in async context
        query = select(Conversation).where(
            Conversation.id == conversation.id
        ).options(selectinload(Conversation.chats))
        
        result = await db.execute(query)
        return result.scalar_one()
    
    @staticmethod
    async def get_conversation(
        db: AsyncSession,
        conversation_id: str,
        include_messages: bool = True
    ) -> Optional[Conversation]:
        """Get a conversation by ID"""
        query = select(Conversation).where(Conversation.id == conversation_id)
        
        if include_messages:
            query = query.options(selectinload(Conversation.chats))
        
        result = await db.execute(query)
        return result.scalar_one_or_none()
    
    @staticmethod
    async def list_conversations(
        db: AsyncSession,
        page: int = 1,
        per_page: int = 20
    ) -> Tuple[List[Conversation], int]:
        """List conversations with pagination"""
        # Get total count
        count_query = select(func.count(Conversation.id))
        count_result = await db.execute(count_query)
        total = count_result.scalar() or 0
        
        # Get conversations
        query = (
            select(Conversation)
            .options(selectinload(Conversation.chats))
            .order_by(desc(Conversation.updated_at))
            .offset((page - 1) * per_page)
            .limit(per_page)
        )
        
        result = await db.execute(query)
        conversations = result.scalars().all()
        
        return conversations, total
    
    @staticmethod
    async def create_chat(
        db: AsyncSession,
        conversation_id: str,
        prompt: str,
        model_name: str,
        response: Optional[str] = None,
        input_tokens: int = 0,
        output_tokens: int = 0,
        processing_time: float = 0,
        success: bool = False,
        error_message: Optional[str] = None
    ) -> Chat:
        """Create a new chat (prompt-response) in a conversation"""
        
        # Calculate costs
        input_cost = Decimal("0")
        output_cost = Decimal("0")
        total_cost = Decimal("0")
        
        model = await ModelService.get_model(db, model_name)
        if model and success:
            input_cost = Decimal(input_tokens) * model.price_per_input_token
            output_cost = Decimal(output_tokens) * model.price_per_output_token
            total_cost = input_cost + output_cost
        
        chat = Chat(
            id=ConversationService.generate_chat_id(),
            conversation_id=conversation_id,
            timestamp=datetime.utcnow(),
            prompt=prompt,
            response=response,
            model_name=model_name,
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            input_cost=input_cost,
            output_cost=output_cost,
            total_cost=total_cost,
            processing_time=processing_time,
            success=success,
            error_message=error_message
        )
        
        db.add(chat)
        
        # Update conversation timestamp
        conversation = await ConversationService.get_conversation(db, conversation_id, include_messages=False)
        if conversation:
            conversation.updated_at = datetime.utcnow()
        
        await db.commit()
        await db.refresh(chat)
        return chat
    
    @staticmethod
    async def get_conversation_stats(
        db: AsyncSession,
        conversation_id: str
    ) -> Optional[ConversationStats]:
        """Get usage statistics for a conversation"""
        # Get exchange counts and usage stats in one query
        stats_query = select(
            func.count(Chat.id).label("total_exchanges"),
            func.sum(func.case([(Chat.success == True, 1)], else_=0)).label("successful_exchanges"),
            func.coalesce(func.sum(Chat.input_tokens), 0).label("total_input_tokens"),
            func.coalesce(func.sum(Chat.output_tokens), 0).label("total_output_tokens"),
            func.coalesce(func.sum(Chat.total_cost), 0).label("total_cost"),
            func.coalesce(func.avg(Chat.processing_time), 0).label("avg_processing_time")
        ).where(Chat.conversation_id == conversation_id)
        
        stats_result = await db.execute(stats_query)
        stats_row = stats_result.one()
        
        if not stats_row.total_exchanges:
            return None
        
        # Get models used
        models_query = select(
            Chat.model_name,
            func.count(Chat.id).label("count")
        ).where(Chat.conversation_id == conversation_id).group_by(Chat.model_name)
        
        models_result = await db.execute(models_query)
        models_used = {row.model_name: row.count for row in models_result}
        
        return ConversationStats(
            total_chats=stats_row.total_exchanges or 0,
            successful_chats=stats_row.successful_exchanges or 0,
            failed_chats=(stats_row.total_exchanges or 0) - (stats_row.successful_exchanges or 0),
            total_tokens=(stats_row.total_input_tokens or 0) + (stats_row.total_output_tokens or 0),
            input_tokens=stats_row.total_input_tokens or 0,
            output_tokens=stats_row.total_output_tokens or 0,
            total_cost=float(stats_row.total_cost or 0),
            avg_processing_time=float(stats_row.avg_processing_time or 0),
            models_used=models_used
        )
    
    @staticmethod
    async def delete_conversation(
        db: AsyncSession,
        conversation_id: str
    ) -> bool:
        """Delete a conversation and all its chats/usage records"""
        conversation = await ConversationService.get_conversation(db, conversation_id, include_messages=False)
        if not conversation:
            return False
        
        await db.delete(conversation)
        await db.commit()
        return True
    
