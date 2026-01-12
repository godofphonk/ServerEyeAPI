package storage

import (
	"context"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
)

// StorageAdapter adapts new repository interfaces to old Storage interface
type StorageAdapter struct {
	keyRepo    interfaces.GeneratedKeyRepository
	serverRepo interfaces.ServerRepository
}

// NewStorageAdapter creates a new storage adapter
func NewStorageAdapter(keyRepo interfaces.GeneratedKeyRepository, serverRepo interfaces.ServerRepository) *StorageAdapter {
	return &StorageAdapter{
		keyRepo:    keyRepo,
		serverRepo: serverRepo,
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

// StoreMetric - not implemented in new architecture yet
func (s *StorageAdapter) StoreMetric(ctx context.Context, serverID string, metrics *models.ServerMetrics) error {
	// TODO: Implement metrics repository
	return nil
}

// GetMetric - not implemented in new architecture yet
func (s *StorageAdapter) GetMetric(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	// TODO: Implement metrics repository
	return nil, nil
}

// GetServerMetrics - not implemented in new architecture yet
func (s *StorageAdapter) GetServerMetrics(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	// TODO: Implement server status repository
	return nil, nil
}

// SetServerStatus - not implemented in new architecture yet
func (s *StorageAdapter) SetServerStatus(ctx context.Context, serverID string, status *models.ServerStatus) error {
	// TODO: Implement server status repository
	return nil
}

// GetServerStatus - not implemented in new architecture yet
func (s *StorageAdapter) GetServerStatus(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	// TODO: Implement server status repository
	return nil, nil
}

// StoreCommand - not implemented in new architecture yet
func (s *StorageAdapter) StoreCommand(ctx context.Context, serverID string, command map[string]interface{}) error {
	// TODO: Implement commands repository
	return nil
}

// GetCommands - not implemented in new architecture yet
func (s *StorageAdapter) GetCommands(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	// TODO: Implement commands repository
	return nil, nil
}

// StoreDLQ - not implemented in new architecture yet
func (s *StorageAdapter) StoreDLQ(ctx context.Context, topic, message string, metadata map[string]interface{}) error {
	// TODO: Implement DLQ repository
	return nil
}

// GetDLQ - not implemented in new architecture yet
func (s *StorageAdapter) GetDLQ(ctx context.Context, topic string, limit int) ([]map[string]interface{}, error) {
	// TODO: Implement DLQ repository
	return nil, nil
}

// StoreConnection - not implemented in new architecture yet
func (s *StorageAdapter) StoreConnection(ctx context.Context, serverID string, connectionInfo map[string]interface{}) error {
	// TODO: Implement connections repository
	return nil
}

// GetConnections - not implemented in new architecture yet
func (s *StorageAdapter) GetConnections(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	// TODO: Implement connections repository
	return nil, nil
}

// Close cleanup
func (s *StorageAdapter) Close() error {
	// Repositories don't need explicit closing
	return nil
}

// StoreDLQMessage stores in PostgreSQL
func (s *StorageAdapter) StoreDLQMessage(ctx context.Context, topic string, partition int, offset int64, message []byte, errorMsg string) error {
	// TODO: Implement DLQ repository
	return nil
}

// GetPendingCommands retrieves from Redis
func (s *StorageAdapter) GetPendingCommands(ctx context.Context, serverID string) ([]string, error) {
	// TODO: Implement commands repository
	return nil, nil
}

// Ping checks database connectivity through repository
func (s *StorageAdapter) Ping() error {
	ctx := context.Background()
	return s.keyRepo.Ping(ctx)
}
