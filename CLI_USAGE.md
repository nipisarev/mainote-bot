# Mainote CLI Usage Guide

The `mainote-cli` command provides a convenient interface to manage the mainote-bot project from anywhere in your system.

## Installation

Install the CLI globally:
```bash
# From the project directory
make install-cli

# Or directly
sudo ln -sf "$(pwd)/mainote-cli" /usr/local/bin/mainote-cli
```

## Quick Commands

The CLI provides convenient shortcuts for common operations:

```bash
mainote-cli start          # Start Docker development environment
mainote-cli stop           # Stop Docker development environment  
mainote-cli logs           # Show Docker logs
mainote-cli status         # Show Docker services status
mainote-cli shell          # Open shell in Python bot container
```

## All Available Commands

### Production
```bash
mainote-cli prod-start     # Start production environment (via supervisord)
```

### Development (Local)
```bash
mainote-cli dev-python     # Start Python bot in development mode
mainote-cli dev-go         # Start Go backend in development mode
```

### Development (Docker)
```bash
mainote-cli docker-start      # Start all services with Docker Compose
mainote-cli docker-stop       # Stop all Docker services
mainote-cli docker-restart    # Restart all Docker services
mainote-cli docker-logs       # Show logs from all services
mainote-cli docker-build      # Build all Docker services
mainote-cli docker-clean      # Clean up containers, networks, and volumes
mainote-cli docker-status     # Show status of all services
mainote-cli docker-shell-bot  # Open shell in Python bot container
mainote-cli docker-shell-go   # Open shell in Go backend container
mainote-cli docker-help       # Show detailed Docker management options
```

### Project Management
```bash
mainote-cli scripts-help   # Show scripts directory structure
mainote-cli help           # Show CLI help
mainote-cli help-full      # Show all make targets
```

## Usage Examples

```bash
# Start development environment
mainote-cli start

# Check service status
mainote-cli status

# View logs from all services
mainote-cli logs

# Clean and rebuild everything
mainote-cli docker-clean
mainote-cli docker-build
mainote-cli start

# Start only Python bot locally (without Docker)
mainote-cli dev-python

# Open shell in running container for debugging
mainote-cli shell
```

## Alternative: Direct Make Commands

You can also use make commands directly from the project directory:

```bash
make help              # Show all available commands
make docker-start      # Same as: mainote-cli start
make docker-status     # Same as: mainote-cli status
```

## Uninstallation

To remove the CLI:
```bash
sudo rm /usr/local/bin/mainote-cli
```
