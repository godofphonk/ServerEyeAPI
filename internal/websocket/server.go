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

package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/config"
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
	config   *config.Config
}

// NewServer creates a new WebSocket server
func NewServer(storage storage.Storage, logger *logrus.Logger, cfg *config.Config) *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
		clients: make(map[string]*Client),
		storage: storage,
		logger:  logger,
		config:  cfg,
	}
}

// HandleConnection handles WebSocket connection requests
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
	s.logger.WithFields(logrus.Fields{
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.UserAgent(),
		"method":      r.Method,
		"url":         r.URL.String(),
	}).Info("WebSocket connection request received")

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade connection")
		return
	}

	s.logger.WithField("remote_addr", r.RemoteAddr).Info("WebSocket connection established")

	client := NewClient(conn, s.logger, s.config)
	go s.handleClient(client)
}

// handleClient handles a WebSocket client
func (s *Server) handleClient(client *Client) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.WithFields(logrus.Fields{
				"server_id": client.ServerID,
				"panic":     r,
			}).Error("Recovered from panic in WebSocket client handler")
		}

		// Unregister client
		s.mutex.Lock()
		delete(s.clients, client.ServerID)
		s.mutex.Unlock()

		s.logger.WithField("server_id", client.ServerID).Info("WebSocket client disconnected")
		client.Close()
	}()

	s.logger.Info("Waiting for authentication message...")

	// Wait for authentication
	authMsg, err := client.ReadMessage()
	if err != nil {
		s.logger.WithError(err).Error("Failed to read auth message")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"message_type": authMsg.Type,
		"server_id":    authMsg.ServerID,
		"has_key":      authMsg.ServerKey != "",
	}).Info("Received authentication message")

	if authMsg.Type != models.WSMessageTypeAuth {
		s.logger.WithField("received_type", authMsg.Type).Error("Invalid message type, expected auth")
		client.SendMessage(models.WSMessage{
			Type: models.WSMessageTypeError,
			Data: map[string]interface{}{
				"error": "Authentication required",
			},
		})
		return
	}

	// Validate authentication
	s.logger.WithFields(logrus.Fields{
		"server_id":  authMsg.ServerID,
		"server_key": authMsg.ServerKey[:10] + "...", // Log only first 10 chars for security
	}).Info("Validating authentication")

	if !s.authenticate(authMsg.ServerID, authMsg.ServerKey) {
		s.logger.WithField("server_id", authMsg.ServerID).Error("Authentication failed")
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

	// Create channels for non-blocking message handling
	messageChan := make(chan models.WSMessage, 10)
	errorChan := make(chan error, 1)

	// Start message reader in separate goroutine
	go func() {
		defer close(messageChan)
		defer close(errorChan)

		for {
			// Set read deadline before each read attempt
			deadline := time.Now().Add(s.config.WebSocket.PongWait)
			client.conn.SetReadDeadline(deadline)

			s.logger.WithFields(logrus.Fields{
				"server_id": client.ServerID,
				"deadline":  deadline.Format("15:04:05"),
			}).Debug("Waiting for WebSocket message")

			msg, err := client.ReadMessage()
			if err != nil {
				s.logger.WithFields(logrus.Fields{
					"server_id": client.ServerID,
					"error":     err.Error(),
				}).Info("WebSocket read error, exiting reader")
				errorChan <- err
				return
			}

			s.logger.WithFields(logrus.Fields{
				"server_id":    client.ServerID,
				"message_type": msg.Type,
			}).Debug("Received message in reader")
			messageChan <- msg
		}
	}()

	// Start ping ticker
	pingTicker := time.NewTicker(s.config.WebSocket.PingInterval)
	defer pingTicker.Stop()

	// Main event loop
	ctx := context.Background()
	for {
		select {
		case <-pingTicker.C:
			// Send ping to keep connection alive
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				s.logger.WithFields(logrus.Fields{
					"server_id":  client.ServerID,
					"error":      err.Error(),
					"error_type": "ping_failed",
				}).Error("Failed to send ping, connection unstable")
				return
			}
			s.logger.WithField("server_id", client.ServerID).Debug("Sent ping to client")

		case msg := <-messageChan:
			// Handle incoming message
			s.handleMessage(ctx, client, msg)

		case err := <-errorChan:
			// Handle read error with better classification
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.WithFields(logrus.Fields{
					"server_id":  client.ServerID,
					"error":      err.Error(),
					"error_type": "unexpected_close",
				}).Error("WebSocket closed unexpectedly")
			} else {
				s.logger.WithFields(logrus.Fields{
					"server_id":  client.ServerID,
					"error":      err.Error(),
					"error_type": "read_error",
				}).Error("WebSocket read error, disconnecting")
			}
			return
		}
	}
}

// authenticate validates server credentials
func (s *Server) authenticate(serverID, serverKey string) bool {
	// Check server key in database
	serverInfo, err := s.storage.GetServerByKey(context.Background(), serverKey)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"server_key": serverKey,
			"error":      err.Error(),
		}).Warn("Authentication failed: key not found")
		return false
	}

	// Verify server ID matches
	if serverInfo.ServerID != serverID {
		s.logger.WithFields(logrus.Fields{
			"server_id":  serverID,
			"stored_id":  serverInfo.ServerID,
			"server_key": serverKey,
			"hostname":   serverInfo.Hostname,
		}).Warn("Authentication failed: server ID mismatch")
		return false
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"hostname":  serverInfo.Hostname,
	}).Info("WebSocket authentication successful")

	return true
}

// handleMessage handles incoming WebSocket messages
func (s *Server) handleMessage(ctx context.Context, client *Client, msg models.WSMessage) {
	s.logger.WithFields(logrus.Fields{
		"server_id":    client.ServerID,
		"message_type": msg.Type,
		"has_data":     msg.Data != nil,
		"timestamp":    msg.Timestamp,
	}).Info("Received WebSocket message")

	switch msg.Type {
	case models.WSMessageTypeMetrics:
		s.logger.WithFields(logrus.Fields{
			"server_id":    client.ServerID,
			"message_type": msg.Type,
		}).Info("Processing metrics message")
		s.handleMetrics(ctx, client, msg)
	case models.WSMessageTypeHeartbeat:
		s.logger.WithFields(logrus.Fields{
			"server_id":    client.ServerID,
			"message_type": msg.Type,
		}).Info("Processing heartbeat message")
		s.handleHeartbeat(ctx, client, msg)
	default:
		s.logger.WithFields(logrus.Fields{
			"server_id":    client.ServerID,
			"unknown_type": msg.Type,
		}).Warn("Unknown message type")
	}
}

// handleMetrics handles metrics messages
func (s *Server) handleMetrics(ctx context.Context, client *Client, msg models.WSMessage) {
	s.logger.WithFields(logrus.Fields{
		"server_id":    client.ServerID,
		"message_type": msg.Type,
	}).Info("Starting handleMetrics")

	if msg.Data == nil {
		s.logger.WithField("server_id", client.ServerID).Warn("Metrics message has no data")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": client.ServerID,
		"data_keys": len(msg.Data),
	}).Info("Metrics message has data, parsing...")

	// Parse metrics message
	var metricsMsg models.MetricsMessage
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		s.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to marshal metrics data")
		return
	}

	if err := json.Unmarshal(dataBytes, &metricsMsg); err != nil {
		s.logger.WithError(err).WithField("server_id", client.ServerID).Error("Invalid metrics message format")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": metricsMsg.ServerID,
		"cpu":       metricsMsg.Metrics.CPU,
		"memory":    metricsMsg.Metrics.Memory,
		"disk":      metricsMsg.Metrics.Disk,
	}).Info("Parsed metrics message, storing in Redis")

	// Store metrics in Redis
	if err := s.storage.StoreMetric(ctx, client.ServerID, &metricsMsg.Metrics); err != nil {
		s.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to store metrics")
		return
	}

	s.logger.WithField("server_id", client.ServerID).Info("✅ Successfully stored metrics from WebSocket")
}

// handleHeartbeat handles heartbeat messages
func (s *Server) handleHeartbeat(ctx context.Context, client *Client, msg models.WSMessage) {
	// Update server status
	status := &models.ServerStatus{
		Online:   true,
		LastSeen: time.Unix(msg.Timestamp, 0),
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

// Close gracefully shuts down the WebSocket server
func (s *Server) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Close all client connections
	for _, client := range s.clients {
		client.Close()
	}

	// Clear clients map
	s.clients = make(map[string]*Client)

	s.logger.Info("WebSocket server closed gracefully")
	return nil
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
