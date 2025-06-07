#!/bin/bash

# Wait for database to be ready
echo "Waiting for database to be ready..."
sleep 5

# Install dependencies
echo "Installing dependencies..."
pip install -r requirements.txt

# Run migrations
echo "Running database migrations..."
export SQLALCHEMY_URL="$DATABASE_URL"
alembic upgrade head

# Start the application
echo "Starting application..."
exec gunicorn --bind 0.0.0.0:8080 --worker-class uvicorn.workers.UvicornWorker mainote_bot.main:app 