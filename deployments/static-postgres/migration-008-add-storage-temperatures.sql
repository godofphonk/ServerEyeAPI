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
