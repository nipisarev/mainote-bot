# GitHub Copilot Instructions - Mainote Bot Project

## Project Overview

This is the **Mainote Bot** project, consisting of two main services:
1. **Python Bot Service** - Telegram bot with FastAPI webhook endpoints
2. **Go Backend Service** - REST API following Clean Architecture principles

The project uses dual-service architecture deployed on fly.io with shared PostgreSQL database.

## Project Context

- **Primary Purpose**: Telegram bot for saving notes to Notion with voice message support
- **Architecture**: Dual-service (Python + Go) with Clean Architecture principles
- **Database**: PostgreSQL (production), PostgreSQL (development)
- **Deployment**: Docker containers on fly.io platform
- **Development**: Docker Compose with live reload
- **Project Management**: CLI-first approach with `mainote-cli` and `Makefile`
- **Monitoring**: Sentry integration for error tracking

## Directory Structure & Conventions

```
mainote-bot/
├── mainote_bot/            # Python Telegram bot service
│   ├── bot/               # Telegram handlers (commands, callbacks, messages)
│   ├── notion/            # Notion API integration
│   ├── scheduler/         # Notification scheduling
│   ├── webhook/           # FastAPI webhook handlers
│   ├── utils/             # Shared utilities
│   ├── config.py          # Environment configuration
│   ├── database.py        # PostgreSQL database operations
│   └── main.py           # FastAPI application entry point
├── mainote_server/        # Go backend service
│   ├── cmd/server/        # Application entry points
│   ├── internal/          # Private application code
│   │   ├── config/       # Configuration management
│   │   ├── delivery/http/ # HTTP transport layer
│   │   ├── domain/       # Business entities and rules
│   │   └── usecase/      # Application business logic
│   └── .air.toml         # Live reload configuration
├── migrations/            # Database migrations (Alembic)
├── docker-compose.yml     # Development environment
├── extra/                 # Build configurations and templates
│   └── build/dockerfile/  # Dockerfiles organized by environment
│       ├── dev/          # Development Dockerfiles
│       │   ├── Dockerfile.bot  # Python bot development
│       │   └── Dockerfile.server      # Go backend development
│       └── production/   # Production Dockerfiles
│           └── Dockerfile  # Multi-service production build
├── fly.toml              # Fly.io deployment configuration
├── Makefile              # Main build system with project commands
├── mainote-cli           # CLI wrapper for project management
└── scripts/              # Legacy scripts (maintained for compatibility)
```

## Technology Stack & Dependencies

### Python Service
- **Python**: 3.11+
- **Framework**: FastAPI with Uvicorn workers
- **Bot Library**: python-telegram-bot
- **Database**: asyncpg (PostgreSQL), SQLAlchemy/Alembic
- **APIs**: Notion API, OpenAI Whisper
- **Scheduling**: APScheduler

### Go Service
- **Go Version**: 1.24+
- **HTTP Router**: `github.com/gorilla/mux`
- **Error Monitoring**: `github.com/getsentry/sentry-go`
- **Development**: Air for live reload

## Development Commands

### Primary CLI (Recommended)

```bash
# Start development environment
mainote-cli start

# Stop all services
mainote-cli stop

# View logs (all services or specific)
mainote-cli logs
mainote-cli logs-bot
mainote-cli logs-go

# Rebuild containers after changes
mainote-cli docker-build

# Clean up (remove containers, volumes, networks)
mainote-cli docker-clean

# Access container shell
mainote-cli shell

# Check service status
mainote-cli status

# Full command reference
mainote-cli help-full
```

### Legacy Docker Commands (Alternative)

```bash
# Legacy script usage (still supported)
./dev-docker.sh start
./dev-docker.sh stop
./dev-docker.sh logs
./dev-docker.sh logs python-bot
./dev-docker.sh logs go-backend
./dev-docker.sh logs postgres
./dev-docker.sh build
./dev-docker.sh clean
./dev-docker.sh shell python-bot
./dev-docker.sh shell go-backend
```

> **Recommendation**: Use `mainote-cli` for all development tasks

### CLI Installation & Usage

```bash
# Install CLI globally for use from any directory
make install-cli

# Use from any directory once installed
mainote-cli start
mainote-cli status
mainote-cli logs
mainote-cli stop

# Available command categories
mainote-cli help          # Show basic commands
mainote-cli help-full     # Show all available commands

# Shortcut commands
mainote-cli start         # → docker-start (start dev environment)
mainote-cli stop          # → docker-stop (stop all services)
mainote-cli status        # → docker-status (check service status)
mainote-cli logs          # → logs (show all service logs)
mainote-cli shell         # → shell (access Python container)

# Full commands (examples)
mainote-cli docker-build  # Build all containers
mainote-cli docker-clean  # Clean containers/volumes
mainote-cli dev-python    # Run Python bot locally
mainote-cli prod-start    # Production deployment commands
```

### Service URLs (Development)
- 🐍 **Python Bot**: http://localhost:8080
- 🐹 **Go Backend**: http://localhost:8081
- 🗄️ **PostgreSQL**: localhost:5433 (external port)

### Development Features
- ✅ Live reload for both Python and Go services
- ✅ PostgreSQL database with persistent storage
- ✅ Health check endpoints
- ✅ Automatic database migrations
- ✅ Volume caching for dependencies

## Production Deployment (Fly.io)

### Prerequisites
```bash
# Install Fly CLI
curl -L https://fly.io/install.sh | sh

# Login to Fly.io
fly auth login

# Set environment variables
fly secrets set TELEGRAM_BOT_TOKEN=your_token_here
fly secrets set NOTION_API_KEY=your_notion_key
fly secrets set NOTION_DATABASE_ID=your_database_id
fly secrets set OPENAI_API_KEY=your_openai_key
fly secrets set SENTRY_DSN=your_sentry_dsn
```

### Deployment Commands
```bash
# Deploy to production
fly deploy

# Deploy with specific config
fly deploy --config fly.toml

# Check deployment status
fly status

# View logs
fly logs

# SSH into production container
fly ssh console

# Scale resources
fly scale memory 1gb
fly scale count 2

# Restart application
fly apps restart mainote-bot

# View PostgreSQL status
fly postgres connect -a mainote-bot-db

# Backup database
fly postgres backup list -a mainote-bot-db
```

### Environment Variables (Production)
```bash
# Required secrets
TELEGRAM_BOT_TOKEN       # Telegram bot token
NOTION_API_KEY          # Notion integration API key  
NOTION_DATABASE_ID      # Notion database ID
OPENAI_API_KEY          # OpenAI API key for Whisper
WEBHOOK_URL             # Automatically set by Fly.io
SENTRY_DSN              # Optional: Sentry error tracking

# Auto-configured
DATABASE_URL            # PostgreSQL connection (fly.io managed)
PORT                    # Application port (8080)
```

## Code Style & Standards

### Python Guidelines (Bot Service)
1. **Language**: Russian for user messages, English for code/comments
2. **Error Handling**: Always wrap async operations in try-catch blocks
3. **Logging**: Use centralized logger from `mainote_bot.utils.logging`
4. **Type Hints**: Always include type hints for functions
5. **Async/Await**: All bot handlers and external API calls must be async

### Go Guidelines (Backend Service)
1. **Formatting**: Always use `gofmt` and follow Go conventions
2. **Naming**: Use clear, descriptive names (avoid abbreviations)
3. **Error Handling**: Never ignore errors; always handle explicitly
4. **Contexts**: Use `context.Context` for cancellation and timeouts
5. **Interfaces**: Keep interfaces small and focused

## Architecture Patterns

### Python Telegram Handler Pattern
```python
async def command_name(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /command_name command."""
    try:
        chat_id = update.effective_chat.id
        # Handler logic here
        logger.info(f"Handled command for {chat_id}")
    except Exception as e:
        logger.error(f"Error in command_name: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=update.effective_chat.id,
            text="Произошла ошибка. Попробуйте позже."
        )
```

### Go HTTP Handler Pattern
```go
func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    result, err := h.usecase.CheckHealth(ctx)
    if err != nil {
        sentry.CaptureException(err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### Database Operations Pattern
```python
async def db_operation():
    """Database operation with proper connection handling."""
    conn = None
    try:
        conn = get_db_connection()
        cursor = conn.cursor()
        # Database operations here
        conn.commit()
        return result
    except Exception as e:
        if conn:
            conn.rollback()
        logger.error(f"Database error: {str(e)}", exc_info=True)
        raise
    finally:
        if conn:
            conn.close()
```

## Monitoring & Debugging

### Health Checks

#### Comprehensive Health Check (Recommended)
The main health endpoint checks all services in one request:

```bash
# Development - checks Python bot, PostgreSQL, and Go backend
curl http://localhost:8080/health

# Production - comprehensive health check
curl https://mainote-bot.fly.dev/health
```

**Sample Response (All services healthy):**
```json
{
    "status": "healthy",
    "timestamp": "2025-06-05T22:35:02Z",
    "services": {
        "python_bot": {
            "status": "healthy",
            "bot_initialized": true,
            "application_initialized": true
        },
        "database": {
            "status": "healthy",
            "type": "postgresql",
            "connection": "active"
        },
        "go_backend": {
            "status": "healthy",
            "response_time_ms": 3.4,
            "backend_status": "healthy",
            "backend_version": "1.0.0"
        }
    },
    "summary": {
        "healthy_services": 3,
        "total_services": 3,
        "health_percentage": 100.0
    }
}
```

**Sample Response (Degraded state):**
```json
{
    "status": "degraded",
    "services": {
        "python_bot": {"status": "healthy"},
        "database": {"status": "healthy"},
        "go_backend": {
            "status": "unhealthy",
            "error": "Connection timeout"
        }
    },
    "summary": {
        "healthy_services": 2,
        "total_services": 3,
        "health_percentage": 66.7
    }
}
```

#### Individual Service Health Checks
```bash
# Go backend health only (development)
curl http://localhost:8080/health

# Direct PostgreSQL check
docker exec mainote_postgres pg_isready -U mainote -d mainote
```

#### Health Check Status Codes
- **200 OK**: All services healthy
- **503 Service Unavailable**: One or more services degraded/unhealthy

#### Monitoring Integration
The comprehensive health check can be used with monitoring tools:

```bash
# Check overall status code
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:8080/health

# Monitor specific service health percentage
curl -s http://localhost:8080/health | jq '.summary.health_percentage'

# Get unhealthy services
curl -s http://localhost:8080/health | jq '.services | to_entries | map(select(.value.status != "healthy")) | from_entries'
```

### Logging
```bash
# Development logs (recommended)
mainote-cli logs | grep ERROR
mainote-cli logs-bot | grep ERROR
mainote-cli logs-go | grep ERROR

# Legacy development logs
./dev-docker.sh logs python-bot | grep ERROR
./dev-docker.sh logs go-backend | grep ERROR

# Production logs
fly logs --app mainote-bot
fly logs --app mainote-bot | grep ERROR
```

## Database Management

### Development (PostgreSQL)
```bash
# Access PostgreSQL directly
docker exec -it mainote_postgres psql -U mainote -d mainote

# Run migrations
docker exec mainote_python_bot alembic upgrade head

# Create new migration
docker exec mainote_python_bot alembic revision --autogenerate -m "description"
```

### Production Database
```bash
# Connect to production DB
fly postgres connect -a mainote-bot-db

# Run migrations in production (automatic on deploy)
fly ssh console -a mainote-bot
alembic upgrade head
```

## Troubleshooting

### Common Issues
1. **Port conflicts**: PostgreSQL port 5432 → use 5433 for development
2. **Volume permissions**: Ensure proper permissions for mounted volumes
3. **Database connections**: Check PostgreSQL service health before starting apps
4. **Webhook errors**: HTTPS required for Telegram webhooks (dev uses localhost)
5. **Go build errors**: Ensure `.air.toml` exists and Go modules are downloaded

### Debug Commands
```bash
# Check container status
docker ps

# Inspect container logs
docker logs mainote_python_bot
docker logs mainote_go_backend

# Shell access for debugging (recommended)
mainote-cli shell

# Legacy shell access
./dev-docker.sh shell python-bot
./dev-docker.sh shell go-backend

# Test database connection
docker exec mainote_postgres pg_isready -U mainote -d mainote
```

## Security & Best Practices

### Environment Variables
- Never commit secrets to git
- Use `.env.example` for documentation
- Store secrets securely in Fly.io secrets
- Validate required environment variables at startup

### Error Handling
- Log all errors with context
- Use Sentry for production error tracking
- Provide user-friendly error messages in Russian
- Implement proper fallback behaviors

### Performance
- Use connection pooling for database operations
- Implement proper caching strategies
- Handle concurrent requests efficiently
- Monitor memory usage in production

When generating code, prioritize:
1. **Error handling** as a first-class concern
2. **Async patterns** for I/O operations (Python)
3. **Context handling** for cancellation (Go)
4. **Proper logging** with structured information
5. **Russian user messages** with English code comments
6. **Type safety** and clear interfaces
7. **Testability** through dependency injection 