import { DataTypes } from 'sequelize';

export function initChatModel(sequelize) {
  const Chat = sequelize.define('Chat', {
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
  }, {
    tableName: 'chats',
    timestamps: true,
    createdAt: 'created_at',
    updatedAt: 'updated_at',
    indexes: [
      {
        fields: ['repository_full_name']
      },
      {
        fields: ['last_accessed_at']
      }
    ]
  });

  return Chat;
}