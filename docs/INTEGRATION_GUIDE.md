# C# Backend Integration Guide

## 🎯 Overview

This guide describes the integration between the C# Web Backend and Go API for ServerEye monitoring system.

## 🔐 Authentication

### API Key Authentication

The C# backend uses API Key authentication to communicate with Go API.

**Default API Key (Development):**
```text
sk_csharp_backend_development_key_change_in_production
```

**⚠️ IMPORTANT:** Change this key in production!

### Configuration

**C# Backend (appsettings.json):**
```json
{
  "GoApiSettings": {
    "BaseUrl": "http://localhost:8080",
    "ProductionUrl": "https://api.servereye.dev",
    "ApiKey": "sk_csharp_backend_development_key_change_in_production",
    "ServiceId": "csharp-backend",
    "TimeoutSeconds": 30
  }
}
```

**Go API:**
- API Key is stored in PostgreSQL database
- Hashed with bcrypt
- Validated via middleware on protected endpoints

## 📊 Available Endpoints

### 1. Tiered Metrics (Auto-Granularity)

**Endpoint:** `GET /api/servers/{server_id}/metrics/tiered`

**Headers:**
```http
X-API-Key: sk_csharp_backend_development_key_change_in_production
```

**Query Parameters:**
- `start` (required): RFC3339 timestamp
- `end` (required): RFC3339 timestamp

**Example Request:**
```bash
curl -H "X-API-Key: sk_csharp_backend_development_key_change_in_production" \
     "http://localhost:8080/api/servers/srv_a3d881f1/metrics/tiered?start=2026-02-15T15:00:00Z&end=2026-02-15T16:00:00Z"
```

**Response:**
```json
{
  "server_id": "srv_a3d881f1",
  "granularity": "1m",
  "total_points": 23,
  "metrics": [...]
}
```

**Granularity Strategy:**
- Last 1 hour: 1-minute intervals
- Last 3 hours: 5-minute intervals
- Last 24 hours: 10-minute intervals
- Last 30 days: 1-hour intervals

### 2. Real-Time Metrics

**Endpoint:** `GET /api/servers/{server_id}/metrics/realtime`

**Headers:**
```http
X-API-Key: sk_csharp_backend_development_key_change_in_production
```

**Query Parameters:**
- `duration` (optional): Duration string (default: "1h", max: "1h")

**Example:**
```bash
curl -H "X-API-Key: sk_csharp_backend_development_key_change_in_production" \
     "http://localhost:8080/api/servers/srv_a3d881f1/metrics/realtime?duration=30m"
```

### 3. Historical Metrics

**Endpoint:** `GET /api/servers/{server_id}/metrics/historical`

**Headers:**
```http
X-API-Key: sk_csharp_backend_development_key_change_in_production
```

**Query Parameters:**
- `start` (required): RFC3339 timestamp
- `end` (required): RFC3339 timestamp
- `granularity` (optional): "1m", "5m", "10m", "1h"

### 4. Dashboard Metrics

**Endpoint:** `GET /api/servers/{server_id}/metrics/dashboard`

**Headers:**
```http
X-API-Key: sk_csharp_backend_development_key_change_in_production
```

**Returns:** Optimized metrics for dashboard display

### 5. Metrics Summary

**Endpoint:** `GET /api/metrics/summary`

**Headers:**
```http
X-API-Key: sk_csharp_backend_development_key_change_in_production
```

**Returns:** Storage statistics across all servers

## 🔌 WebSocket Integration

### JWT Token Generation

C# backend generates JWT tokens for WebSocket connections.

**Token Claims:**
```json
{
  "user_id": "123",
  "server_id": "srv_a3d881f1",
  "jti": "unique-token-id",
  "iat": 1708012345,
  "exp": 1708014145
}
```

**Token TTL:** 30 minutes

**Shared Secret:**
Both C# and Go API must use the same `JWT_SECRET` for token validation.

### WebSocket Connection

**Endpoint:** `ws://localhost:8080/ws`

**Query Parameters:**
```
?token=<jwt_token>
```

**Example:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=' + jwtToken);
```

**Go API Validation:**
- Validates JWT signature
- Checks expiration
- Extracts user_id and server_id from claims
- Streams real-time metrics

## 🔒 Security Implementation

### 1. Server Key Encryption (C# Backend)

**Algorithm:** AES-256-CBC

**Features:**
- Random IV for each encryption
- SHA-256 key derivation
- PKCS7 padding
- IV stored with encrypted data

**C# Implementation:**
```csharp
public class EncryptionService : IEncryptionService
{
    private readonly byte[] key;

    public EncryptionService(IOptions<EncryptionSettings> settings)
    {
        this.key = SHA256.HashData(Encoding.UTF8.GetBytes(settings.Value.Key));
    }

    public string Encrypt(string plainText)
    {
        using var aes = Aes.Create();
        aes.Key = this.key;
        aes.GenerateIV(); // Random IV for each encryption
        
        // ... encryption logic
    }
}
```

### 2. API Key Storage (Go API)

**Storage:** PostgreSQL database

**Hashing:** bcrypt (cost factor 10)

**Audit Logging:**
- Every API key usage is logged
- IP address tracking
- User agent tracking
- Timestamp recording

**Database Schema:**
```sql
CREATE TABLE api_keys (
    key_id VARCHAR(255) PRIMARY KEY,
    key_hash TEXT NOT NULL,
    service_id VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    permissions TEXT[] NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_by VARCHAR(255),
    notes TEXT
);
```

### 3. JWT Token Validation (Go API)

**Algorithm:** HMAC-SHA256

**Validation Steps:**
1. Parse JWT token
2. Verify signature with shared secret
3. Check expiration
4. Extract claims
5. Validate user_id and server_id

**Go Implementation:**
```go
func (a *WebSocketAuthenticator) ValidateToken(tokenString string) (*WebSocketClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &WebSocketClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(a.jwtSecret), nil
    })
    
    if claims, ok := token.Claims.(*WebSocketClaims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}
```

## 🔄 Data Flow

### Adding a Server

```
Frontend → C# API: ServerKey
    ↓
C# API → Go API: Validate ServerKey (X-API-Key header)
    ↓
C# API → PostgreSQL: Store encrypted ServerKey (AES-256)
    ↓
Frontend ← C# API: Success response
```

### Fetching Metrics

```
Frontend → C# API: JWT token
    ↓
C# API: Check user permissions
    ↓
C# API → Redis: Check cache
    ↓ (cache miss)
C# API → Go API: Request metrics (X-API-Key header)
    ↓
Go API → TimescaleDB: Query metrics
    ↓
Go API → C# API: Return metrics
    ↓
C# API → Redis: Cache metrics (smart TTL)
    ↓
C# API → Frontend: Return metrics
```

### WebSocket Live Data

```
Frontend → C# API: Request WebSocket token
    ↓
C# API: Generate JWT token (30 min TTL)
    ↓
Frontend ← C# API: JWT token
    ↓
Frontend → Go API WebSocket: Connect with JWT
    ↓
Go API: Validate JWT
    ↓
Go API → Frontend: Stream real-time metrics
```

## 📦 Caching Strategy

### Redis Cache TTL

**C# Backend Configuration:**
```json
{
  "CacheSettings": {
    "LiveMetrics": "00:01:00",      // 1 minute
    "HourMetrics": "00:05:00",      // 5 minutes
    "DayMetrics": "00:15:00",       // 15 minutes
    "MonthMetrics": "01:00:00",     // 1 hour
    "ServerList": "00:10:00"        // 10 minutes
  }
}
```

**Cache Keys:**
```
ServerEye:metrics:live:{server_id}
ServerEye:metrics:hour:{server_id}:{start}:{end}
ServerEye:metrics:day:{server_id}:{start}:{end}
ServerEye:metrics:month:{server_id}:{start}:{end}
```

## 🚀 Testing Integration

### 1. Test API Key

```bash
# Test API key works
curl -H "X-API-Key: sk_csharp_backend_development_key_change_in_production" \
     http://localhost:8080/api/admin/keys
```

**Expected:** List of API keys

### 2. Test Metrics Endpoint

```bash
# Test metrics retrieval
curl -H "X-API-Key: sk_csharp_backend_development_key_change_in_production" \
     "http://localhost:8080/api/servers/srv_a3d881f1/metrics/tiered?start=2026-02-15T15:00:00Z&end=2026-02-15T16:00:00Z"
```

**Expected:** JSON with metrics data

### 3. Test WebSocket Token

**C# Backend:**
```csharp
var token = await webSocketTokenService.GenerateTokenAsync(userId, serverId);
```

**Frontend:**
```javascript
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);
ws.onmessage = (event) => {
    const metrics = JSON.parse(event.data);
    console.log('Real-time metrics:', metrics);
};
```

## ⚠️ Production Checklist

### Before Deployment

- [ ] Change `GoApiSettings.ApiKey` to production key
- [ ] Change `Encryption.Key` to secure 32-character key
- [ ] Ensure `JwtSettings.SecretKey` matches Go API
- [ ] Configure production Redis connection
- [ ] Apply database migrations
- [ ] Test API key authentication
- [ ] Test WebSocket JWT validation
- [ ] Test server key encryption/decryption
- [ ] Configure HTTPS for all endpoints
- [ ] Set up monitoring and alerting
- [ ] Configure rate limiting
- [ ] Set up health checks
- [ ] Review security audit logs

### Environment Variables

**C# Backend:**
```bash
export GoApiSettings__ApiKey="production-api-key"
export Encryption__Key="secure-32-character-encryption-key"
export JwtSettings__SecretKey="same-as-go-api-jwt-secret"
export Redis__ConnectionString="production-redis:6379"
```

**Go API:**
```bash
export JWT_SECRET="same-as-csharp-backend-jwt-secret"
export DATABASE_URL="postgresql://user:pass@host:5432/db"
export TIMESCALEDB_URL="postgresql://user:pass@host:5432/timescaledb"
```

## 📊 Available Test Servers

```
srv_a3d881f1  - Primary test server
srv_6fb4cb4e  - Additional server
srv_7d8cfe79  - Additional server
srv_bd84f46e  - Additional server
srv_e92c5907  - Additional server
```

## 🔧 Troubleshooting

### API Key Not Working

**Check:**
1. API key is correct in appsettings.json
2. X-API-Key header is being sent
3. Go API is running
4. API key exists in database: `SELECT * FROM api_keys WHERE service_id = 'csharp-backend';`

### WebSocket Connection Failed

**Check:**
1. JWT secret matches between C# and Go
2. Token is not expired (30 min TTL)
3. Token includes correct claims (user_id, server_id)
4. WebSocket endpoint is accessible

### Metrics Not Loading

**Check:**
1. Server ID exists in Go API
2. Time range is valid
3. Redis is running
4. Go API can access TimescaleDB

## 📞 Support

**Go API Issues:**
- Check logs: `docker-compose logs servereye-api`
- Health check: `curl http://localhost:8080/health`

**C# Backend Issues:**
- Check logs in console output
- Health check: `curl http://localhost:5000/health`

**Database Issues:**
- PostgreSQL: `docker-compose logs servereye-postgres`
- TimescaleDB: `docker-compose logs servereye-timescaledb`
- Redis: `redis-cli ping`

## 🎉 Success Criteria

Integration is successful when:

✅ C# backend can authenticate with Go API using X-API-Key
✅ C# backend can fetch metrics from Go API
✅ Server keys are encrypted in C# database
✅ WebSocket JWT tokens are validated by Go API
✅ Real-time metrics stream through WebSocket
✅ Redis caching works correctly
✅ All endpoints return expected data

---

**Last Updated:** 2026-02-15

**Version:** 1.0.0
