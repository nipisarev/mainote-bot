#!/bin/bash

# Production startup script with Flyway migrations

set -e

# Wait for database to be ready and run migrations
echo "Running database migrations..."
/app/scripts/production/migrate.sh migrate

# Start the application
echo "Starting application..."
exec gunicorn --bind 0.0.0.0:8080 --worker-class uvicorn.workers.UvicornWorker mainote_bot.main:app 