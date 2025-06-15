# Mainote Bot - Project Management Makefile
# Organized script execution for development and production environments

.PHONY: help prod-start dev-python dev-go docker-start docker-stop docker-restart docker-logs docker-build docker-clean docker-status docker-shell-bot docker-shell-go docker-help scripts-help generate-api install-cli restart reset server-build server-run server-test server-clean server-fmt server-lint server-validate server-dev server-deps server-endpoints server-test-api server-structure server-docker-build server-docker-run install-tools db-migrate db-info db-create db-validate db-repair db-reset db-undo db-test db-clean

# Default target
.DEFAULT_GOAL := help

help:
	@echo "\033[34m🚀 Mainote Bot - Project Commands\033[0m"
	@echo ""
	@echo "Development (Local):"
	@echo "  make dev-python        Start Python bot in development mode"
	@echo "  make dev-go            Start Go backend in development mode"
	@echo ""
	@echo "Development (Docker):"
	@echo "  make docker-start      Start all services with Docker Compose"
	@echo "  make docker-stop       Stop all Docker services"
	@echo "  make docker-restart    Restart all Docker services"
	@echo "  make restart           Restart Docker services (shortcut)"
	@echo "  make reset             Reset Docker environment (clean + build + start)"
	@echo "  make docker-logs       Show logs from all services"
	@echo "  make docker-build      Build all Docker services"
	@echo "  make docker-clean      Clean up containers, networks, and volumes"
	@echo "  make docker-status     Show status of all services"
	@echo "  make docker-shell-bot  Open shell in Python bot container"
	@echo "  make docker-shell-go   Open shell in Go backend container"
	@echo "  make docker-help       Show detailed Docker management options"
	@echo ""
	@echo "Go Server Development:"
	@echo "  make server-build      Build Go server binary"
	@echo "  make server-run        Run Go server in development mode"
	@echo "  make server-test       Run Go server tests"
	@echo "  make server-clean      Clean Go server generated files and binaries"
	@echo "  make server-fmt        Format Go server code"
	@echo "  make server-lint       Lint Go server code"
	@echo "  make server-validate   Validate OpenAPI specification"
	@echo "  make server-deps       Update Go server dependencies"
	@echo "  make server-structure  Show Go server project structure"
	@echo "  make server-docker-build  Build Go server Docker image"
	@echo "  make server-docker-run    Run Go server Docker container"
	@echo ""
	@echo "Database Management:"
	@echo "  make db-migrate        Apply database migrations using Flyway"
	@echo "  make db-info           Show migration status and information"
	@echo "  make db-create         Create new migration file"
	@echo "  make db-validate       Validate applied migrations"
	@echo "  make db-repair         Repair migration metadata table"
	@echo "  make db-reset          Reset database (development only - DANGEROUS!)"
	@echo "  make db-undo           Undo last applied migration"
	@echo "  make db-test           Test migrations (apply and rollback)"
	@echo ""
	@echo "Project Management:"
	@echo "  make generate-api      Generate Go API code from OpenAPI specification"
	@echo "  make install-tools     Check required tools are installed"
	@echo "  make scripts-help      Show scripts directory structure"
	@echo "  make install-cli       Install mainote-cli command globally"
	@echo ""


# Development Commands (Local)
dev-python:
	@echo "\033[33mStarting Python bot in development mode...\033[0m"
	@./scripts/development/start-dev.sh

dev-go:
	@echo "\033[33mStarting Go backend in development mode...\033[0m"
	@./scripts/development/start-go-dev.sh

# Docker Development Commands
docker-start:
	@echo "\033[34mStarting Docker development environment...\033[0m"
	@./scripts/docker/dev-docker.sh start

docker-stop:
	@echo "\033[34mStopping Docker development environment...\033[0m"
	@./scripts/docker/dev-docker.sh stop

docker-restart:
	@echo "\033[34mRestarting Docker development environment...\033[0m"
	@./scripts/docker/dev-docker.sh restart

docker-logs:
	@echo "\033[34mShowing Docker logs...\033[0m"
	@./scripts/docker/dev-docker.sh logs

docker-build:
	@echo "\033[34mBuilding Docker services...\033[0m"
	@./scripts/docker/dev-docker.sh build

docker-clean:
	@echo "\033[31mCleaning Docker environment...\033[0m"
	@./scripts/docker/dev-docker.sh clean

docker-status:
	@echo "\033[34mShowing Docker services status...\033[0m"
	@./scripts/docker/dev-docker.sh status

docker-shell-bot:
	@echo "\033[34mOpening shell in Python bot container...\033[0m"
	@./scripts/docker/dev-docker.sh shell-bot

docker-shell-go:
	@echo "\033[34mOpening shell in Go backend container...\033[0m"
	@./scripts/docker/dev-docker.sh shell-go

docker-help:
	@echo "\033[34mDocker management help:\033[0m"
	@./scripts/docker/dev-docker.sh help

# Go Server Development Commands
server-build: server-generate ## Build Go server binary
	@echo "\033[32m🔨 Building Go server...\033[0m"
	@cd mainote_server && go build -o bin/server cmd/server/main.go
	@echo "\033[32m✅ Server built successfully\033[0m"

server-run: ## Run Go server in development mode
	@echo "\033[32m🚀 Starting Go server on port $(APP_PORT)...\033[0m"
	@cd mainote_server && APP_PORT=$(APP_PORT) go run cmd/server/main.go

server-test: ## Run Go server tests
	@echo "\033[32m🧪 Running Go server tests...\033[0m"
	@cd mainote_server && go test ./...

server-clean: ## Clean Go server generated files and build artifacts
	@echo "\033[31m🧹 Cleaning Go server files...\033[0m"
	@source ./scripts/development/openapi.sh && clean
	@cd mainote_server && rm -f bin/server
	@echo "\033[32m✅ Go server cleanup completed\033[0m"

server-fmt: ## Format Go server code
	@echo "\033[32m🎨 Formatting Go server code...\033[0m"
	@cd mainote_server && go fmt ./...

server-lint: ## Lint Go server code (requires golangci-lint)
	@echo "\033[32m🔍 Linting Go server code...\033[0m"
	@cd mainote_server && golangci-lint run

server-validate: ## Validate Go server OpenAPI specification
	@echo "\033[32m✅ Validating Go server OpenAPI spec...\033[0m"
	@cd mainote_server && oapi-codegen -generate types -o /dev/null api/src.yaml
	@echo "\033[32m✅ OpenAPI specification is valid\033[0m"

server-deps: ## Update Go server dependencies
	@echo "\033[32m📦 Updating Go server dependencies...\033[0m"
	@cd mainote_server && go mod tidy
	@cd mainote_server && go mod download
	@echo "\033[32m✅ Go server dependencies updated\033[0m"

server-structure: ## Show Go server project structure
	@echo "\033[32m📁 Go server project structure:\033[0m"
	@cd mainote_server && (tree -I 'node_modules|.git|*.gen.go' . || ls -la)

server-docker-build: ## Build Go server Docker image
	@echo "\033[32m🐳 Building Go server Docker image...\033[0m"
	@docker build -f extra/build/dockerfile/dev/Dockerfile.server -t mainote-server:dev mainote_server

server-docker-run: ## Run Go server Docker container
	@echo "\033[32m🐳 Running Go server Docker container...\033[0m"
	@cd mainote_server && docker run -p $(APP_PORT):8081 mainote-server:dev

server-generate: install-tools ## Generate Go server API code from OpenAPI specification
	@echo "\033[32m🔄 Generating Go server API code from OpenAPI specification...\033[0m"
	@source ./scripts/development/openapi.sh && generate_api
	@echo "\033[32m✅ Go server API code generation completed successfully!\033[0m"

# Project Management Commands
install-tools: ## Check required tools are installed
	@echo "\033[32m🔍 Checking required tools...\033[0m"
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed. Please install Docker."; exit 1; }
	@echo "\033[32m✅ Docker found\033[0m"
	@echo "\033[32m✅ All required tools are available\033[0m"

# API Generation (alias for server-generate for backward compatibility)
generate-api: server-generate ## Generate Go API code from OpenAPI specification (alias for server-generate)

# CLI Installation
install-cli: ## Install mainote-cli command globally
	@echo "\033[32mInstalling mainote-cli command...\033[0m"
	@chmod +x ./mainote-cli
	@sudo ln -sf "$(PWD)/mainote-cli" /usr/local/bin/mainote-cli
	@echo "\033[32m✅ mainote-cli installed successfully!\033[0m"
	@echo "You can now use 'mainote-cli' from anywhere in your terminal."

# Database Management Commands
db-migrate: ## Apply database migrations using Flyway
	@echo "\033[32mApplying database migrations...\033[0m"
	@./scripts/database/flyway.sh migrate

db-info: ## Show migration status and information
	@echo "\033[32mShowing migration information...\033[0m"
	@./scripts/database/flyway.sh info

db-create: ## Create new migration file (usage: make db-create DESCRIPTION="Add users table")
	@echo "\033[32mCreating new migration file...\033[0m"
	@if [ -z "$(DESCRIPTION)" ]; then \
		echo "\033[31mError: DESCRIPTION is required\033[0m"; \
		echo "Usage: make db-create DESCRIPTION=\"Add users table\""; \
		exit 1; \
	fi
	@./scripts/database/flyway.sh create "$(DESCRIPTION)"

db-validate: ## Validate applied migrations against available ones
	@echo "\033[32mValidating migrations...\033[0m"
	@./scripts/database/flyway.sh validate

db-repair: ## Repair migration metadata table
	@echo "\033[32mRepairing migration metadata...\033[0m"
	@./scripts/database/flyway.sh repair

db-reset: ## Reset database (development only - DANGEROUS!)
	@echo "\033[31m⚠️  WARNING: This will destroy all data in the database!\033[0m"
	@FORCE_RESET=$(FORCE_RESET) ./scripts/database/flyway.sh reset

db-undo: ## Undo last applied migration
	@echo "\033[32mUndoing last migration...\033[0m"
	@./scripts/database/flyway.sh undo

db-test: ## Test migrations (apply and rollback)
	@echo "\033[32mTesting migrations...\033"
