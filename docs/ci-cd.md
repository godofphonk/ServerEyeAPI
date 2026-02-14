# CI/CD Pipeline Documentation

## Overview

This document describes the CI/CD pipeline for ServerEyeAPI, which automates testing, building, and deployment of the application.

## Pipeline Structure

The CI/CD pipeline is defined in `.github/workflows/ci.yml` and consists of the following jobs:

### 1. Test Job

- **Triggers**: Push to master/production, Pull Requests
- **Actions**:
  - Downloads Go dependencies
  - Runs unit tests with race detection
  - Uploads coverage reports to Codecov
  - Performs security scanning with gosec
  - Checks for vulnerabilities with govulncheck

### 2. Lint Job

- **Triggers**: Push to master/production, Pull Requests
- **Actions**:
  - Runs golangci-lint for code quality
  - Checks code formatting with gofmt

### 3. Build Job

- **Triggers**: Only on push to production branch
- **Actions**:
  - Builds Docker image with multi-stage build
  - Pushes to GitHub Container Registry
  - Verifies multi-tier metrics implementation
  - Tags images with version, branch, and commit SHA

### 4. Deploy Job

- **Triggers**: Only on push to production branch
- **Environment**: production
- **Actions**:
  - Creates backup of current deployment
  - Pulls latest Docker image
  - Updates docker-compose configuration
  - Performs rolling deployment
  - Runs health checks
  - Cleans up old Docker images

### 5. Post-deployment Tests

- **Triggers**: After successful deployment
- **Actions**:
  - Tests health endpoint
  - Validates metrics endpoints
  - Confirms service availability

## Environment Variables

### Required Secrets

Configure these in GitHub repository settings:

- `PROD_HOST`: Production server hostname
- `PROD_USER`: SSH username for production server
- `PROD_SSH_KEY`: SSH private key for deployment
- `POSTGRES_PASSWORD`: Database password
- `JWT_SECRET`: JWT signing secret
- `WEBHOOK_SECRET`: Webhook signing secret

### Optional Variables

- `WEB_URL`: Base URL for the service (default: <https://api.servereye.dev>)

## Docker Image Management

### Image Tags

- `production`: Latest production build
- `latest`: Latest build from default branch
- `{branch}-{commit}`: Specific commit build
- `pr-{number}`: Pull request build

### Registry

Images are stored in GitHub Container Registry:

```text
ghcr.io/godofphonk/servereyeapi
```

## Deployment Process

### Pre-deployment

1. All tests must pass
2. Code must be properly formatted
3. Security scans must pass
4. Docker image must build successfully

### Deployment Steps

1. Create backup of current deployment
2. Pull new Docker image
3. Update docker-compose configuration
4. Stop existing services
5. Start new services
6. Wait for health checks
7. Verify deployment success
8. Clean up old resources

### Post-deployment

1. Run health endpoint tests
2. Verify metrics endpoints
3. Monitor service logs
4. Clean up old Docker images

## Multi-tier Metrics Verification

The pipeline includes specific verification for the multi-tier metrics implementation:

### File Checks

- `internal/services/tiered_metrics.go` - Service layer
- `internal/services/metrics_commands.go` - Command handlers
- `internal/storage/timescaledb/multi_tier_metrics.go` - Storage layer

### Implementation Checks

- `TieredMetricsService` implementation
- `MetricsCommandsService` implementation
- Proper integration with dependency injection

## Local Development

### Running Tests

```bash
make test              # Run all tests
make test-coverage     # Run with coverage
make test-all          # Run tests with coverage
```

### Building

```bash
make build             # Build the application
make docker-build      # Build Docker image
make release           # Build release binary
```

### Code Quality

```bash
make fmt               # Format code
make lint              # Run linter
make security          # Security scan
make vuln-check        # Vulnerability check
```

## Troubleshooting

### Common Issues

#### 1. Build Failures

- Check Go version compatibility (requires 1.21+)
- Verify all dependencies are downloaded
- Check Wire generation: `go generate ./internal/wire`

#### 2. Test Failures

- Ensure database is available for integration tests
- Check environment variables are set
- Verify test data setup

#### 3. Deployment Failures

- Check SSH key permissions
- Verify production server connectivity
- Check Docker daemon status on production
- Review service logs

### Debug Commands

```bash
# Check deployment status
docker-compose -f docker-compose.prod.yml ps

# View service logs
docker-compose -f docker-compose.prod.yml logs -f

# Restart services
docker-compose -f docker-compose.prod.yml restart

# Check health endpoint
curl http://localhost:8080/health
```

## Rollback Procedure

If deployment fails:

### 1. Automatic rollback

The pipeline includes health checks that prevent failed deployments

### 2. Manual rollback

```bash
# Stop current deployment
docker-compose -f docker-compose.prod.yml down

# Restore from backup
sudo cp -r /opt/servereye.backup.TIMESTAMP /opt/servereye

# Start previous version
docker-compose -f docker-compose.prod.yml up -d
```

## Monitoring and Alerts

### Health Checks

- Application health: `/health`
- Database connectivity: Built into application
- Service dependencies: Docker health checks

### Logs

- Application logs: `/app/logs/`
- Docker logs: `docker-compose logs`
- System logs: Available via SSH

## Security Considerations

1. **Secrets Management**: All secrets stored in GitHub Secrets
2. **Image Security**: Multi-stage builds minimize attack surface
3. **Network Security**: Services isolated in Docker networks
4. **Access Control**: SSH key-based authentication only

## Performance Optimizations

1. **Build Caching**: GitHub Actions cache for faster builds
2. **Docker Layer Caching**: Optimized Dockerfile structure
3. **Parallel Execution**: Jobs run in parallel where possible
4. **Resource Limits**: Defined resource constraints in docker-compose

## Future Improvements

1. **Blue-Green Deployment**: Zero-downtime deployments
2. **Canary Releases**: Gradual rollout with monitoring
3. **Automated Testing**: Add integration and E2E tests
4. **Monitoring Integration**: Prometheus/Grafana setup
5. **Alerting**: Slack/email notifications for failures
