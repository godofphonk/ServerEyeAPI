package services

import (
	"context"
	"fmt"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/godofphonk/ServerEyeAPI/internal/utils"
	"github.com/sirupsen/logrus"
)

// AuthService handles authentication operations
type AuthService struct {
	storage storage.Storage
	logger  *logrus.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(storage storage.Storage, logger *logrus.Logger) *AuthService {
	return &AuthService{
		storage: storage,
		logger:  logger,
	}
}

// RegisterKey registers a new server key
func (s *AuthService) RegisterKey(ctx context.Context, req *models.RegisterKeyRequest) (*models.RegisterKeyResponse, error) {
	// Generate server ID and key
	serverID := utils.GenerateServerID()
	serverKey := utils.GenerateServerKey()

	s.logger.WithFields(logrus.Fields{
		"hostname":         req.Hostname,
		"operating_system": req.OperatingSystem,
		"agent_version":    req.AgentVersion,
	}).Info("Registering new server key")

	// Store in database (without secret_key)
	if err := s.storage.InsertGeneratedKeyWithIDs(ctx, "", serverID, serverKey, req.AgentVersion, req.OperatingSystem, req.Hostname); err != nil {
		s.logger.WithError(err).WithField("server_id", serverID).Error("Failed to insert generated key with IDs")
		return nil, fmt.Errorf("failed to register key: %w", err)
	}

	response := &models.RegisterKeyResponse{
		ServerID:  serverID,
		ServerKey: serverKey,
		Status:    "registered",
	}

	s.logger.WithFields(logrus.Fields{
		"server_id":        serverID,
		"agent_version":    req.AgentVersion,
		"operating_system": req.OperatingSystem,
		"hostname":         req.Hostname,
	}).Info("Server key registered successfully")

	return response, nil
}
