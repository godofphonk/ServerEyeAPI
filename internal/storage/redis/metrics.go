package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// StoreMetric stores metric data with TTL of 60 seconds
func (c *Client) StoreMetric(ctx context.Context, serverID string, data map[string]interface{}) error {
	key := fmt.Sprintf("metrics:%s", serverID)

	// Add timestamp to data
	data["timestamp"] = time.Now().Unix()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal metric data: %w", err)
	}

	// Store with 60 seconds TTL
	if err := c.client.Set(ctx, key, jsonData, 60*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to store metric: %w", err)
	}

	c.logger.WithField("server_id", serverID).Debug("Metric stored in Redis")
	return nil
}

// GetMetric retrieves metric data for a server
func (c *Client) GetMetric(ctx context.Context, serverID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("metrics:%s", serverID)

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metric data: %w", err)
	}

	c.logger.WithField("server_id", serverID).Debug("Metric retrieved from Redis")
	return data, nil
}

// GetAllMetrics retrieves metrics for all servers
func (c *Client) GetAllMetrics(ctx context.Context) (map[string]map[string]interface{}, error) {
	pattern := "metrics:*"

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get metric keys: %w", err)
	}

	metrics := make(map[string]map[string]interface{})

	for _, key := range keys {
		serverID := key[len("metrics:"):]

		data, err := c.GetMetric(ctx, serverID)
		if err != nil {
			c.logger.WithError(err).WithField("server_id", serverID).Warn("Failed to get server metrics")
			continue
		}

		metrics[serverID] = data
	}

	return metrics, nil
}
