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
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// SetServerStatus stores server status with TTL of 5 minutes
func (c *Client) SetServerStatus(ctx context.Context, serverID string, status *models.ServerStatus) error {
	key := fmt.Sprintf("status:%s", serverID)

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
		"online":    status.Online,
	}).Debug("Server status stored in Redis")

	return nil
}

// GetServerStatus retrieves server status
func (c *Client) GetServerStatus(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	key := fmt.Sprintf("status:%s", serverID)

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return &models.ServerStatus{
				Online:   false,
				LastSeen: time.Time{},
			}, nil
		}
		return nil, fmt.Errorf("failed to get server status: %w", err)
	}

	var status models.ServerStatus
	if err := json.Unmarshal([]byte(result), &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal server status: %w", err)
	}

	c.logger.WithField("server_id", serverID).Debug("Server status retrieved from Redis")
	return &status, nil
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
