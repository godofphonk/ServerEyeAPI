# ServerEyeAPI

[![Go Report Card](https://goreportcard.com/badge/github.com/godofphonk/ServerEyeAPI)](https://goreportcard.com/report/github.com/godofphonk/ServerEyeAPI)

## Base URL

```text
https://api.servereye.dev
```

## Authentication

ServerEyeAPI supports multiple authentication methods:

### 1. API Key Authentication (Recommended)

For service-to-service communication.

**Headers:**

```text
X-API-Key: sk_your_api_key_here
```

**Default Development Key:**

```text
sk_csharp_backend_development_key_change_in_production
```

### 2. Server Key Authentication

For server agents and basic endpoints using server-specific keys.

### 3. Bearer Token Authentication

For protected admin endpoints (marked with 🔒).

## Core Endpoints

### 🔓 Health Check

**Endpoint:** `GET /health`

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2026-02-18T09:12:12Z",
  "version": "1.0.0"
}
```

---

### 🔓 Server Registration

**Endpoint:** `POST /RegisterKey`

**Request:**

```json
{
  "hostname": "server-01",
  "operating_system": "Ubuntu 22.04",
  "agent_version": "1.0.0"
}
```

**Response:**

```json
{
  "server_id": "srv_123456789",
  "server_key": "sk_abcdef123456",
  "status": "registered"
}
```

---

### 🔓 Metrics - Unified Tiered Endpoint

**Endpoint:** `GET /api/servers/{server_id}/metrics/tiered`

**Query Parameters:**
- `start` (string, required): Start time (RFC3339 format)
- `end` (string, required): End time (RFC3339 format)

**Auto-Granularity Strategy:**
- **Last hour**: 1-minute intervals

- **Last 3 hours**: 5-minute intervals

- **Last 24 hours**: 10-minute intervals

- **Last 30 days**: 1-hour intervals

**Response:**

```json
{
  "server_id": "srv_d1dc36d8",
  "start_time": "2026-02-17T18:00:00Z",
  "end_time": "2026-02-17T19:00:00Z",
  "granularity": "1m",
  "data_points": [
    {
      "timestamp": "2026-02-17T18:00:00Z",
      "cpu_avg": 3.31,
      "cpu_max": 3.35,
      "cpu_min": 3.28,
      "memory_avg": 38.35,
      "memory_max": 38.85,
      "memory_min": 37.84,
      "disk_avg": 68,
      "disk_max": 68,
      "network_avg": 1.24,
      "network_max": 5.67,
      "temp_avg": 58.89,
      "temp_max": 72.87,
      "load_avg": 2.12,
      "load_max": 2.38,
      "sample_count": 60
    }
  ],
  "total_points": 61,
  "network_details": {
    "interfaces": [
      {
        "name": "enp111s0",
        "status": "up",
        "rx_bytes": 1674887389,
        "tx_bytes": 148772743,
        "rx_speed_mbps": 0.024,
        "tx_speed_mbps": 0.023
      }
    ],
    "total_rx_mbps": 0.095,
    "total_tx_mbps": 0.183
  },
  "disk_details": {
    "disks": [
      {
        "path": "/",
        "free_gb": 171,
        "used_gb": 354,
        "total_gb": 553,
        "filesystem": "/dev/nvme0n1p2",
        "used_percent": 68
      }
    ]
  },
  "temperature_details": {
    "cpu_temperature": 72.87,
    "gpu_temperature": 49,
    "system_temperature": 0,
    "storage_temperatures": {},
    "highest_temperature": 72.87,
    "temperature_unit": "celsius"
  }
}
```

**Usage Examples:**

```bash
# Realtime (1 hour)
curl "http://localhost:8080/api/servers/srv_d1dc36d8/metrics/tiered?start=2026-02-17T18:00:00Z&end=2026-02-17T19:00:00Z"

# Historical (24 hours)
curl "http://localhost:8080/api/servers/srv_d1dc36d8/metrics/tiered?start=2026-02-16T19:00:00Z&end=2026-02-17T19:00:00Z"

# Historical (7 days)
curl "http://localhost:8080/api/servers/srv_d1dc36d8/metrics/tiered?start=2026-02-10T19:00:00Z&end=2026-02-17T19:00:00Z"
```

**Note:** If the requested time period has no data, the API returns the latest available metrics with a `message` field.

---

### 🔓 Server Sources Management

**Add Source:** `POST /api/servers/{server_id}/sources`

```json
{
  "source": "TGBot"  // or "Web"
}
```

**Get Sources:** `GET /api/servers/{server_id}/sources`

**Remove Source:** `DELETE /api/servers/{server_id}/sources/{source}`

---

### � Static Server Information

**Endpoint:** `POST/PUT /api/servers/{server_id}/static-info`

Update static/persistent server information (hardware, system details).

**Request:**

```json
{
  "server_info": {
    "hostname": "gospodin-A620M-Pro-RS",
    "os": "Ubuntu",
    "os_version": "25.10",
    "kernel": "6.17.0-14-generic",
    "architecture": "x86_64"
  },
  "hardware_info": {
    "cpu_model": "AMD Ryzen 5 5600X",
    "cpu_cores": 6,
    "cpu_threads": 12,
    "cpu_frequency_mhz": 3700,
    "gpu_model": "NVIDIA GeForce RTX 3080",
    "gpu_driver": "550.120",
    "gpu_memory_gb": 10,
    "total_memory_gb": 32,
    "motherboard": "ASRock A620M Pro RS",
    "bios_version": "1.20"
  },
  "network_interfaces": [
    {
      "interface_name": "eth0",
      "mac_address": "00:11:22:33:44:55",
      "interface_type": "ethernet",
      "speed_mbps": 1000,
      "vendor": "Realtek",
      "driver": "r8169",
      "is_physical": true
    }
  ],
  "disk_info": [
    {
      "device_name": "/dev/nvme0n1",
      "model": "Samsung 980 PRO",
      "serial_number": "S5GXNX0T123456",
      "size_gb": 1000,
      "disk_type": "nvme",
      "interface_type": "nvme",
      "filesystem": "ext4",
      "mount_point": "/",
      "is_system_disk": true
    }
  ]
}
```

**Response:**

```json
{
  "message": "Static information updated successfully",
  "server_id": "srv_d1dc36d8"
}
```

**Get Static Info:** `GET /api/servers/{server_id}/static-info`

**Get Hardware Only:** `GET /api/servers/{server_id}/static-info/hardware`

**Get Network Interfaces:** `GET /api/servers/{server_id}/static-info/network`

**Get Disk Info:** `GET /api/servers/{server_id}/static-info/disks`

---

### �🔐 API Key Management

**Create Key:** `POST /api/admin/keys`

```json
{
  "service_id": "csharp-backend",
  "service_name": "C# Web Backend",
  "permissions": ["metrics:read", "servers:read"],
  "expires_days": 365
}
```

**List Keys:** `GET /api/admin/keys`

**Get Key Details:** `GET /api/admin/keys/{keyId}`

**Revoke Key:** `DELETE /api/admin/keys/{keyId}`

---

### 🔒 Server Management (Bearer Token Required)

**List All Servers:** `GET /api/servers`

**Get Server Status:** `GET /api/servers/{server_id}/status`

## Quick Start

1. **Register a server:**
   ```bash
   curl -X POST http://localhost:8080/RegisterKey \
     -H "Content-Type: application/json" \
     -d '{"hostname": "my-server", "operating_system": "Ubuntu 22.04"}'
   ```

2. **Get metrics:**
   ```bash
   curl "http://localhost:8080/api/servers/YOUR_SERVER_ID/metrics/tiered?start=$(date -d '1 hour ago' -Iseconds)&end=$(date -Iseconds)"
   ```

3. **Use API key for backend integration:**
   ```bash
   curl "http://localhost:8080/api/servers/YOUR_SERVER_ID/metrics/tiered?start=2026-02-17T18:00:00Z&end=2026-02-17T19:00:00Z" \
     -H "X-API-Key: sk_csharp_backend_development_key_change_in_production"
   ```

## Performance

- **Response time:** <40ms for complex queries
- **Auto-granularity:** Optimized based on time range
- **Data retention:** 90 days (configurable)
- **Max time range:** 30 days per request

## Error Codes

- `400` - Bad Request (missing/invalid parameters)
- `401` - Unauthorized (invalid/missing credentials)
- `404` - Not Found (server doesn't exist)
- `500` - Internal Server Error
