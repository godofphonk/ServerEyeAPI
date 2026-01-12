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

	-- Create servers table for metadata
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

	c.logger.Info("Database schema initialized successfully")
	return nil
}
