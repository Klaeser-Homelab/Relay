from sqlalchemy import create_engine, Column, Integer, String, Float, Boolean, DateTime, DECIMAL, Text, ForeignKey
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
import os
from datetime import datetime

# Database URL
DATABASE_URL = "sqlite+aiosqlite:///./app.db"

# Create async engine
engine = create_async_engine(DATABASE_URL, echo=False)

# Create async session factory
async_session_maker = async_sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)

# Create base class for models
Base = declarative_base()


class Conversation(Base):
    """Conversation model"""
    __tablename__ = "conversations"

    id = Column(String(50), primary_key=True, index=True)  # e.g., "conv_1234567890_abc123"
    title = Column(String(200), nullable=False) # Title for the conversation, 
    repository_name = Column(String(200), nullable=False) # Repository name for the conversation, GitHub repo name (e.g., "rklaeser/PeoplePerson")
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)
    
    # Relationships
    chats = relationship("Chat", back_populates="conversation", cascade="all, delete-orphan")


class ModelConfig(Base):
    """Configuration for which model"""
    __tablename__ = "ModelConfig"

    id = Column(String(50), primary_key=True, index=True)  # e.g., "conv_1234567890_abc123"
    model_role = Column(String(200), nullable=False)
    model_name = Column(String(50), ForeignKey("models.name"), nullable=False, index=True)

    model_ref = relationship("Model")

class Chat(Base):
    """SQLAlchemy model for conversation exchanges (prompt + response)"""
    __tablename__ = "chats"

    id = Column(String(50), primary_key=True, index=True)  # e.g., "msg_1234567890_abc123"
    conversation_id = Column(String(50), ForeignKey("conversations.id"), nullable=False, index=True)
    timestamp = Column(DateTime, default=datetime.utcnow, nullable=False)
    
    # User prompt
    prompt = Column(Text, nullable=False)
    
    # Assistant response
    response = Column(Text, nullable=True)  # Nullable in case of failed API calls
    
    # Model and usage tracking
    model_name = Column(String(50), ForeignKey("models.name"), nullable=False, index=True)
    input_tokens = Column(Integer, nullable=False, default=0)
    output_tokens = Column(Integer, nullable=False, default=0)
    input_cost = Column(DECIMAL(10, 6), nullable=False, default=0)
    output_cost = Column(DECIMAL(10, 6), nullable=False, default=0)
    total_cost = Column(DECIMAL(10, 6), nullable=False, default=0)
    processing_time = Column(Float, nullable=False, default=0)
    success = Column(Boolean, nullable=False, default=False)
    error_message = Column(Text, nullable=True)  # Store error if API call failed
    
    # Relationships
    conversation = relationship("Conversation", back_populates="chats")
    model_ref = relationship("Model")


class Model(Base):
    """SQLAlchemy model for AI model pricing and metadata"""
    __tablename__ = "models"

    name = Column(String(100), primary_key=True, index=True)  # e.g., "gpt-4.1-mini"
    provider = Column(String(50), nullable=False)  # e.g., "openai", "anthropic"
    price_per_input_token = Column(DECIMAL(10, 8), nullable=False)  # Price per token
    price_per_output_token = Column(DECIMAL(10, 8), nullable=False)  # Price per token
    max_tokens = Column(Integer, nullable=True)  # Maximum context/output tokens
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)
    is_active = Column(Boolean, default=True, nullable=False)  # Can be used for requests
    
    # Relationships
    chats = relationship("Chat", back_populates="model_ref")


async def init_db():
    """Initialize the database by creating all tables and seed data"""
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    
    # Seed the database with initial data
    await seed_database()


async def seed_database():
    """Seed the database with initial model data"""
    async with async_session_maker() as session:
        try:
            # Check if models already exist
            from sqlalchemy import select
            existing_models = await session.execute(select(Model))
            if existing_models.first():
                print("Models already exist, skipping seed data")
                return
            
            # Create model entries
            import uuid
            
            nano_model = Model(
                name="gpt-4.1-nano",
                provider="openai",
                price_per_input_token=0.0000001,  # Example pricing
                price_per_output_token=0.0000002,
                max_tokens=4096,
                is_active=True
            )
            
            mini_model = Model(
                name="gpt-4.1-mini", 
                provider="openai",
                price_per_input_token=0.0000005,  # Example pricing
                price_per_output_token=0.000001,
                max_tokens=8192,
                is_active=True
            )
            
            session.add(nano_model)
            session.add(mini_model)
            
            # Create model config entries
            triage_config = ModelConfig(
                id=str(uuid.uuid4()),
                model_role="triage",
                model_name="gpt-4.1-nano"
            )
            
            planning_config = ModelConfig(
                id=str(uuid.uuid4()), 
                model_role="planning",
                model_name="gpt-4.1-mini"
            )
            
            session.add(triage_config)
            session.add(planning_config)
            
            await session.commit()
            print("Successfully seeded database with model data")
            
        except Exception as e:
            await session.rollback()
            print(f"Error seeding database: {e}")
            raise


async def get_db() -> AsyncSession:
    """Dependency to get database session"""
    async with async_session_maker() as session:
        try:
            yield session
        finally:
            await session.close()