package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Storage implements in-memory storage for testing
type Storage struct {
	generatedKeys map[string]GeneratedKey
	servers       map[string]Server
	metrics       map[string]map[string]interface{}
	commands      map[string][]map[string]interface{}
	mutex         sync.RWMutex
	logger        *logrus.Logger
}

// GeneratedKey represents a generated key
type GeneratedKey struct {
	SecretKey       string
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
		generatedKeys: make(map[string]GeneratedKey),
		servers:       make(map[string]Server),
		metrics:       make(map[string]map[string]interface{}),
		commands:      make(map[string][]map[string]interface{}),
		logger:        logger,
	}
}

// InsertGeneratedKey inserts a new generated key
func (s *Storage) InsertGeneratedKey(ctx context.Context, secretKey, agentVersion, operatingSystem, hostname string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.generatedKeys[secretKey] = GeneratedKey{
		SecretKey:       secretKey,
		AgentVersion:    agentVersion,
		OperatingSystem: operatingSystem,
		Hostname:        hostname,
		Status:          "generated",
		CreatedAt:       time.Now(),
	}

	s.logger.WithField("secret_key", secretKey).Info("Generated key stored in memory")
	return nil
}

// GetServers retrieves all server IDs
func (s *Storage) GetServers(ctx context.Context) ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var serverIDs []string
	for serverID := range s.servers {
		serverIDs = append(serverIDs, serverID)
	}

	return serverIDs, nil
}

// GetServerMetrics retrieves metrics for a server
func (s *Storage) GetServerMetrics(ctx context.Context, serverID string) (map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if server, exists := s.servers[serverID]; exists {
		return map[string]interface{}{
			"hostname":  server.Hostname,
			"os_info":   server.OSInfo,
			"status":    server.Status,
			"last_seen": server.LastSeen,
		}, nil
	}

	return map[string]interface{}{}, nil
}

// StoreMetric stores metrics for a server (no-op in memory storage)
func (s *Storage) StoreMetric(ctx context.Context, serverID string, data map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.metrics[serverID] = data
	return nil
}

// GetMetric retrieves metrics for a server
func (s *Storage) GetMetric(ctx context.Context, serverID string) (map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if metrics, exists := s.metrics[serverID]; exists {
		return metrics, nil
	}

	return map[string]interface{}{}, nil
}

// SetServerStatus sets server status
func (s *Storage) SetServerStatus(ctx context.Context, serverID string, status map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if server, exists := s.servers[serverID]; exists {
		if online, ok := status["online"].(bool); ok {
			server.Status = "offline"
			if online {
				server.Status = "online"
			}
		}
		if lastSeen, ok := status["last_seen"].(int64); ok {
			server.LastSeen = time.Unix(lastSeen, 0)
		}
		s.servers[serverID] = server
	}

	return nil
}

// GetServerStatus retrieves server status
func (s *Storage) GetServerStatus(ctx context.Context, serverID string) (map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if server, exists := s.servers[serverID]; exists {
		return map[string]interface{}{
			"online":    server.Status == "online",
			"last_seen": server.LastSeen.Unix(),
		}, nil
	}

	return map[string]interface{}{"online": false}, nil
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

// StoreDLQMessage stores a message in the dead letter queue (no-op in memory)
func (s *Storage) StoreDLQMessage(ctx context.Context, topic string, partition int, offset int64, message []byte, errorMsg string) error {
	s.logger.WithFields(logrus.Fields{
		"topic": topic,
		"error": errorMsg,
	}).Warn("DLQ message (memory storage)")
	return nil
}

// Ping checks if storage is available
func (s *Storage) Ping() error {
	return nil
}

// Close closes the storage (no-op for memory)
func (s *Storage) Close() error {
	return nil
}
