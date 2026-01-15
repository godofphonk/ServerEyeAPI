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
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Client wraps PostgreSQL client with additional functionality
type Client struct {
	db     *sql.DB
	logger *logrus.Logger
}

// DB returns the underlying database connection
func (c *Client) DB() *sql.DB {
	return c.db
}

// NewClient creates a new PostgreSQL client
func NewClient(databaseURL string, logger *logrus.Logger) (*Client, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for production
	db.SetMaxOpenConns(50)                  // Increase for higher concurrency
	db.SetMaxIdleConns(10)                  // Keep more idle connections ready
	db.SetConnMaxLifetime(30 * time.Minute) // Allow longer connection reuse
	db.SetConnMaxIdleTime(5 * time.Minute)  // Close idle connections faster

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL successfully")

	client := &Client{
		db:     db,
		logger: logger,
	}

	// Initialize schema
	if err := client.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return client, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// GetDB returns the underlying database connection
func (c *Client) GetDB() *sql.DB {
	return c.db
}

// Ping checks if database is available
func (c *Client) Ping() error {
	return c.db.Ping()
}
