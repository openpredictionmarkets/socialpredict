#!/bin/bash

# populate_markets.sh
# Script to populate the SocialPredict database with example markets

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to load .env file
load_env() {
    local env_file="${1:-.env}"
    
    if [ -f "$env_file" ]; then
        print_status "Loading configuration from $env_file"
        
        # Export variables from .env file, ignoring comments and empty lines
        # Use a safer approach that handles special characters and comments
        while IFS= read -r line; do
            # Skip empty lines and comments
            if [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]]; then
                continue
            fi
            
            # Export the variable if it contains an equals sign
            if [[ "$line" =~ = ]]; then
                # Strip quotes from values
                line=$(echo "$line" | sed "s/='\([^']*\)'/=\1/g" | sed 's/="\([^"]*\)"/=\1/g')
                export "$line"
            fi
        done < "$env_file"
        
        print_status "Environment variables loaded successfully"
    else
        print_warning ".env file not found at $env_file"
        print_warning "Using default values or existing environment variables"
    fi
}

# Load .env file if it exists (look in project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
load_env "$PROJECT_ROOT/.env"

# Default database configuration (can be overridden by .env file or environment variables)
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_USER=${POSTGRES_USER:-"user"}
DB_PASSWORD=${POSTGRES_PASSWORD:-"password"}
DB_NAME=${POSTGRES_DATABASE:-"socialpredict_db"}

# Debug: Show what values we're using
print_status "Database configuration:"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_USER: $DB_USER"
echo "  DB_NAME: $DB_NAME"

# Function to execute psql command (tries local psql first, then Docker)
exec_psql() {
    local cmd="$1"
    local conn_string="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
    
    # Try local psql first
    if command -v psql &> /dev/null; then
        psql "$conn_string" -c "$cmd" 2>/dev/null
        return $?
    fi
    
    # If no local psql, try Docker
    if command -v docker &> /dev/null; then
        # Look for PostgreSQL container
        local container_name="${POSTGRES_CONTAINER_NAME:-socialpredict-postgres-container}"
        
        if docker ps --format "table {{.Names}}" | grep -q "$container_name"; then
            print_status "Using Docker container: $container_name"
            docker exec "$container_name" psql -U "$DB_USER" -d "$DB_NAME" -c "$cmd" 2>/dev/null
            return $?
        fi
    fi
    
    return 1
}

# Function to execute SQL file (tries local psql first, then Docker)
exec_psql_file() {
    local sql_file="$1"
    local conn_string="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
    
    # Try local psql first
    if command -v psql &> /dev/null; then
        psql "$conn_string" -f "$sql_file"
        return $?
    fi
    
    # If no local psql, try Docker
    if command -v docker &> /dev/null; then
        # Look for PostgreSQL container
        local container_name="${POSTGRES_CONTAINER_NAME:-socialpredict-postgres-container}"
        
        if docker ps --format "table {{.Names}}" | grep -q "$container_name"; then
            print_status "Using Docker container to execute SQL file: $container_name"
            # Copy SQL file into container and execute it
            docker cp "$sql_file" "$container_name:/tmp/$(basename "$sql_file")"
            docker exec "$container_name" psql -U "$DB_USER" -d "$DB_NAME" -f "/tmp/$(basename "$sql_file")"
            local result=$?
            # Clean up
            docker exec "$container_name" rm -f "/tmp/$(basename "$sql_file")"
            return $result
        fi
    fi
    
    return 1
}

# Function to check if database exists and is accessible
check_database() {
    print_status "Checking database connection..."
    
    if ! exec_psql '\q'; then
        print_error "Cannot connect to database. Please check your configuration:"
        echo "  DB_HOST: $DB_HOST"
        echo "  DB_PORT: $DB_PORT"
        echo "  DB_USER: $DB_USER"
        echo "  DB_NAME: $DB_NAME"
        echo ""
        if ! command -v psql &> /dev/null; then
            echo "Note: psql not found locally. Attempting to use Docker container."
            echo "Make sure Docker is running and the PostgreSQL container is up."
        else
            echo "Make sure the database is running and credentials are correct."
        fi
        exit 1
    fi
    
    print_status "Database connection successful!"
}

# Function to check if markets table exists
check_markets_table() {
    print_status "Checking if markets table exists..."
    
    TABLE_EXISTS=$(exec_psql "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'markets');" | grep -o 't\|f' | head -1)
    
    if [ "$TABLE_EXISTS" != "t" ]; then
        print_error "Markets table does not exist!"
        print_warning "Please run the application first to create the database schema via GORM auto-migration."
        print_warning "Or run: go run main.go (this will create the tables automatically)"
        exit 1
    fi
    
    print_status "Markets table found!"
}

# Function to check if admin user exists
check_admin_user() {
    print_status "Checking if admin user exists..."
    
    ADMIN_EXISTS=$(exec_psql "SELECT EXISTS (SELECT 1 FROM users WHERE username = 'admin');" | grep -o 't\|f' | head -1)
    
    if [ "$ADMIN_EXISTS" != "t" ]; then
        print_warning "Admin user does not exist!"
        print_warning "The example markets reference 'admin' as creator_username."
        print_warning "Please create an admin user first or modify the SQL file to use an existing username."
        
        read -p "Do you want to continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status "Aborting market population."
            exit 1
        fi
    else
        print_status "Admin user found!"
    fi
}

# Function to count existing markets
count_existing_markets() {
    MARKET_COUNT=$(exec_psql "SELECT COUNT(*) FROM markets;" | grep -o '[0-9]*' | head -1)
    print_status "Found $MARKET_COUNT existing markets in database."
    
    if [ "$MARKET_COUNT" -gt 0 ]; then
        print_warning "Database already contains markets."
        read -p "Do you want to add the example markets anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status "Aborting market population."
            exit 0
        fi
    fi
}

# Function to populate markets
populate_markets() {
    print_status "Populating database with example markets..."
    
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    SQL_FILE="$SCRIPT_DIR/example_markets.sql"
    
    if [ ! -f "$SQL_FILE" ]; then
        print_error "SQL file not found: $SQL_FILE"
        print_error "Make sure example_markets.sql is in the same directory as this script."
        exit 1
    fi
    
    if exec_psql_file "$SQL_FILE"; then
        print_status "Successfully populated database with example markets!"
        
        # Count markets after insertion
        NEW_MARKET_COUNT=$(exec_psql "SELECT COUNT(*) FROM markets;" | grep -o '[0-9]*' | head -1)
        print_status "Database now contains $NEW_MARKET_COUNT total markets."
    else
        print_error "Failed to populate database with example markets."
        print_error "Check the SQL file for syntax errors or constraint violations."
        exit 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Populates the SocialPredict database with example prediction markets."
    echo ""
    echo "Configuration:"
    echo "  The script looks for a .env file in the project root directory."
    echo "  If no .env file is found, it uses environment variables or defaults."
    echo ""
    echo "Environment variables (.env file or shell):"
    echo "  DB_HOST            Database host (default: localhost)"
    echo "  DB_PORT            Database port (default: 5432)"
    echo "  POSTGRES_USER      Database user (default: user)"
    echo "  POSTGRES_PASSWORD  Database password (default: password)"
    echo "  POSTGRES_DATABASE  Database name (default: devdb)"
    echo ""
    echo "Options:"
    echo "  -h, --help         Show this help message"
    echo "  --check-only       Only check database connection and schema"
    echo "  --env-file FILE    Use specific .env file (default: ../env)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Use .env file or defaults"
    echo "  $0 --env-file /path/to/.env          # Use specific .env file"
    echo "  DB_HOST=db.example.com $0            # Override database host"
    echo "  $0 --check-only                      # Only run checks"
}

# Main script logic
main() {
    case "${1:-}" in
        -h|--help)
            show_usage
            exit 0
            ;;
        --env-file)
            if [ -z "${2:-}" ]; then
                print_error "--env-file requires a file path"
                show_usage
                exit 1
            fi
            load_env "$2"
            # Re-set database variables after loading custom .env
            DB_HOST=${DB_HOST:-"localhost"}
            DB_PORT=${DB_PORT:-"5432"}
            DB_USER=${POSTGRES_USER:-"user"}
            DB_PASSWORD=${POSTGRES_PASSWORD:-"password"}
            DB_NAME=${POSTGRES_DATABASE:-"devdb"}
            
            print_status "Starting market population process with custom .env file..."
            check_database
            check_markets_table
            check_admin_user
            count_existing_markets
            populate_markets
            print_status "Market population completed successfully!"
            ;;
        --check-only)
            check_database
            check_markets_table
            check_admin_user
            print_status "All checks passed! Database is ready for market population."
            exit 0
            ;;
        "")
            print_status "Starting market population process..."
            check_database
            check_markets_table
            check_admin_user
            count_existing_markets
            populate_markets
            print_status "Market population completed successfully!"
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"