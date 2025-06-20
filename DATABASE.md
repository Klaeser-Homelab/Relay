# Relay Database Setup

This document describes how to set up and manage the PostgreSQL database for the Relay project.

## Quick Start

### Development Setup

1. **Start the development database:**
   ```bash
   make dev-db
   ```

2. **Connect to the database:**
   ```bash
   make dev-db-connect
   ```

3. **View database logs:**
   ```bash
   make dev-db-logs
   ```

4. **Stop the database:**
   ```bash
   make dev-db-stop
   ```

### Production Setup

1. **Set required environment variables:**
   ```bash
   export POSTGRES_PASSWORD=your_secure_password
   export OPENAI_API_KEY=your_openai_key
   export GH_TOKEN=your_github_token
   ```

2. **Start production stack:**
   ```bash
   make prod-db-full
   ```

## Database Schema

The database consists of three main tables:

### `chats` Table
Stores individual chat sessions with repositories.

```sql
CREATE TABLE chats (
    id UUID PRIMARY KEY,
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
```

### `messages` Table
Stores individual messages within chats.

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 'user', 'openai', 'claude', etc.
    content TEXT NOT NULL,
    metadata JSONB, -- Additional data (issue info, function results, etc.)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### `chat_snapshots` Table
Stores periodic snapshots of chat state for restoration.

```sql
CREATE TABLE chat_snapshots (
    id UUID PRIMARY KEY,
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    transcriptions JSONB,
    function_results JSONB,
    claude_streaming_texts JSONB,
    claude_todo_writes JSONB,
    repository_issues JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Environment Variables

### Development
- `DATABASE_URL`: `postgresql://relay_user:relay_dev_password@localhost:5432/relay_dev`

### Production
- `POSTGRES_PASSWORD`: Secure password for the database user
- `POSTGRES_DB`: Database name (default: `relay`)
- `POSTGRES_USER`: Database user (default: `relay_user`)
- `POSTGRES_PORT`: Database port (default: `5432`)
- `DATABASE_URL`: Full connection string

## Docker Configuration

### Development (`docker-compose.dev.yml`)
- **Purpose**: Local development with just PostgreSQL
- **Database**: `relay_dev`
- **Port**: `5432`
- **User**: `relay_user`
- **Password**: `relay_dev_password`

### Production (`docker-compose.prod.yml`)
- **Purpose**: Full production stack with PostgreSQL + Application
- **Database**: Configurable via environment variables
- **Security**: Uses environment variables for credentials
- **Dependencies**: Application waits for database health check

## Database Management Commands

### Backup & Restore
```bash
# Create backup
make db-backup

# Restore from backup
make db-restore BACKUP_FILE=backups/relay_dev_20231220_143022.sql

# Dump schema only
make db-schema-dump
```

### Database Operations
```bash
# Reset database (WARNING: destroys all data)
make dev-db-reset

# Check service status
make status

# Create environment template
make env-template
```

## Features Implemented

### Chat Management
- **Recents**: Sidebar shows recent chat sessions like Claude Desktop
- **Persistence**: Chat history is stored in PostgreSQL
- **Per-Repository**: Each repository can have multiple chat sessions
- **Metadata**: Stores active issues, quiet mode settings, etc.

### Message Storage
- **All Message Types**: User inputs, AI responses, function results, etc.
- **Metadata**: Rich context stored as JSONB for flexibility
- **Chronological**: Proper ordering and timestamps

### State Snapshots
- **Periodic Saves**: Chat state snapshots for complex restoration
- **Full Context**: Transcriptions, function results, todo writes, etc.
- **Performance**: Efficient querying with proper indexes

## Architecture

### Frontend Integration
- **Recents Sidebar**: Shows chat history like Claude Desktop
- **Plus Button**: Header button to start new chats
- **State Management**: React state synced with database

### Backend API
- **Chat CRUD**: Create, read, update, delete chat sessions
- **Message Persistence**: Store all conversation messages
- **Snapshot Management**: Periodic state saves

### Database Design
- **Relational**: Proper foreign key relationships
- **Scalable**: Indexed for performance
- **Flexible**: JSONB for evolving metadata requirements

## Security

### Development
- **Isolated**: Development database separate from production
- **Local**: Runs locally via Docker with known credentials

### Production
- **Environment Variables**: All credentials via environment
- **Docker Secrets**: Can be extended to use Docker secrets
- **Network Isolation**: Database only accessible within Docker network

## Monitoring

### Health Checks
- **Database**: Built-in PostgreSQL health checks
- **Application**: Application waits for database readiness

### Logging
- **Docker Logs**: `make dev-db-logs` for database logs
- **Application Logs**: Standard Docker logging for app container

## Troubleshooting

### Common Issues

1. **Port Already in Use**: If port 5432 is occupied
   ```bash
   # Check what's using the port
   lsof -i :5432
   
   # Stop local PostgreSQL if running
   brew services stop postgresql
   ```

2. **Permission Denied**: Docker permission issues
   ```bash
   # Ensure Docker is running
   docker ps
   
   # Check Docker permissions
   docker run hello-world
   ```

3. **Database Connection Failed**: Network issues
   ```bash
   # Check container status
   make status
   
   # Restart database
   make dev-db-stop
   make dev-db
   ```

### Data Recovery
- **Backups**: Regular backups via `make db-backup`
- **Volume Persistence**: Data persists in Docker volumes
- **Schema Recreation**: `database/init/01-schema.sql` recreates schema

## Next Steps

### API Implementation
- [ ] Chat CRUD endpoints in `backend/src/server.js`
- [ ] Message persistence endpoints
- [ ] Chat snapshot endpoints

### Frontend Integration
- [ ] Database API calls in React components
- [ ] Real-time chat loading and saving
- [ ] Recents list populated from database

### Production Deployment
- [ ] Environment variable configuration
- [ ] Database migration scripts
- [ ] Backup automation
- [ ] Monitoring setup