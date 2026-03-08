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
