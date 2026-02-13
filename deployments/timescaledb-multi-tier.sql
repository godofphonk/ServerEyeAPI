-- Multi-tier continuous aggregates for optimized metrics storage
-- Strategy:
-- - Last hour: every 1 minute
-- - Last 3 hours: every 5 minutes  
-- - Last 24 hours: every 10 minutes
-- - Last 30 days: every 1 hour

-- Drop existing aggregates to recreate with new strategy
DROP MATERIALIZED VIEW IF EXISTS metrics_1m_avg CASCADE;
DROP MATERIALIZED VIEW IF EXISTS metrics_5m_avg CASCADE;
DROP MATERIALIZED VIEW IF EXISTS metrics_10m_avg CASCADE;
DROP MATERIALIZED VIEW IF EXISTS metrics_1h_avg CASCADE;

-- Level 1: Ultra-detailed metrics (last hour) - every 1 minute
CREATE MATERIALIZED VIEW metrics_1m_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 minute', time) AS bucket,
    server_id,
    hostname,
    -- Core metrics with full detail
    AVG(cpu_usage) as avg_cpu,
    MAX(cpu_usage) as max_cpu,
    MIN(cpu_usage) as min_cpu,
    STDDEV(cpu_usage) as stddev_cpu,
    
    AVG(memory_usage) as avg_memory,
    MAX(memory_usage) as max_memory,
    MIN(memory_usage) as min_memory,
    STDDEV(memory_usage) as stddev_memory,
    
    AVG(disk_usage) as avg_disk,
    MAX(disk_usage) as max_disk,
    
    AVG(network_usage) as avg_network,
    MAX(network_usage) as max_network,
    
    -- Temperature monitoring
    AVG(cpu_temperature) as avg_cpu_temp,
    MAX(cpu_temperature) as max_cpu_temp,
    AVG(highest_temperature) as avg_highest_temp,
    MAX(highest_temperature) as max_highest_temp,
    
    -- Load averages
    AVG(load_avg_1m) as avg_load_1m,
    MAX(load_avg_1m) as max_load_1m,
    
    -- Process counts
    AVG(processes_total) as avg_processes,
    MAX(processes_total) as max_processes,
    
    COUNT(*) as sample_count,
    MIN(time) as first_seen,
    MAX(time) as last_seen
FROM server_metrics 
GROUP BY bucket, server_id, hostname;

-- Level 2: Detailed metrics (1-3 hours) - every 5 minutes
CREATE MATERIALIZED VIEW metrics_5m_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('5 minutes', time) AS bucket,
    server_id,
    hostname,
    -- Core metrics
    AVG(cpu_usage) as avg_cpu,
    MAX(cpu_usage) as max_cpu,
    MIN(cpu_usage) as min_cpu,
    
    AVG(memory_usage) as avg_memory,
    MAX(memory_usage) as max_memory,
    MIN(memory_usage) as min_memory,
    
    AVG(disk_usage) as avg_disk,
    MAX(disk_usage) as max_disk,
    
    AVG(network_usage) as avg_network,
    MAX(network_usage) as max_network,
    
    -- Temperature
    AVG(cpu_temperature) as avg_cpu_temp,
    MAX(cpu_temperature) as max_cpu_temp,
    
    -- Load averages
    AVG(load_avg_1m) as avg_load_1m,
    MAX(load_avg_1m) as max_load_1m,
    AVG(load_avg_5m) as avg_load_5m,
    
    -- Memory details
    AVG(memory_total_gb) as avg_memory_total,
    AVG(memory_used_gb) as avg_memory_used,
    
    COUNT(*) as sample_count,
    MIN(time) as first_seen,
    MAX(time) as last_seen
FROM server_metrics 
GROUP BY bucket, server_id, hostname;

-- Level 3: Hourly metrics (3-24 hours) - every 10 minutes  
CREATE MATERIALIZED VIEW metrics_10m_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('10 minutes', time) AS bucket,
    server_id,
    hostname,
    os_info,
    -- Core metrics
    AVG(cpu_usage) as avg_cpu,
    MAX(cpu_usage) as max_cpu,
    MIN(cpu_usage) as min_cpu,
    
    AVG(memory_usage) as avg_memory,
    MAX(memory_usage) as max_memory,
    MIN(memory_usage) as min_memory,
    
    AVG(disk_usage) as avg_disk,
    MAX(disk_usage) as max_disk,
    
    AVG(network_usage) as avg_network,
    MAX(network_usage) as max_network,
    
    -- Temperature trends
    AVG(cpu_temperature) as avg_cpu_temp,
    MAX(cpu_temperature) as max_cpu_temp,
    AVG(highest_temperature) as avg_highest_temp,
    MAX(highest_temperature) as max_highest_temp,
    
    -- Load trends
    AVG(load_avg_1m) as avg_load_1m,
    MAX(load_avg_1m) as max_load_1m,
    AVG(load_avg_5m) as avg_load_5m,
    MAX(load_avg_5m) as max_load_5m,
    AVG(load_avg_15m) as avg_load_15m,
    
    -- System info
    AVG(uptime_seconds) as avg_uptime,
    AVG(processes_total) as avg_processes,
    AVG(processes_running) as avg_running,
    
    -- CPU details
    AVG(cpu_cores) as avg_cpu_cores,
    AVG(cpu_frequency) as avg_cpu_freq,
    
    COUNT(*) as sample_count,
    MIN(time) as first_seen,
    MAX(time) as last_seen
FROM server_metrics 
GROUP BY bucket, server_id, hostname, os_info;

-- Level 4: Daily metrics (24 hours - 30 days) - every 1 hour
CREATE MATERIALIZED VIEW metrics_1h_avg WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', time) AS bucket,
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
    
    -- Memory utilization
    AVG(memory_used_gb) as avg_memory_used,
    MAX(memory_used_gb) as max_memory_used,
    AVG(memory_total_gb) as avg_memory_total,
    
    -- System metrics
    AVG(processes_total) as avg_processes,
    MAX(processes_total) as max_processes,
    AVG(processes_running) as avg_running,
    MAX(processes_running) as max_running,
    
    -- Uptime and availability
    MAX(uptime_seconds) as max_uptime,
    MIN(uptime_seconds) as min_uptime,
    
    COUNT(*) as sample_count,
    MIN(time) as first_seen,
    MAX(time) as last_seen
FROM server_metrics 
GROUP BY bucket, server_id, hostname, os_info;

-- Create optimized indexes for each view
CREATE INDEX ON metrics_1m_avg (server_id, bucket DESC);
CREATE INDEX ON metrics_5m_avg (server_id, bucket DESC);
CREATE INDEX ON metrics_10m_avg (server_id, bucket DESC);
CREATE INDEX ON metrics_1h_avg (server_id, bucket DESC);

-- Partial indexes for alert conditions
CREATE INDEX ON metrics_1m_avg (server_id, bucket DESC) WHERE max_cpu > 80 OR max_memory > 85;
CREATE INDEX ON metrics_5m_avg (server_id, bucket DESC) WHERE max_cpu > 80 OR max_memory > 85;
CREATE INDEX ON metrics_10m_avg (server_id, bucket DESC) WHERE max_cpu > 80 OR max_memory > 85;

-- Refresh policies for each level
-- Level 1: Refresh every minute for real-time data
SELECT add_continuous_aggregate_policy('metrics_1m_avg',
    start_offset => INTERVAL '5 minutes',
    end_offset => INTERVAL '30 seconds',
    schedule_interval => INTERVAL '1 minute'
);

-- Level 2: Refresh every 2 minutes
SELECT add_continuous_aggregate_policy('metrics_5m_avg',
    start_offset => INTERVAL '15 minutes',
    end_offset => INTERVAL '2 minutes',
    schedule_interval => INTERVAL '2 minutes'
);

-- Level 3: Refresh every 5 minutes
SELECT add_continuous_aggregate_policy('metrics_10m_avg',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes'
);

-- Level 4: Refresh every 10 minutes
SELECT add_continuous_aggregate_policy('metrics_1h_avg',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '10 minutes',
    schedule_interval => INTERVAL '10 minutes'
);

-- Add compression policies for aggregates
SELECT add_compression_policy('metrics_1m_avg', INTERVAL '2 hours');
SELECT add_compression_policy('metrics_5m_avg', INTERVAL '6 hours');
SELECT add_compression_policy('metrics_10m_avg', INTERVAL '1 day');
SELECT add_compression_policy('metrics_1h_avg', INTERVAL '2 days');

-- Retention policies
-- Keep 1-minute data for 3 hours
SELECT add_retention_policy('metrics_1m_avg', INTERVAL '3 hours');

-- Keep 5-minute data for 24 hours  
SELECT add_retention_policy('metrics_5m_avg', INTERVAL '24 hours');

-- Keep 10-minute data for 7 days
SELECT add_retention_policy('metrics_10m_avg', INTERVAL '7 days');

-- Keep hourly data for 90 days
SELECT add_retention_policy('metrics_1h_avg', INTERVAL '90 days');

-- Create helper function to get metrics with appropriate granularity
CREATE OR REPLACE FUNCTION get_metrics_by_granularity(
    p_server_id TEXT,
    p_start_time TIMESTAMPTZ,
    p_end_time TIMESTAMPTZ
)
RETURNS TABLE (
    bucket TIMESTAMPTZ,
    avg_cpu DOUBLE PRECISION,
    max_cpu DOUBLE PRECISION,
    min_cpu DOUBLE PRECISION,
    avg_memory DOUBLE PRECISION,
    max_memory DOUBLE PRECISION,
    min_memory DOUBLE PRECISION,
    avg_disk DOUBLE PRECISION,
    max_disk DOUBLE PRECISION,
    avg_network DOUBLE PRECISION,
    max_network DOUBLE PRECISION,
    sample_count BIGINT,
    granularity TEXT
) AS $$
BEGIN
    -- Use 1-minute data for last hour
    IF p_end_time - p_start_time <= INTERVAL '1 hour' THEN
        RETURN QUERY
        SELECT 
            bucket, avg_cpu, max_cpu, min_cpu,
            avg_memory, max_memory, min_memory,
            avg_disk, max_disk, avg_network, max_network,
            sample_count, '1m'::TEXT
        FROM metrics_1m_avg
        WHERE server_id = p_server_id 
        AND bucket BETWEEN p_start_time AND p_end_time
        ORDER BY bucket;
    
    -- Use 5-minute data for 1-3 hours
    ELSIF p_end_time - p_start_time <= INTERVAL '3 hours' THEN
        RETURN QUERY
        SELECT 
            bucket, avg_cpu, max_cpu, min_cpu,
            avg_memory, max_memory, min_memory,
            avg_disk, max_disk, avg_network, max_network,
            sample_count, '5m'::TEXT
        FROM metrics_5m_avg
        WHERE server_id = p_server_id 
        AND bucket BETWEEN p_start_time AND p_end_time
        ORDER BY bucket;
    
    -- Use 10-minute data for 3-24 hours
    ELSIF p_end_time - p_start_time <= INTERVAL '24 hours' THEN
        RETURN QUERY
        SELECT 
            bucket, avg_cpu, max_cpu, min_cpu,
            avg_memory, max_memory, min_memory,
            avg_disk, max_disk, avg_network, max_network,
            sample_count, '10m'::TEXT
        FROM metrics_10m_avg
        WHERE server_id = p_server_id 
        AND bucket BETWEEN p_start_time AND p_end_time
        ORDER BY bucket;
    
    -- Use hourly data for >24 hours
    ELSE
        RETURN QUERY
        SELECT 
            bucket, avg_cpu, max_cpu, min_cpu,
            avg_memory, max_memory, min_memory,
            avg_disk, max_disk, avg_network, max_network,
            sample_count, '1h'::TEXT
        FROM metrics_1h_avg
        WHERE server_id = p_server_id 
        AND bucket BETWEEN p_start_time AND p_end_time
        ORDER BY bucket;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create view for current system status using 1-minute data
CREATE OR REPLACE VIEW current_system_status AS
SELECT 
    server_id,
    hostname,
    os_info,
    bucket as last_update,
    avg_cpu,
    max_cpu,
    avg_memory,
    max_memory,
    avg_disk,
    avg_cpu_temp,
    max_cpu_temp,
    avg_load_1m,
    max_load_1m,
    sample_count
FROM metrics_1m_avg
WHERE bucket >= NOW() - INTERVAL '5 minutes'
ORDER BY bucket DESC;

COMMIT;
