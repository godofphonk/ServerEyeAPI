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

package storage

import (
	"context"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/postgres"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/redis"
)

// Storage defines the interface for all storage operations
type Storage interface {
	// Key operations
	InsertGeneratedKey(ctx context.Context, secretKey, agentVersion, operatingSystem, hostname string) error
	InsertGeneratedKeyWithIDs(ctx context.Context, secretKey, serverID, serverKey, agentVersion, operatingSystem, hostname string) error
	GetServerByKey(ctx context.Context, serverKey string) (*models.ServerInfo, error)
	GetServers(ctx context.Context) ([]string, error)

	// Metrics operations
	StoreMetric(ctx context.Context, serverID string, metrics *models.ServerMetrics) error
	GetMetric(ctx context.Context, serverID string) (*models.ServerMetrics, error)
	GetServerMetrics(ctx context.Context, serverID string) (*models.ServerStatus, error)

	// Server status operations
	SetServerStatus(ctx context.Context, serverID string, status *models.ServerStatus) error
	GetServerStatus(ctx context.Context, serverID string) (*models.ServerStatus, error)

	// Command operations
	StoreCommand(ctx context.Context, serverID string, command map[string]interface{}) error
	GetCommands(ctx context.Context, serverID string) ([]map[string]interface{}, error)

	// DLQ operations
	StoreDLQ(ctx context.Context, topic, message string, metadata map[string]interface{}) error
	GetDLQ(ctx context.Context, topic string, limit int) ([]map[string]interface{}, error)

	// Connection operations
	StoreConnection(ctx context.Context, serverID string, connectionInfo map[string]interface{}) error
	GetConnections(ctx context.Context, serverID string) ([]map[string]interface{}, error)

	// Health check operations
	Ping() error

	// Close cleanup
	Close() error
}

// CombinedStorage combines PostgreSQL and Redis storage
type CombinedStorage struct {
	postgres *postgres.Client
	redis    *redis.Client
}

// NewCombinedStorage creates a new combined storage
func NewCombinedStorage(pg *postgres.Client, r *redis.Client) *CombinedStorage {
	return &CombinedStorage{
		postgres: pg,
		redis:    r,
	}
}

// InsertGeneratedKey stores in PostgreSQL
func (s *CombinedStorage) InsertGeneratedKey(ctx context.Context, secretKey, agentVersion, operatingSystem, hostname string) error {
	return s.postgres.InsertGeneratedKey(ctx, secretKey, agentVersion, operatingSystem, hostname)
}

// InsertGeneratedKeyWithIDs stores in PostgreSQL with server_id and server_key
func (s *CombinedStorage) InsertGeneratedKeyWithIDs(ctx context.Context, secretKey, serverID, serverKey, agentVersion, operatingSystem, hostname string) error {
	return s.postgres.InsertGeneratedKeyWithIDs(ctx, secretKey, serverID, serverKey, agentVersion, operatingSystem, hostname)
}

// GetServerByKey retrieves from PostgreSQL
func (s *CombinedStorage) GetServerByKey(ctx context.Context, serverKey string) (*models.ServerInfo, error) {
	return s.postgres.GetServerByKey(ctx, serverKey)
}

// GetServers retrieves from PostgreSQL
func (s *CombinedStorage) GetServers(ctx context.Context) ([]string, error) {
	return s.postgres.GetServers(ctx)
}

// StoreMetric stores in Redis
func (s *CombinedStorage) StoreMetric(ctx context.Context, serverID string, metrics *models.ServerMetrics) error {
	return s.redis.StoreMetric(ctx, serverID, metrics)
}

// GetMetric retrieves from Redis
func (s *CombinedStorage) GetMetric(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	return s.redis.GetMetric(ctx, serverID)
}

// GetServerMetrics retrieves from PostgreSQL
func (s *CombinedStorage) GetServerMetrics(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	return s.postgres.GetServerMetrics(ctx, serverID)
}

// SetServerStatus stores in Redis
func (s *CombinedStorage) SetServerStatus(ctx context.Context, serverID string, status *models.ServerStatus) error {
	return s.redis.SetServerStatus(ctx, serverID, status)
}

// GetServerStatus retrieves from Redis
func (s *CombinedStorage) GetServerStatus(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	return s.redis.GetServerStatus(ctx, serverID)
}

// StoreCommand stores in Redis
func (s *CombinedStorage) StoreCommand(ctx context.Context, serverID string, command map[string]interface{}) error {
	return s.redis.StoreCommand(ctx, serverID, command)
}

// GetCommands retrieves from Redis
func (s *CombinedStorage) GetCommands(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	return s.redis.GetCommands(ctx, serverID)
}

// GetPendingCommands retrieves from Redis
func (s *CombinedStorage) GetPendingCommands(ctx context.Context, serverID string) ([]string, error) {
	return s.redis.GetPendingCommands(ctx, serverID)
}

// StoreDLQ stores in Redis
func (s *CombinedStorage) StoreDLQ(ctx context.Context, topic, message string, metadata map[string]interface{}) error {
	return s.redis.StoreDLQ(ctx, topic, message, metadata)
}

// GetDLQ retrieves from Redis
func (s *CombinedStorage) GetDLQ(ctx context.Context, topic string, limit int) ([]map[string]interface{}, error) {
	return s.redis.GetDLQ(ctx, topic, limit)
}

// StoreConnection stores in Redis
func (s *CombinedStorage) StoreConnection(ctx context.Context, serverID string, connectionInfo map[string]interface{}) error {
	return s.redis.StoreConnection(ctx, serverID, connectionInfo)
}

// GetConnections retrieves from Redis
func (s *CombinedStorage) GetConnections(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	return s.redis.GetConnections(ctx, serverID)
}

// StoreDLQMessage stores in PostgreSQL
func (s *CombinedStorage) StoreDLQMessage(ctx context.Context, topic string, partition int, offset int64, message []byte, errorMsg string) error {
	return s.postgres.StoreDLQMessage(ctx, topic, partition, offset, message, errorMsg)
}

// Ping checks both connections
func (s *CombinedStorage) Ping() error {
	ctx := context.Background()
	if err := s.postgres.Ping(); err != nil {
		return err
	}
	return s.redis.Ping(ctx)
}

// Close closes both connections
func (s *CombinedStorage) Close() error {
	if err := s.postgres.Close(); err != nil {
		return err
	}
	return s.redis.Close()
}
