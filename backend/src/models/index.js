import { Sequelize } from 'sequelize';
import { initChatModel } from './chat.js';
import { initMessageModel } from './message.js';
import { initChatSnapshotModel } from './chatSnapshot.js';

const sequelize = new Sequelize(
  process.env.DATABASE_URL || 'postgresql://relay_user:relay_dev_password@localhost:5432/relay_dev',
  {
    dialect: 'postgres',
    logging: process.env.NODE_ENV === 'development' ? console.log : false,
    define: {
      underscored: true,
      timestamps: true,
      createdAt: 'created_at',
      updatedAt: 'updated_at'
    }
  }
);

// Initialize models
const Chat = initChatModel(sequelize);
const Message = initMessageModel(sequelize);
const ChatSnapshot = initChatSnapshotModel(sequelize);

// Define associations
Chat.hasMany(Message, { foreignKey: 'chat_id', onDelete: 'CASCADE' });
Message.belongsTo(Chat, { foreignKey: 'chat_id' });

Chat.hasMany(ChatSnapshot, { foreignKey: 'chat_id', onDelete: 'CASCADE' });
ChatSnapshot.belongsTo(Chat, { foreignKey: 'chat_id' });

export { sequelize, Chat, Message, ChatSnapshot };