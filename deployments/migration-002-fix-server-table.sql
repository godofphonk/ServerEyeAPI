-- Migration 002: Add missing columns to servers table
-- This adds server_key and updated_at columns if they don't exist

-- Add server_key column if it doesn't exist
ALTER TABLE servers 
ADD COLUMN IF NOT EXISTS server_key TEXT UNIQUE;

-- Add updated_at column if it doesn't exist  
ALTER TABLE servers 
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Create index for server_key if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_servers_server_key ON servers (server_key);

-- Add comments for documentation
COMMENT ON COLUMN servers.server_key IS 'Server authentication key for WebSocket';
COMMENT ON COLUMN servers.updated_at IS 'Last update timestamp';
