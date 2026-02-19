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
