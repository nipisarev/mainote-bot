#!/bin/bash

# Database management script for Flyway migrations
# Usage: ./scripts/database/flyway.sh [command] [options]

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
FLYWAY_IMAGE="flyway/flyway:10.10.0"
NETWORK_NAME="mainote-bot_mainote_network"

# Database connection settings
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5433}"
DB_NAME="${DB_NAME:-mainote}"
DB_USER="${DB_USER:-mainote}"
DB_PASSWORD="${DB_PASSWORD:-mainote_dev_password}"

# For Docker Compose environment
DOCKER_DB_HOST="postgres"
DOCKER_DB_PORT="5432"

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

# Function to check if Docker network exists
check_docker_network() {
    if ! docker network ls | grep -q "$NETWORK_NAME"; then
        print_warning "Docker network $NETWORK_NAME not found. Starting Docker Compose..."
        cd "$PROJECT_ROOT"
        docker-compose up -d postgres
        sleep 5
    fi
}

# Function to run Flyway command
run_flyway() {
    local command="$1"
    shift
    local extra_args="$@"
    
    print_info "Running Flyway $command..."
    
    # Check if we're in Docker Compose environment
    if docker-compose ps postgres | grep -q "Up"; then
        # Use Docker Compose network
        docker run --rm \
            --network "$NETWORK_NAME" \
            -v "$PROJECT_ROOT/db/sql:/flyway/sql" \
            -v "$PROJECT_ROOT/db/flyway.conf:/flyway/conf/flyway.conf" \
            "$FLYWAY_IMAGE" \
            -url="jdbc:postgresql://$DOCKER_DB_HOST:$DOCKER_DB_PORT/$DB_NAME" \
            -user="$DB_USER" \
            -password="$DB_PASSWORD" \
            -locations="filesystem:/flyway/sql" \
            $extra_args \
            "$command"
    else
        # Use local database connection
        print_warning "PostgreSQL container not running. Using local connection."
        docker run --rm \
            --network host \
            -v "$PROJECT_ROOT/db/sql:/flyway/sql" \
            -v "$PROJECT_ROOT/db/flyway.conf:/flyway/conf/flyway.conf" \
            "$FLYWAY_IMAGE" \
            -url="jdbc:postgresql://$DB_HOST:$DB_PORT/$DB_NAME" \
            -user="$DB_USER" \
            -password="$DB_PASSWORD" \
            -locations="filesystem:/flyway/sql" \
            $extra_args \
            "$command"
    fi
}

# Function to show help
show_help() {
    cat << EOF
Database Management Script (Flyway)

Usage: $0 [COMMAND] [OPTIONS]

COMMANDS:
    migrate         Apply all pending migrations
    info            Show migration status and information
    validate        Validate applied migrations against available ones
    baseline        Baseline an existing database
    repair          Repair the schema history table
    clean           Drop all objects in configured schemas (DANGEROUS!)
    undo            Undo the most recently applied versioned migration
    
    # Development helpers
    reset           Clean database and re-apply all migrations (DANGEROUS!)
    status          Show current migration status
    create          Create a new migration file
    test            Test migration by applying and rolling back

OPTIONS:
    -h, --help      Show this help message
    -v, --verbose   Verbose output
    --dry-run       Show what would be executed without running

EXAMPLES:
    $0 migrate                    # Apply all pending migrations
    $0 info                       # Show migration status
    $0 create "Add users table"   # Create new migration file
    $0 reset                      # Reset database (development only)

For production use, set environment variables:
    DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD
EOF
}

# Function to create new migration file
create_migration() {
    local description="$1"
    if [ -z "$description" ]; then
        print_error "Migration description is required"
        echo "Usage: $0 create \"Description of migration\""
        exit 1
    fi
    
    # Get next version number
    local last_version=$(ls "$PROJECT_ROOT/db/sql"/V*.sql 2>/dev/null | sed 's/.*V\([0-9]*\)__.*/\1/' | sort -n | tail -1)
    local next_version=$((${last_version:-0} + 1))
    
    # Create filename
    local filename="V${next_version}__$(echo "$description" | sed 's/[^a-zA-Z0-9]/_/g').sql"
    local filepath="$PROJECT_ROOT/db/sql/$filename"
    
    # Create migration file
    cat > "$filepath" << EOF
-- ================================
-- Migration: $filename
-- Description: $description
-- Author: $(whoami)
-- Date: $(date '+%Y-%m-%d')
-- ================================

-- Your SQL migration goes here


-- Add comments for documentation
-- COMMENT ON TABLE table_name IS 'Description of table';
-- COMMENT ON COLUMN table_name.column_name IS 'Description of column';
EOF
    
    print_success "Created migration file: $filename"
    print_info "Edit the file at: $filepath"
}

# Function to reset database (development only)
reset_database() {
    print_warning "This will DESTROY all data in the database!"
    read -p "Are you sure you want to continue? (yes/no): " -r
    
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        print_info "Operation cancelled."
        exit 0
    fi
    
    print_info "Cleaning database..."
    run_flyway clean
    
    print_info "Re-applying all migrations..."
    run_flyway migrate
    
    print_success "Database reset completed!"
}

# Function to test migrations
test_migrations() {
    print_info "Testing migrations (apply and rollback)..."
    
    # Get current state
    print_info "Getting current migration state..."
    run_flyway info
    
    # Apply migrations
    print_info "Applying migrations..."
    run_flyway migrate
    
    # Check if undo migrations exist
    if ls "$PROJECT_ROOT/db/sql"/U*.sql 1> /dev/null 2>&1; then
        print_info "Testing rollback..."
        run_flyway undo
        
        print_info "Re-applying migrations..."
        run_flyway migrate
        
        print_success "Migration test completed successfully!"
    else
        print_warning "No undo migrations found. Skipping rollback test."
        print_success "Forward migration test completed successfully!"
    fi
}

# Main script logic
main() {
    cd "$PROJECT_ROOT"
    
    # Check if flyway directory exists
    if [ ! -d "$PROJECT_ROOT/db" ]; then
        print_error "Database directory not found: $PROJECT_ROOT/db"
        exit 1
    fi
    
    # Parse command
    case "${1:-help}" in
        migrate)
            check_docker_network
            run_flyway migrate
            print_success "Migrations applied successfully!"
            ;;
        info|status)
            check_docker_network
            run_flyway info
            ;;
        validate)
            check_docker_network
            run_flyway validate
            print_success "Migrations validated successfully!"
            ;;
        baseline)
            check_docker_network
            run_flyway baseline
            print_success "Database baselined successfully!"
            ;;
        repair)
            check_docker_network
            run_flyway repair
            print_success "Schema history repaired successfully!"
            ;;
        clean)
            print_warning "This will DESTROY all data in the database!"
            read -p "Are you sure you want to continue? (yes/no): " -r
            if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
                check_docker_network
                run_flyway clean
                print_success "Database cleaned successfully!"
            else
                print_info "Operation cancelled."
            fi
            ;;
        undo)
            check_docker_network
            run_flyway undo
            print_success "Last migration undone successfully!"
            ;;
        reset)
            reset_database
            ;;
        create)
            create_migration "$2"
            ;;
        test)
            check_docker_network
            test_migrations
            ;;
        help|-h|--help)
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
