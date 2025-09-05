# Ultra API Makefile
# Provides convenient commands for Docker operations

.PHONY: help build up down dev prod clean logs shell test

# Default target
help: ## Show this help message
	@echo "Ultra API Docker Commands:"
	@echo "========================="
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development commands
dev: ## Start development environment
	docker-compose -f docker-compose.yml up --build

dev-bg: ## Start development environment in background
	docker-compose -f docker-compose.yml up -d --build

dev-full: ## Start development environment with additional dev services
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build

# Production commands
prod: ## Start production environment
	docker-compose -f docker-compose.prod.yml up -d --build

prod-logs: ## View production logs
	docker-compose -f docker-compose.prod.yml logs -f

# Build commands
build: ## Build the application image
	docker-compose build

build-no-cache: ## Build the application image without cache
	docker-compose build --no-cache

# Service management
up: ## Start all services
	docker-compose up -d

down: ## Stop all services
	docker-compose down

restart: ## Restart all services
	docker-compose restart

# Database commands
db-reset: ## Reset databases (WARNING: This will delete all data)
	docker-compose down -v
	docker volume rm ultra_postgres_data ultra_mongo_data 2>/dev/null || true
	docker-compose up -d postgres mongo
	@echo "Waiting for databases to be ready..."
	@sleep 10

db-backup: ## Backup databases
	mkdir -p ./backups
	docker-compose exec postgres pg_dump -U ultra_user ultra_db > ./backups/postgres_backup_$(shell date +%Y%m%d_%H%M%S).sql
	docker-compose exec mongo mongodump --uri="mongodb://ultra_user:ultra_password@localhost:27017/ultra?authSource=admin" --out=/tmp/backup
	docker cp ultra_mongo:/tmp/backup ./backups/mongo_backup_$(shell date +%Y%m%d_%H%M%S)

# Logs and debugging
logs: ## View logs from all services
	docker-compose logs -f

logs-api: ## View logs from API service only
	docker-compose logs -f ultra-api

logs-db: ## View logs from database services
	docker-compose logs -f postgres mongo

shell-api: ## Get shell access to API container
	docker-compose exec ultra-api sh

shell-db: ## Get shell access to PostgreSQL container
	docker-compose exec postgres psql -U ultra_user -d ultra_db

shell-mongo: ## Get shell access to MongoDB container
	docker-compose exec mongo mongosh -u ultra_user -p ultra_password --authenticationDatabase admin ultra

# Testing
test: ## Run tests in container
	docker-compose exec ultra-api go test ./...

test-coverage: ## Run tests with coverage
	docker-compose exec ultra-api go test -coverprofile=coverage.out ./...
	docker-compose exec ultra-api go tool cover -html=coverage.out -o coverage.html

# Cleanup
clean: ## Clean up containers, networks, and volumes
	docker-compose down -v --remove-orphans
	docker system prune -f
	docker volume prune -f

clean-all: ## Clean everything including images
	docker-compose down -v --remove-orphans --rmi all
	docker system prune -af
	docker volume prune -f

# Health checks
health: ## Check health of all services
	@echo "Checking service health..."
	@docker-compose ps
	@echo "\nAPI Health Check:"
	@curl -s http://localhost:8080/health || echo "API not responding"
	@echo "\n\nDatabase Connections:"
	@docker-compose exec postgres pg_isready -U ultra_user -d ultra_db || echo "PostgreSQL not ready"
	@docker-compose exec mongo mongosh --eval "db.adminCommand('ping')" --quiet || echo "MongoDB not ready"

# Development utilities
dev-tools: ## Start additional development tools
	@echo "Starting development tools..."
	@echo "- Adminer (PostgreSQL): http://localhost:8081"
	@echo "- Mongo Express: http://localhost:8082"
	@echo "- API: http://localhost:8080"
	@echo "- Health Check: http://localhost:8080/health"

# Quick start
quick-start: build up dev-tools ## Quick start for development

# Environment setup
setup: ## Setup environment files
	@if [ ! -f .env ]; then cp .env.example .env; echo "Created .env file from template"; fi
	@echo "Environment setup complete"

# Format and lint
format: ## Format Go code
	docker-compose exec ultra-api go fmt ./...

lint: ## Run linting
	docker-compose exec ultra-api golangci-lint run ./... || echo "Linter not available, install golangci-lint in container"

# Database migrations (if you add a migration system later)
migrate: ## Run database migrations
	docker-compose exec ultra-api echo "Migration system not implemented yet"

# Show running services and their ports
status: ## Show status and exposed ports
	@echo "Service Status and Ports:"
	@echo "========================"
	@docker-compose ps --format "table {{.Name}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"