-- Migration 009: Alerts Table
-- Description: Creates table for storing system alerts
-- Author: ServerEye Team
-- Date: 2026-03-04

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    server_id VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    device VARCHAR(255),
    temperature DOUBLE PRECISION,
    threshold DOUBLE PRECISION,
    value DOUBLE PRECISION,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_alerts_server_id ON alerts (server_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts (type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts (severity, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_active ON alerts (server_id, status) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts (created_at DESC);

-- Add comments
COMMENT ON TABLE alerts IS 'Stores system alerts for server monitoring';
COMMENT ON COLUMN alerts.id IS 'Unique alert identifier (UUID)';
COMMENT ON COLUMN alerts.type IS 'Alert type: storage_temperature, cpu_temperature, memory_usage, disk_usage, network_usage, load_average, system_temperature';
COMMENT ON COLUMN alerts.severity IS 'Alert severity: info, warning, critical';
COMMENT ON COLUMN alerts.status IS 'Alert status: active, resolved';
