# Mainote Bot - Project Management Makefile
# Organized script execution for development and production environments

.PHONY: help prod-start dev-python dev-go docker-start docker-stop docker-restart docker-logs docker-build docker-clean docker-status docker-shell-bot docker-shell-go docker-help scripts-help install-cli restart reset

# Default target
.DEFAULT_GOAL := help

help:
	@echo "\033[34m🚀 Mainote Bot - Project Commands\033[0m"
	@echo ""
	@echo "Production:"
	@echo "  make prod-start        Start production environment (via supervisord)"
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
	@echo "Project Management:"
	@echo "  make scripts-help      Show scripts directory structure"
	@echo "  make install-cli       Install mainote-cli command globally"
	@echo ""

# Production Commands
prod-start:
	@echo "\033[32mStarting production environment...\033[0m"
	@./scripts/production/start.sh

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

# Information Commands
scripts-help:
	@echo "\033[32mScripts directory structure:\033[0m"
	@cat ./scripts/README.md

# CLI Installation
install-cli:
	@echo "\033[32mInstalling mainote-cli command...\033[0m"
	@chmod +x ./mainote-cli
	@sudo ln -sf "$(PWD)/mainote-cli" /usr/local/bin/mainote-cli
	@echo "\033[32m✅ mainote-cli installed successfully!\033[0m"
	@echo "You can now use 'mainote-cli' from anywhere in your terminal."

# Convenience Commands
restart:
	@echo "\033[34mRestarting Docker development environment...\033[0m"
	@./scripts/docker/dev-docker.sh restart

reset:
	@echo "\033[31mResetting Docker environment (clean + build + start)...\033[0m"
	@./scripts/docker/dev-docker.sh reset
