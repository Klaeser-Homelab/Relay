import { DataTypes } from 'sequelize';

export async function up(queryInterface) {
  // Create UUID extension if not exists
  await queryInterface.sequelize.query('CREATE EXTENSION IF NOT EXISTS "uuid-ossp";');

  // Create chats table
  await queryInterface.createTable('chats', {
    id: {
      type: DataTypes.UUID,
      defaultValue: DataTypes.UUIDV4,
      primaryKey: true
    },
    title: {
      type: DataTypes.STRING(255),
      allowNull: false
    },
    repository_name: {
      type: DataTypes.STRING(255),
      allowNull: false
    },
    repository_full_name: {
      type: DataTypes.STRING(255),
      allowNull: false
    },
    created_at: {
      type: DataTypes.DATE,
      allowNull: false,
      defaultValue: DataTypes.NOW
    },
    updated_at: {
      type: DataTypes.DATE,
      allowNull: false,
      defaultValue: DataTypes.NOW
    },
    last_accessed_at: {
      type: DataTypes.DATE,
      allowNull: false,
      defaultValue: DataTypes.NOW
    },
    active_issue_number: {
      type: DataTypes.INTEGER,
      allowNull: true
    },
    active_issue_title: {
      type: DataTypes.TEXT,
      allowNull: true
    },
    active_issue_url: {
      type: DataTypes.TEXT,
      allowNull: true
    },
    quiet_mode: {
      type: DataTypes.BOOLEAN,
      allowNull: false,
      defaultValue: false
    }
  });

  // Create messages table
  await queryInterface.createTable('messages', {
    id: {
      type: DataTypes.UUID,
      defaultValue: DataTypes.UUIDV4,
      primaryKey: true
    },
    chat_id: {
      type: DataTypes.UUID,
      allowNull: false,
      references: {
        model: 'chats',
        key: 'id'
      },
      onDelete: 'CASCADE'
    },
    type: {
      type: DataTypes.STRING(50),
      allowNull: false
    },
    content: {
      type: DataTypes.TEXT,
      allowNull: false
    },
    metadata: {
      type: DataTypes.JSONB,
      allowNull: true
    },
    created_at: {
      type: DataTypes.DATE,
      allowNull: false,
      defaultValue: DataTypes.NOW
    }
  });

  // Create chat_snapshots table
  await queryInterface.createTable('chat_snapshots', {
    id: {
      type: DataTypes.UUID,
      defaultValue: DataTypes.UUIDV4,
      primaryKey: true
    },
    chat_id: {
      type: DataTypes.UUID,
      allowNull: false,
      references: {
        model: 'chats',
        key: 'id'
      },
      onDelete: 'CASCADE'
    },
    transcriptions: {
      type: DataTypes.JSONB,
      allowNull: true
    },
    function_results: {
      type: DataTypes.JSONB,
      allowNull: true
    },
    claude_streaming_texts: {
      type: DataTypes.JSONB,
      allowNull: true
    },
    claude_todo_writes: {
      type: DataTypes.JSONB,
      allowNull: true
    },
    repository_issues: {
      type: DataTypes.JSONB,
      allowNull: true
    },
    created_at: {
      type: DataTypes.DATE,
      allowNull: false,
      defaultValue: DataTypes.NOW
    }
  });

  // Create indexes
  await queryInterface.addIndex('chats', ['repository_full_name']);
  await queryInterface.addIndex('chats', ['last_accessed_at']);
  await queryInterface.addIndex('messages', ['chat_id', 'created_at']);
  await queryInterface.addIndex('messages', ['type']);
  await queryInterface.addIndex('chat_snapshots', ['chat_id', 'created_at']);
}

export async function down(queryInterface) {
  await queryInterface.dropTable('chat_snapshots');
  await queryInterface.dropTable('messages');
  await queryInterface.dropTable('chats');
}