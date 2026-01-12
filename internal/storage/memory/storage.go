package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/sirupsen/logrus"
)

// Storage implements in-memory storage for testing
type Storage struct {
	generatedKeys []GeneratedKey
	serverMetrics map[string]*models.ServerMetrics
	serverStatus  map[string]*models.ServerStatus
	commands      map[string][]map[string]interface{}
	dlq           map[string][]map[string]interface{}
	connections   map[string][]map[string]interface{}
	mutex         sync.RWMutex
	logger        *logrus.Logger
}

// GeneratedKey represents a generated key
type GeneratedKey struct {
	SecretKey       string
	ServerID        string
	ServerKey       string
	AgentVersion    string
	OperatingSystem string
	Hostname        string
	Status          string
	CreatedAt       time.Time
}

// Server represents a server
type Server struct {
	ServerID  string
	SecretKey string
	Hostname  string
	OSInfo    string
	Status    string
	LastSeen  time.Time
	CreatedAt time.Time
}

// NewStorage creates a new in-memory storage
func NewStorage(logger *logrus.Logger) *Storage {
	return &Storage{
		generatedKeys: make([]GeneratedKey, 0),
		serverMetrics: make(map[string]*models.ServerMetrics),
		serverStatus:  make(map[string]*models.ServerStatus),
		commands:      make(map[string][]map[string]interface{}),
		dlq:           make(map[string][]map[string]interface{}),
		connections:   make(map[string][]map[string]interface{}),
		logger:        logger,
	}
}

// InsertGeneratedKey inserts a new generated key
func (s *Storage) InsertGeneratedKey(ctx context.Context, secretKey, agentVersion, operatingSystem, hostname string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.generatedKeys = append(s.generatedKeys, GeneratedKey{
		SecretKey:       secretKey,
		AgentVersion:    agentVersion,
		OperatingSystem: operatingSystem,
		Hostname:        hostname,
		Status:          "generated",
		CreatedAt:       time.Now(),
	})

	s.logger.WithField("secret_key", secretKey).Info("Generated key stored in memory")
	return nil
}

// InsertGeneratedKeyWithIDs inserts a new generated key with server_id and server_key
func (s *Storage) InsertGeneratedKeyWithIDs(ctx context.Context, secretKey, serverID, serverKey, agentVersion, operatingSystem, hostname string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if server_id already exists
	for _, key := range s.generatedKeys {
		if key.ServerID == serverID {
			return fmt.Errorf("server_id already exists")
		}
	}

	newKey := GeneratedKey{
		SecretKey:       secretKey, // Can be empty now
		ServerID:        serverID,
		ServerKey:       serverKey,
		AgentVersion:    agentVersion,
		OperatingSystem: operatingSystem,
		Hostname:        hostname,
		Status:          "generated",
		CreatedAt:       time.Now(),
	}

	s.generatedKeys = append(s.generatedKeys, newKey)

	s.logger.WithFields(logrus.Fields{
		"server_id":        serverID,
		"server_key":       serverKey,
		"agent_version":    agentVersion,
		"operating_system": operatingSystem,
		"hostname":         hostname,
	}).Info("Generated key with IDs stored in memory")

	return nil
}

// GetServerByKey retrieves server information by server key
func (s *Storage) GetServerByKey(ctx context.Context, serverKey string) (*models.ServerInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, key := range s.generatedKeys {
		if key.ServerKey == serverKey && key.Status == "generated" {
			return &models.ServerInfo{
				ServerID:  key.ServerID,
				SecretKey: key.SecretKey,
				Hostname:  key.Hostname,
			}, nil
		}
	}

	return nil, fmt.Errorf("server key not found")
}

// GetServers retrieves all server IDs
func (s *Storage) GetServers(ctx context.Context) ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var servers []string
	for _, key := range s.generatedKeys {
		if key.ServerID != "" {
			servers = append(servers, key.ServerID)
		}
	}

	return servers, nil
}

// GetServerMetrics retrieves metrics for a server
func (s *Storage) GetServerMetrics(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Find server info from generated keys
	var serverInfo GeneratedKey
	found := false
	for _, key := range s.generatedKeys {
		if key.ServerID == serverID {
			serverInfo = key
			found = true
			break
		}
	}

	if !found {
		return &models.ServerStatus{Online: false}, nil
	}

	return &models.ServerStatus{
		Online:       serverInfo.Status == "online",
		LastSeen:     serverInfo.CreatedAt,
		Version:      serverInfo.AgentVersion,
		OSInfo:       serverInfo.OperatingSystem,
		AgentVersion: serverInfo.AgentVersion,
		Hostname:     serverInfo.Hostname,
	}, nil
}

// StoreMetric stores metrics for a server
func (s *Storage) StoreMetric(ctx context.Context, serverID string, metrics *models.ServerMetrics) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Store metrics directly
	s.serverMetrics[serverID] = metrics

	s.logger.WithField("server_id", serverID).Debug("Metrics stored in memory")
	return nil
}

// GetMetric retrieves metrics for a server
func (s *Storage) GetMetric(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if metrics, exists := s.serverMetrics[serverID]; exists {
		return metrics, nil
	}

	return nil, fmt.Errorf("metrics not found")
}

// SetServerStatus sets server status
func (s *Storage) SetServerStatus(ctx context.Context, serverID string, status *models.ServerStatus) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Store status directly
	s.serverStatus[serverID] = status

	s.logger.WithField("server_id", serverID).Debug("Server status stored in memory")
	return nil
}

func (s *Storage) GetServerStatus(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if status, exists := s.serverStatus[serverID]; exists {
		return status, nil
	}

	return &models.ServerStatus{Online: false}, nil
}

// StoreCommand stores a command for a server
func (s *Storage) StoreCommand(ctx context.Context, serverID string, command map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.commands[serverID] == nil {
		s.commands[serverID] = []map[string]interface{}{}
	}

	s.commands[serverID] = append(s.commands[serverID], command)
	return nil
}

// GetCommands retrieves commands for a server
func (s *Storage) GetCommands(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if commands, exists := s.commands[serverID]; exists {
		return commands, nil
	}

	return []map[string]interface{}{}, nil
}

// GetPendingCommands retrieves pending commands for a server
func (s *Storage) GetPendingCommands(ctx context.Context, serverID string) ([]string, error) {
	commands, err := s.GetCommands(ctx, serverID)
	if err != nil {
		return nil, err
	}

	var pending []string
	for _, cmd := range commands {
		if status, ok := cmd["status"].(string); ok && status == "pending" {
			cmdStr := fmt.Sprintf("%v", cmd)
			pending = append(pending, cmdStr)
		}
	}

	return pending, nil
}

// StoreDLQ stores a message in the dead letter queue
func (s *Storage) StoreDLQ(ctx context.Context, topic, message string, metadata map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.dlq[topic] == nil {
		s.dlq[topic] = []map[string]interface{}{}
	}

	dlqMessage := map[string]interface{}{
		"message":   message,
		"metadata":  metadata,
		"timestamp": time.Now(),
	}

	s.dlq[topic] = append(s.dlq[topic], dlqMessage)

	s.logger.WithField("topic", topic).Debug("DLQ message stored in memory")
	return nil
}

// StoreDLQMessage stores a message in the dead letter queue (no-op in memory)
func (s *Storage) StoreDLQMessage(ctx context.Context, topic string, partition int, offset int64, message []byte, errorMsg string) error {
	s.logger.WithFields(logrus.Fields{
		"topic": topic,
		"error": errorMsg,
	}).Warn("DLQ message (memory storage)")
	return nil
}

// GetDLQ retrieves DLQ messages for a topic
func (s *Storage) GetDLQ(ctx context.Context, topic string, limit int) ([]map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if messages, exists := s.dlq[topic]; exists {
		if limit > 0 && len(messages) > limit {
			return messages[:limit], nil
		}
		return messages, nil
	}

	return []map[string]interface{}{}, nil
}

// Ping checks if storage is available
func (s *Storage) Ping() error {
	return nil
}

// StoreConnection stores connection info for a server
func (s *Storage) StoreConnection(ctx context.Context, serverID string, connectionInfo map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.connections[serverID] == nil {
		s.connections[serverID] = []map[string]interface{}{}
	}

	s.connections[serverID] = append(s.connections[serverID], connectionInfo)

	s.logger.WithField("server_id", serverID).Debug("Connection info stored in memory")
	return nil
}

// GetConnections retrieves connection info for a server
func (s *Storage) GetConnections(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if connections, exists := s.connections[serverID]; exists {
		return connections, nil
	}

	return []map[string]interface{}{}, nil
}

// Close closes the storage (no-op for memory)
func (s *Storage) Close() error {
	return nil
}
