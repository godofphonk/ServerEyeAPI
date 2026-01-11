package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/godofphonk/ServerEyeAPI/pkg/models"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RedisStorage struct {
	client *redis.Client
	logger *logrus.Logger
}

func NewRedisStorage(redisURL string, logger *logrus.Logger) (*RedisStorage, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("Connected to Redis successfully")

	return &RedisStorage{
		client: client,
		logger: logger,
	}, nil
}

// StoreMetric stores metric data with TTL of 60 seconds
func (r *RedisStorage) StoreMetric(ctx context.Context, serverID string, data map[string]interface{}) error {
	key := fmt.Sprintf("metrics:%s", serverID)

	// Add timestamp to data
	data["timestamp"] = time.Now().Unix()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal metric data: %w", err)
	}

	// Store with 60 seconds TTL
	if err := r.client.Set(ctx, key, jsonData, 60*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to store metric: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"key":       key,
	}).Debug("Metric stored in Redis")

	return nil
}

// GetMetric retrieves latest metric data for a server
func (r *RedisStorage) GetMetric(ctx context.Context, serverID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("metrics:%s", serverID)

	result, err := r.client.Get(ctx, key).Result()
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

	return data, nil
}

// StoreCommand stores command for a server with TTL of 1 hour
func (r *RedisStorage) StoreCommand(ctx context.Context, serverID string, command map[string]interface{}) error {
	key := fmt.Sprintf("commands:%s", serverID)

	// Add timestamp
	command["timestamp"] = time.Now().Unix()

	// Get existing commands
	var commands []map[string]interface{}

	existing, err := r.client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get existing commands: %w", err)
	}

	if existing != "" {
		if err := json.Unmarshal([]byte(existing), &commands); err != nil {
			r.logger.WithError(err).Warn("Failed to unmarshal existing commands, starting fresh")
			commands = []map[string]interface{}{}
		}
	}

	// Add new command
	commands = append(commands, command)

	jsonData, err := json.Marshal(commands)
	if err != nil {
		return fmt.Errorf("failed to marshal commands: %w", err)
	}

	// Store with 1 hour TTL
	if err := r.client.Set(ctx, key, jsonData, time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store commands: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"command":   command,
	}).Debug("Command stored in Redis")

	return nil
}

// GetCommands retrieves pending commands for a server
func (r *RedisStorage) GetCommands(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	key := fmt.Sprintf("commands:%s", serverID)

	result, err := r.client.Get(ctx, key).Result()
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

	return commands, nil
}

// ClearCommands removes all commands for a server
func (r *RedisStorage) ClearCommands(ctx context.Context, serverID string) error {
	key := fmt.Sprintf("commands:%s", serverID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to clear commands: %w", err)
	}

	r.logger.WithField("server_id", serverID).Debug("Commands cleared from Redis")

	return nil
}

// SetServerStatus stores server status with TTL of 5 minutes
func (r *RedisStorage) SetServerStatus(ctx context.Context, serverID string, status map[string]interface{}) error {
	key := fmt.Sprintf("status:%s", serverID)

	// Add last_seen timestamp
	status["last_seen"] = time.Now().Unix()

	jsonData, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal server status: %w", err)
	}

	// Store with 5 minutes TTL
	if err := r.client.Set(ctx, key, jsonData, 5*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to store server status: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"status":    status,
	}).Debug("Server status stored in Redis")

	return nil
}

// GetServerStatus retrieves server status
func (r *RedisStorage) GetServerStatus(ctx context.Context, serverID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("status:%s", serverID)

	result, err := r.client.Get(ctx, key).Result()
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

	return status, nil
}

// GetAllServers returns all servers with status
func (r *RedisStorage) GetAllServers(ctx context.Context) ([]map[string]interface{}, error) {
	pattern := "status:*"

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get server keys: %w", err)
	}

	var servers []map[string]interface{}

	for _, key := range keys {
		serverID := strings.TrimPrefix(key, "status:")

		status, err := r.GetServerStatus(ctx, serverID)
		if err != nil {
			r.logger.WithError(err).WithField("server_id", serverID).Warn("Failed to get server status")
			continue
		}

		servers = append(servers, map[string]interface{}{
			"server_id": serverID,
			"status":    status,
		})
	}

	return servers, nil
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
}

// Implement existing Storage interface methods (stubs for compatibility)

func (r *RedisStorage) StoreMetricLegacy(ctx context.Context, metric *models.Metric) error {
	data := map[string]interface{}{
		"server_id":  metric.ServerID,
		"server_key": metric.ServerKey,
		"type":       metric.Type,
		"value":      metric.Value,
		"tags":       metric.Tags,
		"version":    metric.Version,
	}

	return r.StoreMetric(ctx, metric.ServerID, data)
}

func (r *RedisStorage) GetLatestMetrics(ctx context.Context, serverID string) ([]*models.Metric, error) {
	data, err := r.GetMetric(ctx, serverID)
	if err != nil {
		return []*models.Metric{}, err
	}

	if len(data) == 0 {
		return []*models.Metric{}, nil
	}

	// Convert Redis data to Metric model
	metric := &models.Metric{
		ServerID:  serverID,
		ServerKey: fmt.Sprintf("%v", data["server_key"]),
		Type:      fmt.Sprintf("%v", data["type"]),
		Timestamp: time.Now(),
		Tags:      make(map[string]string),
	}

	if value, ok := data["value"]; ok {
		if v, ok := value.(float64); ok {
			metric.Value = v
		}
	}

	return []*models.Metric{metric}, nil
}

func (r *RedisStorage) GetMetricsHistory(ctx context.Context, serverID string, metricType string, from, to time.Time) ([]*models.Metric, error) {
	r.logger.Warn("GetMetricsHistory not implemented in Redis storage")
	return []*models.Metric{}, nil
}

func (r *RedisStorage) GetServers(ctx context.Context) ([]string, error) {
	servers, err := r.GetAllServers(ctx)
	if err != nil {
		return []string{}, err
	}

	var serverIDs []string
	for _, server := range servers {
		if id, ok := server["server_id"].(string); ok {
			serverIDs = append(serverIDs, id)
		}
	}

	return serverIDs, nil
}

func (r *RedisStorage) GetPendingCommands(ctx context.Context, serverID string) ([]string, error) {
	commands, err := r.GetCommands(ctx, serverID)
	if err != nil {
		return []string{}, err
	}

	var commandStrings []string
	for _, cmd := range commands {
		if cmdJSON, err := json.Marshal(cmd); err == nil {
			commandStrings = append(commandStrings, string(cmdJSON))
		}
	}

	return commandStrings, nil
}

func (r *RedisStorage) StoreDLQMessage(ctx context.Context, topic string, partition int, offset int64, message []byte, errorMsg string) error {
	r.logger.Warn("StoreDLQMessage not implemented in Redis storage")
	return nil
}

func (r *RedisStorage) InsertGeneratedKey(ctx context.Context, secretKey, agentVersion, operatingSystem, hostname string) error {
	r.logger.Warn("InsertGeneratedKey should use PostgreSQL, not Redis")
	return nil
}
