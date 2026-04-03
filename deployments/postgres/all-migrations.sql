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

-- Migration 003: Add sources column to servers table
-- For existing production databases

-- Add sources column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='servers' AND column_name='sources'
    ) THEN
        ALTER TABLE servers ADD COLUMN sources TEXT DEFAULT '';
    END IF;
END $$;

-- Create index for sources column for better performance
CREATE INDEX IF NOT EXISTS idx_servers_sources ON servers (sources);

-- Update existing records to have empty sources if NULL
UPDATE servers SET sources = '' WHERE sources IS NULL;
-- Migration 004: API Keys Management
-- Description: Adds API keys table for service-to-service authentication

-- Create api_keys table
CREATE TABLE IF NOT EXISTS api_keys (
    key_id TEXT PRIMARY KEY,
    key_hash TEXT NOT NULL,
    service_id TEXT NOT NULL,
    service_name TEXT,
    permissions TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT true,
    created_by TEXT,
    notes TEXT
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_api_keys_service_id ON api_keys(service_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at) WHERE expires_at IS NOT NULL;

-- Create audit log table for API key usage
CREATE TABLE IF NOT EXISTS api_key_audit_log (
    id SERIAL PRIMARY KEY,
    key_id TEXT NOT NULL REFERENCES api_keys(key_id) ON DELETE CASCADE,
    accessed_at TIMESTAMPTZ DEFAULT NOW(),
    endpoint TEXT,
    ip_address TEXT,
    user_agent TEXT,
    success BOOLEAN DEFAULT true,
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_api_key_audit_key_id ON api_key_audit_log(key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_audit_accessed_at ON api_key_audit_log(accessed_at);

-- Insert default API key for C# backend (hash for "development_key_change_in_production")
-- Password hash generated with bcrypt cost 10
INSERT INTO api_keys (key_id, key_hash, service_id, service_name, permissions, notes, is_active)
VALUES (
    'key_csharp_backend_001',
    '$2a$10$rQZ9YJZxvJKx7YqXqZxQxOYqXqZxQxOYqXqZxQxOYqXqZxQxOYqXq',
    'csharp-backend',
    'C# Web Backend',
    ARRAY['metrics:read', 'servers:read', 'servers:validate'],
    'Default API key for C# backend - CHANGE IN PRODUCTION',
    true
) ON CONFLICT (key_id) DO NOTHING;

COMMIT;
-- Migration 009: Alerts Table
-- Description: Creates table for storing system alerts
-- Author: ServerEye Team
-- Date: 2026-03-04

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    server_id VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    device VARCHAR(255),
    temperature DOUBLE PRECISION,
    threshold DOUBLE PRECISION,
    value DOUBLE PRECISION,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_alerts_server_id ON alerts (server_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts (type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts (severity, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_active ON alerts (server_id, status) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts (created_at DESC);

-- Add comments
COMMENT ON TABLE alerts IS 'Stores system alerts for server monitoring';
COMMENT ON COLUMN alerts.id IS 'Unique alert identifier (UUID)';
COMMENT ON COLUMN alerts.type IS 'Alert type: storage_temperature, cpu_temperature, memory_usage, disk_usage, network_usage, load_average, system_temperature';
COMMENT ON COLUMN alerts.severity IS 'Alert severity: info, warning, critical';
COMMENT ON COLUMN alerts.status IS 'Alert status: active, resolved';
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

-- Migration 010: Add server source identifiers table
-- For storing multiple identifiers per source type (TG IDs, user IDs, emails)

-- Create server_source_identifiers table
CREATE TABLE IF NOT EXISTS server_source_identifiers (
    id SERIAL PRIMARY KEY,
    server_id VARCHAR(255) NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    identifier VARCHAR(255) NOT NULL,
    identifier_type VARCHAR(50) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (server_id) REFERENCES servers(server_id) ON DELETE CASCADE,
    UNIQUE(server_id, source_type, identifier)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_server_source_identifiers_server_id ON server_source_identifiers(server_id);
CREATE INDEX IF NOT EXISTS idx_server_source_identifiers_source_type ON server_source_identifiers(source_type);
CREATE INDEX IF NOT EXISTS idx_server_source_identifiers_identifier ON server_source_identifiers(identifier);
CREATE INDEX IF NOT EXISTS idx_server_source_identifiers_composite ON server_source_identifiers(server_id, source_type);

-- Create trigger for updated_at timestamp
CREATE OR REPLACE FUNCTION update_server_source_identifiers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_server_source_identifiers_updated_at
    BEFORE UPDATE ON server_source_identifiers
    FOR EACH ROW
    EXECUTE FUNCTION update_server_source_identifiers_updated_at();

-- Add comment
COMMENT ON TABLE server_source_identifiers IS 'Stores identifiers for server sources (TG IDs, user IDs, emails)';
COMMENT ON COLUMN server_source_identifiers.source_type IS 'Type of source: TGBot, Web, Email, etc.';
COMMENT ON COLUMN server_source_identifiers.identifier IS 'Identifier value: Telegram ID, User ID, Email address';
COMMENT ON COLUMN server_source_identifiers.identifier_type IS 'Type of identifier: telegram_id, user_id, email';
COMMENT ON COLUMN server_source_identifiers.metadata IS 'Additional metadata in JSON format';
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

-- Migration 011: Add telegram_id field to server_source_identifiers
-- For linking accounts between Telegram bot and web application

-- Add telegram_id column (optional field)
ALTER TABLE server_source_identifiers 
ADD COLUMN IF NOT EXISTS telegram_id BIGINT;

-- Create index for fast lookups by telegram_id
CREATE INDEX IF NOT EXISTS idx_server_source_identifiers_telegram_id 
ON server_source_identifiers(telegram_id) 
WHERE telegram_id IS NOT NULL;

-- Add comment
COMMENT ON COLUMN server_source_identifiers.telegram_id IS 'Telegram user ID for linking accounts between TG bot and web application';
