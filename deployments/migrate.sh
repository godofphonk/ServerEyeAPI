#!/bin/bash
# Copyright (c) 2026 godofphonk
#
# Migration script for ServerEye PostgreSQL database
# This script handles database migrations safely

set -e

echo "=== Starting PostgreSQL migrations ==="

# Function to check if table exists
table_exists() {
    local table_name=$1
    local result=$(psql -U postgres -d servereye -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = '$table_name' AND table_schema = 'public');" 2>/dev/null | head -1 | tr -d ' ')
    echo "$result"
}

# Function to execute migration file
execute_migration() {
    local migration_file=$1
    echo "Applying migration: $migration_file"
    psql -U postgres -d servereye -f "$migration_file" || echo "Migration $migration_file failed or already applied"
}

# Check database connection
echo "Checking database connection..."
if ! psql -U postgres -d servereye -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ Cannot connect to database"
    exit 1
fi
echo "✅ Database connection successful"

# Apply migrations in order
echo "Applying schema migrations..."

# Check and apply init-schema.sql if needed (only for PostgreSQL, not TimescaleDB)
if [ "$(table_exists 'generated_keys')" != "t" ]; then
    echo "Initial schema not found, applying init-schema.sql"
    execute_migration "init-schema.sql"
else
    echo "✅ Base schema already exists"
fi

# Apply incremental migrations in order
for migration in migration-001-*.sql migration-002-*.sql migration-003-*.sql; do
    if [ -f "$migration" ]; then
        echo "Processing migration file: $migration"
        execute_migration "$migration"
    fi
done

echo "✅ PostgreSQL migrations completed successfully"
