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

// RedisMetricsRepository implements MetricsRepository for Redis
type RedisMetricsRepository struct {
	client *redis.Client
	logger *logrus.Logger
	ttl    time.Duration
}

// NewRedisMetricsRepository creates a new Redis metrics repository
func NewRedisMetricsRepository(client *redis.Client, logger *logrus.Logger, cfg *config.Config) repositories.MetricsRepository {
	return &RedisMetricsRepository{
		client: client,
		logger: logger,
		ttl:    cfg.Redis.TTL,
	}
}

// Store stores server metrics with TTL
func (r *RedisMetricsRepository) Store(ctx context.Context, serverID string, metrics *models.ServerMetrics) error {
	key := fmt.Sprintf("metrics:%s", serverID)

	data, err := json.Marshal(metrics)
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to marshal metrics")
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Store with configured TTL
	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to store metrics in Redis")
		return fmt.Errorf("failed to store metrics: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"timestamp": metrics.Time,
		"ttl":       r.ttl,
	}).Debug("Metrics stored in Redis successfully")
	return nil
}

// Get retrieves metrics for a server
func (r *RedisMetricsRepository) Get(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	key := fmt.Sprintf("metrics:%s", serverID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("metrics not found for server: %s", serverID)
		}
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get metrics from Redis")
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	var metrics models.ServerMetrics
	if err := json.Unmarshal([]byte(data), &metrics); err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to unmarshal metrics")
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	r.logger.WithField("server_id", serverID).Debug("Metrics retrieved from Redis successfully")
	return &metrics, nil
}

// Delete deletes metrics for a server
func (r *RedisMetricsRepository) Delete(ctx context.Context, serverID string) error {
	key := fmt.Sprintf("metrics:%s", serverID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to delete metrics from Redis")
		return fmt.Errorf("failed to delete metrics: %w", err)
	}

	r.logger.WithField("server_id", serverID).Debug("Metrics deleted from Redis successfully")
	return nil
}

// GetLatest retrieves latest metrics for multiple servers
func (r *RedisMetricsRepository) GetLatest(ctx context.Context, serverID string, limit int) ([]*models.ServerMetrics, error) {
	// Redis doesn't support time-based queries easily, so we'll get the latest
	// and return it as a single item list
	metrics, err := r.Get(ctx, serverID)
	if err != nil {
		return nil, err
	}

	return []*models.ServerMetrics{metrics}, nil
}

// GetAll retrieves all latest metrics
func (r *RedisMetricsRepository) GetAll(ctx context.Context) (map[string]*models.ServerMetrics, error) {
	// Get all metrics keys
	pattern := "metrics:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get metrics keys from Redis")
		return nil, fmt.Errorf("failed to get metrics keys: %w", err)
	}

	metricsMap := make(map[string]*models.ServerMetrics)

	// Get metrics for each key
	for _, key := range keys {
		serverID := key[len("metrics:"):]

		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			if err != redis.Nil {
				r.logger.WithError(err).WithField("key", key).Warn("Failed to get metrics data")
			}
			continue
		}

		var metrics models.ServerMetrics
		if err := json.Unmarshal([]byte(data), &metrics); err != nil {
			r.logger.WithError(err).WithField("key", key).Warn("Failed to unmarshal metrics")
			continue
		}

		metricsMap[serverID] = &metrics
	}

	r.logger.WithField("count", len(metricsMap)).Debug("All metrics retrieved from Redis successfully")
	return metricsMap, nil
}

// GetByTimeRange retrieves metrics for a server within a time range
func (r *RedisMetricsRepository) GetByTimeRange(ctx context.Context, serverID string, start, end time.Time) ([]*models.ServerMetrics, error) {
	// Redis doesn't support time-based queries on stored data
	// We'll get the latest metrics and check if it's within the range
	metrics, err := r.Get(ctx, serverID)
	if err != nil {
		return nil, err
	}

	// Check if metrics timestamp is within the requested range
	if metrics.Time.Before(start) || metrics.Time.After(end) {
		return []*models.ServerMetrics{}, nil
	}

	return []*models.ServerMetrics{metrics}, nil
}

// DeleteOlderThan deletes metrics older than specified duration
func (r *RedisMetricsRepository) DeleteOlderThan(ctx context.Context, olderThan time.Duration) error {
	// Redis handles TTL automatically, so we don't need to implement this
	// The metrics will be automatically deleted when their TTL expires
	r.logger.WithField("older_than", olderThan).Debug("Redis handles TTL cleanup automatically")
	return nil
}

// DeleteByTimeRange deletes metrics within a time range for a server
func (r *RedisMetricsRepository) DeleteByTimeRange(ctx context.Context, serverID string, start, end time.Time) error {
	// Get current metrics
	metrics, err := r.Get(ctx, serverID)
	if err != nil {
		// If metrics don't exist, that's fine
		if err.Error() == fmt.Sprintf("metrics not found for server: %s", serverID) {
			return nil
		}
		return err
	}

	// Delete if within time range
	if metrics.Time.After(start) && metrics.Time.Before(end) {
		return r.Delete(ctx, serverID)
	}

	return nil
}

// Ping checks Redis connectivity
func (r *RedisMetricsRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
