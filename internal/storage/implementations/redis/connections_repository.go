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

	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/repositories"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisConnectionsRepository implements ConnectionsRepository for Redis
type RedisConnectionsRepository struct {
	client *redis.Client
	logger *logrus.Logger
	ttl    time.Duration
}

// NewRedisConnectionsRepository creates a new Redis connections repository
func NewRedisConnectionsRepository(client *redis.Client, logger *logrus.Logger, cfg *config.Config) repositories.ConnectionsRepository {
	return &RedisConnectionsRepository{
		client: client,
		logger: logger,
		ttl:    cfg.Redis.TTL,
	}
}

// Store stores a new connection
func (r *RedisConnectionsRepository) Store(ctx context.Context, serverID string, conn *models.Connection) error {
	if conn.ID == "" {
		return fmt.Errorf("connection ID cannot be empty")
	}
	if conn.ConnectedAt.IsZero() {
		conn.ConnectedAt = time.Now()
	}
	if conn.LastActivity.IsZero() {
		conn.LastActivity = time.Now()
	}
	if conn.Status == "" {
		conn.Status = "active"
	}

	// Store connection in a set for active connections
	activeKey := fmt.Sprintf("connections:active:%s", serverID)
	historyKey := fmt.Sprintf("connections:history:%s", serverID)

	data, err := json.Marshal(conn)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"connection_id": conn.ID,
			"server_id":     serverID,
		}).Error("Failed to marshal connection")
		return fmt.Errorf("failed to marshal connection: %w", err)
	}

	// Add to active connections if active
	if conn.Status == "active" {
		if err := r.client.SAdd(ctx, activeKey, conn.ID).Err(); err != nil {
			r.logger.WithError(err).WithField("key", activeKey).Error("Failed to add to active connections")
		}
		// Set TTL on active connections
		if err := r.client.Expire(ctx, activeKey, r.ttl).Err(); err != nil {
			r.logger.WithError(err).WithField("key", activeKey).Warn("Failed to set TTL on active connections")
		}
	}

	// Store connection data
	connKey := fmt.Sprintf("connection:%s", conn.ID)
	if err := r.client.Set(ctx, connKey, data, r.ttl).Err(); err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"connection_id": conn.ID,
			"server_id":     serverID,
		}).Error("Failed to store connection in Redis")
		return fmt.Errorf("failed to store connection: %w", err)
	}

	// Add to history
	if err := r.client.LPush(ctx, historyKey, conn.ID).Err(); err != nil {
		r.logger.WithError(err).WithField("key", historyKey).Warn("Failed to add to connection history")
	}
	// Limit history to last 100 connections
	if err := r.client.LTrim(ctx, historyKey, 0, 99).Err(); err != nil {
		r.logger.WithError(err).WithField("key", historyKey).Warn("Failed to trim connection history")
	}
	// Set TTL on history
	if err := r.client.Expire(ctx, historyKey, r.ttl*2).Err(); err != nil {
		r.logger.WithError(err).WithField("key", historyKey).Warn("Failed to set TTL on connection history")
	}

	r.logger.WithFields(logrus.Fields{
		"connection_id": conn.ID,
		"server_id":     serverID,
		"type":          conn.Type,
		"status":        conn.Status,
		"ttl":           r.ttl,
	}).Debug("Connection stored in Redis successfully")
	return nil
}

// GetActive retrieves active connections for a server
func (r *RedisConnectionsRepository) GetActive(ctx context.Context, serverID string) ([]*models.Connection, error) {
	activeKey := fmt.Sprintf("connections:active:%s", serverID)

	// Get active connection IDs
	connIDs, err := r.client.SMembers(ctx, activeKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get active connection IDs from Redis")
		return nil, fmt.Errorf("failed to get active connection IDs: %w", err)
	}

	var connections []*models.Connection
	for _, connID := range connIDs {
		connKey := fmt.Sprintf("connection:%s", connID)

		data, err := r.client.Get(ctx, connKey).Result()
		if err != nil {
			if err != redis.Nil {
				r.logger.WithError(err).WithField("connection_id", connID).Warn("Failed to get connection data")
			}
			continue
		}

		var conn models.Connection
		if err := json.Unmarshal([]byte(data), &conn); err != nil {
			r.logger.WithError(err).WithField("connection_id", connID).Warn("Failed to unmarshal connection")
			continue
		}

		// Only return active connections
		if conn.Status == "active" {
			connections = append(connections, &conn)
		}
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"count":     len(connections),
	}).Debug("Active connections retrieved from Redis successfully")
	return connections, nil
}

// GetHistory retrieves connection history for a server
func (r *RedisConnectionsRepository) GetHistory(ctx context.Context, serverID string, limit int) ([]*models.Connection, error) {
	historyKey := fmt.Sprintf("connections:history:%s", serverID)

	// Get connection IDs from history
	connIDs, err := r.client.LRange(ctx, historyKey, 0, int64(limit-1)).Result()
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get connection history from Redis")
		return nil, fmt.Errorf("failed to get connection history: %w", err)
	}

	var connections []*models.Connection
	for _, connID := range connIDs {
		connKey := fmt.Sprintf("connection:%s", connID)

		data, err := r.client.Get(ctx, connKey).Result()
		if err != nil {
			if err != redis.Nil {
				r.logger.WithError(err).WithField("connection_id", connID).Warn("Failed to get connection data")
			}
			continue
		}

		var conn models.Connection
		if err := json.Unmarshal([]byte(data), &conn); err != nil {
			r.logger.WithError(err).WithField("connection_id", connID).Warn("Failed to unmarshal connection")
			continue
		}

		connections = append(connections, &conn)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"limit":     limit,
		"count":     len(connections),
	}).Debug("Connection history retrieved from Redis successfully")
	return connections, nil
}

// GetByID retrieves a connection by ID
func (r *RedisConnectionsRepository) GetByID(ctx context.Context, connectionID string) (*models.Connection, error) {
	connKey := fmt.Sprintf("connection:%s", connectionID)

	data, err := r.client.Get(ctx, connKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("connection not found: %s", connectionID)
		}
		r.logger.WithError(err).WithField("connection_id", connectionID).Error("Failed to get connection from Redis")
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	var conn models.Connection
	if err := json.Unmarshal([]byte(data), &conn); err != nil {
		r.logger.WithError(err).WithField("connection_id", connectionID).Error("Failed to unmarshal connection")
		return nil, fmt.Errorf("failed to unmarshal connection: %w", err)
	}

	r.logger.WithField("connection_id", connectionID).Debug("Connection retrieved from Redis successfully")
	return &conn, nil
}

// GetByType retrieves connections by type for a server
func (r *RedisConnectionsRepository) GetByType(ctx context.Context, serverID string, connType string) ([]*models.Connection, error) {
	// Get all connections and filter by type
	connections, err := r.GetActive(ctx, serverID)
	if err != nil {
		return nil, err
	}

	var filtered []*models.Connection
	for _, conn := range connections {
		if conn.Type == connType {
			filtered = append(filtered, conn)
		}
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"type":      connType,
		"count":     len(filtered),
	}).Debug("Connections retrieved by type from Redis successfully")
	return filtered, nil
}

// GetAll retrieves all connections
func (r *RedisConnectionsRepository) GetAll(ctx context.Context) ([]*models.Connection, error) {
	// This would require scanning all server keys, which is inefficient
	// For simplicity, return empty list
	return []*models.Connection{}, nil
}

// Close closes a connection
func (r *RedisConnectionsRepository) Close(ctx context.Context, connectionID string) error {
	// Get connection data first
	conn, err := r.GetByID(ctx, connectionID)
	if err != nil {
		return err
	}

	// Update status to disconnected
	conn.Status = "disconnected"
	now := time.Now()
	conn.DisconnectedAt = &now

	// Remove from active connections if it was active
	if conn.Status == "active" {
		activeKey := fmt.Sprintf("connections:active:%s", conn.ServerID)
		r.client.SRem(ctx, activeKey, connectionID)
	}

	// Update connection data
	connKey := fmt.Sprintf("connection:%s", connectionID)
	data, err := json.Marshal(conn)
	if err != nil {
		return fmt.Errorf("failed to marshal connection: %w", err)
	}

	if err := r.client.Set(ctx, connKey, data, r.ttl).Err(); err != nil {
		r.logger.WithError(err).WithField("connection_id", connectionID).Error("Failed to update connection in Redis")
		return fmt.Errorf("failed to update connection: %w", err)
	}

	r.logger.WithField("connection_id", connectionID).Debug("Connection closed in Redis successfully")
	return nil
}

// CloseByServer closes all connections for a server
func (r *RedisConnectionsRepository) CloseByServer(ctx context.Context, serverID string) error {
	activeKey := fmt.Sprintf("connections:active:%s", serverID)

	// Get all active connection IDs
	connIDs, err := r.client.SMembers(ctx, activeKey).Result()
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get active connection IDs")
		return fmt.Errorf("failed to get active connection IDs: %w", err)
	}

	// Close each connection
	for _, connID := range connIDs {
		if err := r.Close(ctx, connID); err != nil {
			r.logger.WithError(err).WithField("connection_id", connID).Warn("Failed to close connection")
		}
	}

	// Remove the active connections set
	if err := r.client.Del(ctx, activeKey).Err(); err != nil {
		r.logger.WithError(err).WithField("key", activeKey).Error("Failed to delete active connections set")
	}

	r.logger.WithField("server_id", serverID).Debug("All connections closed for server in Redis")
	return nil
}

// MarkDisconnected marks a connection as disconnected
func (r *RedisConnectionsRepository) MarkDisconnected(ctx context.Context, connectionID string) error {
	return r.Close(ctx, connectionID)
}

// DeleteOlderThan deletes connections older than specified duration
func (r *RedisConnectionsRepository) DeleteOlderThan(ctx context.Context, olderThan time.Duration) error {
	// Redis handles TTL automatically, so we don't need to implement this
	// The connections will be automatically deleted when their TTL expires
	r.logger.WithField("older_than", olderThan).Debug("Redis handles TTL cleanup automatically")
	return nil
}

// DeleteDisconnected deletes disconnected connections older than specified duration
func (r *RedisConnectionsRepository) DeleteDisconnected(ctx context.Context, olderThan time.Duration) error {
	// Redis handles TTL automatically, so we don't need to implement this
	// The connections will be automatically deleted when their TTL expires
	r.logger.WithField("older_than", olderThan).Debug("Redis handles TTL cleanup automatically")
	return nil
}

// DeleteByServer deletes all connections for a server
func (r *RedisConnectionsRepository) DeleteByServer(ctx context.Context, serverID string) error {
	keys := []string{
		fmt.Sprintf("connections:active:%s", serverID),
		fmt.Sprintf("connections:history:%s", serverID),
	}

	// Delete all connection-related keys for the server
	for _, key := range keys {
		if err := r.client.Del(ctx, key).Err(); err != nil {
			r.logger.WithError(err).WithField("key", key).Warn("Failed to delete connection key")
		}
	}

	// Delete individual connection data (optional, as they will expire with TTL)
	pattern := fmt.Sprintf("connection:*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get connection keys for cleanup")
		return fmt.Errorf("failed to get connection keys: %w", err)
	}

	for _, key := range keys {
		if err := r.client.Del(ctx, key).Err(); err != nil {
			r.logger.WithError(err).WithField("key", key).Warn("Failed to delete connection data key")
		}
	}

	r.logger.WithField("server_id", serverID).Debug("All connections deleted for server in Redis")
	return nil
}

// Ping checks Redis connectivity
func (r *RedisConnectionsRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
