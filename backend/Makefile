# Database management commands
DB_FILE = app.db

.PHONY: db-init db-destroy db-rebuild db-rename help

help:
	@echo "Available commands:"
	@echo "  db-init     - Initialize database with correct schema"
	@echo "  db-destroy  - Remove database file"
	@echo "  db-rebuild  - Destroy and recreate database"
	@echo "  db-rename   - Rename api_usage.db to app.db"
	@echo "  help        - Show this help message"

db-init:
	@echo "Initializing database..."
	source .venv/bin/activate && python3 -c "from app.core.database import init_db; import asyncio; asyncio.run(init_db())"
	@echo "Database initialized successfully"

db-destroy:
	@echo "Removing database file..."
	@if [ -f $(DB_FILE) ]; then rm $(DB_FILE) && echo "Removed $(DB_FILE)"; else echo "$(DB_FILE) not found"; fi

db-rebuild: db-destroy db-init
	@echo "Database rebuilt successfully"
