package services

import (
	"context"
	"fmt"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/sirupsen/logrus"
)

// CommandsService handles command operations
type CommandsService struct {
	storage storage.Storage
	logger  *logrus.Logger
}

// NewCommandsService creates a new commands service
func NewCommandsService(storage storage.Storage, logger *logrus.Logger) *CommandsService {
	return &CommandsService{
		storage: storage,
		logger:  logger,
	}
}

// SendCommand sends a command to a server
func (s *CommandsService) SendCommand(ctx context.Context, req *models.SendCommandRequest) (*models.CommandResponse, error) {
	// Validate server exists
	servers, err := s.storage.GetServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers: %w", err)
	}

	serverExists := false
	for _, server := range servers {
		if server == req.ServerID {
			serverExists = true
			break
		}
	}

	if !serverExists {
		return nil, fmt.Errorf("server %s not found", req.ServerID)
	}

	// Store command in Redis for WebSocket delivery
	if err := s.storage.StoreCommand(ctx, req.ServerID, req.Command); err != nil {
		s.logger.WithError(err).WithField("server_id", req.ServerID).Error("Failed to store command")
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": req.ServerID,
		"command":   req.Command,
	}).Info("Command sent to server")

	response := &models.CommandResponse{
		Message:  "Command sent successfully",
		ServerID: req.ServerID,
		Command:  req.Command,
	}

	return response, nil
}
