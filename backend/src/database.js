import { Sequelize } from 'sequelize';
import { initMessageModel } from './models/message.js';
import { initChatSnapshotModel } from './models/chatSnapshot.js';
import dotenv from 'dotenv';

dotenv.config();

// Database connection configuration
const sequelize = new Sequelize({
  dialect: 'postgres',
  host: process.env.DB_HOST || 'localhost',
  port: process.env.DB_PORT || 5432,
  database: process.env.DB_NAME || 'relay',
  username: process.env.DB_USER || 'postgres',
  password: process.env.DB_PASSWORD || '',
  logging: process.env.NODE_ENV === 'development' ? console.log : false,
  define: {
    timestamps: true,
    underscored: true,
    createdAt: 'created_at',
    updatedAt: 'updated_at'
  },
  pool: {
    max: 5,
    min: 0,
    acquire: 30000,
    idle: 10000
  }
});

// Initialize models
const Message = initMessageModel(sequelize);
const ChatSnapshot = initChatSnapshotModel(sequelize);

// Define Chat model here since it's simpler
const Chat = sequelize.define('Chat', {
  id: {
    type: Sequelize.UUID,
    defaultValue: Sequelize.UUIDV4,
    primaryKey: true
  },
  title: {
    type: Sequelize.STRING(255),
    allowNull: false
  },
  repository_name: {
    type: Sequelize.STRING(255),
    allowNull: false
  },
  repository_full_name: {
    type: Sequelize.STRING(255),
    allowNull: false
  },
  last_accessed_at: {
    type: Sequelize.DATE,
    allowNull: false,
    defaultValue: Sequelize.NOW
  },
  active_issue_number: {
    type: Sequelize.INTEGER,
    allowNull: true
  },
  active_issue_title: {
    type: Sequelize.TEXT,
    allowNull: true
  },
  active_issue_url: {
    type: Sequelize.TEXT,
    allowNull: true
  },
  quiet_mode: {
    type: Sequelize.BOOLEAN,
    allowNull: false,
    defaultValue: false
  }
}, {
  tableName: 'chats',
  timestamps: true,
  createdAt: 'created_at',
  updatedAt: 'updated_at'
});

// Define associations
Chat.hasMany(Message, { foreignKey: 'chat_id', onDelete: 'CASCADE' });
Message.belongsTo(Chat, { foreignKey: 'chat_id' });

Chat.hasMany(ChatSnapshot, { foreignKey: 'chat_id', onDelete: 'CASCADE' });
ChatSnapshot.belongsTo(Chat, { foreignKey: 'chat_id' });

// Database initialization function
async function initializeDatabase() {
  try {
    // Test the connection
    await sequelize.authenticate();
    console.log('Database connection established successfully.');

    // Sync all models (create tables if they don't exist)
    // In production, you'd use migrations instead
    await sequelize.sync({ alter: false });
    console.log('Database models synchronized.');

    return true;
  } catch (error) {
    console.error('Unable to connect to the database:', error);
    return false;
  }
}

// Export everything
export {
  sequelize,
  Chat,
  Message,
  ChatSnapshot,
  initializeDatabase
};