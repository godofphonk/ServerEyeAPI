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

	"github.com/redis/go-redis/v9"
)

// StoreDLQ stores a message in the dead letter queue
func (c *Client) StoreDLQ(ctx context.Context, topic, message string, metadata map[string]interface{}) error {
	key := fmt.Sprintf("dlq:%s", topic)

	dlqMessage := map[string]interface{}{
		"message":   message,
		"metadata":  metadata,
		"timestamp": time.Now().Unix(),
	}

	jsonData, err := json.Marshal(dlqMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ message: %w", err)
	}

	// Store with 24 hours TTL
	if err := c.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: jsonData,
	}).Err(); err != nil {
		return fmt.Errorf("failed to store DLQ message: %w", err)
	}

	// Set TTL
	if err := c.client.Expire(ctx, key, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set DLQ TTL: %w", err)
	}

	c.logger.WithField("topic", topic).Debug("DLQ message stored in Redis")
	return nil
}

// GetDLQ retrieves DLQ messages for a topic
func (c *Client) GetDLQ(ctx context.Context, topic string, limit int) ([]map[string]interface{}, error) {
	key := fmt.Sprintf("dlq:%s", topic)

	// Get messages by score (timestamp) in descending order
	results, err := c.client.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		if err == redis.Nil {
			return []map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("failed to get DLQ messages: %w", err)
	}

	var messages []map[string]interface{}
	for _, result := range results {
		var message map[string]interface{}
		if err := json.Unmarshal([]byte(result.Member.(string)), &message); err != nil {
			c.logger.WithError(err).Warn("Failed to unmarshal DLQ message")
			continue
		}
		messages = append(messages, message)
	}

	return messages, nil
}
