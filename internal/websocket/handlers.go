package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/sirupsen/logrus"
)

// MessageHandlers handles different types of WebSocket messages
type MessageHandlers struct {
	storage storage.Storage
	logger  *logrus.Logger
}

// NewMessageHandlers creates new message handlers
func NewMessageHandlers(storage storage.Storage, logger *logrus.Logger) *MessageHandlers {
	return &MessageHandlers{
		storage: storage,
		logger:  logger,
	}
}

// HandleAuth handles authentication messages
func (h *MessageHandlers) HandleAuth(ctx context.Context, client *Client, msg models.WSMessage) error {
	// Authentication is handled in the main server
	return nil
}

// HandleMetrics handles metrics messages from agents
func (h *MessageHandlers) HandleMetrics(ctx context.Context, client *Client, msg models.WSMessage) error {
	if msg.Data == nil {
		h.logger.WithField("server_id", client.ServerID).Warn("Metrics message has no data")
		return nil
	}

	// Parse metrics message
	var metricsMsg models.MetricsMessage
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to marshal metrics data")
		return err
	}

	if err := json.Unmarshal(dataBytes, &metricsMsg); err != nil {
		h.logger.WithError(err).WithField("server_id", client.ServerID).Error("Invalid metrics message format")
		return err
	}

	// Store metrics in Redis
	if err := h.storage.StoreMetric(ctx, client.ServerID, &metricsMsg.Metrics); err != nil {
		h.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to store metrics")
		return err
	}

	h.logger.WithField("server_id", client.ServerID).Debug("Stored metrics from WebSocket")
	return nil
}

// HandleHeartbeat handles heartbeat messages from agents
func (h *MessageHandlers) HandleHeartbeat(ctx context.Context, client *Client, msg models.WSMessage) error {
	// Update server status in Redis
	status := &models.ServerStatus{
		Online:   true,
		LastSeen: time.Unix(msg.Timestamp, 0),
	}

	if err := h.storage.SetServerStatus(ctx, client.ServerID, status); err != nil {
		h.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to update server status")
		return err
	}

	h.logger.WithField("server_id", client.ServerID).Debug("Updated server heartbeat")
	return nil
}

// HandleCommand handles command responses from agents
func (h *MessageHandlers) HandleCommand(ctx context.Context, client *Client, msg models.WSMessage) error {
	if msg.Data == nil {
		h.logger.WithField("server_id", client.ServerID).Warn("Command message has no data")
		return nil
	}

	// Store command response
	h.logger.WithFields(logrus.Fields{
		"server_id": client.ServerID,
		"command":   msg.Data,
	}).Info("Received command response from agent")

	return nil
}
