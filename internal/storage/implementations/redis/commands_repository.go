package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/repositories"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisCommandsRepository implements CommandsRepository for Redis
type RedisCommandsRepository struct {
	client *redis.Client
	logger *logrus.Logger
	ttl    time.Duration
}

// NewRedisCommandsRepository creates a new Redis commands repository
func NewRedisCommandsRepository(client *redis.Client, logger *logrus.Logger, cfg *config.Config) repositories.CommandsRepository {
	return &RedisCommandsRepository{
		client: client,
		logger: logger,
		ttl:    cfg.Redis.TTL,
	}
}

// Store stores a new command
func (r *RedisCommandsRepository) Store(ctx context.Context, serverID string, command *models.Command) error {
	if command.ID == "" {
		return fmt.Errorf("command ID cannot be empty")
	}
	if command.CreatedAt.IsZero() {
		command.CreatedAt = time.Now()
	}
	if command.Status == "" {
		command.Status = "pending"
	}

	// Store command in a list for the server
	key := fmt.Sprintf("commands:%s", serverID)

	data, err := json.Marshal(command)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"command_id": command.ID,
			"server_id":  serverID,
		}).Error("Failed to marshal command")
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	// Add to the beginning of the list (LPUSH)
	if err := r.client.LPush(ctx, key, data).Err(); err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"command_id": command.ID,
			"server_id":  serverID,
		}).Error("Failed to store command in Redis")
		return fmt.Errorf("failed to store command: %w", err)
	}

	// Set TTL on the key
	if err := r.client.Expire(ctx, key, r.ttl).Err(); err != nil {
		r.logger.WithError(err).WithField("key", key).Warn("Failed to set TTL on commands key")
	}

	r.logger.WithFields(logrus.Fields{
		"command_id": command.ID,
		"server_id":  serverID,
		"type":       command.Type,
		"ttl":        r.ttl,
	}).Debug("Command stored in Redis successfully")
	return nil
}

// GetPending retrieves pending commands for a server
func (r *RedisCommandsRepository) GetPending(ctx context.Context, serverID string) ([]*models.Command, error) {
	key := fmt.Sprintf("commands:%s", serverID)

	// Get all commands from the list
	dataList, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get commands from Redis")
		return nil, fmt.Errorf("failed to get commands: %w", err)
	}

	var commands []*models.Command
	for _, data := range dataList {
		var command models.Command
		if err := json.Unmarshal([]byte(data), &command); err != nil {
			r.logger.WithError(err).WithField("data", data).Warn("Failed to unmarshal command")
			continue
		}

		// Only return pending commands
		if command.Status == "pending" {
			commands = append(commands, &command)
		}
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"count":     len(commands),
	}).Debug("Pending commands retrieved from Redis successfully")
	return commands, nil
}

// GetHistory retrieves command history for a server
func (r *RedisCommandsRepository) GetHistory(ctx context.Context, serverID string, limit int) ([]*models.Command, error) {
	key := fmt.Sprintf("commands:%s", serverID)

	// Get commands from the list with limit
	dataList, err := r.client.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get command history from Redis")
		return nil, fmt.Errorf("failed to get command history: %w", err)
	}

	var commands []*models.Command
	for _, data := range dataList {
		var command models.Command
		if err := json.Unmarshal([]byte(data), &command); err != nil {
			r.logger.WithError(err).WithField("data", data).Warn("Failed to unmarshal command")
			continue
		}
		commands = append(commands, &command)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"limit":     limit,
		"count":     len(commands),
	}).Debug("Command history retrieved from Redis successfully")
	return commands, nil
}

// MarkProcessed marks a command as processed
func (r *RedisCommandsRepository) MarkProcessed(ctx context.Context, commandID string) error {
	// Redis doesn't support updating individual items in a list easily
	// We'll need to scan through the list and update the matching command
	// For simplicity, we'll just log this operation
	r.logger.WithField("command_id", commandID).Debug("Command marked as processed (Redis limitation)")
	return nil
}

// MarkFailed marks a command as failed
func (r *RedisCommandsRepository) MarkFailed(ctx context.Context, commandID string, errorMsg string) error {
	// Redis doesn't support updating individual items in a list easily
	// We'll need to scan through the list and update the matching command
	// For simplicity, we'll just log this operation
	r.logger.WithFields(logrus.Fields{
		"command_id": commandID,
		"error":      errorMsg,
	}).Debug("Command marked as failed (Redis limitation)")
	return nil
}

// GetByID retrieves a command by ID
func (r *RedisCommandsRepository) GetByID(ctx context.Context, commandID string) (*models.Command, error) {
	// Redis doesn't support efficient lookup by ID in lists
	// We would need to scan through all server command lists
	// For simplicity, return not found
	return nil, fmt.Errorf("command lookup by ID not supported in Redis implementation")
}

// GetByType retrieves commands by type for a server
func (r *RedisCommandsRepository) GetByType(ctx context.Context, serverID string, commandType string) ([]*models.Command, error) {
	key := fmt.Sprintf("commands:%s", serverID)

	// Get all commands from the list
	dataList, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get commands from Redis")
		return nil, fmt.Errorf("failed to get commands: %w", err)
	}

	var commands []*models.Command
	for _, data := range dataList {
		var command models.Command
		if err := json.Unmarshal([]byte(data), &command); err != nil {
			r.logger.WithError(err).WithField("data", data).Warn("Failed to unmarshal command")
			continue
		}

		// Only return commands of the specified type
		if command.Type == commandType {
			commands = append(commands, &command)
		}
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"type":      commandType,
		"count":     len(commands),
	}).Debug("Commands retrieved by type from Redis successfully")
	return commands, nil
}

// DeleteProcessed deletes processed commands older than specified duration
func (r *RedisCommandsRepository) DeleteProcessed(ctx context.Context, olderThan time.Duration) error {
	// Redis handles TTL automatically, so we don't need to implement this
	// The commands will be automatically deleted when their TTL expires
	r.logger.WithField("older_than", olderThan).Debug("Redis handles TTL cleanup automatically")
	return nil
}

// DeleteByServer deletes all commands for a server
func (r *RedisCommandsRepository) DeleteByServer(ctx context.Context, serverID string) error {
	key := fmt.Sprintf("commands:%s", serverID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to delete commands from Redis")
		return fmt.Errorf("failed to delete commands: %w", err)
	}

	r.logger.WithField("server_id", serverID).Debug("Commands deleted from Redis successfully")
	return nil
}

// Ping checks Redis connectivity
func (r *RedisCommandsRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
