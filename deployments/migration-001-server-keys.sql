-- Migration 001: Add server_id and server_key to generated_keys
-- This enables proper WebSocket authentication

-- Add server_id column
ALTER TABLE generated_keys 
ADD COLUMN IF NOT EXISTS server_id TEXT UNIQUE;

-- Add server_key column  
ALTER TABLE generated_keys 
ADD COLUMN IF NOT EXISTS server_key TEXT UNIQUE;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_generated_keys_server_id ON generated_keys (server_id);
CREATE INDEX IF NOT EXISTS idx_generated_keys_server_key ON generated_keys (server_key);

-- Add comments for documentation
COMMENT ON COLUMN generated_keys.server_id IS 'Generated unique server identifier';
COMMENT ON COLUMN generated_keys.server_key IS 'Generated WebSocket authentication key';
