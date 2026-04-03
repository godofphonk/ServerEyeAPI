-- Optimized granularity views for better visualization performance
-- Based on enterprise-level requirements: 30m, 2h, 6h granularities

-- Drop existing optimized views if they exist
DROP MATERIALIZED VIEW IF EXISTS metrics_30m_avg CASCADE;
DROP MATERIALIZED VIEW IF EXISTS metrics_2h_avg CASCADE;
DROP MATERIALIZED VIEW IF EXISTS metrics_6h_avg CASCADE;

-- Level 3.5: 30-minute metrics (6-24 hours) - optimized for dashboard visualization
CREATE MATERIALIZED VIEW metrics_30m_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('30 minutes', time) AS bucket,
    server_id,
    hostname,
    os_info,
    -- Core metrics
    AVG(cpu_usage) as avg_cpu,
    MAX(cpu_usage) as max_cpu,
    MIN(cpu_usage) as min_cpu,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY cpu_usage) as p95_cpu,
    
    AVG(memory_usage) as avg_memory,
    MAX(memory_usage) as max_memory,
    MIN(memory_usage) as min_memory,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY memory_usage) as p95_memory,
    
    AVG(disk_usage) as avg_disk,
    MAX(disk_usage) as max_disk,
    MIN(disk_usage) as min_disk,
    
    -- Network totals
    AVG(network_usage) as avg_network,
    MAX(network_usage) as max_network,
    SUM(network_usage) as total_network,
    
    -- Temperature statistics
    AVG(cpu_temperature) as avg_cpu_temp,
    MAX(cpu_temperature) as max_cpu_temp,
    AVG(highest_temperature) as avg_highest_temp,
    MAX(highest_temperature) as max_highest_temp,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY highest_temperature) as p95_temp,
    
    -- Load averages
    AVG(load_avg_1m) as avg_load_1m,
    MAX(load_avg_1m) as max_load_1m,
    AVG(load_avg_5m) as avg_load_5m,
    MAX(load_avg_5m) as max_load_5m,
    AVG(load_avg_15m) as avg_load_15m,
    MAX(load_avg_15m) as max_load_15m,
    
    -- Memory utilization (calculated from memory fields)
    AVG(CASE WHEN memory_total_gb > 0 THEN (memory_used_gb / memory_total_gb) * 100 ELSE 0 END) as avg_memory_util,
    MAX(CASE WHEN memory_total_gb > 0 THEN (memory_used_gb / memory_total_gb) * 100 ELSE 0 END) as max_memory_util,
    MIN(CASE WHEN memory_total_gb > 0 THEN (memory_used_gb / memory_total_gb) * 100 ELSE 0 END) as min_memory_util,
    
    -- Process count (using processes_total)
    AVG(processes_total) as avg_process_count,
    MAX(processes_total) as max_process_count,
    
    -- Uptime (using uptime_seconds)
    MAX(uptime_seconds) as max_uptime,
    
    -- Sample statistics
    COUNT(*) as sample_count,
    MIN(time) as first_seen,
    MAX(time) as last_seen
FROM server_metrics 
GROUP BY bucket, server_id, hostname, os_info;

-- Level 4.5: 2-hour metrics (1-7 days) - optimized for weekly trends
CREATE MATERIALIZED VIEW metrics_2h_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('2 hours', time) AS bucket,
    server_id,
    hostname,
    os_info,
    -- Core metrics
    AVG(cpu_usage) as avg_cpu,
    MAX(cpu_usage) as max_cpu,
    MIN(cpu_usage) as min_cpu,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY cpu_usage) as p95_cpu,
    
    AVG(memory_usage) as avg_memory,
    MAX(memory_usage) as max_memory,
    MIN(memory_usage) as min_memory,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY memory_usage) as p95_memory,
    
    AVG(disk_usage) as avg_disk,
    MAX(disk_usage) as max_disk,
    MIN(disk_usage) as min_disk,
    
    -- Network totals
    AVG(network_usage) as avg_network,
    MAX(network_usage) as max_network,
    SUM(network_usage) as total_network,
    
    -- Temperature statistics
    AVG(cpu_temperature) as avg_cpu_temp,
    MAX(cpu_temperature) as max_cpu_temp,
    AVG(highest_temperature) as avg_highest_temp,
    MAX(highest_temperature) as max_highest_temp,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY highest_temperature) as p95_temp,
    
    -- Load averages
    AVG(load_avg_1m) as avg_load_1m,
    MAX(load_avg_1m) as max_load_1m,
    AVG(load_avg_5m) as avg_load_5m,
    MAX(load_avg_5m) as max_load_5m,
    AVG(load_avg_15m) as avg_load_15m,
    MAX(load_avg_15m) as max_load_15m,
    
    -- Memory utilization (calculated from memory fields)
    AVG(CASE WHEN memory_total_gb > 0 THEN (memory_used_gb / memory_total_gb) * 100 ELSE 0 END) as avg_memory_util,
    MAX(CASE WHEN memory_total_gb > 0 THEN (memory_used_gb / memory_total_gb) * 100 ELSE 0 END) as max_memory_util,
    MIN(CASE WHEN memory_total_gb > 0 THEN (memory_used_gb / memory_total_gb) * 100 ELSE 0 END) as min_memory_util,
    
    -- Process count (using processes_total)
    AVG(processes_total) as avg_process_count,
    MAX(processes_total) as max_process_count,
    
    -- Uptime (using uptime_seconds)
    MAX(uptime_seconds) as max_uptime,
    
    -- Sample statistics
    COUNT(*) as sample_count,
    MIN(time) as first_seen,
    MAX(time) as last_seen
FROM server_metrics 
GROUP BY bucket, server_id, hostname, os_info;

-- Level 5: 6-hour metrics (7-30+ days) - optimized for monthly analysis
CREATE MATERIALIZED VIEW metrics_6h_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('6 hours', time) AS bucket,
    server_id,
    hostname,
    os_info,
    -- Core metrics
    AVG(cpu_usage) as avg_cpu,
    MAX(cpu_usage) as max_cpu,
    MIN(cpu_usage) as min_cpu,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY cpu_usage) as p95_cpu,
    
    AVG(memory_usage) as avg_memory,
    MAX(memory_usage) as max_memory,
    MIN(memory_usage) as min_memory,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY memory_usage) as p95_memory,
    
    AVG(disk_usage) as avg_disk,
    MAX(disk_usage) as max_disk,
    MIN(disk_usage) as min_disk,
    
    -- Network totals
    AVG(network_usage) as avg_network,
    MAX(network_usage) as max_network,
    SUM(network_usage) as total_network,
    
    -- Temperature statistics
    AVG(cpu_temperature) as avg_cpu_temp,
    MAX(cpu_temperature) as max_cpu_temp,
    AVG(highest_temperature) as avg_highest_temp,
    MAX(highest_temperature) as max_highest_temp,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY highest_temperature) as p95_temp,
    
    -- Load averages
    AVG(load_avg_1m) as avg_load_1m,
    MAX(load_avg_1m) as max_load_1m,
    AVG(load_avg_5m) as avg_load_5m,
    MAX(load_avg_5m) as max_load_5m,
    AVG(load_avg_15m) as avg_load_15m,
    MAX(load_avg_15m) as max_load_15m,
    
    -- Memory utilization (calculated from memory fields)
    AVG(CASE WHEN memory_total_gb > 0 THEN (memory_used_gb / memory_total_gb) * 100 ELSE 0 END) as avg_memory_util,
    MAX(CASE WHEN memory_total_gb > 0 THEN (memory_used_gb / memory_total_gb) * 100 ELSE 0 END) as max_memory_util,
    MIN(CASE WHEN memory_total_gb > 0 THEN (memory_used_gb / memory_total_gb) * 100 ELSE 0 END) as min_memory_util,
    
    -- Process count (using processes_total)
    AVG(processes_total) as avg_process_count,
    MAX(processes_total) as max_process_count,
    
    -- Uptime (using uptime_seconds)
    MAX(uptime_seconds) as max_uptime,
    
    -- Sample statistics
    COUNT(*) as sample_count,
    MIN(time) as first_seen,
    MAX(time) as last_seen
FROM server_metrics 
GROUP BY bucket, server_id, hostname, os_info;

-- Create optimized indexes for new views
CREATE INDEX ON metrics_30m_avg (server_id, bucket DESC);
CREATE INDEX ON metrics_2h_avg (server_id, bucket DESC);
CREATE INDEX ON metrics_6h_avg (server_id, bucket DESC);

-- Partial indexes for alert conditions on new views
CREATE INDEX ON metrics_30m_avg (server_id, bucket DESC) WHERE max_cpu > 80 OR max_memory > 85;
CREATE INDEX ON metrics_2h_avg (server_id, bucket DESC) WHERE max_cpu > 80 OR max_memory > 85;
CREATE INDEX ON metrics_6h_avg (server_id, bucket DESC) WHERE max_cpu > 80 OR max_memory > 85;

-- Refresh policies for new optimized views
-- 30-minute view: Refresh every 5 minutes
SELECT add_continuous_aggregate_policy('metrics_30m_avg',
    start_offset => INTERVAL '2 hours',
    end_offset => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes'
);

-- 2-hour view: Refresh every 15 minutes
SELECT add_continuous_aggregate_policy('metrics_2h_avg',
    start_offset => INTERVAL '6 hours',
    end_offset => INTERVAL '15 minutes',
    schedule_interval => INTERVAL '15 minutes'
);

-- 6-hour view: Refresh every 30 minutes
SELECT add_continuous_aggregate_policy('metrics_6h_avg',
    start_offset => INTERVAL '12 hours',
    end_offset => INTERVAL '30 minutes',
    schedule_interval => INTERVAL '30 minutes'
);

-- Add compression policies for new views
SELECT add_compression_policy('metrics_30m_avg', INTERVAL '6 hours');
SELECT add_compression_policy('metrics_2h_avg', INTERVAL '2 days');
SELECT add_compression_policy('metrics_6h_avg', INTERVAL '5 days');

-- Retention policies for optimized views
-- Keep 30-minute data for 2 days
SELECT add_retention_policy('metrics_30m_avg', INTERVAL '2 days');

-- Keep 2-hour data for 14 days
SELECT add_retention_policy('metrics_2h_avg', INTERVAL '14 days');

-- Keep 6-hour data for 120 days (extended retention)
SELECT add_retention_policy('metrics_6h_avg', INTERVAL '120 days');

-- Grant permissions
GRANT SELECT ON metrics_30m_avg TO server_eye_read;
GRANT SELECT ON metrics_2h_avg TO server_eye_read;
GRANT SELECT ON metrics_6h_avg TO server_eye_read;

-- Add comments for documentation
COMMENT ON MATERIALIZED VIEW metrics_30m_avg IS '30-minute aggregated metrics optimized for 6-24 hour dashboard visualization (48 points max)';
COMMENT ON MATERIALIZED VIEW metrics_2h_avg IS '2-hour aggregated metrics optimized for 1-7 day trend analysis (84 points max)';
COMMENT ON MATERIALIZED VIEW metrics_6h_avg IS '6-hour aggregated metrics optimized for 7-30+ day analysis (120 points max for 30 days)';
