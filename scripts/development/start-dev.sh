#!/bin/bash
echo "=== Development Mode - Python Bot ==="

# Wait for database to be ready
echo "Waiting for database to be ready..."
while ! pg_isready -h postgres -p 5432 -U mainote -d mainote; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done
echo "PostgreSQL is ready!"

echo "Running database migrations..."
export SQLALCHEMY_URL="$DATABASE_URL"
alembic upgrade head

echo "Starting application with auto-reload..."
exec gunicorn --bind 0.0.0.0:8080 \
    --worker-class uvicorn.workers.UvicornWorker \
    --reload \
    --timeout 120 \
    --max-requests 1000 \
    --max-requests-jitter 50 \
    mainote_bot.main:app 