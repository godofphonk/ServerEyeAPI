# ServerEye API

<div align="center">

![ServerEye API](https://img.shields.io/badge/ServerEye-API-orange?style=for-the-badge&logo=servereye)
![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![API Version](https://img.shields.io/badge/API-v1.0-blue?style=for-the-badge)

**High-performance monitoring data collection and processing API**

[Quick Start](#-quick-start) • [API Reference](#-api-reference) • [Deployment](#-deployment) • [Architecture](#-architecture)

</div>

## Overview

ServerEye API is a robust, scalable backend service designed to collect, process, and store monitoring data from ServerEye agents. It provides real-time WebSocket communication, RESTful APIs, and comprehensive data management capabilities for enterprise monitoring environments.

### Key Features

- 🚀 **High Performance** - Concurrent WebSocket connections and efficient data processing
- 🔄 **Real-time Communication** - WebSocket support with automatic reconnection
- 📊 **Data Management** - PostgreSQL storage with Redis caching
- 🔐 **Enterprise Security** - JWT authentication, rate limiting, and TLS support
- 🐳 **Container Ready** - Docker deployment with health checks
- 📈 **Scalable Architecture** - Horizontal scaling with load balancer support

## 🚀 Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 12+
- Redis 6+
- Docker (optional)

### Installation

#### Using Docker Compose (Recommended)

```bash
# Clone repository
git clone https://github.com/godofphonk/ServerEyeAPI.git
cd ServerEyeAPI

# Start services
docker-compose up -d

# Check status
docker-compose ps
```

#### Manual Installation

```bash
# Clone repository
git clone https://github.com/godofphonk/ServerEyeAPI.git
cd ServerEyeAPI

# Install dependencies
go mod download

# Build application
go build -o servereye-api ./cmd/api

# Run with environment file
./servereye-api
```

### Configuration

Create a `.env` file based on `.env.example`:

```bash
# Server Configuration
HOST=0.0.0.0
PORT=8080

# Database
DATABASE_URL=postgres://user:password@localhost:5432/servereye?sslmode=disable
REDIS_URL=redis://localhost:6379

# Security
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
WEBHOOK_SECRET=your-webhook-secret

# Kafka (optional)
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=servereye-api
```

### Verification

```bash
# Health check
curl http://localhost:8080/health

# API version
curl http://localhost:8080/api/v1/version
```

## 📡 API Reference

### Authentication

All API endpoints require JWT authentication except for health checks and agent registration.

```bash
# Get JWT token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# Use token in requests
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/servers
```

### Endpoints

#### Health & System

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Service health status |
| GET | `/api/v1/version` | API version information |
| GET | `/api/v1/metrics` | Application metrics |

#### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | User authentication |
| POST | `/api/v1/auth/refresh` | Refresh JWT token |
| POST | `/api/v1/auth/logout` | User logout |

#### Servers

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/servers` | List all servers |
| GET | `/api/v1/servers/{id}` | Get server details |
| POST | `/api/v1/servers` | Register new server |
| PUT | `/api/v1/servers/{id}` | Update server |
| DELETE | `/api/v1/servers/{id}` | Delete server |

#### Metrics

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/metrics/servers/{id}` | Get server metrics |
| GET | `/api/v1/metrics/servers/{id}/latest` | Latest metrics |
| POST | `/api/v1/metrics` | Submit metrics (agent) |

#### Commands

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/commands` | List commands |
| POST | `/api/v1/commands` | Execute command |
| GET | `/api/v1/commands/{id}/status` | Command status |

#### WebSocket

| Endpoint | Description |
|----------|-------------|
| `/ws` | Real-time metrics and commands |

### WebSocket API

#### Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

// Authentication
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'auth',
    token: 'your-jwt-token'
  }));
};
```

#### Message Format

```json
{
  "id": "uuid",
  "type": "metric|command|heartbeat",
  "server_id": "server-uuid",
  "timestamp": "2024-01-01T00:00:00Z",
  "payload": {
    // Metric data or command response
  }
}
```

## 🏗️ Architecture

### System Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   ServerEye     │    │   Load Balancer │    │   ServerEye     │
│     Agent       │◄──►│    (Nginx)      │◄──►│      API        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                      │
                       ┌─────────────────┐            │
                       │   Redis Cache   │◄───────────┤
                       └─────────────────┘            │
                                                      │
                       ┌─────────────────┐            │
                       │   PostgreSQL    │◄───────────┘
                       │   Database      │
                       └─────────────────┘
```

### Data Flow

1. **Agent Registration** - Agents register with unique secret keys
2. **WebSocket Connection** - Persistent connection for real-time data
3. **Metrics Collection** - Periodic data transmission from agents
4. **Data Processing** - Validation, storage, and caching
5. **Command Distribution** - Remote commands sent to agents

### Database Schema

#### Servers Table
- `id` - Primary key (UUID)
- `name` - Server name
- `secret_key` - Authentication key
- `created_at` - Registration timestamp
- `last_seen` - Last activity timestamp

#### Metrics Table
- `id` - Primary key (UUID)
- `server_id` - Foreign key
- `metric_type` - Type of metric
- `value` - Metric value
- `timestamp` - Collection time

#### Commands Table
- `id` - Primary key (UUID)
- `server_id` - Target server
- `command_type` - Command type
- `status` - Execution status
- `created_at` - Command creation time

## 🚀 Deployment

### Docker Deployment

#### Docker Compose

```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://postgres:password@db:5432/servereye
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: servereye
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

#### Production Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./api"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: servereye-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: servereye-api
  template:
    metadata:
      labels:
        app: servereye-api
    spec:
      containers:
      - name: api
        image: servereye/api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: servereye-secrets
              key: database-url
---
apiVersion: v1
kind: Service
metadata:
  name: servereye-api-service
spec:
  selector:
    app: servereye-api
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### Environment Configuration

#### Development

```bash
# Development environment
export ENV=development
export LOG_LEVEL=debug
export HOST=localhost
export PORT=8080
```

#### Production

```bash
# Production environment
export ENV=production
export LOG_LEVEL=info
export HOST=0.0.0.0
export PORT=8080
export TLS_CERT_FILE=/path/to/cert.pem
export TLS_KEY_FILE=/path/to/key.pem
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `HOST` | Server bind address | `0.0.0.0` | No |
| `PORT` | Server port | `8080` | No |
| `DATABASE_URL` | PostgreSQL connection string | - | Yes |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` | No |
| `JWT_SECRET` | JWT signing secret | - | Yes |
| `WEBHOOK_SECRET` | Webhook validation secret | - | Yes |
| `KAFKA_BROKERS` | Kafka broker addresses | - | No |
| `METRICS_TOPIC` | Kafka metrics topic | `metrics` | No |

### Advanced Configuration

```yaml
# config.yaml (alternative to env vars)
server:
  host: "0.0.0.0"
  port: 8080
  timeout: 30s

database:
  url: "postgres://user:pass@localhost/db"
  max_connections: 25
  max_idle_connections: 5

redis:
  url: "redis://localhost:6379"
  ttl: 5m
  pool_size: 10

websocket:
  buffer_size: 256
  write_timeout: 30s
  read_timeout: 300s
  ping_interval: 60s

rate_limit:
  limit: 100
  window: 1m
```

## 🔐 Security

### Authentication

- **JWT Tokens**: Secure token-based authentication
- **Secret Keys**: Per-agent unique authentication
- **Rate Limiting**: Configurable request limits
- **CORS**: Cross-origin resource sharing control

### TLS Configuration

```bash
# Enable TLS
export TLS_CERT_FILE=/path/to/cert.pem
export TLS_KEY_FILE=/path/to/key.pem

# Or use Let's Encrypt
export TLS_AUTO_CERT=true
export TLS_DOMAIN=api.servereye.com
```

### Security Headers

```go
// Security middleware configuration
security := middleware.Security{
  AllowedHosts:     []string{"api.servereye.com"},
  SSLRedirect:      true,
  STSSeconds:       31536000,
  FrameDeny:        true,
  ContentTypeNosniff: true,
}
```

## 📊 Monitoring

### Application Metrics

The API exposes internal metrics for monitoring:

```bash
# Get application metrics
curl http://localhost:8080/api/v1/metrics
```

Available metrics:
- `http_requests_total` - HTTP request count
- `http_request_duration_seconds` - Request duration
- `websocket_connections_active` - Active WebSocket connections
- `database_connections_active` - Active database connections
- `redis_connections_active` - Active Redis connections

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check
curl http://localhost:8080/health/detailed
```

Health check response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0",
  "checks": {
    "database": "healthy",
    "redis": "healthy",
    "websocket": "healthy"
  }
}
```

### Logging

Structured JSON logging with configurable levels:

```bash
# Log levels: debug, info, warn, error
export LOG_LEVEL=info

# Log format: json, text
export LOG_FORMAT=json
```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/...
```

### Integration Tests

```bash
# Run integration tests
make test-integration

# With Docker Compose
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### API Testing

```bash
# Run API tests
make test-api

# Load testing
make test-load
```

## 🚨 Troubleshooting

### Common Issues

#### Database Connection

```bash
# Check database connectivity
psql $DATABASE_URL -c "SELECT 1"

# Verify connection string
echo $DATABASE_URL
```

#### Redis Connection

```bash
# Test Redis connection
redis-cli -u $REDIS_URL ping

# Check Redis logs
docker logs redis-container
```

#### WebSocket Issues

```bash
# Test WebSocket connection
wscat -c ws://localhost:8080/ws

# Check connection limits
ss -an | grep :8080
```

### Debug Mode

```bash
# Enable debug logging
export LOG_LEVEL=debug
export DEBUG=true

# Run with debug flags
./servereye-api --debug --log-level=debug
```

### Performance Tuning

#### Database Optimization

```sql
-- Create indexes
CREATE INDEX idx_metrics_server_timestamp ON metrics(server_id, timestamp);
CREATE INDEX idx_commands_status ON commands(status);

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM metrics WHERE server_id = $1;
```

#### Redis Optimization

```bash
# Configure Redis memory
redis-cli CONFIG SET maxmemory 256mb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

## 📚 Development

### Project Structure

```
cmd/
  api/                 # Application entry point
internal/
  api/                 # HTTP handlers and middleware
  config/              # Configuration management
  handlers/            # Business logic handlers
  models/              # Data models
  services/            # Business services
  storage/             # Database repositories
  wire/                # Dependency injection
pkg/                   # Public packages
scripts/               # Build and deployment scripts
```

### Adding New Features

1. **Model**: Define data structures in `internal/models/`
2. **Repository**: Implement storage in `internal/storage/`
3. **Service**: Add business logic in `internal/services/`
4. **Handler**: Create HTTP handlers in `internal/handlers/`
5. **Routes**: Register endpoints in `internal/api/routes.go`
6. **Tests**: Add unit and integration tests

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Security scan
gosec ./...

# Vulnerability check
govulncheck ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📖 [API Documentation](https://docs.servereye.com/api)
- 🐛 [Issue Tracker](https://github.com/godofphonk/ServerEyeAPI/issues)
- 💬 [Discussions](https://github.com/godofphonk/ServerEyeAPI/discussions)
- 📧 [Email Support](mailto:api-support@servereye.com)

## 🗺️ Roadmap

- [ ] GraphQL API support
- [ ] Metrics aggregation and analytics
- [ ] Advanced alerting system
- [ ] Multi-tenant support
- [ ] API versioning strategy
- [ ] Performance monitoring dashboard

---

<div align="center">

**Built with ❤️ by the ServerEye Team**

[![ServerEye](https://img.shields.io/badge/ServerEye-Enterprise%20Monitoring-orange?style=flat-square)](https://servereye.com)

</div>
