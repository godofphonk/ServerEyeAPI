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

// RedisServerRepository implements ServerRepository for Redis (status management only)
type RedisServerRepository struct {
	client *redis.Client
	logger *logrus.Logger
	ttl    time.Duration
}

// NewRedisServerRepository creates a new Redis server repository
func NewRedisServerRepository(client *redis.Client, logger *logrus.Logger, cfg *config.Config) repositories.ServerRepository {
	return &RedisServerRepository{
		client: client,
		logger: logger,
		ttl:    cfg.Redis.TTL,
	}
}

// Create creates a new server (not implemented - use PostgreSQL)
func (r *RedisServerRepository) Create(ctx context.Context, server *models.Server) error {
	return fmt.Errorf("server creation not supported in Redis implementation - use PostgreSQL")
}

// GetByID retrieves a server by ID (not implemented - use PostgreSQL)
func (r *RedisServerRepository) GetByID(ctx context.Context, id string) (*models.Server, error) {
	return nil, fmt.Errorf("server retrieval by ID not supported in Redis implementation - use PostgreSQL")
}

// GetByKey retrieves a server by server key (not implemented - use PostgreSQL)
func (r *RedisServerRepository) GetByKey(ctx context.Context, serverKey string) (*models.Server, error) {
	return nil, fmt.Errorf("server retrieval by key not supported in Redis implementation - use PostgreSQL")
}

// Update updates a server (not implemented - use PostgreSQL)
func (r *RedisServerRepository) Update(ctx context.Context, server *models.Server) error {
	return fmt.Errorf("server update not supported in Redis implementation - use PostgreSQL")
}

// Delete deletes a server (not implemented - use PostgreSQL)
func (r *RedisServerRepository) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("server deletion not supported in Redis implementation - use PostgreSQL")
}

// List retrieves all servers (not implemented - use PostgreSQL)
func (r *RedisServerRepository) List(ctx context.Context) ([]*models.Server, error) {
	return nil, fmt.Errorf("server listing not supported in Redis implementation - use PostgreSQL")
}

// ListByStatus retrieves servers by status (not implemented - use PostgreSQL)
func (r *RedisServerRepository) ListByStatus(ctx context.Context, status string) ([]*models.Server, error) {
	return nil, fmt.Errorf("server listing by status not supported in Redis implementation - use PostgreSQL")
}

// ListByHostname retrieves servers by hostname (not implemented - use PostgreSQL)
func (r *RedisServerRepository) ListByHostname(ctx context.Context, hostname string) ([]*models.Server, error) {
	return nil, fmt.Errorf("server listing by hostname not supported in Redis implementation - use PostgreSQL")
}

// UpdateStatus updates server status
func (r *RedisServerRepository) UpdateStatus(ctx context.Context, serverID string, status string) error {
	key := fmt.Sprintf("server:status:%s", serverID)

	// Create status data
	statusData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	data, err := json.Marshal(statusData)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"server_id": serverID,
			"status":    status,
		}).Error("Failed to marshal server status")
		return fmt.Errorf("failed to marshal server status: %w", err)
	}

	// Store status with TTL
	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"server_id": serverID,
			"status":    status,
		}).Error("Failed to store server status in Redis")
		return fmt.Errorf("failed to store server status: %w", err)
	}

	// Also add to status sets for quick lookup
	statusKey := fmt.Sprintf("servers:status:%s", status)
	if err := r.client.SAdd(ctx, statusKey, serverID).Err(); err != nil {
		r.logger.WithError(err).WithField("status", status).Warn("Failed to add server to status set")
	}
	// Set TTL on status set
	if err := r.client.Expire(ctx, statusKey, r.ttl).Err(); err != nil {
		r.logger.WithError(err).WithField("status", status).Warn("Failed to set TTL on status set")
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"status":    status,
		"ttl":       r.ttl,
	}).Debug("Server status updated in Redis successfully")
	return nil
}

// UpdateLastSeen updates server last seen timestamp
func (r *RedisServerRepository) UpdateLastSeen(ctx context.Context, serverID string, lastSeen time.Time) error {
	key := fmt.Sprintf("server:last_seen:%s", serverID)

	// Create last seen data
	lastSeenData := map[string]interface{}{
		"last_seen":  lastSeen,
		"updated_at": time.Now(),
	}

	data, err := json.Marshal(lastSeenData)
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to marshal last seen data")
		return fmt.Errorf("failed to marshal last seen data: %w", err)
	}

	// Store last seen with TTL
	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to store last seen in Redis")
		return fmt.Errorf("failed to store last seen: %w", err)
	}

	// Also add to active servers set
	if err := r.client.SAdd(ctx, "servers:active", serverID).Err(); err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Warn("Failed to add server to active set")
	}
	// Set TTL on active set
	if err := r.client.Expire(ctx, "servers:active", r.ttl).Err(); err != nil {
		r.logger.WithError(err).Warn("Failed to set TTL on active servers set")
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"last_seen": lastSeen,
		"ttl":       r.ttl,
	}).Debug("Server last seen updated in Redis successfully")
	return nil
}

// Ping checks Redis connectivity
func (r *RedisServerRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
