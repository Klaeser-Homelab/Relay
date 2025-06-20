import { sequelize, Chat, Message, ChatSnapshot } from '../models/index.js';

export async function connectDatabase() {
  try {
    await sequelize.authenticate();
    console.log('✓ Database connection established successfully');
    return true;
  } catch (error) {
    console.error('✗ Unable to connect to the database:', error);
    return false;
  }
}

export async function syncDatabase(options = {}) {
  try {
    const { force = false, alter = false } = options;
    
    await sequelize.sync({ force, alter });
    
    if (force) {
      console.log('✓ Database synced with force=true (all tables recreated)');
    } else if (alter) {
      console.log('✓ Database synced with alter=true (schema updated)');
    } else {
      console.log('✓ Database synced successfully');
    }
    
    return true;
  } catch (error) {
    console.error('✗ Error syncing database:', error);
    return false;
  }
}

export async function closeDatabase() {
  try {
    await sequelize.close();
    console.log('✓ Database connection closed');
    return true;
  } catch (error) {
    console.error('✗ Error closing database connection:', error);
    return false;
  }
}

export { sequelize, Chat, Message, ChatSnapshot };