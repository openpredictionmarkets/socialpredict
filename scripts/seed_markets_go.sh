#!/bin/bash

# seed_markets_go.sh
# Wrapper script to run the Go market seeder with proper module setup
# Usage: ./seed_markets_go.sh [SQL_FILE]
# Default: example_markets.sql

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

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

print_status "SocialPredict Go Market Seeder"
echo "==============================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    print_error "Please install Go from https://golang.org/dl/"
    exit 1
fi

print_status "Go found: $(go version)"

# Navigate to scripts directory
cd "$SCRIPT_DIR"

# Check if backend go.mod exists
if [ ! -f "$PROJECT_ROOT/backend/go.mod" ]; then
    print_error "Backend go.mod not found at $PROJECT_ROOT/backend/"
    print_error "Make sure you're running this from the correct directory"
    exit 1
fi

# Initialize go module in scripts directory if needed
if [ ! -f "go.mod" ]; then
    print_status "Initializing Go module for scripts..."
    go mod init socialpredict-scripts
    go mod edit -replace socialpredict="$PROJECT_ROOT/backend"
fi

# Download dependencies
print_status "Downloading Go dependencies..."
go mod tidy

# Build and run the seeder
print_status "Building and running market seeder..."
if go run seed_markets.go "$@"; then
    print_status "Market seeding completed successfully!"
else
    print_error "Market seeding failed!"
    exit 1
fi