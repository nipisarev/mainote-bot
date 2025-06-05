#!/bin/sh
echo "=== Development Mode - Go Backend ==="

# Wait for database to be ready
echo "Waiting for database to be ready..."
while ! pg_isready -h postgres -p 5432 -U mainote -d mainote; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done
echo "PostgreSQL is ready!"

echo "Starting Go backend with live reload..."
exec air 