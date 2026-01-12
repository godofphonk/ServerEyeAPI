package redis

import (
	"context"
	"fmt"

	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Client wraps Redis client with additional functionality
type Client struct {
	client *redis.Client
	logger *logrus.Logger
	config *config.Config
}

// NewClient creates a new Redis client
func NewClient(addr, password string, db int, logger *logrus.Logger, cfg *config.Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection with configured timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Redis.ConnTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis successfully")

	return &Client{
		client: rdb,
		logger: logger,
		config: cfg,
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
