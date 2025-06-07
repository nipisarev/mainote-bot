# Mainote Bot

Telegram bot for saving notes to Notion with voice message support and morning notifications.

## Features

- 📝 Save text notes to Notion
- 🎤 Voice message recognition via Whisper
- 📊 Note categorization (idea, task, personal)
- 🌅 Daily morning notifications with active notes

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/mainote-bot.git
cd mainote-bot
```

2. Install the package:
```bash
# Development installation
pip install -e .

# Or install from repository
pip install git+https://github.com/yourusername/mainote-bot.git
```

3. Create a `.env` file based on `config.py` and fill in the required environment variables:
```
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
NOTION_API_KEY=your_notion_api_key
NOTION_DATABASE_ID=your_notion_database_id
OPENAI_API_KEY=your_openai_api_key
WEBHOOK_URL=your_webhook_url
MORNING_NOTIFICATION_TIME=08:00
NOTIFICATION_CHAT_IDS=your_chat_id1,your_chat_id2
ENABLE_MORNING_NOTIFICATIONS=true
```

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
HTTP API service with clean architecture:

```
mainote_server/
├── cmd/server/           # Application entry point
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── delivery/http/   # HTTP handlers and middleware
│   ├── domain/          # Business entities and rules
│   └── usecase/         # Application business logic
├── go.mod               # Go module
└── go.sum               # Dependencies
```

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
- `logs-go` - Show logs of Go backend only (using `mainote-cli logs-go`)
- `shell-bot` - Open shell in Python bot container (using `mainote-cli shell`)
- `build` - Build all services (using `mainote-cli docker-build`)
- `clean` - Stop and remove all containers (using `mainote-cli docker-clean`)

**Services will be available at:**
- **Python Bot**: http://localhost:8080
- **Go Backend**: http://localhost:8081
- **Health Check**: http://localhost:8081/health

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
