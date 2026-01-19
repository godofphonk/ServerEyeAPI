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