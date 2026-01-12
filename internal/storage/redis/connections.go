package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// StoreConnection stores connection info for a server
func (c *Client) StoreConnection(ctx context.Context, serverID string, connectionInfo map[string]interface{}) error {
	key := fmt.Sprintf("connections:%s", serverID)

	// Add timestamp to connection info
	connectionInfo["timestamp"] = time.Now().Unix()

	jsonData, err := json.Marshal(connectionInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal connection info: %w", err)
	}

	// Store with 1 hour TTL
	if err := c.client.Set(ctx, key, jsonData, time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store connection info: %w", err)
	}

	c.logger.WithField("server_id", serverID).Debug("Connection info stored in Redis")
	return nil
}

// GetConnections retrieves connection info for a server
func (c *Client) GetConnections(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	key := fmt.Sprintf("connections:%s", serverID)

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("failed to get connection info: %w", err)
	}

	var connectionInfo map[string]interface{}
	if err := json.Unmarshal([]byte(result), &connectionInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal connection info: %w", err)
	}

	return []map[string]interface{}{connectionInfo}, nil
}
