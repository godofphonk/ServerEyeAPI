package services

import (
	"context"
	"fmt"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	"github.com/sirupsen/logrus"
)

// AuthService handles authentication operations using repositories directly
type AuthService struct {
	keyRepo    interfaces.GeneratedKeyRepository
	serverRepo interfaces.ServerRepository
	logger     *logrus.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(keyRepo interfaces.GeneratedKeyRepository, serverRepo interfaces.ServerRepository, logger *logrus.Logger) *AuthService {
	return &AuthService{
		keyRepo:    keyRepo,
		serverRepo: serverRepo,
		logger:     logger,
	}
}

// RegisterKey registers a new server key
func (s *AuthService) RegisterKey(ctx context.Context, req *models.RegisterKeyRequest) (*models.RegisterKeyResponse, error) {
	// Use ServerService for registration
	serverService := NewServerService(s.serverRepo, s.keyRepo, s.logger)

	serverReq := &RegisterServerRequest{
		Hostname:        req.Hostname,
		OperatingSystem: req.OperatingSystem,
		AgentVersion:    req.AgentVersion,
	}

	response, err := serverService.RegisterServer(ctx, serverReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register key: %w", err)
	}

	return &models.RegisterKeyResponse{
		ServerID:  response.ServerID,
		ServerKey: response.ServerKey,
		Status:    response.Status,
	}, nil
}

// AuthenticateServer authenticates a server using ServerService
func (s *AuthService) AuthenticateServer(ctx context.Context, serverID, serverKey string) (*models.GeneratedKey, error) {
	// Use ServerService for authentication
	serverService := NewServerService(s.serverRepo, s.keyRepo, s.logger)

	server, err := serverService.AuthenticateWebSocket(ctx, serverID, serverKey)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Convert Server to GeneratedKey for compatibility
	key := &models.GeneratedKey{
		ServerID:     server.ID,
		ServerKey:    server.ServerKey,
		AgentVersion: server.AgentVersion,
		OSInfo:       server.OSInfo,
		Hostname:     server.Hostname,
		Status:       server.Status,
		CreatedAt:    server.CreatedAt,
	}

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
	serverService := NewServerService(s.serverRepo, s.keyRepo, s.logger)
	return serverService.UpdateServerStatus(ctx, serverID, status)
}

// ListServers retrieves all servers
func (s *AuthService) ListServers(ctx context.Context) ([]*models.Server, error) {
	serverService := NewServerService(s.serverRepo, s.keyRepo, s.logger)
	return serverService.ListServers(ctx, "")
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
