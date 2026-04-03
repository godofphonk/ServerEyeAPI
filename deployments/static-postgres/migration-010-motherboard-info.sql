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
