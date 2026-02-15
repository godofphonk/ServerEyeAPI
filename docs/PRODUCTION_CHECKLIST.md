# Production Deployment Checklist

## 🎯 Pre-Deployment Checklist

### 🔐 Security Configuration

#### Go API

- [ ] **Change JWT Secret**
  ```bash
  export JWT_SECRET="$(openssl rand -base64 32)"
  ```

- [ ] **Generate Production API Key**
  ```bash
  curl -X POST http://localhost:8080/api/admin/keys \
    -H "X-API-Key: sk_csharp_backend_development_key_change_in_production" \
    -H "Content-Type: application/json" \
    -d '{
      "service_id": "csharp-backend-prod",
      "service_name": "C# Web Backend Production",
      "permissions": ["metrics:read", "servers:read", "servers:validate"],
      "expires_days": 365
    }'
  ```

- [ ] **Revoke Development API Keys**
  ```bash
  curl -X DELETE http://localhost:8080/api/admin/keys/key_csharp_backend_001 \
    -H "X-API-Key: <admin_key>"
  ```

- [ ] **Configure HTTPS**
  - SSL/TLS certificates installed
  - HTTP redirects to HTTPS
  - HSTS headers configured

- [ ] **Database Security**
  - Strong PostgreSQL password
  - Strong TimescaleDB password
  - Database accessible only from application servers
  - SSL connections enabled

#### C# Backend

- [ ] **Change Encryption Key**
  ```bash
  # Generate secure 32-character key
  openssl rand -base64 32
  ```
  Update in `appsettings.Production.json`:
  ```json
  {
    "Encryption": {
      "Key": "your-secure-32-character-key-here"
    }
  }
  ```

- [ ] **Update JWT Secret** (must match Go API)
  ```json
  {
    "JwtSettings": {
      "SecretKey": "same-as-go-api-jwt-secret"
    }
  }
  ```

- [ ] **Update Go API Settings**
  ```json
  {
    "GoApiSettings": {
      "BaseUrl": "https://api.servereye.dev",
      "ApiKey": "production-api-key-from-go-api"
    }
  }
  ```

- [ ] **Configure Production Database**
  ```json
  {
    "ConnectionStrings": {
      "DefaultConnection": "Host=prod-db;Database=ServerEyeWeb;Username=app_user;Password=strong_password"
    }
  }
  ```

### 🗄️ Database Setup

#### PostgreSQL (Go API)

- [ ] **Apply Migrations**
  ```bash
  docker exec servereye-postgres psql -U servereye -d servereye -f /app/deployments/migration-001-initial.sql
  docker exec servereye-postgres psql -U servereye -d servereye -f /app/deployments/migration-002-multi-tier.sql
  docker exec servereye-postgres psql -U servereye -d servereye -f /app/deployments/migration-003-audit-logs.sql
  docker exec servereye-postgres psql -U servereye -d servereye -f /app/deployments/migration-004-api-keys.sql
  ```

- [ ] **Verify Tables**
  ```sql
  SELECT table_name FROM information_schema.tables 
  WHERE table_schema = 'public';
  ```

- [ ] **Create Backup User**
  ```sql
  CREATE USER backup_user WITH PASSWORD 'backup_password';
  GRANT SELECT ON ALL TABLES IN SCHEMA public TO backup_user;
  ```

#### TimescaleDB

- [ ] **Verify Hypertables**
  ```sql
  SELECT * FROM timescaledb_information.hypertables;
  ```

- [ ] **Check Continuous Aggregates**
  ```sql
  SELECT * FROM timescaledb_information.continuous_aggregates;
  ```

- [ ] **Verify Compression**
  ```sql
  SELECT * FROM timescaledb_information.compression_settings;
  ```

#### PostgreSQL (C# Backend)

- [ ] **Apply EF Migrations**
  ```bash
  cd ServerEye.API
  dotnet ef database update --project ../ServerEye.Infrastracture
  ```

- [ ] **Verify Schema**
  ```sql
  SELECT table_name FROM information_schema.tables 
  WHERE table_schema = 'public';
  ```

### 🐳 Docker Configuration

#### Go API

- [ ] **Update docker-compose.yml**
  - Production database credentials
  - Production Redis configuration
  - Resource limits set
  - Health checks configured
  - Restart policies set

- [ ] **Environment Variables**
  ```yaml
  environment:
    - DATABASE_URL=postgresql://user:pass@postgres:5432/servereye
    - TIMESCALEDB_URL=postgresql://user:pass@timescaledb:5432/metrics
    - JWT_SECRET=${JWT_SECRET}
    - REDIS_URL=redis:6379
  ```

- [ ] **Volumes for Persistence**
  ```yaml
  volumes:
    - postgres_data:/var/lib/postgresql/data
    - timescaledb_data:/var/lib/postgresql/data
    - redis_data:/data
  ```

#### C# Backend

- [ ] **Create Dockerfile**
  ```dockerfile
  FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS base
  WORKDIR /app
  EXPOSE 80
  EXPOSE 443
  
  FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
  WORKDIR /src
  COPY . .
  RUN dotnet restore
  RUN dotnet build -c Release -o /app/build
  
  FROM build AS publish
  RUN dotnet publish -c Release -o /app/publish
  
  FROM base AS final
  WORKDIR /app
  COPY --from=publish /app/publish .
  ENTRYPOINT ["dotnet", "ServerEye.API.dll"]
  ```

- [ ] **Docker Compose Integration**
  ```yaml
  csharp-backend:
    build: ./ServerEyeWeb
    ports:
      - "5000:80"
    environment:
      - ASPNETCORE_ENVIRONMENT=Production
      - ConnectionStrings__DefaultConnection=${DB_CONNECTION}
    depends_on:
      - postgres
      - redis
  ```

### 📊 Monitoring & Logging

#### Go API

- [ ] **Configure Structured Logging**
  - Log level set to INFO in production
  - JSON format enabled
  - Log rotation configured

- [ ] **Health Checks**
  ```bash
  curl https://api.servereye.dev/health
  ```

- [ ] **Metrics Endpoint** (if using Prometheus)
  ```bash
  curl https://api.servereye.dev/metrics
  ```

#### C# Backend

- [ ] **Configure Serilog/NLog**
  - Production log level (Warning/Error)
  - Structured logging enabled
  - Log aggregation configured

- [ ] **Health Checks**
  ```csharp
  app.MapHealthChecks("/health");
  ```

- [ ] **Application Insights** (optional)
  ```json
  {
    "ApplicationInsights": {
      "InstrumentationKey": "your-key"
    }
  }
  ```

### 🚀 Performance

#### Redis Configuration

- [ ] **Persistence Enabled**
  ```conf
  appendonly yes
  appendfsync everysec
  ```

- [ ] **Memory Limits**
  ```conf
  maxmemory 2gb
  maxmemory-policy allkeys-lru
  ```

- [ ] **Connection Pool**
  - C# backend: Configure StackExchange.Redis
  - Go API: Configure go-redis pool size

#### Database Optimization

- [ ] **PostgreSQL Tuning**
  ```conf
  shared_buffers = 256MB
  effective_cache_size = 1GB
  maintenance_work_mem = 64MB
  checkpoint_completion_target = 0.9
  wal_buffers = 16MB
  default_statistics_target = 100
  random_page_cost = 1.1
  effective_io_concurrency = 200
  work_mem = 4MB
  min_wal_size = 1GB
  max_wal_size = 4GB
  ```

- [ ] **TimescaleDB Optimization**
  ```sql
  -- Set compression policy
  SELECT add_compression_policy('metrics_1m', INTERVAL '7 days');
  SELECT add_compression_policy('metrics_5m', INTERVAL '30 days');
  
  -- Set retention policy
  SELECT add_retention_policy('metrics_1m', INTERVAL '30 days');
  SELECT add_retention_policy('metrics_5m', INTERVAL '90 days');
  ```

- [ ] **Indexes Verified**
  ```sql
  -- Check index usage
  SELECT schemaname, tablename, indexname, idx_scan
  FROM pg_stat_user_indexes
  ORDER BY idx_scan DESC;
  ```

### 🔒 Network Security

- [ ] **Firewall Rules**
  - Only necessary ports open
  - Database ports not exposed publicly
  - Redis not exposed publicly

- [ ] **CORS Configuration**
  ```json
  {
    "AllowedOrigins": ["https://servereye.dev", "https://www.servereye.dev"]
  }
  ```

- [ ] **Rate Limiting**
  - Configured in C# backend
  - Configured in Go API (if applicable)
  - DDoS protection enabled

### 📦 Backup Strategy

- [ ] **Database Backups**
  ```bash
  # PostgreSQL backup script
  pg_dump -h localhost -U servereye servereye > backup_$(date +%Y%m%d).sql
  
  # TimescaleDB backup
  pg_dump -h localhost -U metrics timescaledb > metrics_backup_$(date +%Y%m%d).sql
  ```

- [ ] **Automated Backups**
  - Daily backups scheduled
  - Backup retention policy (30 days)
  - Backup verification process

- [ ] **Disaster Recovery Plan**
  - Restore procedure documented
  - Recovery time objective (RTO) defined
  - Recovery point objective (RPO) defined

### 🧪 Testing

- [ ] **Integration Tests**
  - C# backend can authenticate with Go API
  - Metrics endpoints return data
  - WebSocket connections work
  - Server key encryption/decryption works

- [ ] **Load Testing**
  ```bash
  # Example with Apache Bench
  ab -n 1000 -c 10 -H "X-API-Key: prod-key" \
     https://api.servereye.dev/api/servers/srv_test/metrics/realtime
  ```

- [ ] **Security Testing**
  - SQL injection tests
  - XSS tests
  - CSRF protection verified
  - API key validation tested

### 📝 Documentation

- [ ] **API Documentation Updated**
  - All endpoints documented
  - Authentication methods explained
  - Example requests/responses provided

- [ ] **Deployment Guide**
  - Step-by-step deployment instructions
  - Rollback procedures documented
  - Troubleshooting guide available

- [ ] **Operations Runbook**
  - Common issues and solutions
  - Monitoring alerts explained
  - Escalation procedures defined

## 🚀 Deployment Steps

### 1. Pre-Deployment

```bash
# 1. Backup current databases
./scripts/backup-databases.sh

# 2. Tag current version
git tag -a v1.0.0 -m "Production release v1.0.0"
git push origin v1.0.0

# 3. Build Docker images
docker-compose build

# 4. Run tests
dotnet test ServerEye.Tests
go test ./...
```

### 2. Deployment

```bash
# 1. Stop current services
docker-compose down

# 2. Pull latest images
docker-compose pull

# 3. Apply database migrations
docker-compose run --rm servereye-api migrate
docker-compose run --rm csharp-backend dotnet ef database update

# 4. Start services
docker-compose up -d

# 5. Verify health
curl https://api.servereye.dev/health
curl https://web.servereye.dev/health
```

### 3. Post-Deployment

```bash
# 1. Monitor logs
docker-compose logs -f

# 2. Check metrics
curl https://api.servereye.dev/api/metrics/summary

# 3. Verify WebSocket
# Use browser console or wscat

# 4. Test critical paths
./scripts/smoke-tests.sh
```

## 🔄 Rollback Procedure

If deployment fails:

```bash
# 1. Stop new services
docker-compose down

# 2. Restore previous version
git checkout v0.9.0
docker-compose up -d

# 3. Restore database (if needed)
psql -U servereye servereye < backup_YYYYMMDD.sql

# 4. Verify system
curl https://api.servereye.dev/health
```

## 📊 Post-Deployment Monitoring

### First 24 Hours

- [ ] Monitor error rates
- [ ] Check response times
- [ ] Verify database performance
- [ ] Monitor memory usage
- [ ] Check disk space
- [ ] Review security logs

### First Week

- [ ] Analyze user feedback
- [ ] Review performance metrics
- [ ] Check backup success
- [ ] Verify monitoring alerts
- [ ] Review security incidents

## ✅ Success Criteria

Deployment is successful when:

- ✅ All health checks pass
- ✅ API response times < 200ms (p95)
- ✅ Error rate < 0.1%
- ✅ Database connections stable
- ✅ WebSocket connections stable
- ✅ No security incidents
- ✅ Backups completing successfully
- ✅ Monitoring alerts working

## 🆘 Emergency Contacts

- **DevOps Lead:** [Contact Info]
- **Backend Lead:** [Contact Info]
- **Database Admin:** [Contact Info]
- **Security Team:** [Contact Info]

## 📞 Support Resources

- **Documentation:** https://docs.servereye.dev
- **Status Page:** https://status.servereye.dev
- **Incident Management:** [Tool/Process]
- **On-Call Schedule:** [Link]

---

**Last Updated:** 2026-02-15

**Version:** 1.0.0
