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
	"fmt"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/timescaledb"
	"github.com/sirupsen/logrus"
)

// TimescaleDBStorageAdapter implements Storage interface using TimescaleDB
type TimescaleDBStorageAdapter struct {
	keyRepo     interfaces.GeneratedKeyRepository
	serverRepo  interfaces.ServerRepository
	timescaleDB *timescaledb.Client
	logger      *logrus.Logger
	config      *config.Config
}

// NewTimescaleDBStorageAdapter creates a new storage adapter with TimescaleDB
func NewTimescaleDBStorageAdapter(
	keyRepo interfaces.GeneratedKeyRepository,
	serverRepo interfaces.ServerRepository,
	timescaleDB *timescaledb.Client,
	logger *logrus.Logger,
	config *config.Config,
) *TimescaleDBStorageAdapter {
	return &TimescaleDBStorageAdapter{
		keyRepo:     keyRepo,
		serverRepo:  serverRepo,
		timescaleDB: timescaleDB,
		logger:      logger,
		config:      config,
	}
}

// InsertGeneratedKey stores in PostgreSQL
func (s *TimescaleDBStorageAdapter) InsertGeneratedKey(ctx context.Context, secretKey, agentVersion, operatingSystem, hostname string) error {
	return s.InsertGeneratedKeyWithIDs(ctx, secretKey, "", "", agentVersion, operatingSystem, hostname)
}

// InsertGeneratedKeyWithIDs stores in PostgreSQL with server_id and server_key
func (s *TimescaleDBStorageAdapter) InsertGeneratedKeyWithIDs(ctx context.Context, secretKey, serverID, serverKey, agentVersion, operatingSystem, hostname string) error {
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
func (s *TimescaleDBStorageAdapter) GetServerByKey(ctx context.Context, serverKey string) (*models.ServerInfo, error) {
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
func (s *TimescaleDBStorageAdapter) GetServers(ctx context.Context) ([]string, error) {
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

// StoreMetric stores in TimescaleDB
func (s *TimescaleDBStorageAdapter) StoreMetric(ctx context.Context, serverID string, metrics *models.ServerMetrics) error {
	if s.timescaleDB == nil {
		return fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.StoreMetric(ctx, serverID, metrics)
}

// GetMetric retrieves from TimescaleDB
func (s *TimescaleDBStorageAdapter) GetMetric(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	if s.timescaleDB == nil {
		return nil, fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.GetLatestMetric(ctx, serverID)
}

// GetServerMetrics retrieves from TimescaleDB
func (s *TimescaleDBStorageAdapter) GetServerMetrics(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	if s.timescaleDB == nil {
		return nil, fmt.Errorf("TimescaleDB client not initialized")
	}

	// Get latest status from TimescaleDB
	status, err := s.timescaleDB.GetServerStatus(ctx, serverID)
	if err != nil {
		return nil, err
	}

	return status, nil
}

// SetServerStatus stores in TimescaleDB
func (s *TimescaleDBStorageAdapter) SetServerStatus(ctx context.Context, serverID string, status *models.ServerStatus) error {
	if s.timescaleDB == nil {
		return fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.SetServerStatus(ctx, serverID, status)
}

// GetServerStatus retrieves from TimescaleDB
func (s *TimescaleDBStorageAdapter) GetServerStatus(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	if s.timescaleDB == nil {
		return nil, fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.GetServerStatus(ctx, serverID)
}

// StoreCommand stores in TimescaleDB
func (s *TimescaleDBStorageAdapter) StoreCommand(ctx context.Context, serverID string, command map[string]interface{}) error {
	if s.timescaleDB == nil {
		return fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.StoreCommand(ctx, serverID, command)
}

// GetCommands retrieves from TimescaleDB
func (s *TimescaleDBStorageAdapter) GetCommands(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	if s.timescaleDB == nil {
		return nil, fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.GetCommands(ctx, serverID, 50) // default limit
}

// GetPendingCommands retrieves pending commands from TimescaleDB
func (s *TimescaleDBStorageAdapter) GetPendingCommands(ctx context.Context, serverID string) ([]string, error) {
	if s.timescaleDB == nil {
		return nil, fmt.Errorf("TimescaleDB client not initialized")
	}

	commands, err := s.timescaleDB.GetPendingCommands(ctx, serverID)
	if err != nil {
		return nil, err
	}

	var commandIDs []string
	for _, cmd := range commands {
		if id, ok := cmd["command_id"].(string); ok {
			commandIDs = append(commandIDs, id)
		}
	}

	return commandIDs, nil
}

// StoreDLQ stores in TimescaleDB (regular table)
func (s *TimescaleDBStorageAdapter) StoreDLQ(ctx context.Context, topic, message string, metadata map[string]interface{}) error {
	if s.timescaleDB == nil {
		return fmt.Errorf("TimescaleDB client not initialized")
	}

	// Convert metadata to JSON and store in dead_letter_queue table
	query := `
	INSERT INTO dead_letter_queue (topic, message, error, server_id)
	VALUES ($1, $2, $3, $4)`

	var serverID string
	if sid, ok := metadata["server_id"].(string); ok {
		serverID = sid
	}

	err := s.timescaleDB.ExecuteCommand(ctx, query, topic, message, "DLQ storage", serverID)
	return err
}

// GetDLQ retrieves from TimescaleDB
func (s *TimescaleDBStorageAdapter) GetDLQ(ctx context.Context, topic string, limit int) ([]map[string]interface{}, error) {
	if s.timescaleDB == nil {
		return nil, fmt.Errorf("TimescaleDB client not initialized")
	}

	query := `
	SELECT id, topic, message, error, server_id, created_at, attempts
	FROM dead_letter_queue 
	WHERE topic = $1 
	ORDER BY created_at DESC 
	LIMIT $2`

	rows, err := s.timescaleDB.ExecuteQuery(ctx, query, topic, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dlqItems []map[string]interface{}
	for rows.Next() {
		var id int64
		var topic, message, error, serverID string
		var createdAt time.Time
		var attempts int

		if err := rows.Scan(&id, &topic, &message, &error, &serverID, &createdAt, &attempts); err != nil {
			continue
		}

		item := map[string]interface{}{
			"id":         id,
			"topic":      topic,
			"message":    message,
			"error":      error,
			"server_id":  serverID,
			"created_at": createdAt,
			"attempts":   attempts,
		}

		dlqItems = append(dlqItems, item)
	}

	return dlqItems, nil
}

// StoreConnection stores in TimescaleDB as server events
func (s *TimescaleDBStorageAdapter) StoreConnection(ctx context.Context, serverID string, connectionInfo map[string]interface{}) error {
	if s.timescaleDB == nil {
		return fmt.Errorf("TimescaleDB client not initialized")
	}

	// Store as connection event
	query := `
	INSERT INTO server_events (server_id, event_type, event_data, level, source)
	VALUES ($1, 'connection', $2, 'info', 'websocket')`

	err := s.timescaleDB.ExecuteCommand(ctx, query, serverID, connectionInfo)
	return err
}

// GetConnections retrieves from TimescaleDB as server events
func (s *TimescaleDBStorageAdapter) GetConnections(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	if s.timescaleDB == nil {
		return nil, fmt.Errorf("TimescaleDB client not initialized")
	}

	query := `
	SELECT event_data, time
	FROM server_events 
	WHERE server_id = $1 AND event_type = 'connection'
	ORDER BY time DESC 
	LIMIT 50`

	rows, err := s.timescaleDB.ExecuteQuery(ctx, query, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []map[string]interface{}
	for rows.Next() {
		var eventData map[string]interface{}
		var eventTime time.Time

		if err := rows.Scan(&eventData, &eventTime); err != nil {
			continue
		}

		connection := map[string]interface{}{
			"event_data": eventData,
			"time":       eventTime,
		}

		connections = append(connections, connection)
	}

	return connections, nil
}

// StoreDLQMessage stores in TimescaleDB dead_letter_queue
func (s *TimescaleDBStorageAdapter) StoreDLQMessage(ctx context.Context, topic string, partition int, offset int64, message []byte, errorMsg string) error {
	if s.timescaleDB == nil {
		return fmt.Errorf("TimescaleDB client not initialized")
	}

	query := `
	INSERT INTO dead_letter_queue (topic, partition, message_offset, message, error)
	VALUES ($1, $2, $3, $4, $5)`

	err := s.timescaleDB.ExecuteCommand(ctx, query, topic, partition, offset, message, errorMsg)
	return err
}

// Close cleanup
func (s *TimescaleDBStorageAdapter) Close() error {
	if s.timescaleDB != nil {
		return s.timescaleDB.Close()
	}
	return nil
}

// Ping checks database connectivity
func (s *TimescaleDBStorageAdapter) Ping() error {
	ctx := context.Background()

	// Check PostgreSQL repositories
	if err := s.keyRepo.Ping(ctx); err != nil {
		return fmt.Errorf("key repository ping failed: %w", err)
	}

	// Check TimescaleDB
	if s.timescaleDB != nil {
		if err := s.timescaleDB.Ping(ctx); err != nil {
			return fmt.Errorf("TimescaleDB ping failed: %w", err)
		}
	}

	return nil
}

// Additional TimescaleDB-specific methods

// GetMetricsHistory retrieves historical metrics from TimescaleDB
func (s *TimescaleDBStorageAdapter) GetMetricsHistory(ctx context.Context, serverID string, start, end time.Time, interval time.Duration) ([]*models.ServerMetrics, error) {
	if s.timescaleDB == nil {
		return nil, fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.GetMetricsHistory(ctx, serverID, start, end, interval)
}

// GetAggregatedMetrics retrieves aggregated metrics from TimescaleDB
func (s *TimescaleDBStorageAdapter) GetAggregatedMetrics(ctx context.Context, serverID string, period string) (interface{}, error) {
	if s.timescaleDB == nil {
		return nil, fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.GetAggregatedMetrics(ctx, serverID, period)
}

// GetActiveServers retrieves active servers from TimescaleDB
func (s *TimescaleDBStorageAdapter) GetActiveServers(ctx context.Context) ([]map[string]interface{}, error) {
	if s.timescaleDB == nil {
		return nil, fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.GetActiveServers(ctx)
}

// UpdateCommandStatus updates command status in TimescaleDB
func (s *TimescaleDBStorageAdapter) UpdateCommandStatus(ctx context.Context, commandID string, status string, response map[string]interface{}, errorMessage string) error {
	if s.timescaleDB == nil {
		return fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.UpdateCommandStatus(ctx, commandID, status, response, errorMessage)
}

// GetServerUptime retrieves server uptime from TimescaleDB
func (s *TimescaleDBStorageAdapter) GetServerUptime(ctx context.Context, serverID string) (float64, error) {
	if s.timescaleDB == nil {
		return 0, fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.GetServerUptime(ctx, serverID)
}

// UpdateServerHeartbeat updates server heartbeat in TimescaleDB
func (s *TimescaleDBStorageAdapter) UpdateServerHeartbeat(ctx context.Context, serverID string) error {
	if s.timescaleDB == nil {
		return fmt.Errorf("TimescaleDB client not initialized")
	}
	return s.timescaleDB.UpdateServerHeartbeat(ctx, serverID)
}
