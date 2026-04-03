-- Copyright (c) 2026 godofphonk
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

-- TimescaleDB initialization script for ServerEye API
-- This script creates hypertables for time-series data storage

-- Create TimescaleDB extension with version check
DO $$
BEGIN
    -- Check if TimescaleDB extension is already installed
    IF NOT EXISTS (
        SELECT 1 FROM pg_extension WHERE extname = 'timescaledb'
    ) THEN
        -- Create the extension
        CREATE EXTENSION timescaledb;
        RAISE NOTICE 'TimescaleDB extension created successfully';
    ELSE
        RAISE NOTICE 'TimescaleDB extension already exists';
    END IF;
    
    -- Verify TimescaleDB is properly loaded
    PERFORM * FROM pg_proc WHERE proname = 'create_hypertable';
    IF NOT FOUND THEN
        RAISE EXCEPTION 'TimescaleDB extension not properly loaded';
    END IF;
END $$;

-- Create indexes for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Server Metrics Hypertable
-- Stores detailed server performance metrics over time
CREATE TABLE IF NOT EXISTS server_metrics (
    time TIMESTAMPTZ NOT NULL,
    server_id TEXT NOT NULL,
    
    -- Basic metrics
    cpu_usage DOUBLE PRECISION,
    memory_usage DOUBLE PRECISION,
    disk_usage DOUBLE PRECISION,
    network_usage DOUBLE PRECISION,
    
    -- Detailed CPU metrics
    cpu_usage_total DOUBLE PRECISION,
    cpu_usage_user DOUBLE PRECISION,
    cpu_usage_system DOUBLE PRECISION,
    cpu_usage_idle DOUBLE PRECISION,
    cpu_cores INTEGER,
    cpu_frequency DOUBLE PRECISION,
    
    -- Load averages
    load_avg_1m DOUBLE PRECISION,
    load_avg_5m DOUBLE PRECISION,
    load_avg_15m DOUBLE PRECISION,
    
    -- Detailed memory metrics
    memory_total_gb DOUBLE PRECISION,
    memory_used_gb DOUBLE PRECISION,
    memory_available_gb DOUBLE PRECISION,
    memory_free_gb DOUBLE PRECISION,
    memory_buffers_gb DOUBLE PRECISION,
    memory_cached_gb DOUBLE PRECISION,
    
    -- Detailed disk metrics (JSON array for multiple disks)
    disk_details JSONB,
    
    -- Detailed network metrics (JSON array for multiple interfaces)
    network_details JSONB,
    
    -- Temperature metrics
    cpu_temperature DOUBLE PRECISION,
    gpu_temperature DOUBLE PRECISION,
    system_temperature DOUBLE PRECISION,
    highest_temperature DOUBLE PRECISION,
    temperature_unit TEXT DEFAULT 'celsius',
    
    -- System information
    hostname TEXT,
    os_info TEXT,
    kernel TEXT,
    architecture TEXT,
    uptime_seconds BIGINT,
    uptime_human TEXT,
    boot_time TIMESTAMPTZ,
    
    -- Process information
    processes_total INTEGER,
    processes_running INTEGER,
    processes_sleeping INTEGER,
    
    -- Metadata
    agent_version TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Server Status Hypertable
-- Stores server online/offline status and basic information
CREATE TABLE IF NOT EXISTS server_status (
    time TIMESTAMPTZ NOT NULL,
    server_id TEXT NOT NULL,
    online BOOLEAN NOT NULL,
    last_seen TIMESTAMPTZ,
    version TEXT,
    os_info TEXT,
    agent_version TEXT,
    hostname TEXT,
    ip_address INET,
    response_time_ms INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Server Commands Hypertable
-- Stores commands sent to servers and their responses
CREATE TABLE IF NOT EXISTS server_commands (
    time TIMESTAMPTZ NOT NULL,
    server_id TEXT NOT NULL,
    command_id UUID DEFAULT gen_random_uuid(),
    command_type TEXT NOT NULL,
    command_data JSONB,
    status TEXT DEFAULT 'pending', -- pending, sent, executed, failed, timeout
    response JSONB,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    sent_at TIMESTAMPTZ,
    executed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    retry_count INTEGER DEFAULT 0,
    PRIMARY KEY (time, command_id) -- Composite primary key with time for hypertable
);

-- Server Events Hypertable
-- Stores various events like connections, disconnections, errors, alerts
CREATE TABLE IF NOT EXISTS server_events (
    time TIMESTAMPTZ NOT NULL,
    server_id TEXT NOT NULL,
    event_type TEXT NOT NULL, -- connection, disconnection, error, alert, heartbeat
    event_data JSONB,
    level TEXT DEFAULT 'info', -- debug, info, warn, error, critical
    source TEXT, -- agent, api, websocket, system
    message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Dead Letter Queue (Regular table, not hypertable)
-- Stores failed messages for retry processing
CREATE TABLE IF NOT EXISTS dead_letter_queue (
    id BIGSERIAL PRIMARY KEY,
    time TIMESTAMPTZ DEFAULT NOW(),
    topic TEXT NOT NULL,
    partition INTEGER,
    message_offset BIGINT,
    message JSONB NOT NULL,
    error TEXT NOT NULL,
    attempts INTEGER DEFAULT 0,
    server_id TEXT,
    processed_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create hypertables
SELECT create_hypertable('server_metrics', 'time', 
    chunk_time_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

SELECT create_hypertable('server_status', 'time', 
    chunk_time_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

SELECT create_hypertable('server_commands', 'time', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

SELECT create_hypertable('server_events', 'time', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_server_metrics_server_id_time ON server_metrics (server_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_server_metrics_time ON server_metrics (time DESC);
CREATE INDEX IF NOT EXISTS idx_server_metrics_hostname ON server_metrics (hostname);
CREATE INDEX IF NOT EXISTS idx_server_metrics_cpu_usage ON server_metrics (cpu_usage) WHERE cpu_usage > 80;
CREATE INDEX IF NOT EXISTS idx_server_metrics_memory_usage ON server_metrics (memory_usage) WHERE memory_usage > 85;

CREATE INDEX IF NOT EXISTS idx_server_status_server_id_time ON server_status (server_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_server_status_online ON server_status (online) WHERE online = TRUE;
CREATE INDEX IF NOT EXISTS idx_server_status_last_seen ON server_status (last_seen DESC);

CREATE INDEX IF NOT EXISTS idx_server_commands_server_id_time ON server_commands (server_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_server_commands_status ON server_commands (status, time DESC);
CREATE INDEX IF NOT EXISTS idx_server_commands_type ON server_commands (command_type, time DESC);
CREATE INDEX IF NOT EXISTS idx_server_commands_expires ON server_commands (expires_at) WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_server_events_server_id_time ON server_events (server_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_server_events_type ON server_events (event_type, time DESC);
CREATE INDEX IF NOT EXISTS idx_server_events_level ON server_events (level, time DESC) WHERE level IN ('error', 'critical');

CREATE INDEX IF NOT EXISTS idx_dlq_server_id ON dead_letter_queue (server_id);
CREATE INDEX IF NOT EXISTS idx_dlq_topic ON dead_letter_queue (topic);
CREATE INDEX IF NOT EXISTS idx_dlq_next_retry ON dead_letter_queue (next_retry_at) WHERE attempts < 5;

-- Skip continuous aggregates creation here - will be created after cleanup

-- Create useful functions
ALTER TABLE server_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'server_id',
    timescaledb.compress_orderby = 'time DESC'
);

ALTER TABLE server_status SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'server_id',
    timescaledb.compress_orderby = 'time DESC'
);

-- Add retention policies
-- Keep metrics for 30 days
SELECT add_retention_policy('server_metrics', INTERVAL '30 days', if_not_exists => TRUE);

-- Keep status for 7 days
SELECT add_retention_policy('server_status', INTERVAL '7 days', if_not_exists => TRUE);

-- Keep commands for 14 days
SELECT add_retention_policy('server_commands', INTERVAL '14 days', if_not_exists => TRUE);

-- Keep events for 14 days
SELECT add_retention_policy('server_events', INTERVAL '14 days', if_not_exists => TRUE);

-- Add compression policies after 1 day for metrics
SELECT add_compression_policy('server_metrics', INTERVAL '1 day', if_not_exists => TRUE);

-- Add compression policies after 6 hours for status
SELECT add_compression_policy('server_status', INTERVAL '6 hours', if_not_exists => TRUE);

-- Add refresh policies for continuous aggregates
-- FORCE DROP existing policies first to avoid conflicts
DO $$
BEGIN
    -- Drop existing continuous aggregates if they exist
    EXECUTE 'DROP MATERIALIZED VIEW IF EXISTS metrics_5m_avg CASCADE';
    EXECUTE 'DROP MATERIALIZED VIEW IF EXISTS metrics_1h_avg CASCADE';
    EXECUTE 'DROP MATERIALIZED VIEW IF EXISTS server_uptime_daily CASCADE';
    EXECUTE 'DROP MATERIALIZED VIEW IF EXISTS alert_stats_hourly CASCADE';
    
    -- Drop existing policies if they exist
    PERFORM remove_continuous_aggregate_policy('metrics_5m_avg', if_exists => TRUE);
    PERFORM remove_continuous_aggregate_policy('metrics_1h_avg', if_exists => TRUE);
    PERFORM remove_continuous_aggregate_policy('server_uptime_daily', if_exists => TRUE);
    PERFORM remove_continuous_aggregate_policy('alert_stats_hourly', if_exists => TRUE);
EXCEPTION
    WHEN OTHERS THEN
        NULL; -- Ignore any errors during cleanup
END;
$$;

-- Create continuous aggregates after cleanup
DO $$
BEGIN
    -- Check and create continuous aggregates safely
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'metrics_5m_avg') THEN
        EXECUTE '
        CREATE MATERIALIZED VIEW metrics_5m_avg WITH (timescaledb.continuous) AS
        SELECT 
            time_bucket(''5 minutes'', time) AS bucket,
            server_id,
            AVG(cpu_usage) as avg_cpu,
            AVG(memory_usage) as avg_memory,
            AVG(disk_usage) as avg_disk,
            AVG(network_usage) as avg_network,
            AVG(cpu_temperature) as avg_temp
        FROM server_metrics
        GROUP BY bucket, server_id';
        RAISE NOTICE 'Created metrics_5m_avg continuous aggregate';
    ELSE
        RAISE NOTICE 'metrics_5m_avg already exists, skipping creation';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'server_uptime_daily') THEN
        EXECUTE '
        CREATE MATERIALIZED VIEW server_uptime_daily WITH (timescaledb.continuous) AS
        SELECT 
            time_bucket(''1 day'', time) AS bucket,
            server_id,
            hostname,
            AVG(CASE WHEN online THEN 1 ELSE 0 END) * 100 as uptime_percentage,
            COUNT(*) as status_checks,
            AVG(response_time_ms) as avg_response_time
        FROM server_status
        GROUP BY bucket, server_id, hostname';
        RAISE NOTICE 'Created server_uptime_daily continuous aggregate';
    ELSE
        RAISE NOTICE 'server_uptime_daily already exists, skipping creation';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'alert_stats_hourly') THEN
        EXECUTE '
        CREATE MATERIALIZED VIEW alert_stats_hourly WITH (timescaledb.continuous) AS
        SELECT 
            time_bucket(''1 hour'', time) AS bucket,
            server_id,
            COUNT(*) as alert_count,
            COUNT(CASE WHEN level = ''critical'' THEN 1 END) as critical_count,
            COUNT(CASE WHEN level = ''error'' THEN 1 END) as error_count,
            COUNT(CASE WHEN level = ''warn'' THEN 1 END) as warning_count
        FROM server_events
        WHERE level IN (''warn'', ''error'', ''critical'')
        GROUP BY bucket, server_id';
        RAISE NOTICE 'Created alert_stats_hourly continuous aggregate';
    ELSE
        RAISE NOTICE 'alert_stats_hourly already exists, skipping creation';
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error creating continuous aggregates: %', SQLERRM;
END;
$$;

-- Add refresh policies with conservative windows
DO $$
BEGIN
    -- Check and add policies safely
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.views 
        WHERE table_name = 'continuous_aggregate_policies' 
        AND table_schema = 'timescaledb_information'
    ) THEN
        RAISE NOTICE 'timescaledb_information.continuous_aggregate_policies not available, skipping policy checks';
    ELSE
        IF NOT EXISTS (
            SELECT 1 FROM timescaledb_information.continuous_aggregate_policies 
            WHERE hypertable_name = 'server_metrics' 
            AND view_name = 'metrics_5m_avg'
        ) THEN
            PERFORM add_continuous_aggregate_policy('metrics_5m_avg', 
                start_offset => INTERVAL '3 hours',
                end_offset => INTERVAL '1 hour',
                schedule_interval => INTERVAL '1 minute'
            );
            RAISE NOTICE 'Created metrics_5m_avg refresh policy';
        ELSE
            RAISE NOTICE 'metrics_5m_avg refresh policy already exists';
        END IF;
        
        -- Note: metrics_1h_avg view creation was removed from above, so we'll skip its policy
        IF EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'metrics_1h_avg') AND
           NOT EXISTS (
               SELECT 1 FROM timescaledb_information.continuous_aggregate_policies 
               WHERE hypertable_name = 'server_metrics' 
               AND view_name = 'metrics_1h_avg'
           ) THEN
            PERFORM add_continuous_aggregate_policy('metrics_1h_avg', 
                start_offset => INTERVAL '3 days',
                end_offset => INTERVAL '3 hours',
                schedule_interval => INTERVAL '5 minutes'
            );
            RAISE NOTICE 'Created metrics_1h_avg refresh policy';
        END IF;
        
        IF NOT EXISTS (
            SELECT 1 FROM timescaledb_information.continuous_aggregate_policies 
            WHERE hypertable_name = 'server_status' 
            AND view_name = 'server_uptime_daily'
        ) THEN
            PERFORM add_continuous_aggregate_policy('server_uptime_daily', 
                start_offset => INTERVAL '3 days',
                end_offset => INTERVAL '3 hours',
                schedule_interval => INTERVAL '1 hour'
            );
            RAISE NOTICE 'Created server_uptime_daily refresh policy';
        ELSE
            RAISE NOTICE 'server_uptime_daily refresh policy already exists';
        END IF;
        
        IF NOT EXISTS (
            SELECT 1 FROM timescaledb_information.continuous_aggregate_policies 
            WHERE hypertable_name = 'server_events' 
            AND view_name = 'alert_stats_hourly'
        ) THEN
            PERFORM add_continuous_aggregate_policy('alert_stats_hourly', 
                start_offset => INTERVAL '3 days',
                end_offset => INTERVAL '3 hours',
                schedule_interval => INTERVAL '10 minutes'
            );
            RAISE NOTICE 'Created alert_stats_hourly refresh policy';
        ELSE
            RAISE NOTICE 'alert_stats_hourly refresh policy already exists';
        END IF;
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error creating refresh policies: %', SQLERRM;
END;
$$;

-- Create useful functions
-- Function to get latest metrics for a server
CREATE OR REPLACE FUNCTION get_latest_metrics(p_server_id TEXT)
RETURNS TABLE (
    metric_time TIMESTAMPTZ,
    cpu_usage DOUBLE PRECISION,
    memory_usage DOUBLE PRECISION,
    disk_usage DOUBLE PRECISION,
    network_usage DOUBLE PRECISION,
    cpu_temperature DOUBLE PRECISION,
    highest_temperature DOUBLE PRECISION
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        time AS metric_time,
        cpu_usage,
        memory_usage,
        disk_usage,
        network_usage,
        cpu_temperature,
        highest_temperature
    FROM server_metrics
    WHERE server_id = p_server_id
    ORDER BY time DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to get server uptime for last 24 hours
CREATE OR REPLACE FUNCTION get_uptime_last_24h(p_server_id TEXT)
RETURNS DOUBLE PRECISION AS $$
DECLARE
    uptime_percent DOUBLE PRECISION;
BEGIN
    SELECT COALESCE(AVG(CASE WHEN online THEN 1 ELSE 0 END) * 100, 0)
    INTO uptime_percent
    FROM server_status
    WHERE server_id = p_server_id 
    AND time >= NOW() - INTERVAL '24 hours';
    
    RETURN uptime_percent;
END;
$$ LANGUAGE plpgsql;

-- Function to get metrics history with aggregation
CREATE OR REPLACE FUNCTION get_metrics_history(
    p_server_id TEXT,
    p_start_time TIMESTAMPTZ,
    p_end_time TIMESTAMPTZ,
    p_interval INTERVAL DEFAULT INTERVAL '1 hour'
)
RETURNS TABLE (
    time_bucket TIMESTAMPTZ,
    avg_cpu DOUBLE PRECISION,
    max_cpu DOUBLE PRECISION,
    avg_memory DOUBLE PRECISION,
    max_memory DOUBLE PRECISION,
    avg_disk DOUBLE PRECISION,
    max_disk DOUBLE PRECISION,
    avg_network DOUBLE PRECISION,
    max_network DOUBLE PRECISION,
    sample_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        time_bucket(p_interval, time) as time_bucket,
        AVG(cpu_usage) as avg_cpu,
        MAX(cpu_usage) as max_cpu,
        AVG(memory_usage) as avg_memory,
        MAX(memory_usage) as max_memory,
        AVG(disk_usage) as avg_disk,
        MAX(disk_usage) as max_disk,
        AVG(network_usage) as avg_network,
        MAX(network_usage) as max_network,
        COUNT(*) as sample_count
    FROM server_metrics
    WHERE server_id = p_server_id 
    AND time BETWEEN p_start_time AND p_end_time
    GROUP BY time_bucket
    ORDER BY time_bucket;
END;
$$ LANGUAGE plpgsql;

-- Create view for active servers (seen in last 5 minutes)
CREATE OR REPLACE VIEW active_servers AS
SELECT DISTINCT 
    server_id,
    MAX(hostname) as hostname,
    MAX(os_info) as os_info,
    MAX(agent_version) as agent_version,
    MAX(last_seen) as last_seen,
    CASE WHEN MAX(last_seen) >= NOW() - INTERVAL '5 minutes' THEN TRUE ELSE FALSE END as online
FROM server_status
WHERE time >= NOW() - INTERVAL '1 hour'
GROUP BY server_id;

-- Create view for servers with alerts
CREATE OR REPLACE VIEW servers_with_alerts AS
SELECT 
    server_id,
    COUNT(*) as alert_count,
    COUNT(CASE WHEN level = 'critical' THEN 1 END) as critical_count,
    MAX(time) as last_alert_time
FROM server_events
WHERE level IN ('warn', 'error', 'critical')
AND time >= NOW() - INTERVAL '24 hours'
GROUP BY server_id
HAVING COUNT(*) > 0;

-- Grant permissions (adjust as needed)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO servereye_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO servereye_user;

-- Create sample data for testing (optional)
-- INSERT INTO server_metrics (
--     server_id, cpu_usage, memory_usage, disk_usage, network_usage,
--     cpu_temperature, highest_temperature, hostname, os_info
-- ) VALUES (
--     'test-server-001', 75.5, 60.2, 45.8, 120.4,
--     65.2, 68.1, 'test-server', 'Ubuntu 22.04'
-- );

COMMIT;
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
