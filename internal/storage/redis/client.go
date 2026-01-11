package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Client wraps Redis client with additional functionality
type Client struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewClient creates a new Redis client
func NewClient(addr, password string, db int, logger *logrus.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis successfully")

	return &Client{
		client: rdb,
		logger: logger,
	}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying Redis client
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// Ping checks if Redis is available
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
