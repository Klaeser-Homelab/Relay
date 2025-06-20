-- Relay Chat Database Schema
-- This script initializes the database schema for chat history and recents

-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Chats table - represents individual chat sessions with repositories
CREATE TABLE chats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    repository_name VARCHAR(255) NOT NULL,
    repository_full_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    active_issue_number INTEGER,
    active_issue_title TEXT,
    active_issue_url TEXT,
    quiet_mode BOOLEAN DEFAULT FALSE
);

-- Messages table - stores individual messages in chats
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 'user', 'openai', 'claude', 'claude_streaming', 'claude_todowrite'
    content TEXT NOT NULL,
    metadata JSONB, -- Store additional data like issue info, function results, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_messages_chat_id (chat_id),
    INDEX idx_messages_created_at (created_at)
);

-- Chat snapshots table - store periodic snapshots of chat state
CREATE TABLE chat_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    transcriptions JSONB,
    function_results JSONB,
    claude_streaming_texts JSONB,
    claude_todo_writes JSONB,
    repository_issues JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_snapshots_chat_id (chat_id),
    INDEX idx_snapshots_created_at (created_at)
);

-- Function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at on chats
CREATE TRIGGER update_chats_updated_at 
    BEFORE UPDATE ON chats 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to update last_accessed_at when messages are added
CREATE OR REPLACE FUNCTION update_chat_last_accessed()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE chats 
    SET last_accessed_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.chat_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to update last_accessed_at when messages are inserted
CREATE TRIGGER update_chat_last_accessed_trigger
    AFTER INSERT ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_chat_last_accessed();

-- Create indexes for better performance
CREATE INDEX idx_chats_repository_full_name ON chats(repository_full_name);
CREATE INDEX idx_chats_last_accessed_at ON chats(last_accessed_at DESC);
CREATE INDEX idx_chats_created_at ON chats(created_at DESC);

-- Sample data for development (optional)
-- INSERT INTO chats (title, repository_name, repository_full_name) 
-- VALUES ('Relay Project Voice-Enabled AI Planning', 'Relay', 'rklaeser/Relay');