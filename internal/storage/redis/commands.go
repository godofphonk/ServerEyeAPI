package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// StoreCommand stores command for a server with TTL of 1 hour
func (c *Client) StoreCommand(ctx context.Context, serverID string, command map[string]interface{}) error {
	key := fmt.Sprintf("commands:%s", serverID)

	// Add timestamp
	command["timestamp"] = time.Now().Unix()

	// Get existing commands
	var commands []map[string]interface{}

	existing, err := c.client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get existing commands: %w", err)
	}

	if existing != "" {
		if err := json.Unmarshal([]byte(existing), &commands); err != nil {
			return fmt.Errorf("failed to unmarshal existing commands: %w", err)
		}
	}

	// Add new command
	commands = append(commands, command)

	// Limit to last 100 commands
	if len(commands) > 100 {
		commands = commands[len(commands)-100:]
	}

	jsonData, err := json.Marshal(commands)
	if err != nil {
		return fmt.Errorf("failed to marshal commands: %w", err)
	}

	// Store with 1 hour TTL
	if err := c.client.Set(ctx, key, jsonData, time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store commands: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"command":   command,
	}).Debug("Command stored in Redis")

	return nil
}

// GetCommands retrieves commands for a server
func (c *Client) GetCommands(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	key := fmt.Sprintf("commands:%s", serverID)

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("failed to get commands: %w", err)
	}

	var commands []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &commands); err != nil {
		return nil, fmt.Errorf("failed to unmarshal commands: %w", err)
	}

	c.logger.WithField("server_id", serverID).Debug("Commands retrieved from Redis")
	return commands, nil
}

// GetPendingCommands retrieves pending commands for a server
func (c *Client) GetPendingCommands(ctx context.Context, serverID string) ([]string, error) {
	commands, err := c.GetCommands(ctx, serverID)
	if err != nil {
		return nil, err
	}

	var pending []string
	for _, cmd := range commands {
		if status, ok := cmd["status"].(string); ok && status == "pending" {
			cmdStr, _ := json.Marshal(cmd)
			pending = append(pending, string(cmdStr))
		}
	}

	return pending, nil
}
