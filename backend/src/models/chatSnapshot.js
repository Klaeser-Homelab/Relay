import { DataTypes } from 'sequelize';

export function initChatSnapshotModel(sequelize) {
  const ChatSnapshot = sequelize.define('ChatSnapshot', {
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
      allowNull: true,
      comment: 'Snapshot of transcriptions array'
    },
    function_results: {
      type: DataTypes.JSONB,
      allowNull: true,
      comment: 'Snapshot of function results array'
    },
    claude_streaming_texts: {
      type: DataTypes.JSONB,
      allowNull: true,
      comment: 'Snapshot of Claude streaming texts array'
    },
    claude_todo_writes: {
      type: DataTypes.JSONB,
      allowNull: true,
      comment: 'Snapshot of Claude todo writes array'
    },
    repository_issues: {
      type: DataTypes.JSONB,
      allowNull: true,
      comment: 'Snapshot of repository issues array'
    }
  }, {
    tableName: 'chat_snapshots',
    timestamps: true,
    createdAt: 'created_at',
    updatedAt: false,
    indexes: [
      {
        fields: ['chat_id', 'created_at']
      }
    ]
  });

  return ChatSnapshot;
}