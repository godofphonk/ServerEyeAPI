-- Migration 001: Remove secret_key and add server_id/server_key
-- This removes secret_key dependency and adds proper server identification

-- Drop old indexes
DROP INDEX IF EXISTS idx_generated_keys_secret;

-- Remove secret_key column and add server_id/server_key
ALTER TABLE generated_keys 
DROP COLUMN IF EXISTS secret_key,
ADD COLUMN IF NOT EXISTS server_id TEXT UNIQUE,
ADD COLUMN IF NOT EXISTS server_key TEXT UNIQUE;

-- Update servers table
ALTER TABLE servers 
DROP COLUMN IF EXISTS secret_key;

-- Create new indexes
CREATE INDEX IF NOT EXISTS idx_generated_keys_server_id ON generated_keys (server_id);
CREATE INDEX IF NOT EXISTS idx_generated_keys_server_key ON generated_keys (server_key);

-- Add comments for documentation
COMMENT ON COLUMN generated_keys.server_id IS 'Generated unique server identifier';
COMMENT ON COLUMN generated_keys.server_key IS 'Generated WebSocket authentication key';
