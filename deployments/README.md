# Database Migrations

This directory contains database migrations organized by database type.

## Structure

```
deployments/
├── postgres/           # Main PostgreSQL database migrations
├── timescaledb/        # TimescaleDB (metrics) migrations
├── static-postgres/    # Static data PostgreSQL database migrations
└── README.md
```

## Migration Files

### PostgreSQL (Main Database)
**Location:** `deployments/postgres/`

- `migration-001-server-keys.sql` - Server keys table
- `migration-002-fix-server-table.sql` - Server table fixes
- `migration-003-add-sources-column.sql` - Add sources column
- `migration-004-api-keys.sql` - API keys management
- `migration-009-alerts-table.sql` - Alerts system
- `migration-010-server-source-identifiers.sql` - Server source identifiers
- `migration-011-add-telegram-id.sql` - Telegram ID for account linking

### TimescaleDB (Metrics Database)
**Location:** `deployments/timescaledb/`

- `timescaledb-init.sql` - Initial TimescaleDB setup
- `timescaledb-multi-tier.sql` - Multi-tier metrics with auto-granularity

### Static PostgreSQL (Static Data Database)
**Location:** `deployments/static-postgres/`

- `migration-005-static-data.sql` - Static server data schema
- `migration-006-memory-motherboard.sql` - Memory and motherboard info
- `migration-007-fix-hardware-info.sql` - Hardware info fixes
- `migration-008-add-storage-temperatures.sql` - Storage temperature tracking

## Migration Naming Convention

- **PostgreSQL:** `migration-XXX-description.sql`
- **TimescaleDB:** `timescaledb-description.sql`
- **Static PostgreSQL:** `migration-XXX-description.sql`

## Applying Migrations

Migrations are automatically applied during deployment via CI/CD pipeline.

### Manual Application (Development)

```bash
# PostgreSQL (main)
docker exec -i ServereyeAPI-postgres psql -U postgres -d servereye < deployments/postgres/migration-XXX.sql

# TimescaleDB
docker exec -i ServereyeAPI-timescaledb psql -U postgres -d servereye < deployments/timescaledb/timescaledb-XXX.sql

# Static PostgreSQL
docker exec -i ServereyeAPI-postgres-static psql -U postgres -d servereye < deployments/static-postgres/migration-XXX.sql
```

## Database Purposes

### Main PostgreSQL
- Server registration and metadata
- API keys and authentication
- Server sources and identifiers
- Alerts configuration

### TimescaleDB
- Time-series metrics data
- Multi-tier aggregations (1m, 5m, 10m, 1h)
- Historical metrics storage

### Static PostgreSQL
- Server hardware information
- Network interfaces configuration
- Disk information
- Motherboard and memory details
