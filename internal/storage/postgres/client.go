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

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

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
