package storage

import (
	"context"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/postgres"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/redis"
)

// Storage defines the interface for storage operations
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
	GetPendingCommands(ctx context.Context, serverID string) ([]string, error)

	// DLQ operations
	StoreDLQMessage(ctx context.Context, topic string, partition int, offset int64, message []byte, errorMsg string) error

	// Connection operations
	Ping() error
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
