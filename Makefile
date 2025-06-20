# Relay Project Makefile
# Database management and development commands

.PHONY: help dev-db dev-db-stop dev-db-logs dev-db-reset prod-db prod-db-stop db-migrate db-rollback db-sync db-backup db-restore

# Default environment variables for development
DEV_DB_NAME=relay_dev
DEV_DB_USER=relay_user
DEV_DB_PASSWORD=relay_dev_password
DEV_DB_HOST=localhost
DEV_DB_PORT=5432

# Help command
help: ## Show this help message
	@echo "Relay Project Database Commands"
	@echo "==============================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development Database Commands
dev-db: ## Start development PostgreSQL database
	@echo "Starting development database..."
	docker-compose -f docker-compose.dev.yml up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Database is ready!"
	@echo "Connection details:"
	@echo "  Host: $(DEV_DB_HOST)"
	@echo "  Port: $(DEV_DB_PORT)"
	@echo "  Database: $(DEV_DB_NAME)"
	@echo "  User: $(DEV_DB_USER)"
	@echo "  Password: $(DEV_DB_PASSWORD)"

dev-db-stop: ## Stop development database
	@echo "Stopping development database..."
	docker-compose -f docker-compose.dev.yml down

dev-db-logs: ## Show development database logs
	docker-compose -f docker-compose.dev.yml logs -f postgres

dev-db-reset: ## Reset development database (WARNING: destroys all data)
	@echo "WARNING: This will destroy all data in the development database!"
	@read -p "Are you sure? [y/N]: " confirm && [ "$$confirm" = "y" ] || exit 1
	docker-compose -f docker-compose.dev.yml down -v
	docker volume rm relay_postgres_dev_data 2>/dev/null || true
	$(MAKE) dev-db

dev-db-connect: ## Connect to development database with psql
	@echo "Connecting to development database..."
	docker exec -it relay-postgres-dev psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME)

# Production Database Commands
prod-db: ## Start production database (requires environment variables)
	@echo "Starting production database..."
	@if [ -z "$$POSTGRES_PASSWORD" ]; then \
		echo "Error: POSTGRES_PASSWORD environment variable is required"; \
		exit 1; \
	fi
	docker-compose -f docker-compose.prod.yml up -d postgres

prod-db-stop: ## Stop production database
	docker-compose -f docker-compose.prod.yml down

prod-db-full: ## Start full production stack (database + application)
	@echo "Starting full production stack..."
	@if [ -z "$$POSTGRES_PASSWORD" ]; then \
		echo "Error: POSTGRES_PASSWORD environment variable is required"; \
		exit 1; \
	fi
	@if [ -z "$$OPENAI_API_KEY" ]; then \
		echo "Error: OPENAI_API_KEY environment variable is required"; \
		exit 1; \
	fi
	docker-compose -f docker-compose.prod.yml up -d

# Sequelize Database Management Commands
db-migrate: ## Run Sequelize migrations
	@echo "Running database migrations..."
	cd backend && node -e "import('./src/utils/migrate.js').then(m => m.runMigrations().then(process.exit))"

db-rollback: ## Rollback last Sequelize migration
	@echo "Rolling back last migration..."
	cd backend && node -e "import('./src/utils/migrate.js').then(m => m.rollbackLastMigration().then(process.exit))"

db-sync: ## Sync Sequelize models with database (development only)
	@echo "Syncing database models..."
	cd backend && node -e "import('./src/utils/database.js').then(db => db.connectDatabase().then(() => db.syncDatabase({ alter: true })).then(process.exit))"

db-sync-force: ## Force sync Sequelize models (WARNING: destroys all data)
	@echo "WARNING: This will destroy all data in the database!"
	@read -p "Are you sure? [y/N]: " confirm && [ "$$confirm" = "y" ] || exit 1
	cd backend && node -e "import('./src/utils/database.js').then(db => db.connectDatabase().then(() => db.syncDatabase({ force: true })).then(process.exit))"

# Database Management Commands
db-backup: ## Backup development database
	@echo "Creating database backup..."
	@mkdir -p backups
	docker exec relay-postgres-dev pg_dump -U $(DEV_DB_USER) -d $(DEV_DB_NAME) > backups/relay_dev_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Backup created in backups/ directory"

db-restore: ## Restore database from backup (Usage: make db-restore BACKUP_FILE=path/to/backup.sql)
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "Usage: make db-restore BACKUP_FILE=path/to/backup.sql"; \
		exit 1; \
	fi
	@echo "Restoring database from $(BACKUP_FILE)..."
	docker exec -i relay-postgres-dev psql -U $(DEV_DB_USER) -d $(DEV_DB_NAME) < $(BACKUP_FILE)
	@echo "Database restored successfully"

# Schema Management
db-schema-dump: ## Dump current database schema
	@echo "Dumping database schema..."
	@mkdir -p backups
	docker exec relay-postgres-dev pg_dump -U $(DEV_DB_USER) -d $(DEV_DB_NAME) --schema-only > backups/schema_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Schema dumped to backups/ directory"

# Development Environment
dev-setup: ## Set up complete development environment
	@echo "Setting up development environment..."
	$(MAKE) dev-db
	@echo "Waiting for database to be ready..."
	@sleep 5
	$(MAKE) db-migrate
	@echo "Development environment is ready!"

dev-teardown: ## Tear down development environment
	@echo "Tearing down development environment..."
	$(MAKE) dev-db-stop
	@echo "Development environment stopped"

# Environment file template
env-template: ## Create .env template file
	@echo "Creating .env template..."
	@echo "# Development Environment Variables" > .env.template
	@echo "DATABASE_URL=postgresql://relay_user:relay_dev_password@localhost:5432/relay_dev" >> .env.template
	@echo "" >> .env.template
	@echo "# Production Environment Variables (required for production)" >> .env.template
	@echo "POSTGRES_PASSWORD=your_secure_password_here" >> .env.template
	@echo "POSTGRES_DB=relay" >> .env.template
	@echo "POSTGRES_USER=relay_user" >> .env.template
	@echo "POSTGRES_PORT=5432" >> .env.template
	@echo "" >> .env.template
	@echo "# Application Environment Variables" >> .env.template
	@echo "OPENAI_API_KEY=your_openai_api_key_here" >> .env.template
	@echo "ANTHROPIC_API_KEY=your_anthropic_api_key_here" >> .env.template
	@echo "GEMINI_API_KEY=your_gemini_api_key_here" >> .env.template
	@echo "GH_TOKEN=your_github_token_here" >> .env.template
	@echo "DEFAULT_CODE_PATH=/Users/your_username/Code" >> .env.template
	@echo "" >> .env.template
	@echo "# Node Environment" >> .env.template
	@echo "NODE_ENV=development" >> .env.template
	@echo "PORT=8080" >> .env.template
	@echo ".env template created! Copy to .env and fill in your values."

# Status check
status: ## Check status of all services
	@echo "Development Database Status:"
	@docker-compose -f docker-compose.dev.yml ps postgres 2>/dev/null || echo "  Not running"
	@echo ""
	@echo "Production Services Status:"
	@docker-compose -f docker-compose.prod.yml ps 2>/dev/null || echo "  Not running"