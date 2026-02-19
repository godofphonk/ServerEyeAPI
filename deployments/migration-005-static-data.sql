-- Migration 005: Static Server Data Schema
-- Description: Creates schema and tables for storing static/persistent server information
-- Author: ServerEye Team
-- Date: 2026-02-19

-- Create schema for static data
CREATE SCHEMA IF NOT EXISTS static_data;

-- Server basic information
CREATE TABLE IF NOT EXISTS static_data.server_info (
    server_id VARCHAR(50) PRIMARY KEY,
    hostname VARCHAR(255) NOT NULL,
    os VARCHAR(100),
    os_version VARCHAR(100),
    kernel VARCHAR(100),
    architecture VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Hardware information
CREATE TABLE IF NOT EXISTS static_data.hardware_info (
    server_id VARCHAR(50) PRIMARY KEY REFERENCES static_data.server_info(server_id) ON DELETE CASCADE,
    cpu_model VARCHAR(255),
    cpu_cores INTEGER,
    cpu_threads INTEGER,
    cpu_frequency_mhz INTEGER,
    gpu_model VARCHAR(255),
    gpu_driver VARCHAR(100),
    gpu_memory_gb INTEGER,
    total_memory_gb INTEGER,
    motherboard VARCHAR(255),
    bios_version VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Network interfaces (static configuration)
CREATE TABLE IF NOT EXISTS static_data.network_interfaces (
    id SERIAL PRIMARY KEY,
    server_id VARCHAR(50) REFERENCES static_data.server_info(server_id) ON DELETE CASCADE,
    interface_name VARCHAR(100) NOT NULL,
    mac_address VARCHAR(17),
    interface_type VARCHAR(50), -- ethernet, wifi, virtual, loopback
    speed_mbps INTEGER,
    vendor VARCHAR(255),
    driver VARCHAR(100),
    is_physical BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(server_id, interface_name)
);

-- Disk information (static configuration)
CREATE TABLE IF NOT EXISTS static_data.disk_info (
    id SERIAL PRIMARY KEY,
    server_id VARCHAR(50) REFERENCES static_data.server_info(server_id) ON DELETE CASCADE,
    device_name VARCHAR(100) NOT NULL,
    model VARCHAR(255),
    serial_number VARCHAR(100),
    size_gb BIGINT,
    disk_type VARCHAR(50), -- ssd, hdd, nvme, raid
    interface_type VARCHAR(50), -- sata, nvme, usb
    filesystem VARCHAR(100),
    mount_point VARCHAR(255),
    is_system_disk BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(server_id, device_name)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_server_info_hostname ON static_data.server_info(hostname);
CREATE INDEX IF NOT EXISTS idx_hardware_info_server_id ON static_data.hardware_info(server_id);
CREATE INDEX IF NOT EXISTS idx_network_interfaces_server_id ON static_data.network_interfaces(server_id);
CREATE INDEX IF NOT EXISTS idx_disk_info_server_id ON static_data.disk_info(server_id);
CREATE INDEX IF NOT EXISTS idx_disk_info_mount_point ON static_data.disk_info(mount_point);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION static_data.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic updated_at updates
CREATE TRIGGER update_server_info_updated_at
    BEFORE UPDATE ON static_data.server_info
    FOR EACH ROW
    EXECUTE FUNCTION static_data.update_updated_at_column();

CREATE TRIGGER update_hardware_info_updated_at
    BEFORE UPDATE ON static_data.hardware_info
    FOR EACH ROW
    EXECUTE FUNCTION static_data.update_updated_at_column();

CREATE TRIGGER update_network_interfaces_updated_at
    BEFORE UPDATE ON static_data.network_interfaces
    FOR EACH ROW
    EXECUTE FUNCTION static_data.update_updated_at_column();

CREATE TRIGGER update_disk_info_updated_at
    BEFORE UPDATE ON static_data.disk_info
    FOR EACH ROW
    EXECUTE FUNCTION static_data.update_updated_at_column();

-- Grant permissions
GRANT USAGE ON SCHEMA static_data TO servereye;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA static_data TO servereye;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA static_data TO servereye;

-- Comments for documentation
COMMENT ON SCHEMA static_data IS 'Schema for storing static/persistent server information';
COMMENT ON TABLE static_data.server_info IS 'Basic server system information';
COMMENT ON TABLE static_data.hardware_info IS 'Hardware specifications and components';
COMMENT ON TABLE static_data.network_interfaces IS 'Network interface static configuration';
COMMENT ON TABLE static_data.disk_info IS 'Disk and storage device information';
