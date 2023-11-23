#!/bin/bash
set -e

POSTGRES_USER=user
POSTGRES_DATABASE=devdb

# Function to check if the database already exists
function database_exists {
    psql -tAc "SELECT 1 FROM pg_database WHERE datname = '$1'"
}

# Create the database if it doesn't exist
if database_exists ${POSTGRES_DB}; then
	echo "Database devdb already exists"
else
    echo "Creating database: ${POSTGRES_DATABASE}"
	psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DATABASE}" <<-EOSQL
		CREATE DATABASE devdb;
	EOSQL
fi