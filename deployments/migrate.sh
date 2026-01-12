#!/bin/bash

# Database migration script for ServerEye API
# This script applies all pending migrations in order

set -e

# Database connection details
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-servereye}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${POSTGRES_PASSWORD}"

# Wait for database to be ready
echo "Waiting for database to be ready..."
until docker exec servereye-postgres pg_isready -U "$DB_USER"; do
    echo "Database is not ready yet..."
    sleep 5
done

echo "Database is ready. Applying migrations..."

# Function to apply migration
apply_migration() {
    local migration_file=$1
    local migration_name=$2
    
    echo "Applying $migration_name..."
    if docker exec servereye-postgres psql -U "$DB_USER" -d "$DB_NAME" -f "$migration_file"; then
        echo "✅ $migration_name applied successfully"
    else
        echo "⚠️  $migration_name failed or already applied"
    fi
}

# Apply migrations in order
apply_migration "/migrations/init-schema.sql" "Initial schema"
apply_migration "/migrations/migration-001-server-keys.sql" "Server keys migration"  
apply_migration "/migrations/migration-002-fix-server-table.sql" "Server table fix"

echo "🎉 All migrations completed!"
