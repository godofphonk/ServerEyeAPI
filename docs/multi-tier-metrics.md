# Multi-Tier Metrics Architecture

## Overview

The ServerEyeAPI now implements a sophisticated multi-tier metrics storage system that automatically optimizes data granularity based on the time range being queried. This approach provides both real-time precision for recent data and efficient storage for historical data.

## Granularity Strategy

### Tier 1: Ultra-Detailed (Last Hour)
- **Granularity**: Every 1 minute
- **Retention**: 3 hours
- **Use Case**: Real-time monitoring, immediate alerts
- **Features**: Full metrics with min/max/avg/stddev

### Tier 2: Detailed (1-3 Hours)
- **Granularity**: Every 5 minutes
- **Retention**: 24 hours
- **Use Case**: Recent performance analysis
- **Features**: Core metrics with temperature and load data

### Tier 3: Hourly (3-24 Hours)
- **Granularity**: Every 10 minutes
- **Retention**: 7 days
- **Use Case**: Daily performance review
- **Features**: Extended metrics with system information

### Tier 4: Daily (24 Hours - 30 Days)
- **Granularity**: Every 1 hour
- **Retention**: 90 days
- **Features**: Percentiles (95th), totals, comprehensive stats

## API Endpoints

### 1. Get Metrics with Auto-Granularity
```http
GET /api/servers/{server_id}/metrics/tiered?start=2026-02-13T15:00:00Z&end=2026-02-13T16:00:00Z
```
Automatically selects the best granularity based on the time range.

### 2. Real-Time Metrics
```http
GET /api/servers/{server_id}/metrics/realtime?duration=30m
```
Returns the most recent metrics with 1-minute granularity.

### 3. Historical Metrics
```http
GET /api/servers/{server_id}/metrics/historical?start=...&end=...&granularity=5m
```
Explicitly specifies the granularity for historical analysis.

### 4. Dashboard Metrics
```http
GET /api/servers/{server_id}/metrics/dashboard
```
Optimized endpoint for dashboard displays with current status and 24h trends.

### 5. Metrics Comparison
```http
GET /api/servers/{server_id}/metrics/comparison?period1_start=...&period1_end=...&period2_start=...&period2_end=...
```
Compares metrics between two time periods.

### 6. Metrics Heatmap
```http
GET /api/servers/{server_id}/metrics/heatmap?start=...&end=...
```
Returns data formatted for heatmap visualization.

### 7. Metrics Summary
```http
GET /api/metrics/summary
```
Returns storage statistics across all granularity levels.

## Performance Benefits

### Storage Optimization
- **Raw data**: 1-minute intervals for 1 hour = 60 points
- **Tier 2**: 5-minute intervals for 3 hours = 36 points (40% reduction)
- **Tier 3**: 10-minute intervals for 24 hours = 144 points (75% reduction)
- **Tier 4**: 1-hour intervals for 30 days = 720 points (96% reduction vs raw)

### Query Performance
- Automatic index selection based on granularity
- Pre-aggregated data eliminates runtime calculations
- Continuous aggregates refresh in background

### Memory Efficiency
- Compression policies automatically compress old data
- Retention policies prevent unlimited storage growth
- Chunked storage enables efficient time-range queries

## Database Schema

### Continuous Aggregates

#### metrics_1m_avg
```sql
CREATE MATERIALIZED VIEW metrics_1m_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 minute', time) AS bucket,
    server_id,
    hostname,
    AVG(cpu_usage) as avg_cpu,
    MAX(cpu_usage) as max_cpu,
    MIN(cpu_usage) as min_cpu,
    STDDEV(cpu_usage) as stddev_cpu,
    -- ... other metrics
FROM server_metrics 
GROUP BY bucket, server_id, hostname;
```

#### metrics_5m_avg
```sql
CREATE MATERIALIZED VIEW metrics_5m_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('5 minutes', time) AS bucket,
    server_id,
    hostname,
    -- Core metrics with avg/max/min
    -- ...
FROM server_metrics 
GROUP BY bucket, server_id, hostname;
```

#### metrics_10m_avg
```sql
CREATE MATERIALIZED VIEW metrics_10m_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('10 minutes', time) AS bucket,
    server_id,
    hostname,
    os_info,
    -- Extended metrics including system info
    -- ...
FROM server_metrics 
GROUP BY bucket, server_id, hostname, os_info;
```

#### metrics_1h_avg
```sql
CREATE MATERIALIZED VIEW metrics_1h_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', time) AS bucket,
    server_id,
    hostname,
    os_info,
    -- Comprehensive metrics with percentiles
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY cpu_usage) as p95_cpu,
    -- ...
FROM server_metrics 
GROUP BY bucket, server_id, hostname, os_info;
```

## Refresh Policies

### Automatic Refresh Intervals
- **1-minute view**: Every 1 minute (30-second lag)
- **5-minute view**: Every 2 minutes (2-minute lag)
- **10-minute view**: Every 5 minutes (5-minute lag)
- **1-hour view**: Every 10 minutes (10-minute lag)

### Background Processing
Continuous aggregates are refreshed in the background without impacting write performance.

## Usage Examples

### Go Client Usage
```go
// Get metrics with auto-granularity
response, err := tieredService.GetMetricsWithAutoGranularity(
    ctx, 
    "server-001", 
    time.Now().Add(-2*time.Hour), 
    time.Now(),
)

// Get real-time metrics
realtime, err := tieredService.GetRealTimeMetrics(
    ctx, 
    "server-001", 
    30*time.Minute,
)

// Get dashboard metrics
dashboard, err := tieredService.GetDashboardMetrics(ctx, "server-001")
```

### Direct SQL Usage
```sql
-- Automatic granularity selection
SELECT * FROM get_metrics_by_granularity(
    'server-001', 
    NOW() - INTERVAL '2 hours', 
    NOW()
);

-- Query specific granularity
SELECT bucket, avg_cpu, max_cpu 
FROM metrics_5m_avg 
WHERE server_id = 'server-001' 
AND bucket BETWEEN NOW() - INTERVAL '2 hours' AND NOW();
```

## Monitoring and Maintenance

### Storage Statistics
```sql
SELECT 
    view_name,
    COUNT(*) as records,
    COUNT(DISTINCT server_id) as servers,
    pg_size_pretty(pg_total_relation_size(view_name)) as size
FROM (
    SELECT 'metrics_1m_avg' as view_name, COUNT(*) FROM metrics_1m_avg
    UNION ALL SELECT 'metrics_5m_avg', COUNT(*) FROM metrics_5m_avg
    UNION ALL SELECT 'metrics_10m_avg', COUNT(*) FROM metrics_10m_avg
    UNION ALL SELECT 'metrics_1h_avg', COUNT(*) FROM metrics_1h_avg
) stats;
```

### Compression Status
```sql
SELECT 
    hypertable_name,
    compression_status,
    compressed_chunks,
    uncompressed_chunks
FROM timescaledb_information.compressed_hypertable_stats;
```

## Best Practices

1. **For Real-Time Alerts**: Use 1-minute data with thresholds on max values
2. **For Performance Analysis**: Use 5-minute data for recent trends
3. **For Capacity Planning**: Use hourly data with percentiles
4. **For Long-Term Trends**: Use daily aggregated data

## Future Enhancements

1. **Adaptive Granularity**: Automatically adjust based on data volume
2. **Machine Learning Integration**: Anomaly detection on tiered data
3. **Custom Aggregations**: User-defined aggregation windows
4. **Cross-Server Analytics**: Multi-server tiered aggregations

## Migration Guide

To migrate from single-tier to multi-tier:

1. Deploy the new SQL schema
2. Continuous aggregates will backfill automatically
3. Update application code to use new endpoints
4. Monitor storage usage and adjust retention policies
5. Gradually increase retention periods as needed

The migration is non-disruptive - existing data continues to be stored in `server_metrics` while aggregates are built in the background.
