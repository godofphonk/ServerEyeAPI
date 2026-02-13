# Tiered Metrics API Endpoints

This document describes the API endpoints for accessing multi-tier metrics with automatic granularity selection.

## Base URL
```
http://localhost:8080
```

## Authentication
All tiered metrics endpoints are **public** (no authentication required) for monitoring and dashboard access.

## Endpoints

### 1. Get Metrics with Auto-Granularity
Automatically selects the best granularity based on the time range.

```http
GET /api/servers/{server_id}/metrics/tiered
```

**Query Parameters:**
- `server_id` (path): Server identifier
- `start` (query): Start time in RFC3339 format
- `end` (query): End time in RFC3339 format

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/tiered?start=2026-02-13T15:00:00Z&end=2026-02-13T16:00:00Z"
```

**Response:**
```json
{
  "server_id": "test-server-001",
  "start_time": "2026-02-13T15:00:00Z",
  "end_time": "2026-02-13T16:00:00Z",
  "granularity": "1m",
  "data_points": [
    {
      "timestamp": "2026-02-13T15:01:00Z",
      "cpu_avg": 75.5,
      "cpu_max": 82.1,
      "cpu_min": 45.2,
      "memory_avg": 60.2,
      "memory_max": 65.8,
      "memory_min": 55.3,
      "disk_avg": 45.8,
      "disk_max": 45.8,
      "network_avg": 120.4,
      "network_max": 120.4,
      "sample_count": 1
    }
  ],
  "total_points": 60
}
```

### 2. Get Real-Time Metrics
Get the most recent metrics with 1-minute granularity.

```http
GET /api/servers/{server_id}/metrics/realtime
```

**Query Parameters:**
- `server_id` (path): Server identifier
- `duration` (query): Duration string (default: "1h", max: "1h")

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/realtime?duration=30m"
```

### 3. Get Historical Metrics
Get historical metrics with specified granularity.

```http
GET /api/servers/{server_id}/metrics/historical
```

**Query Parameters:**
- `server_id` (path): Server identifier
- `start` (query): Start time in RFC3339 format
- `end` (query): End time in RFC3339 format
- `granularity` (query): Optional granularity ("1m", "5m", "10m", "1h")

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/historical?start=2026-02-12T00:00:00Z&end=2026-02-13T00:00:00Z&granularity=1h"
```

### 4. Get Dashboard Metrics
Optimized endpoint for dashboard displays.

```http
GET /api/servers/{server_id}/metrics/dashboard
```

**Response includes:**
- Current system status
- 24-hour trends
- Heatmap data
- Performance indicators

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/dashboard"
```

### 5. Compare Metrics Between Periods
Compare metrics between two time periods.

```http
GET /api/servers/{server_id}/metrics/comparison
```

**Query Parameters:**
- `server_id` (path): Server identifier
- `period1_start` (query): First period start time
- `period1_end` (query): First period end time
- `period2_start` (query): Second period start time
- `period2_end` (query): Second period end time

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/comparison?period1_start=2026-02-12T00:00:00Z&period1_end=2026-02-12T12:00:00Z&period2_start=2026-02-12T12:00:00Z&period2_end=2026-02-13T00:00:00Z"
```

**Response:**
```json
{
  "server_id": "test-server-001",
  "period1": {
    "start": "2026-02-12T00:00:00Z",
    "end": "2026-02-12T12:00:00Z",
    "granularity": "10m"
  },
  "period2": {
    "start": "2026-02-12T12:00:00Z",
    "end": "2026-02-13T00:00:00Z",
    "granularity": "10m"
  },
  "averages1": {
    "cpu_avg": 65.5,
    "memory_avg": 55.2,
    "disk_avg": 45.8,
    "network_avg": 100.4
  },
  "averages2": {
    "cpu_avg": 70.2,
    "memory_avg": 58.9,
    "disk_avg": 45.8,
    "network_avg": 110.7
  },
  "changes": {
    "cpu_change": 7.2,
    "memory_change": 6.7,
    "disk_change": 0.0,
    "network_change": 10.2
  }
}
```

### 6. Get Metrics Heatmap
Get metrics data formatted for heatmap visualization.

```http
GET /api/servers/{server_id}/metrics/heatmap
```

**Query Parameters:**
- `server_id` (path): Server identifier
- `start` (query): Start time in RFC3339 format
- `end` (query): End time in RFC3339 format

**Example:**
```bash
curl "http://localhost:8080/api/servers/test-server-001/metrics/heatmap?start=2026-02-13T00:00:00Z&end=2026-02-13T23:59:59Z"
```

### 7. Get Metrics Summary
Get storage statistics across all granularity levels.

```http
GET /api/metrics/summary
```

**Example:**
```bash
curl "http://localhost:8080/api/metrics/summary"
```

**Response:**
```json
{
  "granularity_stats": {
    "1m": {
      "total_records": 180000,
      "unique_servers": 25,
      "earliest_record": "2026-02-13T15:00:00Z",
      "latest_record": "2026-02-13T16:00:00Z",
      "table_size": "245 MB"
    },
    "5m": {
      "total_records": 86400,
      "unique_servers": 25,
      "earliest_record": "2026-02-10T00:00:00Z",
      "latest_record": "2026-02-13T16:00:00Z",
      "table_size": "156 MB"
    }
  },
  "total_data_points": 524880,
  "total_servers": 25,
  "storage_size": "512 MB",
  "last_updated": "2026-02-13T16:05:00Z"
}
```

## Granularity Strategy

The system automatically selects granularity based on the time range:

| Time Range | Granularity | Data Points | Retention |
|------------|-------------|-------------|-----------|
| ≤ 1 hour | 1 minute | Up to 60 | 3 hours |
| ≤ 3 hours | 5 minutes | Up to 36 | 24 hours |
| ≤ 24 hours | 10 minutes | Up to 144 | 7 days |
| > 24 hours | 1 hour | Up to 720 | 90 days |

## Error Responses

All endpoints return standard error responses:

```json
{
  "error": "Error description"
}
```

Common error codes:
- `400 Bad Request`: Invalid parameters
- `404 Not Found`: Server not found
- `500 Internal Server Error`: Server error

## Time Format

All time parameters must be in **RFC3339** format:
- `2026-02-13T15:30:00Z` (UTC)
- `2026-02-13T15:30:00+03:00` (with timezone)

## Rate Limits

Endpoints are rate-limited to prevent abuse:
- 100 requests per minute per IP
- 1000 requests per hour per IP

## Examples

### Python
```python
import requests
import datetime

# Get last hour metrics
server_id = "test-server-001"
end_time = datetime.datetime.now(datetime.timezone.utc)
start_time = end_time - datetime.timedelta(hours=1)

url = f"http://localhost:8080/api/servers/{server_id}/metrics/tiered"
params = {
    "start": start_time.isoformat(),
    "end": end_time.isoformat()
}

response = requests.get(url, params=params)
data = response.json()
print(f"Got {data['total_points']} data points with {data['granularity']} granularity")
```

### JavaScript
```javascript
// Get real-time metrics
const serverId = 'test-server-001';
const duration = '30m';

fetch(`http://localhost:8080/api/servers/${serverId}/metrics/realtime?duration=${duration}`)
  .then(response => response.json())
  .then(data => {
    console.log('Latest metrics:', data.data_points[data.data_points.length - 1]);
  });
```

### Go
```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type TieredMetricsResponse struct {
    ServerID    string `json:"server_id"`
    Granularity string `json:"granularity"`
    TotalPoints int64  `json:"total_points"`
    DataPoints  []struct {
        Timestamp time.Time `json:"timestamp"`
        CPUAvg    float64   `json:"cpu_avg"`
        MemoryAvg float64   `json:"memory_avg"`
    } `json:"data_points"`
}

func getMetrics(serverID string) (*TieredMetricsResponse, error) {
    end := time.Now()
    start := end.Add(-time.Hour)
    
    url := fmt.Sprintf("http://localhost:8080/api/servers/%s/metrics/tiered?start=%s&end=%s",
        serverID, start.Format(time.RFC3339), end.Format(time.RFC3339))
    
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    var data TieredMetricsResponse
    err = json.Unmarshal(body, &data)
    return &data, err
}
```

## WebSocket Integration

For real-time updates, use the WebSocket endpoint:
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => {
    const metrics = JSON.parse(event.data);
    // Update dashboard with new metrics
};
```

## Best Practices

1. **Use appropriate time ranges**: Don't request more data than needed
2. **Cache responses**: Metrics don't change for past time periods
3. **Use dashboard endpoint**: For UI dashboards, use the optimized endpoint
4. **Handle errors gracefully**: Always check response status
5. **Use auto-granularity**: Let the system select the best granularity
