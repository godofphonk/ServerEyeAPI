#!/bin/bash
# Copyright (c) 2026 godofphonk
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.


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
apply_migration "/migrations/migration-003-add-sources-column.sql" "Add sources column"

echo "🎉 All migrations completed!"
