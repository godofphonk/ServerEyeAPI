# ServerEyeAPI

[![Go Report Card](https://goreportcard.com/badge/github.com/godofphonk/ServerEyeAPI)](https://goreportcard.com/report/github.com/godofphonk/ServerEyeAPI)

## Base URL

```text
https://api.servereye.dev
```

## Authentication

ServerEyeAPI supports multiple authentication methods:

### 1. Server Key Authentication

For server agents and basic endpoints using server-specific keys.

### 2. API Key Authentication

For service-to-service communication (recommended for backend integration).

**Headers:**

```text
X-API-Key: sk_your_api_key_here
```

**Default C# Backend API Key:**

```text
sk_csharp_backend_development_key_change_in_production
```

### 3. Bearer Token Authentication

For protected admin endpoints (marked with 🔒).

Protected endpoints are marked with 🔒 in the documentation.

## Endpoints

### 🔓 Authentication

#### Register Server Key
Registers a new server and generates authentication credentials.

**Endpoint:** `POST /RegisterKey`

**Request Body:**

```json
{
  "hostname": "server-01",
  "operating_system": "Ubuntu 22.04",
  "agent_version": "1.0.0"
}
```

**Response (201 Created):**

```json
{
  "server_id": "srv_123456789",
  "server_key": "sk_abcdef123456",
  "status": "registered"
}
```

---

### 🔓 Health Check

#### System Health

Checks the health status of the API server and its dependencies.

**Endpoint:** `GET /health`

**Response (200 OK):**

```json
{
  "status": "healthy",
  "timestamp": "2026-01-19T05:29:00Z",
  "version": "1.0.0",
  "clients": 5
}
```


---

### 🔓 Metrics (Public)

#### Get Server Metrics by ID

Retrieves current metrics and status for a specific server.

**Endpoint:** `GET /api/servers/{server_id}/metrics` **KEY FOR TEST - "key_954492a7"**

**Path Parameters:**

- `server_id` (string, required): Unique server identifier

**Response (200 OK):**

```json
{
  "server_id": "srv_123456789",
  "status": {
    "online": true,
    "last_seen": "2026-01-19T05:29:00Z",
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 23.1,
    "network_rx": 1024,
    "network_tx": 2048
  },
  "metrics": {
    "timestamp": "2026-01-19T05:29:00Z",
    "uptime": 86400,
    "load_average": [0.5, 0.3, 0.2],
    "processes": 156
  }
}
```

**Error Responses:**

- `400 Bad Request` - Missing server_id
- `404 Not Found` - Server not found
- `500 Internal Server Error` - Failed to retrieve metrics

#### Get Server Metrics by Key
Retrieves metrics using server key instead of ID (for Telegram bot integration).

**Endpoint:** `GET /api/servers/by-key/{server_key}/metrics` 

**Path Parameters:**
- `server_key` (string, required): Server authentication **KEY FOR TEST - "key_954492a7"**

**Response (200 OK):**
```json
{
  "server_id": "srv_123456789",
  "server_key": "sk_abcdef123456",
  "status": {
    "online": true,
    "last_seen": "2026-01-19T05:29:00Z",
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 23.1
  },
  "metrics": {
    "timestamp": "2026-01-19T05:29:00Z",
    "uptime": 86400,
    "load_average": [0.5, 0.3, 0.2]
  }
}
```

---

### 🔓 Server Sources Management (Public)

#### Add Server Source by ID
Adds a notification source for a specific server.

**Endpoint:** `POST /api/servers/{server_id}/sources`

**Path Parameters:**
- `server_id` (string, required): Unique server identifier

**Request Body:**
```json
{
  "source": "TGBot"
}
```

**Source Values:**
- `"TGBot"` - Telegram Bot 
- `"Web"` - Web dashboard 

**Response (200 OK):**
```json
{
  "server_id": "srv_123456789",
  "source": "TGBot",
  "message": "Source added successfully"
}
```

#### Get Server Sources by ID
Retrieves all notification sources for a server.

**Endpoint:** `GET /api/servers/{server_id}/sources`

**Path Parameters:**
- `server_id` (string, required): Unique server identifier **KEY FOR TEST - "key_954492a7"**

**Response (200 OK):**
```json
{
  "server_id": "srv_123456789",
  "sources": ["TGBot", "Web"]
}
```

#### Remove Server Source by ID
Removes a notification source from a server.

**Endpoint:** `DELETE /api/servers/{server_id}/sources/{source}`

**Path Parameters:**
- `server_id` (string, required): Unique server identifier
- `source` (string, required): Source type ("TGBot" or "Web")

**Response (200 OK):**
```json
{
  "server_id": "srv_123456789",
  "source": "TGBot",
  "message": "Source removed successfully"
}
```

#### Add Server Source by Key
Adds notification source using server key.

**Endpoint:** `POST /api/servers/by-key/{server_key}/sources`

**Path Parameters:**
- `server_key` (string, required): Server authentication key

**Request Body:**
```json
{
  "source": "Web"
}
```

**Response (200 OK):**
```json
{
  "message": "Source added successfully",
  "server_id": "srv_123456789",
  "server_key": "sk_abcdef123456",
  "source": "Web"
}
```

#### Get Server Sources by Key
Retrieves notification sources using server key.

**Endpoint:** `GET /api/servers/by-key/{server_key}/sources`

**Path Parameters:**
- `server_key` (string, required): Server authentication key

**Response (200 OK):**
```json
{
  "server_id": "srv_123456789",
  "server_key": "sk_abcdef123456",
  "sources": ["TGBot", "Web"]
}
```

#### Remove Server Source by Key
Removes notification source using server key.

**Endpoint:** `DELETE /api/servers/by-key/{server_key}/sources/{source}`

**Path Parameters:**
- `server_key` (string, required): Server authentication key
- `source` (string, required): Source type ("TGBot" or "Web")

**Response (200 OK):**
```json
{
  "message": "Source removed successfully",
  "server_id": "srv_123456789",
  "server_key": "sk_abcdef123456",
  "source": "TGBot"
}
```

---

### 🔐 API Key Management

#### Create API Key
Creates a new API key for service authentication.

**Endpoint:** `POST /api/admin/keys`

**Headers:**
- `X-API-Key: <admin_api_key>`

**Request Body:**
```json
{
  "service_id": "csharp-backend",
  "service_name": "C# Web Backend",
  "permissions": ["metrics:read", "servers:read", "servers:validate"],
  "expires_days": 365
}
```

**Response (201 Created):**
```json
{
  "api_key": "sk_VhausxMPKH40oH66je21EWErL3JmTH8S",
  "key_id": "key_j6mdji4Kjm_UiIGn26XQVg",
  "service_id": "csharp-backend",
  "service_name": "C# Web Backend",
  "permissions": ["metrics:read", "servers:read", "servers:validate"],
  "created_at": "2026-02-15T16:51:37.881571749Z"
}
```

#### List API Keys
Retrieves all API keys.

**Endpoint:** `GET /api/admin/keys`

**Headers:**
- `X-API-Key: <admin_api_key>`

**Response (200 OK):**
```json
[
  {
    "key_id": "key_j6mdji4Kjm_UiIGn26XQVg",
    "service_id": "csharp-backend",
    "service_name": "C# Web Backend",
    "permissions": ["metrics:read", "servers:read"],
    "created_at": "2026-02-15T16:51:37.882303Z",
    "is_active": true,
    "last_used_at": "2026-02-15T16:52:00Z"
  }
]
```

#### Get API Key Details
Retrieves details for a specific API key.

**Endpoint:** `GET /api/admin/keys/{keyId}`

**Headers:**
- `X-API-Key: <admin_api_key>`

#### Revoke API Key
Deactivates an API key.

**Endpoint:** `DELETE /api/admin/keys/{keyId}`

**Headers:**
- `X-API-Key: <admin_api_key>`

---

## Metrics Endpoint

### Overview
The ServerEyeAPI provides a unified metrics endpoint with automatic granularity selection based on time ranges.

### Granularity Strategy
- **Last hour**: 1-minute intervals
- **Last 3 hours**: 5-minute intervals
- **Last 24 hours**: 10-minute intervals
- **Last 30 days**: 1-hour intervals

### Get Metrics with Auto-Granularity
Unified endpoint for all metrics queries. Automatically selects the best granularity based on time range.

**Endpoint:** `GET /api/servers/{server_id}/metrics/tiered`

**Query Parameters:**
- `start` (string, required): Start time (RFC3339 format)
- `end` (string, required): End time (RFC3339 format)

**Response (with data):**
```json
{
  "server_id": "srv_71453434",
  "start_time": "2026-02-15T19:00:00Z",
  "end_time": "2026-02-15T20:00:00Z",
  "granularity": "1m",
  "data_points": [
    {
      "timestamp": "2026-02-15T19:18:00Z",
      "cpu_avg": 18.53,
      "cpu_max": 18.54,
      "cpu_min": 18.52,
      "memory_avg": 71.31,
      "memory_max": 71.85,
      "memory_min": 70.84,
      "disk_avg": 66,
      "disk_max": 66,
      "network_avg": 0.37,
      "network_max": 2.12,
      "temp_avg": 48.08,
      "temp_max": 60.25,
      "load_avg": 3.12,
      "load_max": 3.75,
      "sample_count": 50
    }
  ],
  "total_points": 26
}
```

**Response (showing available data when requested period is empty):**
```json
{
  "server_id": "srv_71453434",
  "start_time": "2026-02-14T19:00:00Z",
  "end_time": "2026-02-15T19:00:00Z",
  "granularity": "1m",
  "data_points": [
    {
      "timestamp": "2026-02-15T18:18:00Z",
      "cpu_avg": 18.53,
      "memory_avg": 71.31,
      "..."
    }
  ],
  "total_points": 26,
  "message": "Showing available data (requested period had no data)"
}
```

**Note:** If the requested time period has no data (e.g., server was recently installed), the API will automatically return the latest available metrics with a `message` field explaining the situation. This ensures the frontend always has data to display.

### Usage Examples

**Dashboard (5 minutes):**
```bash
curl "http://localhost:8080/api/servers/srv_71453434/metrics/tiered?start=2026-02-15T19:00:00Z&end=2026-02-15T19:05:00Z"
```

**Realtime (1 hour):**
```bash
curl "http://localhost:8080/api/servers/srv_71453434/metrics/tiered?start=2026-02-15T18:00:00Z&end=2026-02-15T19:00:00Z"
```

**Historical (6 hours):**
```bash
curl "http://localhost:8080/api/servers/srv_71453434/metrics/tiered?start=2026-02-15T13:00:00Z&end=2026-02-15T19:00:00Z"
```

**Historical (24 hours):**
```bash
curl "http://localhost:8080/api/servers/srv_71453434/metrics/tiered?start=2026-02-14T19:00:00Z&end=2026-02-15T19:00:00Z"
```

**Historical (7 days):**
```bash
curl "http://localhost:8080/api/servers/srv_71453434/metrics/tiered?start=2026-02-08T19:00:00Z&end=2026-02-15T19:00:00Z"
```

**Historical (30 days):**
```bash
curl "http://localhost:8080/api/servers/srv_71453434/metrics/tiered?start=2026-01-16T19:00:00Z&end=2026-02-15T19:00:00Z"
```

---

## Metrics Management Commands

The ServerEyeAPI provides commands for managing the multi-tier metrics system.

### Available Commands

1. **Refresh Aggregates** - Update continuous aggregates with latest data
2. **Rebuild Aggregates** - Rebuild aggregates from scratch
3. **Cleanup Old Metrics** - Remove old data based on retention policies
4. **Compression Policy** - Apply compression to save storage
5. **RetentionPolicy** - Configure automatic data deletion
6. **Metrics Statistics** - Get storage and performance statistics
7. **Analyze Performance** - Analyze query performance
8. **Export/Import Metrics** - Backup and restore metrics data
9. **Validate Metrics** - Check data integrity
10. **Optimize Storage** - Optimize TimescaleDB performance

### Example Usage

```bash
# Refresh all aggregates
curl -X POST http://localhost:8080/api/servers/management/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "server_id": "management",
    "type": "refresh_aggregates",
    "payload": {"granularity": "all"}
  }'

# Get metrics statistics
curl -X POST http://localhost:8080/api/servers/management/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "server_id": "management",
    "type": "metrics_stats",
    "payload": {}
  }'

# Cleanup old metrics (dry run)
curl -X POST http://localhost:8080/api/servers/management/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "server_id": "management",
    "type": "cleanup_old_metrics",
    "payload": {
      "older_than": "90 days",
      "dry_run": true
    }
  }'
```

For detailed documentation, see [Metrics Management Commands](docs/metrics-commands.md).

---

### 🔒 Server Management (Protected)

#### List All Servers
Retrieves list of all registered servers with their status.

**Endpoint:** `GET /api/servers`

**Headers:**
- `Authorization: Bearer <token>`

**Response (200 OK):**
```json
{
  "count": 2,
  "servers": [
    {
      "server_id": "srv_123456789",
      "status": {
        "online": true,
        "last_seen": "2026-01-19T05:29:00Z",
        "cpu_usage": 45.2,
        "memory_usage": 67.8
      }
    },
    {
      "server_id": "srv_987654321",
      "status": {
        "online": false,
        "last_seen": "2026-01-19T05:20:00Z"
      }
    }
  ],
  "timestamp": "2026-01-19T05:29:00Z"
}
```

#### Get Server Status
Retrieves detailed status information for a specific server.

**Endpoint:** `GET /api/servers/{server_id}/status`

**Headers:**
- `Authorization: Bearer <token>`

**Path Parameters:**
- `server_id` (string, required): Unique server identifier

**Response (200 OK):**
```json
{
  "server_id": "srv_123456789",
  "status": {
    "online": true,
    "last_seen": "2026-01-19T05:29:00Z",
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 23.1,
    "network_rx": 1024,
    "network_tx": 2048,
    "uptime": 86400,
    "load_average": [0.5, 0.3, 0.2],
    "processes": 156
  }
}
```