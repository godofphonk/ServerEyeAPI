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
-- Migration 006: Memory and Motherboard Information
-- Description: Adds detailed memory modules and motherboard information tables
-- Author: ServerEye Team
-- Date: 2026-02-19

-- Memory modules information
CREATE TABLE IF NOT EXISTS static_data.memory_modules (
    id SERIAL PRIMARY KEY,
    server_id VARCHAR(50) REFERENCES static_data.server_info(server_id) ON DELETE CASCADE,
    slot_name VARCHAR(50) NOT NULL,
    size_gb INTEGER NOT NULL,
    memory_type VARCHAR(50), -- DDR3, DDR4, DDR5
    frequency_mhz INTEGER,
    manufacturer VARCHAR(255),
    part_number VARCHAR(255),
    speed_mts INTEGER, -- For DDR5 (MT/s instead of MHz)
    voltage REAL, -- Memory voltage (e.g., 1.35V)
    timings VARCHAR(100), -- CAS timings (e.g., 16-18-18-38)
    ecc BOOLEAN DEFAULT false, -- ECC memory
    registered BOOLEAN DEFAULT false, -- Registered memory
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(server_id, slot_name)
);

-- Extended motherboard information
CREATE TABLE IF NOT EXISTS static_data.motherboard_info (
    server_id VARCHAR(50) PRIMARY KEY REFERENCES static_data.server_info(server_id) ON DELETE CASCADE,
    manufacturer VARCHAR(255),
    model VARCHAR(255),
    chipset VARCHAR(100),
    bios_version VARCHAR(100),
    bios_date DATE,
    bios_vendor VARCHAR(255),
    form_factor VARCHAR(100), -- ATX, Micro-ATX, Mini-ITX, etc.
    max_memory_gb INTEGER,
    memory_slots INTEGER,
    supported_memory_types TEXT[], -- Array of supported types: ['DDR4', 'DDR5']
    onboard_video BOOLEAN DEFAULT false,
    onboard_audio BOOLEAN DEFAULT true,
    onboard_network BOOLEAN DEFAULT true,
    sata_ports INTEGER,
    sata_speed VARCHAR(50), -- SATA 3.0, SATA 6.0
    m2_slots INTEGER,
    pcie_slots TEXT[], -- Array of PCIe slots: ['x16', 'x8', 'x4']
    usb_ports_total INTEGER,
    usb_ports_2_0 INTEGER,
    usb_ports_3_0 INTEGER,
    usb_ports_c INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Update hardware_info table to include memory summary
ALTER TABLE static_data.hardware_info 
ADD COLUMN IF NOT EXISTS memory_type VARCHAR(50),
ADD COLUMN IF NOT EXISTS memory_frequency_mhz INTEGER,
ADD COLUMN IF NOT EXISTS memory_slots_total INTEGER,
ADD COLUMN IF NOT EXISTS memory_slots_used INTEGER;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_memory_modules_server_id ON static_data.memory_modules(server_id);
CREATE INDEX IF NOT EXISTS idx_memory_modules_slot ON static_data.memory_modules(server_id, slot_name);
CREATE INDEX IF NOT EXISTS idx_motherboard_info_manufacturer ON static_data.motherboard_info(manufacturer);
CREATE INDEX IF NOT EXISTS idx_motherboard_info_model ON static_data.motherboard_info(model);

-- Create triggers for automatic updated_at timestamps
CREATE TRIGGER update_memory_modules_updated_at
    BEFORE UPDATE ON static_data.memory_modules
    FOR EACH ROW
    EXECUTE FUNCTION static_data.update_updated_at_column();

CREATE TRIGGER update_motherboard_info_updated_at
    BEFORE UPDATE ON static_data.motherboard_info
    FOR EACH ROW
    EXECUTE FUNCTION static_data.update_updated_at_column();

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA static_data TO servereye;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA static_data TO servereye;

-- Comments for documentation
COMMENT ON TABLE static_data.memory_modules IS 'Detailed information about individual memory modules';
COMMENT ON TABLE static_data.motherboard_info IS 'Extended motherboard specifications and capabilities';
COMMENT ON COLUMN static_data.memory_modules.speed_mts IS 'Transfer rate for DDR5 (MT/s instead of MHz)';
COMMENT ON COLUMN static_data.memory_modules.timings IS 'CAS latency timings (e.g., 16-18-18-38)';
COMMENT ON COLUMN static_data.motherboard_info.supported_memory_types IS 'Array of supported memory types';
COMMENT ON COLUMN static_data.motherboard_info.pcie_slots IS 'Array of available PCIe slots';
-- Migration 007: Fix HardwareInfo table structure
-- Description: Remove duplicate fields and fix data types
-- Author: ServerEye Team
-- Date: 2026-02-19

-- Remove duplicate fields from hardware_info table
ALTER TABLE static_data.hardware_info 
DROP COLUMN IF EXISTS motherboard,
DROP COLUMN IF EXISTS bios_version;

-- Fix data types for frequency and memory
ALTER TABLE static_data.hardware_info 
ALTER COLUMN cpu_frequency_mhz TYPE DOUBLE PRECISION USING cpu_frequency_mhz::DOUBLE PRECISION,
ALTER COLUMN total_memory_gb TYPE DOUBLE PRECISION USING total_memory_gb::DOUBLE PRECISION;

-- Add comments for clarity
COMMENT ON COLUMN static_data.hardware_info.cpu_frequency_mhz IS 'CPU frequency in MHz (can be fractional)';
COMMENT ON COLUMN static_data.hardware_info.total_memory_gb IS 'Total system memory in GB (can be fractional)';
COMMENT ON COLUMN static_data.hardware_info.gpu_model IS 'GPU model name (empty if no GPU)';
COMMENT ON COLUMN static_data.hardware_info.gpu_driver IS 'GPU driver version (empty if no GPU)';
COMMENT ON COLUMN static_data.hardware_info.gpu_memory_gb IS 'GPU memory in GB (0 if no GPU)';

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA static_data TO servereye;
-- Migration 008: Add storage_temperatures to server_metrics
-- Add storage temperatures field for tracking individual storage device temperatures

-- Add storage_temperatures column to server_metrics table
ALTER TABLE server_metrics 
ADD COLUMN IF NOT EXISTS storage_temperatures JSONB;

-- Add comment for documentation
COMMENT ON COLUMN server_metrics.storage_temperatures IS 'JSON array of storage device temperatures with device name, type, and temperature';

-- Update existing records to have empty array for storage_temperatures
UPDATE server_metrics 
SET storage_temperatures = '[]' 
WHERE storage_temperatures IS NULL;

-- Add index for storage temperature queries (optional, for performance)
CREATE INDEX IF NOT EXISTS idx_server_metrics_storage_temperatures 
ON server_metrics USING GIN (storage_temperatures) 
WHERE storage_temperatures IS NOT NULL;

-- Grant permissions
GRANT ALL ON TABLE server_metrics TO servereye_user;
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

-- Migration 009: Fix numeric fields to support decimal values
-- For hardware info that can have decimal values (frequency, memory, etc.)

-- Convert integer fields to numeric to support decimal values
ALTER TABLE static_data.hardware_info 
ALTER COLUMN cpu_frequency_mhz TYPE NUMERIC(10,2) USING cpu_frequency_mhz::NUMERIC(10,2);

ALTER TABLE static_data.hardware_info 
ALTER COLUMN gpu_memory_gb TYPE NUMERIC(10,2) USING gpu_memory_gb::NUMERIC(10,2);

ALTER TABLE static_data.hardware_info 
ALTER COLUMN total_memory_gb TYPE NUMERIC(10,2) USING total_memory_gb::NUMERIC(10,2);

-- Add comments for clarity
COMMENT ON COLUMN static_data.hardware_info.cpu_frequency_mhz IS 'CPU frequency in MHz (can be decimal)';
COMMENT ON COLUMN static_data.hardware_info.gpu_memory_gb IS 'GPU memory in GB (can be decimal)';
COMMENT ON COLUMN static_data.hardware_info.total_memory_gb IS 'Total system memory in GB (can be decimal)';
-- Migration 010: Add motherboard_info table
-- This table stores motherboard and BIOS information

-- Create motherboard_info table
CREATE TABLE IF NOT EXISTS static_data.motherboard_info (
    server_id VARCHAR(255) PRIMARY KEY,
    manufacturer VARCHAR(255),
    model VARCHAR(255),
    chipset VARCHAR(255),
    bios_version VARCHAR(255),
    bios_date DATE,
    bios_vendor VARCHAR(255),
    form_factor VARCHAR(255),
    max_memory_gb INTEGER,
    memory_slots INTEGER,
    supported_memory_types TEXT[], -- Array of supported memory types
    onboard_video BOOLEAN DEFAULT FALSE,
    onboard_audio BOOLEAN DEFAULT FALSE,
    onboard_network BOOLEAN DEFAULT FALSE,
    sata_ports INTEGER,
    sata_speed VARCHAR(50), -- e.g., "6 Gbps", "3 Gbps"
    m2_slots INTEGER,
    pcie_slots TEXT[], -- Array of PCIe slots info
    usb_ports_total INTEGER,
    usb_ports_2_0 INTEGER,
    usb_ports_3_0 INTEGER,
    usb_ports_c INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign key constraint to server_info table
    CONSTRAINT fk_motherboard_server 
        FOREIGN KEY (server_id) 
        REFERENCES static_data.server_info(server_id)
        ON DELETE CASCADE
);

-- Create indexes for faster queries
CREATE INDEX IF NOT EXISTS idx_motherboard_info_manufacturer 
ON static_data.motherboard_info(manufacturer);

CREATE INDEX IF NOT EXISTS idx_motherboard_info_model 
ON static_data.motherboard_info(model);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION static_data.update_motherboard_info_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_update_motherboard_info_timestamp
    BEFORE UPDATE ON static_data.motherboard_info
    FOR EACH ROW
    EXECUTE FUNCTION static_data.update_motherboard_info_timestamp();

-- Add comment
COMMENT ON TABLE static_data.motherboard_info IS 'Stores motherboard and BIOS information for each server';
COMMENT ON COLUMN static_data.motherboard_info.server_id IS 'Foreign key to server_info';
COMMENT ON COLUMN static_data.motherboard_info.motherboard IS 'Motherboard model/name';
COMMENT ON COLUMN static_data.motherboard_info.bios_version IS 'BIOS firmware version';
COMMENT ON COLUMN static_data.motherboard_info.bios_manufacturer IS 'BIOS manufacturer';
COMMENT ON COLUMN static_data.motherboard_info.bios_release_date IS 'BIOS release date';
COMMENT ON COLUMN static_data.motherboard_info.bios_characteristics IS 'Array of BIOS characteristics/features';
