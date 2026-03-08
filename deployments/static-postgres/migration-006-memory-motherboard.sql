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
