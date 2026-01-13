package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/redis"
)

// StorageAdapter adapts new repository interfaces to old Storage interface
type StorageAdapter struct {
	keyRepo     interfaces.GeneratedKeyRepository
	serverRepo  interfaces.ServerRepository
	redisClient *redis.Client
}

// NewStorageAdapter creates a new storage adapter
func NewStorageAdapter(keyRepo interfaces.GeneratedKeyRepository, serverRepo interfaces.ServerRepository) *StorageAdapter {
	return &StorageAdapter{
		keyRepo:    keyRepo,
		serverRepo: serverRepo,
	}
}

// NewStorageAdapterWithRedis creates a new storage adapter with Redis support
func NewStorageAdapterWithRedis(keyRepo interfaces.GeneratedKeyRepository, serverRepo interfaces.ServerRepository, redisClient *redis.Client) *StorageAdapter {
	return &StorageAdapter{
		keyRepo:     keyRepo,
		serverRepo:  serverRepo,
		redisClient: redisClient,
	}
}

// InsertGeneratedKey stores in PostgreSQL
func (s *StorageAdapter) InsertGeneratedKey(ctx context.Context, secretKey, agentVersion, operatingSystem, hostname string) error {
	// This method is deprecated, use InsertGeneratedKeyWithIDs instead
	return s.InsertGeneratedKeyWithIDs(ctx, secretKey, "", "", agentVersion, operatingSystem, hostname)
}

// InsertGeneratedKeyWithIDs stores in PostgreSQL with server_id and server_key
func (s *StorageAdapter) InsertGeneratedKeyWithIDs(ctx context.Context, secretKey, serverID, serverKey, agentVersion, operatingSystem, hostname string) error {
	key := &models.GeneratedKey{
		ServerID:     serverID,
		ServerKey:    serverKey,
		AgentVersion: agentVersion,
		OSInfo:       operatingSystem,
		Hostname:     hostname,
		Status:       "generated",
	}
	return s.keyRepo.Create(ctx, key)
}

// GetServerByKey retrieves from PostgreSQL
func (s *StorageAdapter) GetServerByKey(ctx context.Context, serverKey string) (*models.ServerInfo, error) {
	key, err := s.keyRepo.GetByKey(ctx, serverKey)
	if err != nil {
		return nil, err
	}

	return &models.ServerInfo{
		ServerID: key.ServerID,
		Hostname: key.Hostname,
	}, nil
}

// GetServers retrieves from PostgreSQL
func (s *StorageAdapter) GetServers(ctx context.Context) ([]string, error) {
	servers, err := s.serverRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	var serverIDs []string
	for _, server := range servers {
		serverIDs = append(serverIDs, server.ID)
	}

	return serverIDs, nil
}

// StoreMetric stores in Redis
func (s *StorageAdapter) StoreMetric(ctx context.Context, serverID string, metrics *models.ServerMetrics) error {
	if s.redisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.StoreMetric(ctx, serverID, metrics)
}

// GetMetric retrieves from Redis
func (s *StorageAdapter) GetMetric(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.GetMetric(ctx, serverID)
}

// GetServerMetrics retrieves from PostgreSQL (placeholder)
func (s *StorageAdapter) GetServerMetrics(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	// TODO: Implement server status repository - for now return basic status
	return &models.ServerStatus{
		Online:   true,
		LastSeen: time.Now(),
	}, nil
}

// SetServerStatus stores in Redis
func (s *StorageAdapter) SetServerStatus(ctx context.Context, serverID string, status *models.ServerStatus) error {
	if s.redisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.SetServerStatus(ctx, serverID, status)
}

// GetServerStatus retrieves from Redis
func (s *StorageAdapter) GetServerStatus(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.GetServerStatus(ctx, serverID)
}

// StoreCommand stores in Redis
func (s *StorageAdapter) StoreCommand(ctx context.Context, serverID string, command map[string]interface{}) error {
	if s.redisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.StoreCommand(ctx, serverID, command)
}

// GetCommands retrieves from Redis
func (s *StorageAdapter) GetCommands(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.GetCommands(ctx, serverID)
}

// StoreDLQ stores in Redis
func (s *StorageAdapter) StoreDLQ(ctx context.Context, topic, message string, metadata map[string]interface{}) error {
	if s.redisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.StoreDLQ(ctx, topic, message, metadata)
}

// GetDLQ retrieves from Redis
func (s *StorageAdapter) GetDLQ(ctx context.Context, topic string, limit int) ([]map[string]interface{}, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.GetDLQ(ctx, topic, limit)
}

// StoreConnection stores in Redis
func (s *StorageAdapter) StoreConnection(ctx context.Context, serverID string, connectionInfo map[string]interface{}) error {
	if s.redisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.StoreConnection(ctx, serverID, connectionInfo)
}

// GetConnections retrieves from Redis
func (s *StorageAdapter) GetConnections(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.GetConnections(ctx, serverID)
}

// Close cleanup
func (s *StorageAdapter) Close() error {
	// Repositories don't need explicit closing
	return nil
}

// StoreDLQMessage stores in PostgreSQL (placeholder)
func (s *StorageAdapter) StoreDLQMessage(ctx context.Context, topic string, partition int, offset int64, message []byte, errorMsg string) error {
	// TODO: Implement DLQ repository in PostgreSQL
	return fmt.Errorf("DLQ storage not yet implemented")
}

// GetPendingCommands retrieves from Redis
func (s *StorageAdapter) GetPendingCommands(ctx context.Context, serverID string) ([]string, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}
	return s.redisClient.GetPendingCommands(ctx, serverID)
}

// Ping checks database connectivity through repository
func (s *StorageAdapter) Ping() error {
	ctx := context.Background()
	return s.keyRepo.Ping(ctx)
}
