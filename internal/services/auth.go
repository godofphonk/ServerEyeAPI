package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
)

// AuthService handles authentication and key generation
type AuthService struct {
	keyRepo    interfaces.GeneratedKeyRepository
	serverRepo interfaces.ServerRepository
	logger     *logrus.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(keyRepo interfaces.GeneratedKeyRepository, serverRepo interfaces.ServerRepository, logger *logrus.Logger) *AuthService {
	return &AuthService{
		keyRepo:    keyRepo,
		serverRepo: serverRepo,
		logger:     logger,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Hostname        string `json:"hostname" validate:"required"`
	OperatingSystem string `json:"operating_system" validate:"required"`
	AgentVersion    string `json:"agent_version" validate:"required"`
}

// RegisterResponse represents a registration response
type RegisterResponse struct {
	ServerID  string `json:"server_id"`
	ServerKey string `json:"server_key"`
	Status    string `json:"status"`
}

// RegisterKey registers a new server and generates credentials
func (s *AuthService) RegisterKey(ctx context.Context, req *models.RegisterKeyRequest) (*models.RegisterKeyResponse, error) {
	// Generate unique server ID and key
	serverID := fmt.Sprintf("srv_%s", uuid.New().String()[:8])
	serverKey := fmt.Sprintf("key_%s", uuid.New().String()[:24])

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

	err := s.keyRepo.Create(ctx, key)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create generated key")
		return nil, fmt.Errorf("failed to register key: %w", err)
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

	err = s.serverRepo.Create(ctx, server)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create server record")
		return nil, fmt.Errorf("failed to register key: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"server_id":  serverID,
		"server_key": serverKey,
		"hostname":   req.Hostname,
	}).Info("Server registered successfully")

	return &models.RegisterKeyResponse{
		ServerID:  serverID,
		ServerKey: serverKey,
		Status:    "registered",
	}, nil
}

// AuthenticateServer authenticates a server using server_id and server_key
func (s *AuthService) AuthenticateServer(ctx context.Context, serverID, serverKey string) (*models.GeneratedKey, error) {
	key, err := s.keyRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if key.ServerKey != serverKey {
		return nil, fmt.Errorf("authentication failed: invalid server key")
	}

	// Update server status to online
	err = s.serverRepo.UpdateStatus(ctx, serverID, "online")
	if err != nil {
		s.logger.WithError(err).Warn("Failed to update server status to online")
		// Don't fail authentication if status update fails
	}

	// Update last seen
	err = s.serverRepo.UpdateLastSeen(ctx, serverID, time.Now())
	if err != nil {
		s.logger.WithError(err).Warn("Failed to update server last seen")
		// Don't fail authentication if last seen update fails
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": serverID,
	}).Info("Server authenticated successfully")

	return key, nil
}

// GetServerByKey retrieves server information by server key
func (s *AuthService) GetServerByKey(ctx context.Context, serverKey string) (*models.GeneratedKey, error) {
	return s.keyRepo.GetByKey(ctx, serverKey)
}

// GetServerByID retrieves server information by server ID
func (s *AuthService) GetServerByID(ctx context.Context, serverID string) (*models.Server, error) {
	return s.serverRepo.GetByID(ctx, serverID)
}

// UpdateServerStatus updates server status
func (s *AuthService) UpdateServerStatus(ctx context.Context, serverID, status string) error {
	return s.serverRepo.UpdateStatus(ctx, serverID, status)
}

// ListServers retrieves all servers
func (s *AuthService) ListServers(ctx context.Context) ([]*models.Server, error) {
	return s.serverRepo.List(ctx)
}

// Ping checks repository connectivity
func (s *AuthService) Ping(ctx context.Context) error {
	if err := s.keyRepo.Ping(ctx); err != nil {
		return fmt.Errorf("key repository ping failed: %w", err)
	}

	if err := s.serverRepo.Ping(ctx); err != nil {
		return fmt.Errorf("server repository ping failed: %w", err)
	}

	return nil
}
