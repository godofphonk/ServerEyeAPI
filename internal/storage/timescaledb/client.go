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

package timescaledb

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// Client represents a TimescaleDB client
type Client struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
	config *ClientConfig
}

// ClientConfig holds TimescaleDB client configuration
type ClientConfig struct {
	MaxConnections      int           `json:"max_connections"`
	ConnTimeout         time.Duration `json:"conn_timeout"`
	QueryTimeout        time.Duration `json:"query_timeout"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

// DefaultClientConfig returns default configuration for TimescaleDB client
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		MaxConnections:      20,
		ConnTimeout:         30 * time.Second,
		QueryTimeout:        10 * time.Second,
		HealthCheckInterval: 30 * time.Second,
	}
}

// NewClient creates a new TimescaleDB client
func NewClient(databaseURL string, logger *logrus.Logger, config *ClientConfig) (*Client, error) {
	if config == nil {
		config = DefaultClientConfig()
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnTimeout)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(config.MaxConnections)
	poolConfig.HealthCheckPeriod = config.HealthCheckInterval
	poolConfig.MinConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	client := &Client{
		pool:   pool,
		logger: logger,
		config: config,
	}

	// Test connection
	if err := client.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to TimescaleDB: %w", err)
	}

	logger.Info("Successfully connected to TimescaleDB")
	return client, nil
}

// Ping checks database connectivity
func (c *Client) Ping(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	var result string
	err := c.pool.QueryRow(ctx, "SELECT 'pong'").Scan(&result)
	if err != nil {
		return fmt.Errorf("TimescaleDB ping failed: %w", err)
	}

	if result != "pong" {
		return fmt.Errorf("unexpected ping response: %s", result)
	}

	c.logger.Debug("TimescaleDB ping successful")
	return nil
}

// Close closes the database connection pool
func (c *Client) Close() error {
	if c.pool != nil {
		c.pool.Close()
		c.logger.Info("TimescaleDB connection pool closed")
	}
	return nil
}

// GetPool returns the underlying connection pool
func (c *Client) GetPool() *pgxpool.Pool {
	return c.pool
}

// GetStats returns connection pool statistics
func (c *Client) GetStats() *pgxpool.Stat {
	stats := c.pool.Stat()
	return stats
}

// Health performs comprehensive health check
func (c *Client) Health(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Check basic connectivity
	if err := c.Ping(ctx); err != nil {
		return fmt.Errorf("basic connectivity check failed: %w", err)
	}

	// Check TimescaleDB extension
	var extensionExists bool
	err := c.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'timescaledb')").
		Scan(&extensionExists)
	if err != nil {
		return fmt.Errorf("failed to check TimescaleDB extension: %w", err)
	}

	if !extensionExists {
		return fmt.Errorf("TimescaleDB extension not found")
	}

	// Check hypertables
	var hypertableCount int
	err = c.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM timescaledb_information.hypertables").
		Scan(&hypertableCount)
	if err != nil {
		return fmt.Errorf("failed to check hypertables: %w", err)
	}

	if hypertableCount == 0 {
		return fmt.Errorf("no hypertables found")
	}

	c.logger.WithFields(logrus.Fields{
		"hypertables": hypertableCount,
	}).Info("TimescaleDB health check passed")

	return nil
}

// ExecuteQuery executes a query and returns the results
func (c *Client) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return rows, nil
}

// ExecuteQueryRow executes a query that returns a single row
func (c *Client) ExecuteQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	return c.pool.QueryRow(ctx, query, args...)
}

// ExecuteCommand executes a command that doesn't return results
func (c *Client) ExecuteCommand(ctx context.Context, command string, args ...interface{}) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	_, err := c.pool.Exec(ctx, command, args...)
	if err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// BeginTransaction starts a new transaction
func (c *Client) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return tx, nil
}

// GetServerTime returns the current database server time
func (c *Client) GetServerTime(ctx context.Context) (time.Time, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var serverTime time.Time
	err := c.pool.QueryRow(ctx, "SELECT NOW()").Scan(&serverTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get server time: %w", err)
	}

	return serverTime, nil
}

// GetVersion returns TimescaleDB version information
func (c *Client) GetVersion(ctx context.Context) (map[string]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	version := make(map[string]string)

	// Get PostgreSQL version
	var pgVersion string
	err := c.pool.QueryRow(ctx, "SELECT version()").Scan(&pgVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get PostgreSQL version: %w", err)
	}
	version["postgresql"] = pgVersion

	// Get TimescaleDB version
	var tsVersion string
	err = c.pool.QueryRow(ctx, "SELECT extversion FROM pg_extension WHERE extname = 'timescaledb'").Scan(&tsVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get TimescaleDB version: %w", err)
	}
	version["timescaledb"] = tsVersion

	return version, nil
}

// GetHypertables returns information about all hypertables
func (c *Client) GetHypertables(ctx context.Context) ([]HypertableInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		SELECT 
			hypertable_name,
			associated_schema_name,
			associated_table_name,
			num_chunks,
			Chunk_storage_size
		FROM timescaledb_information.hypertables
		ORDER BY hypertable_name
	`

	rows, err := c.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get hypertables: %w", err)
	}
	defer rows.Close()

	var hypertables []HypertableInfo
	for rows.Next() {
		var ht HypertableInfo
		if err := rows.Scan(
			&ht.Name,
			&ht.Schema,
			&ht.TableName,
			&ht.NumChunks,
			&ht.StorageSize,
		); err != nil {
			return nil, fmt.Errorf("failed to scan hypertable row: %w", err)
		}
		hypertables = append(hypertables, ht)
	}

	return hypertables, nil
}

// HypertableInfo contains information about a hypertable
type HypertableInfo struct {
	Name        string `json:"name"`
	Schema      string `json:"schema"`
	TableName   string `json:"table_name"`
	NumChunks   int    `json:"num_chunks"`
	StorageSize int64  `json:"storage_size"`
}
