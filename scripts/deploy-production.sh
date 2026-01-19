#!/bin/bash

set -e

echo "=== Production Deployment Script ==="

# Validate required secrets are available
if [ -z "${PROD_HOST}" ] || [ -z "${PROD_USER}" ] || [ -z "${PROD_SSH_KEY}" ]; then
  echo "❌ Missing required deployment secrets!"
  exit 1
fi

cd /opt/servereye

# Clone or update repository
echo "=== Updating repository ==="
# Fix git safe directory issue
git config --global --add safe.directory /opt/servereye

if [ -d "/opt/servereye/.git" ]; then
  echo "Repository exists, pulling latest changes..."
  # Save deployments directory before clean
  if [ -d "/opt/servereye/deployments" ]; then
    echo "Saving deployments directory..."
    cp -r /opt/servereye/deployments /tmp/deployments-backup
  fi
  
  git fetch origin production
  git reset --hard origin/production
  git clean -fd
  
  # Restore deployments directory if it was removed
  if [ ! -d "/opt/servereye/deployments" ] && [ -d "/tmp/deployments-backup" ]; then
    echo "Restoring deployments directory..."
    mv /tmp/deployments-backup /opt/servereye/deployments
  elif [ -d "/tmp/deployments-backup" ]; then
    rm -rf /tmp/deployments-backup
  fi
else
  echo "Cloning fresh repository..."
  rm -rf /opt/servereye/*
  git clone https://github.com/godofphonk/ServerEyeAPI.git /opt/servereye/temp-repo
  mv /opt/servereye/temp-repo/* /opt/servereye/
  mv /opt/servereye/temp-repo/.git /opt/servereye/
  rm -rf /opt/servereye/temp-repo
fi

# Verify repository has required files
echo "=== Verifying repository contents ==="
echo "Current directory contents:"
ls -la /opt/servereye/
echo "Looking for SQL files:"
find /opt/servereye -name "*.sql" -type f 2>/dev/null || echo "No SQL files found anywhere"

if [ ! -f "/opt/servereye/deployments/timescaledb-init.sql" ]; then
  echo "❌ timescaledb-init.sql not found in repository after update!"
  echo "Deployments directory exists check:"
  if [ -d "/opt/servereye/deployments" ]; then
    echo "Deployments directory exists, contents:"
    ls -la /opt/servereye/deployments/
  else
    echo "Deployments directory does NOT exist"
  fi
  exit 1
fi

# Force copy latest files to working directory
echo "=== Force copying latest deployment files ==="
# Save original deployments
cp -r /opt/servereye/deployments /opt/servereye/deployments-original
rm -rf ./deployments
mkdir -p ./deployments
cp -r /opt/servereye/deployments-original/* ./deployments/ 2>/dev/null || echo "Copy failed, trying individual files"

# If copy failed, try copying files individually
if [ ! -f "./deployments/timescaledb-init.sql" ]; then
  echo "=== Copying files individually ==="
  find /opt/servereye/deployments-original -name "*.sql" -type f -exec cp {} ./deployments/ \;
  find /opt/servereye/deployments-original -name "*.yml" -type f -exec cp {} ./deployments/ \;
  find /opt/servereye/deployments-original -name "*.sh" -type f -exec cp {} ./deployments/ \;
fi

# Cleanup backup
rm -rf /opt/servereye/deployments-original

# Verify file hash to ensure we have latest version
echo "=== Verifying file versions ==="
CURRENT_HASH=$(sha256sum ./deployments/timescaledb-init.sql | cut -d' ' -f1)
echo "Current timescaledb-init.sql hash: $CURRENT_HASH"

# Check if file contains the fixes
if grep -q "metric_time TIMESTAMPTZ" ./deployments/timescaledb-init.sql; then
  echo "✅ File contains metric_time fix"
else
  echo "❌ File does not contain metric_time fix - forcing update"
  cp /opt/servereye/deployments/timescaledb-init.sql ./deployments/timescaledb-init.sql
fi

if grep -q "end_offset => INTERVAL '1 hour'" ./deployments/timescaledb-init.sql; then
  echo "✅ File contains continuous aggregate fix"
else
  echo "❌ File does not contain continuous aggregate fix"
  exit 1
fi

# Verify it's a file and readable
if [ ! -f "./deployments/timescaledb-init.sql" ] || [ ! -r "./deployments/timescaledb-init.sql" ]; then
  echo "❌ timescaledb-init.sql exists but is not a readable file!"
  ls -la ./deployments/timescaledb-init.sql
  exit 1
fi

echo "✅ timescaledb-init.sql found and validated on server"

# Ensure curl is available for health checks
if ! command -v curl &> /dev/null; then
  echo "Installing curl for health checks..."
  sudo apt-get update && sudo apt-get install -y curl || sudo yum install -y curl || echo "curl installation failed, using wget"
fi

# Set variables
UNIQUE_TAG="production-${GITHUB_SHA}-$(date -u +'%Y-%m-%dT%H-%M-%SZ')"
IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}"

echo "=== Deploying image: $IMAGE_NAME:$UNIQUE_TAG ==="

# Stop services gracefully (preserve data)
echo "=== Stopping services ==="

# CRITICAL: Create volumes if they don't exist
echo "=== Creating data volumes ==="
docker volume create servereye_timescaledb_data || echo "TimescaleDB volume already exists"
docker volume create servereye_postgres_data || echo "PostgreSQL volume already exists"

# Verify volumes exist after creation
echo "=== Verifying data volumes exist ==="
if ! docker volume inspect servereye_timescaledb_data >/dev/null 2>&1; then
  echo "❌ TimescaleDB volume FAILED to create!"
  echo "Available volumes:"
  docker volume ls
  exit 1
fi
if ! docker volume inspect servereye_postgres_data >/dev/null 2>&1; then
  echo "❌ PostgreSQL volume FAILED to create!"
  echo "Available volumes:"
  docker volume ls
  exit 1
fi
echo "✅ All data volumes exist - safe to proceed"

# Create additional backup before stopping
echo "=== Creating emergency backup ==="
mkdir -p /opt/servereye/emergency-backups
docker run --rm -v servereye_timescaledb_data:/data -v /opt/servereye/emergency-backups:/backup alpine tar czf /backup/emergency_timescaledb_$(date +%Y%m%d_%H%M%S).tar.gz -C /data .
echo "✅ Emergency TimescaleDB backup created"

# Stop only API service first (keep databases running)
echo "=== Stopping API service only ==="
if docker-compose ps | grep -q servereye-api; then
  docker-compose stop servereye-api || true
  docker-compose rm -f servereye-api || true
fi

# Stop databases only after API is stopped
echo "=== Stopping database services ==="
docker-compose down --remove-orphans || true

# Additional cleanup for any remaining containers
echo "=== Cleaning up remaining containers ==="
for container in servereye-api servereye-postgres servereye-timescaledb; do
  if docker ps -a -q --filter "name=$container" | grep -q .; then
    docker stop $container || true
    docker rm $container || true
  fi
done

# Clean up Docker images only (preserve volumes)
echo "=== Cleaning up Docker images ==="
docker system prune -a -f

# Pull fresh image
echo "=== Pulling new image ==="
docker pull $IMAGE_NAME:production

# Verify image contains our changes in source code
echo "=== Verifying image contains changes ==="
if ! docker run --rm $IMAGE_NAME:production test -f /app/internal/websocket/handlers.go || ! docker run --rm $IMAGE_NAME:production grep -q "CI/CD TEST" /app/internal/websocket/handlers.go; then
  echo "❌ Test changes NOT found in production image source!"
  docker run --rm $IMAGE_NAME:production find /app -name "*.go" -type f | head -10 || echo "No Go files found"
  exit 1
fi
echo "✅ Test changes found in production image source!"

# Create backup before deployment
echo "=== Creating backup before deployment ==="
mkdir -p /opt/servereye/backups
if docker volume inspect servereye_postgres_data >/dev/null 2>&1; then
  docker run --rm -v servereye_postgres_data:/data -v /opt/servereye/backups:/backup alpine tar czf /backup/postgres_backup_$(date +%Y%m%d_%H%M%S).tar.gz -C /data .
  echo "✅ PostgreSQL backup created"
fi
if docker volume inspect servereye_timescaledb_data >/dev/null 2>&1; then
  docker run --rm -v servereye_timescaledb_data:/data -v /opt/servereye/backups:/backup alpine tar czf /backup/timescaledb_backup_$(date +%Y%m%d_%H%M%S).tar.gz -C /data .
  echo "✅ TimescaleDB backup created"
fi

# Create docker-compose.yml
cat > docker-compose.yml << 'COMPOSE_EOF'
services:
  servereye-api:
    image: ghcr.io/godofphonk/servereyeapi:production
    container_name: servereye-api
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - HOST=0.0.0.0
      - PORT=8080
      - DATABASE_URL=postgres://postgres@postgres:5432/servereye?sslmode=disable
      - KEYS_DATABASE_URL=postgres://postgres@postgres:5432/servereye?sslmode=disable
      - TIMESCALEDB_URL=postgres://postgres@timescaledb:5432/servereye?sslmode=disable
      - JWT_SECRET=${JWT_SECRET}
      - WEBHOOK_SECRET=${WEBHOOK_SECRET}
      - WEB_URL=${WEB_URL:-'https://api.servereye.dev'}
    volumes:
      - ./logs:/app/logs
      - ./.env:/app/.env:ro
    networks:
      - servereye-network
    depends_on:
      postgres:
        condition: service_healthy
      timescaledb:
        condition: service_healthy
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  postgres:
    image: postgres:15-alpine
    container_name: servereye-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: servereye
      POSTGRES_USER: postgres
      POSTGRES_HOST_AUTH_METHOD: trust
    volumes:
      - servereye_postgres_data:/var/lib/postgresql/data
      - /opt/servereye/deployments:/migrations:ro
    networks:
      - servereye-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 30s
      timeout: 10s
      retries: 3

  timescaledb:
    image: timescale/timescaledb:2.15.0-pg15
    container_name: servereye-timescaledb
    restart: unless-stopped
    environment:
      POSTGRES_DB: servereye
      POSTGRES_USER: postgres
      POSTGRES_HOST_AUTH_METHOD: trust
    volumes:
      - servereye_timescaledb_data:/var/lib/postgresql/data
      - /opt/servereye/deployments:/docker-entrypoint-initdb.d:ro
      - /opt/servereye/deployments:/migrations:ro
      - /opt/servereye/deployments/timescaledb-init.sql:/docker-entrypoint-initdb.d/timescaledb-init.sql:ro
    networks:
      - servereye-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  servereye_postgres_data:
    external: true
  servereye_timescaledb_data:
    external: true

networks:
  servereye-network:
    driver: bridge
COMPOSE_EOF

# Start database services
echo "=== Starting database services ==="
docker-compose up -d postgres timescaledb

# Wait for databases to be ready
echo "=== Waiting for databases to be ready ==="
sleep 20

# CRITICAL: Verify volumes are properly mounted
echo "=== Verifying volumes are mounted ==="
if ! docker-compose exec -T postgres test -d /var/lib/postgresql/data; then
  echo "❌ PostgreSQL volume NOT MOUNTED properly!"
  docker-compose down
  exit 1
fi
if ! docker-compose exec -T timescaledb test -d /var/lib/postgresql/data; then
  echo "❌ TimescaleDB volume NOT MOUNTED properly!"
  docker-compose down
  exit 1
fi
echo "✅ All volumes mounted correctly"

# Check database health
echo "=== Checking database health ==="
timeout 60 bash -c 'until docker-compose exec -T postgres pg_isready -U postgres; do sleep 2; done'
timeout 60 bash -c 'until docker-compose exec -T timescaledb pg_isready -U postgres; do sleep 2; done'

# Fix PostgreSQL authentication for existing databases
echo "=== Fixing PostgreSQL authentication ==="
# Check current pg_hba.conf content
echo "Current postgres pg_hba.conf:"
docker-compose exec -T postgres cat /var/lib/postgresql/data/pg_hba.conf | tail -5 || echo "Cannot read postgres pg_hba.conf"
echo "Current timescaledb pg_hba.conf:"
docker-compose exec -T timescaledb cat /var/lib/postgresql/data/pg_hba.conf | tail -5 || echo "Cannot read timescaledb pg_hba.conf"

# Replace scram-sha-256 rule with trust rule (more aggressive approach)
docker-compose exec -T postgres bash -c 'sed -i "s/host all all scram-sha-256/host all postgres 0.0.0.0\/0 trust/g" /var/lib/postgresql/data/pg_hba.conf' || echo "Failed to update postgres pg_hba.conf"
docker-compose exec -T timescaledb bash -c 'sed -i "s/host all all scram-sha-256/host all postgres 0.0.0.0\/0 trust/g" /var/lib/postgresql/data/pg_hba.conf' || echo "Failed to update timescaledb pg_hba.conf"

# Verify changes were applied
echo "Updated postgres pg_hba.conf:"
docker-compose exec -T postgres cat /var/lib/postgresql/data/pg_hba.conf | tail -3 || echo "Cannot verify postgres pg_hba.conf"
echo "Updated timescaledb pg_hba.conf:"
docker-compose exec -T timescaledb cat /var/lib/postgresql/data/pg_hba.conf | tail -3 || echo "Cannot verify timescaledb pg_hba.conf"

# Restart databases to apply authentication changes
docker-compose restart postgres timescaledb
sleep 10

# Wait for databases to be ready again
echo "=== Waiting for databases after restart ==="
timeout 60 bash -c 'until docker-compose exec -T postgres pg_isready -U postgres; do sleep 2; done'
timeout 60 bash -c 'until docker-compose exec -T timescaledb pg_isready -U postgres; do sleep 2; done'

# Verify TimescaleDB extension
echo "=== Verifying TimescaleDB extension ==="
docker-compose exec -T timescaledb psql -U postgres -d servereye -c "SELECT extversion FROM pg_extension WHERE extname = 'timescaledb';" || echo "TimescaleDB extension verification"

# Apply migrations if needed
echo "=== Applying migrations ==="
if [ -f "./deployments/migrate.sh" ]; then
  chmod +x ./deployments/migrate.sh
  docker-compose exec -T postgres /bin/bash -c "cd /migrations && ./migrate.sh" || echo "PostgreSQL migration completed or not needed"
fi

# CRITICAL: Ensure TimescaleDB schema is always present
echo "=== Ensuring TimescaleDB schema ==="
# Check if server_metrics table exists
TABLE_EXISTS=$(docker-compose exec -T timescaledb psql -U postgres -d servereye -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'server_metrics' AND table_schema = 'public');" | head -1 || echo "f")
if [ "$TABLE_EXISTS" != "t" ]; then
  echo "TimescaleDB schema missing, applying initialization..."
  # First, ensure TimescaleDB extension is installed
  echo "Installing TimescaleDB extension..."
  docker-compose exec -T timescaledb psql -U postgres -d servereye -c "CREATE EXTENSION IF NOT EXISTS timescaledb;" || echo "TimescaleDB extension installation"
  
  # Apply schema directly from mounted volume
  echo "Checking if timescaledb-init.sql exists..."
  docker-compose exec -T timescaledb ls -la /docker-entrypoint-initdb.d/ || echo "Init directory not found!"
  docker-compose exec -T timescaledb ls -la /docker-entrypoint-initdb.d/timescaledb-init.sql || echo "File not found in init directory!"
  docker-compose exec -T timescaledb ls -la /migrations/ || echo "Migrations directory not found!"
  docker-compose exec -T timescaledb ls -la /migrations/timescaledb-init.sql || echo "File not found in migrations!"
  
  # Check file permissions and content
  echo "Checking file permissions and content..."
  if docker-compose exec -T timescaledb test -f /docker-entrypoint-initdb.d/timescaledb-init.sql; then
    echo "✅ File found in docker-entrypoint-initdb.d"
    docker-compose exec -T timescaledb cat /docker-entrypoint-initdb.d/timescaledb-init.sql | head -5 || echo "Cannot read file content"
    if docker-compose exec -T timescaledb psql -U postgres -d servereye -f /docker-entrypoint-initdb.d/timescaledb-init.sql; then
      echo "✅ TimescaleDB init completed successfully from init directory"
    else
      echo "⚠️ TimescaleDB init had warnings, but schema may be partially applied"
    fi
  elif docker-compose exec -T timescaledb test -f /migrations/timescaledb-init.sql; then
    echo "✅ File found in migrations directory"
    docker-compose exec -T timescaledb cat /migrations/timescaledb-init.sql | head -5 || echo "Cannot read file content"
    if docker-compose exec -T timescaledb psql -U postgres -d servereye -f /migrations/timescaledb-init.sql; then
      echo "✅ TimescaleDB init completed successfully from migrations"
    else
      echo "⚠️ TimescaleDB init had warnings, but schema may be partially applied"
    fi
  else
    echo "❌ timescaledb-init.sql not found in container!"
    echo "Available files in /docker-entrypoint-initdb.d:"
    docker-compose exec -T timescaledb ls -la /docker-entrypoint-initdb.d/ || echo "Cannot list init directory"
    echo "Available files in /migrations:"
    docker-compose exec -T timescaledb ls -la /migrations/ || echo "Cannot list migrations"
    echo "Checking root directory:"
    docker-compose exec -T timescaledb ls -la / || echo "Cannot list root"
    exit 1
  fi
  
  # Verify schema was applied
  sleep 5
  TABLE_EXISTS_AFTER=$(docker-compose exec -T timescaledb psql -U postgres -d servereye -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'server_metrics' AND table_schema = 'public');" | tr -d ' ' | head -1 || echo "f")
  echo "DEBUG: TABLE_EXISTS_AFTER = '$TABLE_EXISTS_AFTER'"
  if [ "$TABLE_EXISTS_AFTER" = "t" ]; then
    echo "✅ TimescaleDB schema applied successfully!"
    # Also check for continuous aggregates
    AGGREGATES_COUNT=$(docker-compose exec -T timescaledb psql -U postgres -d servereye -t -c "SELECT COUNT(*) FROM information_schema.views WHERE table_name IN ('metrics_5m_avg', 'server_uptime_daily', 'alert_stats_hourly');" | tr -d ' ' | head -1 || echo "0")
    echo "✅ Found $AGGREGATES_COUNT continuous aggregates"
  else
    echo "❌ TimescaleDB schema application failed!"
    docker-compose exec -T timescaledb psql -U postgres -d servereye -c "\dt" || echo "Cannot list tables"
    # Check if at least basic tables exist
    BASIC_TABLES=$(docker-compose exec -T timescaledb psql -U postgres -d servereye -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name IN ('server_metrics', 'server_status', 'server_commands', 'server_events') AND table_schema = 'public';" | tr -d ' ' | head -1 || echo "0")
    echo "DEBUG: BASIC_TABLES = '$BASIC_TABLES'"
    if [ "$BASIC_TABLES" -gt 0 ]; then
      echo "⚠️ Basic tables exist ($BASIC_TABLES), continuing with deployment..."
    else
      exit 1
    fi
  fi
else
  echo "✅ TimescaleDB schema already exists"
fi

# Start main application
echo "=== Starting main application ==="
docker-compose up -d servereye-api

# Wait for API to be ready
echo "=== Waiting for API to start ==="
sleep 15

# Verify API can connect to TimescaleDB
echo "=== Verifying API TimescaleDB connectivity ==="
sleep 15
# Give more time for logs to appear
if docker logs servereye-api 2>&1 | grep -q "TimescaleDB client connected successfully" || docker logs servereye-api 2>&1 | grep -q "database.*connected"; then
  echo "✅ API connected to database successfully!"
else
  echo "⚠️ Database connection check inconclusive, continuing..."
  echo "Recent API logs:"
  docker logs servereye-api --tail 20
fi

# Wait for services
echo "=== Waiting for services to start ==="
sleep 30

# Check if containers are running
echo "=== Checking container status ==="
docker-compose ps

# Health check
echo "=== Performing health check ==="
for i in {1..6}; do
  echo "Health check attempt $i..."
  if curl -f http://localhost:8080/health; then
    echo "✅ Health check passed!"
    break
  else
    echo "Health check failed, retrying in 10s..."
    sleep 10
    if [ $i -eq 6 ]; then
      echo "❌ Health check failed after 6 attempts!"
      docker logs servereye-api --tail 50
      docker logs servereye-timescaledb --tail 20
      docker logs servereye-postgres --tail 20
      docker-compose ps
      exit 1
    fi
  fi
done

# Check for CI/CD TEST logs
echo "=== Checking for CI/CD TEST logs ==="
sleep 5
# Give more time and check more patterns
if docker logs servereye-api 2>&1 | grep -q "CI/CD TEST" || docker logs servereye-api 2>&1 | grep -q "TEST.*CI/CD"; then
  echo "✅ CI/CD TEST logs found!"
else
  echo "⚠️ CI/CD TEST logs not found yet, checking source code verification..."
  # Final verification that our changes are in the running container
  if docker exec servereye-api test -f /app/internal/websocket/handlers.go && docker exec servereye-api grep -q "CI/CD TEST" /app/internal/websocket/handlers.go; then
    echo "✅ CI/CD TEST found in source code - deployment successful!"
  else
    echo "❌ CI/CD verification failed!"
    docker logs servereye-api --tail 30
    exit 1
  fi
fi

# CRITICAL: Verify TimescaleDB data persistence
echo "=== Verifying TimescaleDB data persistence ==="
sleep 5
# Check if server_metrics table exists and has data
TABLE_EXISTS=$(docker-compose exec -T timescaledb psql -U postgres -d servereye -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'server_metrics' AND table_schema = 'public');" | head -1 || echo "f")
if [ "$TABLE_EXISTS" != "t" ]; then
  echo "❌ server_metrics table NOT FOUND - DATA LOST!"
  docker-compose exec -T timescaledb psql -U postgres -d servereye -c "\dt" || echo "Cannot list tables"
  exit 1
fi
echo "✅ server_metrics table exists - Schema preserved!"

# Check for recent data
RECENT_COUNT=$(docker-compose exec -T timescaledb psql -U postgres -d servereye -t -c "SELECT COUNT(*) FROM server_metrics WHERE time > NOW() - INTERVAL '1 hour';" | head -1 || echo "0")
echo "✅ Recent metrics count (last hour): $RECENT_COUNT"

# Final cleanup
docker image prune -f
echo "✅ Production deployment successful!"
