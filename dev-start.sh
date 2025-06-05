#!/bin/bash

# Development startup script for local testing
# This script starts both the Python bot and Go backend locally

echo "=== Starting Mainote Bot with Go Backend (Development Mode) ==="

# Check if required environment variables are set
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "Warning: TELEGRAM_BOT_TOKEN not set"
fi

if [ -z "$DATABASE_URL" ]; then
    echo "Warning: DATABASE_URL not set"
fi

# Build Go backend
echo "Building Go backend..."
go build -o go-backend cmd/server/main.go
if [ $? -ne 0 ]; then
    echo "Failed to build Go backend"
    exit 1
fi

# Set environment variables for Go backend
export GO_PORT=8081

# Start Go backend in background
echo "Starting Go backend on port $GO_PORT..."
./go-backend &
GO_PID=$!

# Install Python dependencies
echo "Installing Python dependencies..."
pip install -r requirements.txt

# Run database migrations
echo "Running database migrations..."
export SQLALCHEMY_URL="$DATABASE_URL"
alembic upgrade head

# Start Python bot
echo "Starting Python bot on port 8080..."
gunicorn --bind 0.0.0.0:8080 --worker-class uvicorn.workers.UvicornWorker mainote_bot.main:app &
PYTHON_PID=$!

echo "=== Both services started ==="
echo "Python bot: http://localhost:8080"
echo "Go backend: http://localhost:8081/health"
echo ""
echo "Press Ctrl+C to stop both services"

# Wait for Ctrl+C
trap 'echo "Stopping services..."; kill $GO_PID $PYTHON_PID; exit' INT

wait 