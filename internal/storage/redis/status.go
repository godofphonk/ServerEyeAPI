package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// SetServerStatus stores server status with TTL of 5 minutes
func (c *Client) SetServerStatus(ctx context.Context, serverID string, status map[string]interface{}) error {
	key := fmt.Sprintf("status:%s", serverID)

	// Add last_seen timestamp
	status["last_seen"] = time.Now().Unix()

	jsonData, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal server status: %w", err)
	}

	// Store with 5 minutes TTL
	if err := c.client.Set(ctx, key, jsonData, 5*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to store server status: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"status":    status,
	}).Debug("Server status stored in Redis")

	return nil
}

// GetServerStatus retrieves server status
func (c *Client) GetServerStatus(ctx context.Context, serverID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("status:%s", serverID)

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return map[string]interface{}{"online": false}, nil
		}
		return nil, fmt.Errorf("failed to get server status: %w", err)
	}

	var status map[string]interface{}
	if err := json.Unmarshal([]byte(result), &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal server status: %w", err)
	}

	c.logger.WithField("server_id", serverID).Debug("Server status retrieved from Redis")
	return status, nil
}

// GetAllServers returns all servers with status
func (c *Client) GetAllServers(ctx context.Context) ([]map[string]interface{}, error) {
	pattern := "status:*"

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get status keys: %w", err)
	}

	var servers []map[string]interface{}

	for _, key := range keys {
		serverID := key[len("status:"):]

		status, err := c.GetServerStatus(ctx, serverID)
		if err != nil {
			c.logger.WithError(err).WithField("server_id", serverID).Warn("Failed to get server status")
			continue
		}

		servers = append(servers, map[string]interface{}{
			"server_id": serverID,
			"status":    status,
		})
	}

	return servers, nil
}
