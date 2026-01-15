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

package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
)

// ServerService handles server-related business logic
type ServerService struct {
	serverRepo interfaces.ServerRepository
	keyRepo    interfaces.GeneratedKeyRepository
	logger     *logrus.Logger
}

// NewServerService creates a new server service
func NewServerService(serverRepo interfaces.ServerRepository, keyRepo interfaces.GeneratedKeyRepository, logger *logrus.Logger) *ServerService {
	return &ServerService{
		serverRepo: serverRepo,
		keyRepo:    keyRepo,
		logger:     logger,
	}
}

// GetServerByKey retrieves server information by server key
func (s *ServerService) GetServerByKey(ctx context.Context, serverKey string) (*models.GeneratedKey, error) {
	return s.keyRepo.GetByKey(ctx, serverKey)
}

// RegisterServerRequest represents a server registration request
type RegisterServerRequest struct {
	Hostname        string `json:"hostname" validate:"required"`
	OperatingSystem string `json:"operating_system" validate:"required"`
	AgentVersion    string `json:"agent_version" validate:"required"`
}

// RegisterServerResponse represents a server registration response
type RegisterServerResponse struct {
	ServerID  string `json:"server_id"`
	ServerKey string `json:"server_key"`
	Status    string `json:"status"`
}

// RegisterServer registers a new server with proper business logic
func (s *ServerService) RegisterServer(ctx context.Context, req *RegisterServerRequest) (*RegisterServerResponse, error) {
	// Validate request
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate unique server ID and key
	serverID := fmt.Sprintf("srv_%s", uuid.New().String()[:8])
	serverKey := fmt.Sprintf("key_%s", uuid.New().String()[:8])

	s.logger.WithFields(logrus.Fields{
		"hostname":         req.Hostname,
		"operating_system": req.OperatingSystem,
		"agent_version":    req.AgentVersion,
		"server_id":        serverID,
	}).Info("Registering new server")

	// Create generated key record
	key := &models.GeneratedKey{
		ServerID:     serverID,
		ServerKey:    serverKey,
		AgentVersion: req.AgentVersion,
		OSInfo:       req.OperatingSystem,
		Hostname:     req.Hostname,
		Status:       "generated",
		CreatedAt:    time.Now(),
	}

	if err := s.keyRepo.Create(ctx, key); err != nil {
		s.logger.WithError(err).Error("Failed to create generated key")
		return nil, fmt.Errorf("failed to register server: %w", err)
	}

	// Create server record
	server := &models.Server{
		ID:           serverID,
		ServerKey:    serverKey,
		Hostname:     req.Hostname,
		OSInfo:       req.OperatingSystem,
		AgentVersion: req.AgentVersion,
		Status:       "offline",
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.serverRepo.Create(ctx, server); err != nil {
		s.logger.WithError(err).Error("Failed to create server record")
		return nil, fmt.Errorf("failed to register server: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"server_id":  serverID,
		"server_key": serverKey,
		"hostname":   req.Hostname,
	}).Info("Server registered successfully")

	return &RegisterServerResponse{
		ServerID:  serverID,
		ServerKey: serverKey,
		Status:    "registered",
	}, nil
}

// AuthenticateWebSocket authenticates a server for WebSocket connection
func (s *ServerService) AuthenticateWebSocket(ctx context.Context, serverID, serverKey string) (*models.Server, error) {
	// Validate input
	if serverID == "" || serverKey == "" {
		return nil, fmt.Errorf("server_id and server_key are required")
	}

	// Get server by key for authentication
	key, err := s.keyRepo.GetByKey(ctx, serverKey)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Verify server ID matches
	if key.ServerID != serverID {
		return nil, fmt.Errorf("authentication failed: server_id mismatch")
	}

	// Get full server information
	server, err := s.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Update server status to online
	if err := s.serverRepo.UpdateStatus(ctx, serverID, "online"); err != nil {
		s.logger.WithError(err).Warn("Failed to update server status to online")
		// Don't fail authentication if status update fails
	}

	// Update last seen
	if err := s.serverRepo.UpdateLastSeen(ctx, serverID, time.Now()); err != nil {
		s.logger.WithError(err).Warn("Failed to update server last seen")
		// Don't fail authentication if last seen update fails
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"hostname":  server.Hostname,
	}).Info("WebSocket authentication successful")

	return server, nil
}

// UpdateServerStatus updates server status with business logic
func (s *ServerService) UpdateServerStatus(ctx context.Context, serverID, status string) error {
	// Validate status
	validStatuses := []string{"online", "offline", "maintenance", "error"}
	isValid := false
	for _, s := range validStatuses {
		if s == status {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Check if server exists
	_, err := s.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Update status
	if err := s.serverRepo.UpdateStatus(ctx, serverID, status); err != nil {
		return fmt.Errorf("failed to update server status: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"status":    status,
	}).Info("Server status updated")

	return nil
}

// GetServerByID retrieves server information
func (s *ServerService) GetServerByID(ctx context.Context, serverID string) (*models.Server, error) {
	server, err := s.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	return server, nil
}

// ListServers retrieves all servers with optional filtering
func (s *ServerService) ListServers(ctx context.Context, status string) ([]*models.Server, error) {
	var servers []*models.Server
	var err error

	if status != "" {
		servers, err = s.serverRepo.ListByStatus(ctx, status)
	} else {
		servers, err = s.serverRepo.List(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	return servers, nil
}

// DeleteServer deletes a server with proper cleanup
func (s *ServerService) DeleteServer(ctx context.Context, serverID string) error {
	// Check if server exists
	server, err := s.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Delete server record
	if err := s.serverRepo.Delete(ctx, serverID); err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	// Delete associated key record
	key, err := s.keyRepo.GetByServerID(ctx, serverID)
	if err == nil {
		if err := s.keyRepo.Delete(ctx, key.ID); err != nil {
			s.logger.WithError(err).Warn("Failed to delete associated key")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"hostname":  server.Hostname,
	}).Info("Server deleted successfully")

	return nil
}

// validateRegisterRequest validates registration request
func (s *ServerService) validateRegisterRequest(req *RegisterServerRequest) error {
	if req.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if req.OperatingSystem == "" {
		return fmt.Errorf("operating_system is required")
	}
	if req.AgentVersion == "" {
		return fmt.Errorf("agent_version is required")
	}
	return nil
}

// Ping checks repository connectivity
func (s *ServerService) Ping(ctx context.Context) error {
	if err := s.serverRepo.Ping(ctx); err != nil {
		return fmt.Errorf("server repository ping failed: %w", err)
	}

	if err := s.keyRepo.Ping(ctx); err != nil {
		return fmt.Errorf("key repository ping failed: %w", err)
	}

	return nil
}

// AddServerSource adds a source (TGBot/Web) to a server
func (s *ServerService) AddServerSource(ctx context.Context, serverID, source string) error {
	// Get current server info
	server, err := s.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Parse current sources
	currentSources := server.Sources
	if currentSources == "" {
		currentSources = source
	} else {
		// Check if source already exists
		sources := strings.Split(currentSources, ",")
		for _, src := range sources {
			if strings.TrimSpace(src) == source {
				return fmt.Errorf("source %s already exists for server", source)
			}
		}
		// Add new source
		currentSources += "," + source
	}

	// Update server sources
	return s.serverRepo.UpdateSources(ctx, serverID, currentSources)
}

// GetServerSources gets all sources for a server
func (s *ServerService) GetServerSources(ctx context.Context, serverID string) ([]string, error) {
	server, err := s.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	if server.Sources == "" {
		return []string{}, nil
	}

	sources := strings.Split(server.Sources, ",")
	for i, src := range sources {
		sources[i] = strings.TrimSpace(src)
	}

	return sources, nil
}

// RemoveServerSource removes a source from a server
func (s *ServerService) RemoveServerSource(ctx context.Context, serverID, source string) error {
	// Get current server info
	server, err := s.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	if server.Sources == "" {
		return fmt.Errorf("no sources to remove")
	}

	// Parse and remove source
	sources := strings.Split(server.Sources, ",")
	var newSources []string
	found := false

	for _, src := range sources {
		src = strings.TrimSpace(src)
		if src == source {
			found = true
			continue // Skip this source
		}
		newSources = append(newSources, src)
	}

	if !found {
		return fmt.Errorf("source %s not found for server", source)
	}

	// Update server sources
	var newSourcesStr string
	if len(newSources) > 0 {
		newSourcesStr = strings.Join(newSources, ",")
	}

	return s.serverRepo.UpdateSources(ctx, serverID, newSourcesStr)
}
