package websocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Server represents the WebSocket server
type Server struct {
	upgrader websocket.Upgrader
	clients  map[string]*Client
	mutex    sync.RWMutex
	storage  storage.Storage
	logger   *logrus.Logger
}

// NewServer creates a new WebSocket server
func NewServer(storage storage.Storage, logger *logrus.Logger) *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
		clients: make(map[string]*Client),
		storage: storage,
		logger:  logger,
	}
}

// HandleConnection handles WebSocket connection requests
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade connection")
		return
	}

	client := NewClient(conn, s.logger)
	go s.handleClient(client)
}

// handleClient handles a WebSocket client
func (s *Server) handleClient(client *Client) {
	defer client.Close()

	// Wait for authentication
	authMsg, err := client.ReadMessage()
	if err != nil {
		s.logger.WithError(err).Error("Failed to read auth message")
		return
	}

	if authMsg.Type != models.WSMessageTypeAuth {
		client.SendMessage(models.WSMessage{
			Type: models.WSMessageTypeError,
			Data: map[string]interface{}{
				"error": "Authentication required",
			},
		})
		return
	}

	// Validate authentication
	if !s.authenticate(authMsg.ServerID, authMsg.ServerKey) {
		client.SendMessage(models.WSMessage{
			Type: models.WSMessageTypeError,
			Data: map[string]interface{}{
				"error": "Invalid credentials",
			},
		})
		return
	}

	// Register client
	client.ServerID = authMsg.ServerID
	client.ServerKey = authMsg.ServerKey
	client.IsAgent = true

	s.mutex.Lock()
	s.clients[client.ServerID] = client
	s.mutex.Unlock()

	s.logger.WithField("server_id", client.ServerID).Info("WebSocket client connected")

	// Send success message
	client.SendMessage(models.WSMessage{
		Type: models.WSMessageTypeAuthSuccess,
		Data: map[string]interface{}{
			"server_id": client.ServerID,
		},
	})

	// Handle messages
	ctx := context.Background()
	for {
		msg, err := client.ReadMessage()
		if err != nil {
			s.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to read message")
			break
		}

		s.handleMessage(ctx, client, msg)
	}

	// Unregister client
	s.mutex.Lock()
	delete(s.clients, client.ServerID)
	s.mutex.Unlock()

	s.logger.WithField("server_id", client.ServerID).Info("WebSocket client disconnected")
}

// authenticate validates server credentials
func (s *Server) authenticate(serverID, serverKey string) bool {
	// For now, just validate format
	// In production, check against database
	if len(serverID) < 5 || len(serverKey) < 10 {
		return false
	}
	return true
}

// handleMessage handles incoming WebSocket messages
func (s *Server) handleMessage(ctx context.Context, client *Client, msg models.WSMessage) {
	switch msg.Type {
	case models.WSMessageTypeMetrics:
		s.handleMetrics(ctx, client, msg)
	case models.WSMessageTypeHeartbeat:
		s.handleHeartbeat(ctx, client, msg)
	default:
		s.logger.WithField("type", msg.Type).Warn("Unknown message type")
	}
}

// handleMetrics handles metrics messages
func (s *Server) handleMetrics(ctx context.Context, client *Client, msg models.WSMessage) {
	if msg.Data == nil {
		return
	}

	// Store metrics in Redis
	if err := s.storage.StoreMetric(ctx, client.ServerID, msg.Data); err != nil {
		s.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to store metrics")
	}

	s.logger.WithField("server_id", client.ServerID).Debug("Stored metrics from WebSocket")
}

// handleHeartbeat handles heartbeat messages
func (s *Server) handleHeartbeat(ctx context.Context, client *Client, msg models.WSMessage) {
	// Update server status
	status := map[string]interface{}{
		"online":    true,
		"last_seen": time.Now().Unix(),
	}

	if err := s.storage.SetServerStatus(ctx, client.ServerID, status); err != nil {
		s.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to update server status")
	}
}

// BroadcastMessage sends a message to all connected clients
func (s *Server) BroadcastMessage(msg models.WSMessage) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, client := range s.clients {
		client.SendMessage(msg)
	}
}

// SendToClient sends a message to a specific client
func (s *Server) SendToClient(serverID string, msg models.WSMessage) bool {
	s.mutex.RLock()
	client, exists := s.clients[serverID]
	s.mutex.RUnlock()

	if !exists {
		return false
	}

	return client.SendMessage(msg)
}
