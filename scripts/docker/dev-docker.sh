#!/bin/bash

# Development script using Docker Compose
# Part of the organized scripts/ directory structure

set -e

# Load environment variables from .env file if it exists
if [ -f ".env" ]; then
    echo "Loading environment variables from .env file..."
    export $(grep -v '^#' .env | xargs)
fi

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to display help
show_help() {
    echo -e "${BLUE}🐳 Mainote Bot - Docker Development Environment${NC}"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start       Start all services"
    echo "  stop        Stop all services"
    echo "  restart     Restart all services"
    echo "  reset       Reset environment (clean + build + start)"
    echo "  logs        Show logs from all services"
    echo "  logs-bot    Show logs from Python bot only"
    echo "  logs-server     Show logs from Go backend only"
    echo "  build       Build all services"
    echo "  clean       Stop and remove all containers, networks, and volumes"
    echo "  status      Show status of all services"
    echo "  shell-bot   Open shell in Python bot container"
    echo "  shell-go    Open shell in Go backend container"
    echo "  test        Run tests"
    echo "  help        Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  TELEGRAM_BOT_TOKEN    - Required: Your Telegram bot token"
    echo "  NOTION_API_KEY        - Required: Your Notion API key"
    echo "  NOTION_DATABASE_ID    - Required: Your Notion database ID"
    echo "  OPENAI_API_KEY        - Required: Your OpenAI API key"
    echo ""
    echo "Examples:"
    echo "  $0 start                 # Start all services"
    echo "  $0 logs-bot             # Follow Python bot logs"
    echo "  $0 shell-bot            # Open shell in bot container"
}

# Function to check required environment variables
check_env() {
    local missing_vars=()

    if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
        missing_vars+=("TELEGRAM_BOT_TOKEN")
    fi

    if [ -z "$NOTION_API_KEY" ]; then
        missing_vars+=("NOTION_API_KEY")
    fi

    if [ -z "$NOTION_DATABASE_ID" ]; then
        missing_vars+=("NOTION_DATABASE_ID")
    fi

    if [ -z "$OPENAI_API_KEY" ]; then
        missing_vars+=("OPENAI_API_KEY")
    fi

    if [ ${#missing_vars[@]} -ne 0 ]; then
        echo -e "${RED}❌ Missing required environment variables:${NC}"
        for var in "${missing_vars[@]}"; do
            echo -e "  ${YELLOW}$var${NC}"
        done
        echo ""
        echo -e "${BLUE}💡 Create a .env file or export these variables${NC}"
        echo "Example .env file:"
        echo "TELEGRAM_BOT_TOKEN=your_bot_token"
        echo "NOTION_API_KEY=your_notion_key"
        echo "NOTION_DATABASE_ID=your_database_id"
        echo "OPENAI_API_KEY=your_openai_key"
        exit 1
    fi
}

# Function to create data directory
setup_dirs() {
    echo -e "${BLUE}📁 Setting up directories...${NC}"
    mkdir -p data
    chmod 755 data
}

# Main command handling
case "$1" in
"start")
    echo -e "${GREEN}🚀 Starting Mainote Bot Development Environment${NC}"
    check_env
    setup_dirs
    docker-compose up -d
    echo -e "${GREEN}✅ Services started successfully!${NC}"
    echo -e "${BLUE}📋 Service URLs:${NC}"
    echo -e "  🐍 Python Bot:  http://localhost:8080"
    echo -e "  🐹 Go Backend:  http://localhost:8081"
    echo -e "  🏥 Health Check: http://localhost:8080/health"
    echo ""
    echo -e "${YELLOW}💡 Use 'mainote-cli logs' to follow logs${NC}"
    ;;
"stop")
    echo -e "${YELLOW}🛑 Stopping services...${NC}"
    docker-compose down
    echo -e "${GREEN}✅ Services stopped${NC}"
    ;;
"restart")
    echo -e "${YELLOW}🔄 Restarting services...${NC}"
    docker-compose restart
    echo -e "${GREEN}✅ Services restarted${NC}"
    ;;
"logs")
    echo -e "${BLUE}📊 Following logs from all services...${NC}"
    docker-compose logs -f
    ;;
"logs-bot")
    echo -e "${BLUE}📊 Following Python bot logs...${NC}"
    docker-compose logs -f mainote-bot
    ;;
"logs-server")
    echo -e "${BLUE}📊 Following Go backend logs...${NC}"
    docker-compose logs -f mainote-server
    ;;
"build")
    echo -e "${BLUE}🔨 Building services...${NC}"
    setup_dirs
    docker-compose build
    echo -e "${GREEN}✅ Build completed${NC}"
    ;;
"clean")
    echo -e "${YELLOW}🧹 Cleaning up...${NC}"
    docker-compose down -v --remove-orphans
    docker system prune -f
    echo -e "${GREEN}✅ Cleanup completed${NC}"
    ;;
"status")
    echo -e "${BLUE}📋 Service Status:${NC}"
    docker-compose ps
    ;;
"shell-bot")
    echo -e "${BLUE}🐚 Opening shell in Python bot container...${NC}"
    docker-compose exec mainote-bot /bin/bash
    ;;
"shell-go")
    echo -e "${BLUE}🐚 Opening shell in Go backend container...${NC}"
    docker-compose exec mainote-server /bin/sh
    ;;
"test")
    echo -e "${BLUE}🧪 Running tests...${NC}"
    docker-compose exec mainote-bot python -m pytest
    ;;
"reset")
    echo -e "${YELLOW}🔄 Resetting environment (clean + build + start)...${NC}"
    echo -e "${BLUE}Step 1/3: Cleaning up...${NC}"
    docker-compose down -v --remove-orphans
    docker system prune -f
    echo -e "${BLUE}Step 2/3: Building services...${NC}"
    setup_dirs
    docker-compose build
    echo -e "${BLUE}Step 3/3: Starting services...${NC}"
    check_env
    docker-compose up -d
    echo -e "${GREEN}✅ Environment reset completed!${NC}"
    echo -e "${BLUE}📋 Service URLs:${NC}"
    echo -e "  🐍 Python Bot:  http://localhost:8080"
    echo -e "  🐹 Go Backend:  http://localhost:8081"
    echo -e "  🏥 Health Check: http://localhost:8080/health"
    echo ""
    echo -e "${YELLOW}💡 Use 'mainote-cli logs' to follow logs${NC}"
    ;;
"help" | "")
    show_help
    ;;
*)
    echo -e "${RED}❌ Unknown command: $1${NC}"
    echo ""
    show_help
    exit 1
    ;;
esac
