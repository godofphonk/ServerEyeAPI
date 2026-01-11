-- ServerEye Database Schema
-- Initialize PostgreSQL database for ServerEye API

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Registered Keys table (PGRegisteredkeys)
CREATE TABLE IF NOT EXISTS registered_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) UNIQUE NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    operating_system VARCHAR(100) NOT NULL,
    installation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    agent_version VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'non-active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Status constraint
    CONSTRAINT valid_status CHECK (
        status IN ('non-active', 'active(telegram)', 'active(web)', 'active(telegram+web)')
    )
);

-- Metrics table (for real-time data)
CREATE TABLE IF NOT EXISTS metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(10,2) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (key) REFERENCES registered_keys(key) ON DELETE CASCADE
);

-- Commands table (for sending commands to agents)
CREATE TABLE IF NOT EXISTS commands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) NOT NULL,
    command_type VARCHAR(100) NOT NULL,
    command_data JSONB,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    executed_at TIMESTAMP,
    FOREIGN KEY (key) REFERENCES registered_keys(key) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_registered_keys_key ON registered_keys(key);
CREATE INDEX IF NOT EXISTS idx_registered_keys_status ON registered_keys(status);
CREATE INDEX IF NOT EXISTS idx_registered_keys_created_at ON registered_keys(created_at);

CREATE INDEX IF NOT EXISTS idx_metrics_key ON metrics(key);
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_metrics_name_timestamp ON metrics(metric_name, timestamp);

CREATE INDEX IF NOT EXISTS idx_commands_key ON commands(key);
CREATE INDEX IF NOT EXISTS idx_commands_status ON commands(status);

-- Update trigger for registered_keys table
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_registered_keys_updated_at 
    BEFORE UPDATE ON registered_keys 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Functions for status management
CREATE OR REPLACE FUNCTION set_key_status(
    p_key VARCHAR(255),
    p_status VARCHAR(50)
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE registered_keys 
    SET status = p_status, updated_at = CURRENT_TIMESTAMP
    WHERE key = p_key;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Function to get active keys by type
CREATE OR REPLACE FUNCTION get_active_keys_by_type(
    p_type VARCHAR(20) -- 'telegram', 'web', 'telegram+web'
) RETURNS TABLE (
    key VARCHAR(255),
    operating_system VARCHAR(100),
    installation_date TIMESTAMP,
    agent_version VARCHAR(50),
    status VARCHAR(50)
) AS $$
BEGIN
    RETURN QUERY
    SELECT rk.key, rk.operating_system, rk.installation_date, 
           rk.agent_version, rk.status
    FROM registered_keys rk
    WHERE rk.status = CASE 
        WHEN p_type = 'telegram' THEN 'active(telegram)'
        WHEN p_type = 'web' THEN 'active(web)'
        WHEN p_type = 'telegram+web' THEN 'active(telegram+web)'
        ELSE rk.status
    END;
END;
$$ LANGUAGE plpgsql;

-- Sample data for testing (optional)
-- INSERT INTO registered_keys (key, operating_system, agent_version, status) 
-- VALUES 
--     ('test-key-1', 'Ubuntu 22.04', '1.0.0', 'active(web)'),
--     ('test-key-2', 'CentOS 8', '1.0.0', 'active(telegram)'),
--     ('test-key-3', 'Windows 11', '1.0.0', 'active(telegram+web)')
-- ON CONFLICT (key) DO NOTHING;
