import { DataTypes } from 'sequelize';

export async function up(queryInterface) {
  // Add claude_session_id field to chat_snapshots table
  await queryInterface.addColumn('chat_snapshots', 'claude_session_id', {
    type: DataTypes.STRING,
    allowNull: true,
    comment: 'Claude Code SDK session ID for conversation continuation'
  });

  // Add index for claude_session_id for efficient lookups
  await queryInterface.addIndex('chat_snapshots', ['claude_session_id']);
}

export async function down(queryInterface) {
  // Remove index and column
  await queryInterface.removeIndex('chat_snapshots', ['claude_session_id']);
  await queryInterface.removeColumn('chat_snapshots', 'claude_session_id');
}