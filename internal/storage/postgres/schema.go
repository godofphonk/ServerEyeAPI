// Copyright (c) 2026 godofphonk
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package postgres

import (
	"context"
	"fmt"
	"time"
)

// initSchema initializes the database schema
func (c *Client) initSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create essential tables only
	schema := `
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

	CREATE INDEX IF NOT EXISTS idx_generated_keys_server_id ON generated_keys (server_id);
	CREATE INDEX IF NOT EXISTS idx_generated_keys_server_key ON generated_keys (server_key);

	-- Create servers table for metadata (basic version for migration)
	CREATE TABLE IF NOT EXISTS servers (
		id BIGSERIAL PRIMARY KEY,
		server_id TEXT UNIQUE NOT NULL,
		hostname TEXT,
		os_info TEXT,
		agent_version TEXT,
		status TEXT DEFAULT 'offline',
		last_seen TIMESTAMPTZ DEFAULT NOW(),
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_servers_server_id ON servers (server_id);
	CREATE INDEX IF NOT EXISTS idx_servers_last_seen ON servers (last_seen);

	-- Create dead letter queue for failed messages
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

	CREATE INDEX IF NOT EXISTS idx_dlq_created_at ON dead_letter_queue (created_at);
	CREATE INDEX IF NOT EXISTS idx_dlq_topic ON dead_letter_queue (topic);
	`

	_, err := c.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	// Always run migration to ensure columns exist
	migration := `
	-- Add server_key column to servers table if it doesn't exist
	DO $$
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name='servers' AND column_name='server_key'
		) THEN
			ALTER TABLE servers ADD COLUMN server_key TEXT UNIQUE;
		END IF;
		
		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name='servers' AND column_name='updated_at'
		) THEN
			ALTER TABLE servers ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
		END IF;
	END $$;
	`

	_, err = c.db.ExecContext(ctx, migration)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	c.logger.Info("Database schema initialized successfully")
	return nil
}
