# Metrics Management Commands

This document describes the commands available for managing the multi-tier metrics system in ServerEyeAPI.

## Overview

The metrics management system provides commands to maintain, optimize, and analyze TimescaleDB metrics storage. These commands are executed through the standard command endpoint with specific command types.

## Command Types

### 1. Refresh Aggregates
Refreshes continuous aggregates to include the latest data.

**Type:** `refresh_aggregates`

**Payload:**
```json
{
  "granularity": "1m"  // Optional: "1m", "5m", "10m", "1h", or "all"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/servers/management/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "server_id": "management",
    "type": "refresh_aggregates",
    "payload": {
      "granularity": "5m"
    }
  }'
```

### 2. Rebuild Aggregates
Rebuilds continuous aggregates from scratch for a specific time range.

**Type:** `rebuild_aggregates`

**Payload:**
```json
{
  "granularity": "1h",
  "start_time": "2026-02-01T00:00:00Z",
  "end_time": "2026-02-13T00:00:00Z"
}
```

### 3. Cleanup Old Metrics
Removes old metrics data based on retention policies.

**Type:** `cleanup_old_metrics`

**Payload:**
```json
{
  "older_than": "90 days",
  "dry_run": true  // Set to false to actually delete
}
```

### 4. Compression Policy
Applies compression to old data chunks to save storage space.

**Type:** `compression_policy`

**Payload:**
```json
{
  "older_than": "7 days",
  "granularity": "1m"
}
```

### 5. Retention Policy
Applies retention policies to automatically delete old data.

**Type:** `retention_policy`

**Payload:**
```json
{}  // No additional parameters needed
```

### 6. Metrics Statistics
Retrieves current statistics about metrics storage.

**Type:** `metrics_stats`

**Payload:**
```json
{}  // No parameters needed
```

**Response:**
```json
{
  "command_id": "metrics_cmd_1234567890",
  "status": "completed",
  "message": "Retrieved metrics statistics",
  "result": {
    "success": true,
    "output": "Retrieved metrics statistics",
    "metrics": {
      "total_records": 524880,
      "unique_servers": 25,
      "table_size": "512 MB"
    },
    "data": {
      "granularity_stats": {
        "1m": {
          "total_records": 180000,
          "unique_servers": 25,
          "table_size": "245 MB"
        }
      }
    }
  }
}
```

### 7. Analyze Performance
Analyzes query performance and provides optimization recommendations.

**Type:** `analyze_performance`

**Payload:**
```json
{
  "time_range": "24h"  // Optional: "1h", "24h", "7d", "30d"
}
```

### 8. Export Metrics
Exports metrics data to a file for backup or analysis.

**Type:** `export_metrics`

**Payload:**
```json
{
  "format": "csv",        // "csv" or "json"
  "start_time": "2026-02-01T00:00:00Z",
  "end_time": "2026-02-13T00:00:00Z",
  "granularity": "1h"
}
```

### 9. Import Metrics
Imports metrics data from a file.

**Type:** `import_metrics`

**Payload:**
```json
{
  "file_path": "/tmp/metrics_export.csv",
  "format": "csv"
}
```

### 10. Validate Metrics
Validates metrics data integrity and reports issues.

**Type:** `validate_metrics`

**Payload:**
```json
{
  "time_range": "24h"
}
```

### 11. Optimize Storage
Optimizes TimescaleDB storage for better performance.

**Type:** `optimize_storage`

**Payload:**
```json
{
  "operations": [
    "reorder_chunks",
    "apply_compression",
    "vacuum_analyze",
    "update_stats"
  ]
}
```

## Usage Examples

### Daily Maintenance Script
```bash
#!/bin/bash

# Refresh all aggregates
curl -X POST http://localhost:8080/api/servers/management/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "server_id": "management",
    "type": "refresh_aggregates",
    "payload": {"granularity": "all"}
  }'

# Get statistics
curl -X POST http://localhost:8080/api/servers/management/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "server_id": "management",
    "type": "metrics_stats",
    "payload": {}
  }'
```

### Weekly Cleanup
```bash
# Check what would be deleted (dry run)
curl -X POST http://localhost:8080/api/servers/management/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "server_id": "management",
    "type": "cleanup_old_metrics",
    "payload": {
      "older_than": "90 days",
      "dry_run": true
    }
  }'

# Actually delete old data
curl -X POST http://localhost:8080/api/servers/management/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "server_id": "management",
    "type": "cleanup_old_metrics",
    "payload": {
      "older_than": "90 days",
      "dry_run": false
    }
  }'
```

### Performance Analysis
```bash
# Analyze last 24 hours
curl -X POST http://localhost:8080/api/servers/management/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "server_id": "management",
    "type": "analyze_performance",
    "payload": {"time_range": "24h"}
  }'
```

## Response Format

All metrics commands return a response in the following format:

```json
{
  "command_id": "metrics_cmd_<timestamp>",
  "status": "completed|failed|pending",
  "message": "Description of the result",
  "result": {
    "success": true,
    "output": "Human-readable output",
    "error": "Error message if failed",
    "data": {
      // Command-specific data
    },
    "metrics": {
      // Statistics if applicable
    },
    "time": "2026-02-13T16:30:00Z"
  }
}
```

## Best Practices

1. **Use dry_run** for cleanup operations before actually deleting data
2. **Schedule regular refreshes** during off-peak hours
3. **Monitor storage** with metrics_stats command regularly
4. **Export before cleanup** to maintain backup
5. **Analyze performance** weekly to identify optimization opportunities

## Automation

These commands can be automated using cron jobs:

```crontab
# Hourly aggregate refresh
0 * * * * curl -X POST http://localhost:8080/api/servers/management/command -d '{"type":"refresh_aggregates","payload":{"granularity":"1m"}}' -H "Authorization: Bearer $TOKEN"

# Daily statistics
0 8 * * * curl -X POST http://localhost:8080/api/servers/management/command -d '{"type":"metrics_stats","payload":{}}' -H "Authorization: Bearer $TOKEN"

# Weekly optimization
0 2 * * 0 curl -X POST http://localhost:8080/api/servers/management/command -d '{"type":"optimize_storage","payload":{}}' -H "Authorization: Bearer $TOKEN"
```

## Troubleshooting

### Command Failed
- Check if the TimescaleDB connection is active
- Verify the server has sufficient permissions
- Review the error message in the response

### Slow Performance
- Use `analyze_performance` command to identify bottlenecks
- Consider running `optimize_storage`
- Check if compression is properly configured

### Storage Issues
- Run `metrics_stats` to check current usage
- Use `cleanup_old_metrics` with dry_run first
- Verify retention policies are appropriate

## Security

- All metrics commands require authentication
- Use a dedicated management server_id
- Limit command execution to authorized users
- Audit command execution logs regularly
