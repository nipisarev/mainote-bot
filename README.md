# Mainote Bot

Telegram bot for saving notes to Notion with voice message support and morning notifications.

## Features

- 📝 Save text notes to Notion
- 🎤 Voice message recognition via Whisper
- 📊 Note categorization (idea, task, personal)
- 🌅 Daily morning notifications with active notes

## Installation

### 🚀 Quick Start (Recommended)

1. **Clone the repository:**

```bash
git clone https://github.com/yourusername/mainote-bot.git
cd mainote-bot
```

2. **Install the CLI tool globally:**

```bash
make install-cli
```

3. **Create environment configuration:**

Copy the example environment file and fill in your API keys:

```bash
cp .env.example .env
```

Then edit `.env` with your actual values:

```bash
# Required: Telegram Bot Token (get from @BotFather)
TELEGRAM_BOT_TOKEN=your_telegram_bot_token

# Required: Notion Integration (get from https://www.notion.so/my-integrations)
NOTION_API_KEY=your_notion_api_key
NOTION_DATABASE_ID=your_notion_database_id

# Required: OpenAI API Key (for Whisper voice recognition)
OPENAI_API_KEY=your_openai_api_key

# Required: Webhook URL (see Local Development section below)
WEBHOOK_URL=https://your-ngrok-url.ngrok.io

# Optional: Error tracking
SENTRY_DSN=your_sentry_dsn_url

# Optional: Morning notifications
MORNING_NOTIFICATION_TIME=08:00
NOTIFICATION_CHAT_IDS=your_chat_id1,your_chat_id2
ENABLE_MORNING_NOTIFICATIONS=true
```

4. **Start development environment:**

```bash
mainote-cli start
```

### 🐳 Local Development Setup

This project uses **webhook-only architecture** and requires Docker for local development:

**Prerequisites:**

- Docker and Docker Compose installed
- ngrok for webhook tunneling (see instructions below)

**Setup webhook for local development:**

1. **Install ngrok:**

```bash
# macOS (using Homebrew)
brew install ngrok

# Or download from https://ngrok.com/download
```

2. **Start ngrok tunnel:**

```bash
# In a separate terminal, expose local port 8080
ngrok http 8080
```

3. **Update webhook URL:**

Copy the ngrok URL (e.g., `https://abc123.ngrok.io`) and update your `.env` file:

```bash
WEBHOOK_URL=https://abc123.ngrok.io
```

4. **Start the bot:**

```bash
mainote-cli start
```

**Development commands:**

```bash
mainote-cli status        # Check service status
mainote-cli logs          # View logs
mainote-cli logs-bot      # View Python bot logs only
mainote-cli logs-server       # View Go backend logs only
mainote-cli stop          # Stop all services
mainote-cli shell         # Access container shell
```

**Note:** Manual installation requires PostgreSQL setup and is not recommended for development.

## Notion Setup

1. Create a new integration at [Notion Developers](https://www.notion.so/my-integrations)
2. Create a new database in Notion with the following properties:
   - Name (title)
   - Type (select: idea, task, personal)
   - Status (select: active, done)
   - Source (rich text)
   - Created (date)
   - Content (rich text)
3. Share the database with your integration
4. Copy the database ID from the URL

## Project Architecture

The project consists of two main services:

### 🐍 Python Bot (mainote_bot/)
Telegram bot with Notion integration:

```
mainote_bot/
├── bot/                  # Telegram command and message handlers
│   ├── callbacks.py      # Callback query handlers
│   ├── commands.py       # Command handlers
│   └── messages.py       # Message handlers
├── notion/               # Notion integration
│   ├── client.py         # Notion client
│   └── tasks.py          # Task management functions
├── scheduler/            # Notification scheduler
│   ├── notifications.py  # Notification sending functions
│   └── time_utils.py     # Time utilities
├── webhook/              # Webhook request handling
│   ├── routes.py         # FastAPI routes
│   └── setup.py          # Webhook setup
├── utils/                # Utilities
│   └── logging.py        # Logging configuration
├── config.py             # Application configuration
├── database.py           # Database operations
├── user_preferences.py   # User preferences management
└── main.py               # Main entry point
```

### 🐹 Go Backend (mainote_server/)
HTTP API service with clean architecture and OpenAPI-generated code:

```
mainote_server/
├── api/                  # OpenAPI specification and generated code
│   ├── src.yaml         # Main OpenAPI specification
│   ├── generated/       # Generated API code
│   │   ├── private.yaml # Generated OpenAPI specification
│   │   └── private/     # Generated Go server code
│   │       ├── api.go               # Service interfaces
│   │       ├── api_health.go        # Health API controller
│   │       ├── api_notes.go         # Notes API controller
│   │       ├── error.go             # Error handling
│   │       ├── helpers.go           # Helper functions
│   │       ├── impl.go              # Implementation types
│   │       ├── logger.go            # Logging middleware
│   │       ├── routers.go           # HTTP routing
│   │       └── model_*.go           # Data models
│   ├── path/            # API endpoint definitions
│   │   ├── health.yaml  # Health check endpoint
│   │   ├── notes.yaml   # Notes collection endpoints
│   │   └── note_by_id.yaml # Individual note endpoints
│   └── schema/          # Schema definitions
│       ├── components/  # Reusable components
│       ├── requests/    # Request schemas
│       └── responses/   # Response schemas
├── cmd/                 # Application entry points
│   └── server/          # HTTP server
├── internal/            # Private application code
│   ├── config/          # Configuration management
│   ├── delivery/http/   # HTTP handlers and middleware
│   ├── domain/          # Business entities and rules
│   ├── repository/      # Data access layer
│   └── usecase/         # Application business logic
├── scripts/             # Build and development scripts
│   └── openapi.sh       # OpenAPI code generation script
├── go.mod               # Go module
└── go.sum               # Dependencies
```

**API Generation Features:**
- Complete CRUD operations for notes
- OpenAPI 3.0 specification with proper validation
- Generated Go types, interfaces, and handlers using standard OpenAPI Generator
- Automatic code generation from specification via Docker
- Modular schema organization for maintainability
- Standard tooling for consistent, industry-standard code generation

### 🚀 Deployment

**Local Development:**
- Docker Compose with SQLite
- Automatic code reloading
- Isolated environment

**Production (Fly.io):**
- Multi-stage Docker build
- Both services in one container
- Process management via supervisord
- PostgreSQL for production

## Getting Started

### 🚀 Mainote CLI (Recommended)

Install CLI for convenient project management:

```bash
# Install CLI globally
make install-cli

# Now you can use from any directory
mainote-cli start          # Start development environment
mainote-cli status         # Check service status
mainote-cli logs           # Show logs
mainote-cli stop           # Stop all services
```

**Main commands:**
- `mainote-cli start` - Start Docker development environment
- `mainote-cli stop` - Stop all services
- `mainote-cli status` - Show service status
- `mainote-cli logs` - Show logs of all services
- `mainote-cli shell` - Open shell in Python container
- `mainote-cli help` - Show help

**Full command list:**
```bash
mainote-cli help-full      # Show all available commands
```

> 📚 **More details:** See [CLI_USAGE.md](CLI_USAGE.md) for complete guide

### Docker Development (Alternative)

For local development use Docker Compose:

```bash
# Quick start
mainote-cli docker-start

# View logs
mainote-cli logs

# Stop
mainote-cli stop
```

**Available commands:**

- `start` - Start all services (using `mainote-cli docker-start`)
- `stop` - Stop all services (using `mainote-cli docker-stop`)
- `logs` - Show logs of all services (using `mainote-cli logs`)
- `logs-bot` - Show logs of Python bot only (using `mainote-cli logs-bot`)
- `logs-server` - Show logs of Go backend only (using `mainote-cli logs-server`)
- `shell-bot` - Open shell in Python bot container (using `mainote-cli shell`)
- `build` - Build all services (using `mainote-cli docker-build`)
- `clean` - Stop and remove all containers (using `mainote-cli docker-clean`)

**Services will be available at:**
- **Python Bot**: http://localhost:8080
- **Go Backend**: http://localhost:8081
- **Health Check**: http://localhost:8080/health

**Features:**
- ✅ Automatic code reloading on changes
- ✅ SQLite database (file `./data/mainote.db`)
- ✅ Isolated development environment
- ✅ Service health checks

### Running without Docker

After installing the package, you can run the bot in one of the following ways:

```bash
# Using mainote-cli (recommended)
mainote-cli dev-python

# Using run.py script
python run.py

# Using installed package
mainote-bot

# Or directly via module
python -m mainote_bot.main
```

## Deploying to fly.io

The project uses **dual-service deployment** - both services run in one container:

```
┌─────────────────────────────────────────┐
│             Fly.io Container            │
│                                         │
│  ┌─────────────┐    ┌─────────────────┐ │
│  │ Python Bot  │    │   Go Backend    │ │
│  │   (FastAPI) │    │  (HTTP Server)  │ │
│  │  Port 8080  │    │   Port 8081     │ │
│  └─────────────┘    └─────────────────┘ │
│           │                   │         │
│  ┌─────────────────────────────────────┐ │
│  │         Supervisord                 │ │
│  │    (Process Manager)                │ │
│  └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### Setup and deployment:

1. **Install flyctl** and login:
```bash
curl -L https://fly.io/install.sh | sh
flyctl auth login
```

2. **Create application:**
```bash
flyctl launch
```

3. **Set environment variables:**
```bash
flyctl secrets set TELEGRAM_BOT_TOKEN=your_token
flyctl secrets set NOTION_API_KEY=your_key
flyctl secrets set NOTION_DATABASE_ID=your_id
flyctl secrets set OPENAI_API_KEY=your_key
flyctl secrets set WEBHOOK_URL=your_url
flyctl secrets set MORNING_NOTIFICATION_TIME=08:00
flyctl secrets set NOTIFICATION_CHAT_IDS=your_chat_id1,your_chat_id2
flyctl secrets set ENABLE_MORNING_NOTIFICATIONS=true
flyctl secrets set SENTRY_DSN=your_sentry_dsn
```

4. **Deploy application:**
```bash
flyctl deploy
```

### API Endpoints

**Python Bot (Port 8080 - external):**
- Webhook endpoints for Telegram Bot API
- FastAPI automatic documentation

**Go Backend (Port 8081 - internal):**
- `GET /health` - Service health check

### Monitoring

```bash
# View logs
flyctl logs

# Check status
flyctl status

# Check Go backend health (inside container)
flyctl ssh console
curl localhost:8081/health
```

## Usage

1. Find the bot in Telegram by name
2. Send a text message or voice note
3. Select note category using buttons
4. Receive daily morning notifications at 8:00

## Morning Notifications

The bot can send daily morning notifications with a list of active tasks from your Notion database.

### Setup

1. Get your Chat ID in Telegram:
   - Send a message to [@userinfobot](https://t.me/userinfobot)
   - It will reply with your Chat ID (e.g., `123456789`)

2. Add your Chat ID to the `NOTIFICATION_CHAT_IDS` environment variable:
   - For local run: add to `.env` file
   - For fly.io: `flyctl secrets set NOTIFICATION_CHAT_IDS=your_chat_id`

3. Set notification time in the `MORNING_NOTIFICATION_TIME` variable (format: `HH:MM`) or use the `/settime` command to configure notification time through the bot

### Commands

- `/morning` - get daily plan with active tasks
- `/start` - start working with the bot
- `/help` - show command help
- `/settime` - configure morning notification time (you can specify time directly: `/settime 14:30`)
- `/settimezone` - configure timezone for correct notification operation

### Timezones

The bot supports timezone configuration for correct notification operation. This is especially important if the bot server runs in a different timezone.

1. Use the `/settimezone` command to select your timezone
2. Choose your timezone from the list
3. After that, notifications will arrive exactly at the time you specified, taking your timezone into account

## API Development

### OpenAPI Specification & Code Generation

The Go backend uses OpenAPI 3.0 specification with automatic code generation for type-safe API development:

**Generate API code from OpenAPI specification:**

```bash
# Using root Makefile (recommended)
make server-generate

# Or using legacy alias
make generate-api

# Or directly from mainote_server directory
cd mainote_server && ./scripts/openapi.sh generate
```

**OpenAPI Structure:**
- `mainote_server/api/src.yaml` - Main OpenAPI specification
- `mainote_server/api/path/` - API endpoint definitions (health, notes, note_by_id)
- `mainote_server/api/schema/` - Organized schema definitions:
  - `components/` - Reusable components (note, error_response, etc.)
  - `requests/` - Request schemas (create_note_request, update_note_request)
  - `responses/` - Response schemas (notes_list_response, error responses)

**Generated Code:**
- `api/generated/private.yaml` - Generated OpenAPI specification  
- `api/generated/private/` - Generated Go server interfaces and models
- Clean separation between source specifications and generated code

**Available API Endpoints:**
- `GET /health` - Health check
- `GET /notes` - List all notes with pagination
- `POST /notes` - Create a new note
- `GET /notes/{id}` - Get specific note by ID
- `PUT /notes/{id}` - Update specific note
- `DELETE /notes/{id}` - Delete specific note

The API supports full CRUD operations with proper validation, error handling, and OpenAPI documentation.

## Development

### Management Structure

All project commands are organized through `mainote-cli` and `Makefile`:

```
Makefile              # Main project commands
mainote-cli          # CLI wrapper for convenient use
scripts/             # Legacy scripts (maintained for compatibility)
├── production/      # Production scripts  
│   └── start.sh    # Main production script
├── development/     # Development
│   ├── start-dev.sh      # Python bot in dev mode
│   └── start-go-dev.sh   # Go backend in dev mode
└── docker/          # Docker management
    └── dev-docker.sh     # Legacy Docker script
```

> **Recommendation:** Use `mainote-cli` instead of direct script calls

### Docker Compose vs development

Docker environment provides an improved development environment:

| Feature | Local Development | Docker Compose |
|---------|------------------|----------------|
| **Database** | Requires PostgreSQL installation | SQLite (simple and consistent) |
| **Dependencies** | Manual installation | Containerized |
| **Live Reload** | Manual restart | Automatic |
| **Isolation** | Host system | Isolated containers |
| **Cleanup** | Process killing | `docker-compose down` |

### Troubleshooting

**Services not starting:**
```bash
# Check status
mainote-cli status

# Check logs for errors
mainote-cli logs
```

**Database issues:**
```bash
# Reset database (will delete all data!)
rm -f data/mainote.db
mainote-cli docker-restart
```

**Container issues:**
```bash
# Clean everything and start over
mainote-cli docker-clean
mainote-cli docker-build
mainote-cli start
```

**Code changes not reflecting:**

- **Python**: Changes should reflect automatically
- **Go**: Air should rebuild automatically
- If not working: `mainote-cli docker-restart`
