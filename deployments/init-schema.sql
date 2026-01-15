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

-- ServerEye Database Schema
-- Compatible with new architecture

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Generated Keys table (for key registration)
CREATE TABLE IF NOT EXISTS generated_keys (
    id BIGSERIAL PRIMARY KEY,
    server_id TEXT UNIQUE,
    server_key TEXT UNIQUE,
    agent_version TEXT,
    os_info TEXT,
    hostname TEXT,
    status TEXT DEFAULT 'generated',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Servers table (for server metadata)
CREATE TABLE IF NOT EXISTS servers (
    id BIGSERIAL PRIMARY KEY,
    server_id TEXT UNIQUE NOT NULL,
    server_key TEXT UNIQUE,
    hostname TEXT,
    os_info TEXT,
    agent_version TEXT,
    status TEXT DEFAULT 'offline',
    sources TEXT DEFAULT '',           -- TGBot, Web, TGBot,Web etc.
    last_seen TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Dead Letter Queue for failed messages
CREATE TABLE IF NOT EXISTS dead_letter_queue (
    id BIGSERIAL PRIMARY KEY,
    topic TEXT NOT NULL,
    partition INTEGER,
    message_offset BIGINT,
    message JSONB NOT NULL,
    error TEXT NOT NULL,
    attempts INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_generated_keys_server_id ON generated_keys (server_id);
CREATE INDEX IF NOT EXISTS idx_generated_keys_server_key ON generated_keys (server_key);
CREATE INDEX IF NOT EXISTS idx_servers_server_id ON servers (server_id);
CREATE INDEX IF NOT EXISTS idx_servers_server_key ON servers (server_key);
CREATE INDEX IF NOT EXISTS idx_servers_last_seen ON servers (last_seen);
CREATE INDEX IF NOT EXISTS idx_servers_sources ON servers (sources);
CREATE INDEX IF NOT EXISTS idx_dlq_created_at ON dead_letter_queue (created_at);
CREATE INDEX IF NOT EXISTS idx_dlq_topic ON dead_letter_queue (topic);

-- Insert sample data for testing (optional)
-- INSERT INTO generated_keys (server_id, server_key, agent_version, os_info, hostname, status)
-- VALUES ('srv_test123', 'key_test456', '1.0.0', 'Ubuntu 22.04', 'test-server', 'generated')
-- ON CONFLICT (server_id) DO NOTHING;
