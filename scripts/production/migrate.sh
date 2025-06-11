#!/bin/bash

# Production database migration script
# This script runs Flyway migrations in production environment

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to test database connection using Python
test_db_connection() {
    python3 -c "
import sys
import psycopg2
import os
from urllib.parse import urlparse

try:
    database_url = os.environ.get('DATABASE_URL')
    if not database_url:
        print('DATABASE_URL not set')
        sys.exit(1)
    
    # Parse the database URL
    parsed = urlparse(database_url)
    
    # Test connection
    conn = psycopg2.connect(
        host=parsed.hostname,
        port=parsed.port or 5432,
        user=parsed.username,
        password=parsed.password,
        database=parsed.path[1:] if parsed.path else 'postgres',
        connect_timeout=5
    )
    conn.close()
    sys.exit(0)
except Exception as e:
    print(f'Connection failed: {e}')
    sys.exit(1)
" 2>/dev/null
}

# Function to wait for database to be ready
wait_for_database() {
    local max_attempts=10
    local attempt=1
    
    print_info "Testing database connection..."
    
    while [ $attempt -le $max_attempts ]; do
        # Try Python-based connection test
        if test_db_connection; then
            print_success "Database connection successful!"
            return 0
        fi
        
        print_info "Attempt $attempt/$max_attempts: Database not ready, waiting 3 seconds..."
        sleep 3
        attempt=$((attempt + 1))
    done
    
    print_warning "Database connection test failed, but proceeding with migrations..."
    print_info "Flyway will handle connection retries internally"
    return 0
}

# Function to run migrations
run_migrations() {
    print_info "Starting database migrations..."
    
    # Extract database connection info from DATABASE_URL
    if [ -z "$DATABASE_URL" ]; then
        print_error "DATABASE_URL environment variable is required"
        exit 1
    fi
    
    # Parse DATABASE_URL components
    # Extract user, password, host, port, and database from DATABASE_URL
    DB_USER=$(echo "$DATABASE_URL" | sed -n 's/.*\/\/\([^:]*\):.*/\1/p')
    DB_PASSWORD=$(echo "$DATABASE_URL" | sed -n 's/.*\/\/[^:]*:\([^@]*\)@.*/\1/p')
    DB_HOST=$(echo "$DATABASE_URL" | sed -n 's/.*@\([^:]*\):.*/\1/p')
    DB_PORT=$(echo "$DATABASE_URL" | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
    DB_NAME=$(echo "$DATABASE_URL" | sed -n 's/.*\/\([^?]*\).*/\1/p')
    
    # Construct proper JDBC URL
    JDBC_URL="jdbc:postgresql://${DB_HOST}:${DB_PORT}/${DB_NAME}?user=${DB_USER}&password=${DB_PASSWORD}&sslmode=disable"
    
    print_info "JDBC URL: $JDBC_URL"
    print_info "Database user: $DB_USER"
    
    # Check if database has existing tables and needs baseline
    print_info "Checking if database needs baseline..."
    
    # Check if user_preferences table exists (indicates existing database)
    local has_existing_tables=false
    if python3 -c "
import sys
import psycopg2
import os
from urllib.parse import urlparse

try:
    database_url = os.environ.get('DATABASE_URL')
    parsed = urlparse(database_url)
    conn = psycopg2.connect(
        host=parsed.hostname,
        port=parsed.port or 5432,
        user=parsed.username,
        password=parsed.password,
        database=parsed.path[1:] if parsed.path else 'postgres'
    )
    cur = conn.cursor()
    cur.execute(\"SELECT 1 FROM information_schema.tables WHERE table_name = 'user_preferences'\")
    result = cur.fetchone()
    cur.close()
    conn.close()
    sys.exit(0 if result else 1)
except Exception as e:
    sys.exit(1)
" 2>/dev/null; then
        has_existing_tables=true
        print_info "Found existing tables in database"
    fi
    
    # Check if Flyway schema history exists
    local has_flyway_history=false
    if python3 -c "
import sys
import psycopg2
import os
from urllib.parse import urlparse

try:
    database_url = os.environ.get('DATABASE_URL')
    parsed = urlparse(database_url)
    conn = psycopg2.connect(
        host=parsed.hostname,
        port=parsed.port or 5432,
        user=parsed.username,
        password=parsed.password,
        database=parsed.path[1:] if parsed.path else 'postgres'
    )
    cur = conn.cursor()
    cur.execute(\"SELECT 1 FROM flyway_schema_history LIMIT 1\")
    result = cur.fetchone()
    cur.close()
    conn.close()
    sys.exit(0 if result else 1)
except Exception as e:
    sys.exit(1)
" 2>/dev/null; then
        has_flyway_history=true
        print_info "Found existing Flyway schema history"
    fi
    
    # If we have existing tables but no Flyway history, we need to baseline
    if [ "$has_existing_tables" = true ] && [ "$has_flyway_history" = false ]; then
        print_info "Database has existing schema but no Flyway history. Running baseline..."
        flyway \
            -url="$JDBC_URL" \
            -user="$DB_USER" \
            -password="$DB_PASSWORD" \
            -locations="filesystem:/app/db/sql" \
            -configFiles="/app/db/flyway.prod.conf" \
            baseline -baselineVersion=1 -baselineDescription="Baseline existing database with initial schema"
            
        if [ $? -eq 0 ]; then
            print_success "Database baseline completed successfully!"
        else
            print_error "Database baseline failed!"
            exit 1
        fi
    elif [ "$has_existing_tables" = false ]; then
        print_info "Clean database detected, will run all migrations from start"
    else
        print_info "Database already has Flyway history, proceeding with normal migration"
    fi
    
    # Run Flyway migrations
    flyway \
        -url="$JDBC_URL" \
        -user="$DB_USER" \
        -password="$DB_PASSWORD" \
        -locations="filesystem:/app/db/sql" \
        -configFiles="/app/db/flyway.prod.conf" \
        migrate
    
    if [ $? -eq 0 ]; then
        print_success "Database migrations completed successfully!"
    else
        print_error "Database migrations failed!"
        exit 1
    fi
}

# Function to show migration info
show_migration_info() {
    print_info "Current migration status:"
    
    JDBC_URL=$(echo "$DATABASE_URL" | sed -e 's/postgresql:/jdbc:postgresql:/' -e 's/postgres:/jdbc:postgresql:/')
    DB_USER=$(echo "$DATABASE_URL" | sed -n 's/.*\/\/\([^:]*\):.*/\1/p')
    DB_PASSWORD=$(echo "$DATABASE_URL" | sed -n 's/.*\/\/[^:]*:\([^@]*\)@.*/\1/p')
    
    flyway \
        -url="$JDBC_URL" \
        -user="$DB_USER" \
        -password="$DB_PASSWORD" \
        -locations="filesystem:/app/db/sql" \
        -configFiles="/app/db/flyway.prod.conf" \
        info
}

# Main function
main() {
    case "${1:-migrate}" in
        migrate)
            wait_for_database
            run_migrations
            ;;
        info)
            wait_for_database
            show_migration_info
            ;;
        wait)
            wait_for_database
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Usage: $0 [migrate|info|wait]"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
