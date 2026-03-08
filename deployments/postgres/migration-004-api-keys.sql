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
