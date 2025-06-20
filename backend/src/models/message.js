import { DataTypes } from 'sequelize';

export function initMessageModel(sequelize) {
  const Message = sequelize.define('Message', {
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
      allowNull: false,
      validate: {
        isIn: [['user', 'openai', 'claude', 'system', 'function_result', 'error']]
      }
    },
    content: {
      type: DataTypes.TEXT,
      allowNull: false
    },
    metadata: {
      type: DataTypes.JSONB,
      allowNull: true,
      comment: 'Additional data (issue info, function results, etc.)'
    }
  }, {
    tableName: 'messages',
    timestamps: true,
    createdAt: 'created_at',
    updatedAt: false,
    indexes: [
      {
        fields: ['chat_id', 'created_at']
      },
      {
        fields: ['type']
      }
    ]
  });

  return Message;
}