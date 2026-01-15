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

package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/redis/go-redis/v9"
)

// StoreMetric stores metric data with configured TTL
func (c *Client) StoreMetric(ctx context.Context, serverID string, metrics *models.ServerMetrics) error {
	key := fmt.Sprintf("metrics:%s", serverID)

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metric data: %w", err)
	}

	// Store with configured TTL
	if err := c.client.Set(ctx, key, jsonData, c.config.Redis.TTL).Err(); err != nil {
		return fmt.Errorf("failed to store metric: %w", err)
	}

	c.logger.WithField("server_id", serverID).Debug("Metric stored in Redis")
	return nil
}

// GetMetric retrieves metric data for a server
func (c *Client) GetMetric(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	key := fmt.Sprintf("metrics:%s", serverID)

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("metrics not found")
		}
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	var metrics models.ServerMetrics
	if err := json.Unmarshal([]byte(result), &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metric data: %w", err)
	}

	c.logger.WithField("server_id", serverID).Debug("Metric retrieved from Redis")
	return &metrics, nil
}

// GetAllMetrics retrieves metrics for all servers
func (c *Client) GetAllMetrics(ctx context.Context) (map[string]*models.ServerMetrics, error) {
	pattern := "metrics:*"

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get metric keys: %w", err)
	}

	metrics := make(map[string]*models.ServerMetrics)

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
