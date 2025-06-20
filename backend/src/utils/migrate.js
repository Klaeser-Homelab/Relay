import { sequelize } from '../models/index.js';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const migrationsDir = path.join(__dirname, '../migrations');

async function createMigrationsTable() {
  await sequelize.query(`
    CREATE TABLE IF NOT EXISTS sequelize_meta (
      name VARCHAR(255) NOT NULL PRIMARY KEY
    );
  `);
}

async function getAppliedMigrations() {
  const [results] = await sequelize.query('SELECT name FROM sequelize_meta ORDER BY name');
  return results.map(row => row.name);
}

async function markMigrationAsApplied(migrationName) {
  await sequelize.query('INSERT INTO sequelize_meta (name) VALUES (?)', {
    replacements: [migrationName]
  });
}

async function markMigrationAsReverted(migrationName) {
  await sequelize.query('DELETE FROM sequelize_meta WHERE name = ?', {
    replacements: [migrationName]
  });
}

export async function runMigrations() {
  try {
    console.log('Running database migrations...');
    
    await createMigrationsTable();
    const appliedMigrations = await getAppliedMigrations();
    
    const migrationFiles = fs.readdirSync(migrationsDir)
      .filter(file => file.endsWith('.js'))
      .sort();
    
    let migrationsRun = 0;
    
    for (const file of migrationFiles) {
      const migrationName = file;
      
      if (appliedMigrations.includes(migrationName)) {
        console.log(`  ✓ ${migrationName} (already applied)`);
        continue;
      }
      
      console.log(`  → Running ${migrationName}...`);
      
      const migrationPath = path.join(migrationsDir, file);
      const migration = await import(migrationPath);
      
      await migration.up(sequelize.getQueryInterface());
      await markMigrationAsApplied(migrationName);
      
      console.log(`  ✓ ${migrationName} completed`);
      migrationsRun++;
    }
    
    if (migrationsRun === 0) {
      console.log('No new migrations to run.');
    } else {
      console.log(`✓ ${migrationsRun} migration(s) completed successfully.`);
    }
    
    return true;
  } catch (error) {
    console.error('✗ Migration failed:', error);
    return false;
  }
}

export async function rollbackLastMigration() {
  try {
    console.log('Rolling back last migration...');
    
    await createMigrationsTable();
    const appliedMigrations = await getAppliedMigrations();
    
    if (appliedMigrations.length === 0) {
      console.log('No migrations to rollback.');
      return true;
    }
    
    const lastMigration = appliedMigrations[appliedMigrations.length - 1];
    console.log(`  → Rolling back ${lastMigration}...`);
    
    const migrationPath = path.join(migrationsDir, lastMigration);
    const migration = await import(migrationPath);
    
    await migration.down(sequelize.getQueryInterface());
    await markMigrationAsReverted(lastMigration);
    
    console.log(`  ✓ ${lastMigration} rolled back`);
    console.log('✓ Rollback completed successfully.');
    
    return true;
  } catch (error) {
    console.error('✗ Rollback failed:', error);
    return false;
  }
}