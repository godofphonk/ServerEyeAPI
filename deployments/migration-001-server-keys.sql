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
