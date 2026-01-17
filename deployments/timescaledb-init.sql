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

-- Create continuous aggregates for fast analytics
-- 1-hour averages
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_1h_avg
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', time) AS hour,
    server_id,
    AVG(cpu_usage) as avg_cpu,
    MAX(cpu_usage) as max_cpu,
    MIN(cpu_usage) as min_cpu,
    AVG(memory_usage) as avg_memory,
    MAX(memory_usage) as max_memory,
    MIN(memory_usage) as min_memory,
    AVG(disk_usage) as avg_disk,
    MAX(disk_usage) as max_disk,
    MIN(disk_usage) as min_disk,
    AVG(network_usage) as avg_network,
    MAX(network_usage) as max_network,
    AVG(highest_temperature) as avg_temperature,
    MAX(highest_temperature) as max_temperature,
    COUNT(*) as sample_count
FROM server_metrics
GROUP BY hour, server_id;

-- 5-minute averages for real-time monitoring
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_5m_avg
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('5 minutes', time) AS five_min,
    server_id,
    AVG(cpu_usage) as avg_cpu,
    MAX(cpu_usage) as max_cpu,
    AVG(memory_usage) as avg_memory,
    MAX(memory_usage) as max_memory,
    AVG(disk_usage) as avg_disk,
    MAX(disk_usage) as max_disk,
    AVG(network_usage) as avg_network,
    MAX(network_usage) as max_network,
    AVG(highest_temperature) as avg_temperature,
    MAX(highest_temperature) as max_temperature,
    COUNT(*) as sample_count
FROM server_metrics
WHERE time >= NOW() - INTERVAL '7 days'
GROUP BY five_min, server_id;

-- Server uptime summary
CREATE MATERIALIZED VIEW IF NOT EXISTS server_uptime_daily
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', time) AS day,
    server_id,
    hostname,
    AVG(CASE WHEN online THEN 1 ELSE 0 END) * 100 as uptime_percentage,
    COUNT(*) as status_checks,
    AVG(response_time_ms) as avg_response_time
FROM server_status
GROUP BY day, server_id, hostname;

-- Alert statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS alert_stats_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', time) AS hour,
    server_id,
    COUNT(*) as alert_count,
    COUNT(CASE WHEN level = 'critical' THEN 1 END) as critical_count,
    COUNT(CASE WHEN level = 'error' THEN 1 END) as error_count,
    COUNT(CASE WHEN level = 'warn' THEN 1 END) as warning_count
FROM server_events
WHERE level IN ('warn', 'error', 'critical')
GROUP BY hour, server_id;

-- Add compression policies for better storage efficiency
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
-- Refresh 5-minute aggregates every minute
SELECT add_continuous_aggregate_policy('metrics_5m_avg', 
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute',
    if_not_exists => TRUE
);

-- Refresh 1-hour aggregates every 5 minutes
SELECT add_continuous_aggregate_policy('metrics_1h_avg', 
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes',
    if_not_exists => TRUE
);

-- Refresh uptime summary every hour
SELECT add_continuous_aggregate_policy('server_uptime_daily', 
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- Refresh alert stats every 10 minutes
SELECT add_continuous_aggregate_policy('alert_stats_hourly', 
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '10 minutes',
    schedule_interval => INTERVAL '10 minutes',
    if_not_exists => TRUE
);

-- Create useful functions
-- Function to get latest metrics for a server
CREATE OR REPLACE FUNCTION get_latest_metrics(p_server_id TEXT)
RETURNS TABLE (
    time TIMESTAMPTZ,
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
        time,
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
