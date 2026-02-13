# ServerEyeAPI

[![Go Report Card](https://goreportcard.com/badge/github.com/godofphonk/ServerEyeAPI)](https://goreportcard.com/report/github.com/godofphonk/ServerEyeAPI)



## Base URL

```
https://api.servereye.dev
```

## Authentication

Most API endpoints require authentication via Bearer token. Protected endpoints are marked with 🔒 in the documentation.


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

## Tiered Metrics Endpoints

### Overview
The ServerEyeAPI provides multi-tier metrics storage with automatic granularity selection based on time ranges.

### Granularity Strategy
- **Last hour**: 1-minute intervals
- **Last 3 hours**: 5-minute intervals
- **Last 24 hours**: 10-minute intervals
- **Last 30 days**: 1-hour intervals

### Get Metrics with Auto-Granularity
Automatically selects the best granularity based on time range.

**Endpoint:** `GET /api/servers/{server_id}/metrics/tiered`

**Query Parameters:**
- `start` (string, required): Start time (RFC3339 format)
- `end` (string, required): End time (RFC3339 format)

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/tiered?start=2026-02-13T15:00:00Z&end=2026-02-13T16:00:00Z"
```

### Get Real-Time Metrics
Get the most recent metrics with 1-minute granularity.

**Endpoint:** `GET /api/servers/{server_id}/metrics/realtime`

**Query Parameters:**
- `duration` (string, optional): Duration (default: "1h", max: "1h")

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/realtime?duration=30m"
```

### Get Dashboard Metrics
Optimized endpoint for dashboard displays.

**Endpoint:** `GET /api/servers/{server_id}/metrics/dashboard`

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/dashboard"
```

### Get Historical Metrics
Get historical metrics with specified granularity.

**Endpoint:** `GET /api/servers/{server_id}/metrics/historical`

**Query Parameters:**
- `start` (string, required): Start time (RFC3339 format)
- `end` (string, required): End time (RFC3339 format)
- `granularity` (string, optional): "1m", "5m", "10m", "1h"

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/historical?start=2026-02-12T00:00:00Z&end=2026-02-13T00:00:00Z&granularity=1h"
```

### Compare Metrics Between Periods
Compare metrics between two time periods.

**Endpoint:** `GET /api/servers/{server_id}/metrics/comparison`

**Query Parameters:**
- `period1_start` (string, required): First period start time
- `period1_end` (string, required): First period end time
- `period2_start` (string, required): Second period start time
- `period2_end` (string, required): Second period end time

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/comparison?period1_start=2026-02-12T00:00:00Z&period1_end=2026-02-12T12:00:00Z&period2_start=2026-02-12T12:00:00Z&period2_end=2026-02-13T00:00:00Z"
```

### Get Metrics Heatmap
Get metrics data for heatmap visualization.

**Endpoint:** `GET /api/servers/{server_id}/metrics/heatmap`

**Query Parameters:**
- `start` (string, required): Start time (RFC3339 format)
- `end` (string, required): End time (RFC3339 format)

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/heatmap?start=2026-02-13T00:00:00Z&end=2026-02-13T23:59:59Z"
```

### Get Metrics Summary
Get storage statistics across all granularity levels.

**Endpoint:** `GET /api/metrics/summary`

**Example:**
```bash
curl "http://localhost:8080/api/metrics/summary"
```

For detailed documentation, see [Tiered Metrics API Documentation](docs/api/tiered-metrics-endpoints.md).

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