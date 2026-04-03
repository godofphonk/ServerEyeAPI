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
	"net"
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

	// Set read deadline for authentication
	authDeadline := time.Now().Add(30 * time.Second) // 30 seconds timeout for auth
	if err := client.conn.SetReadDeadline(authDeadline); err != nil {
		s.logger.WithError(err).Error("Failed to set auth read deadline")
		return
	}

	// Wait for authentication
	authMsg, err := client.ReadMessage()
	if err != nil {
		s.logger.WithError(err).Error("Failed to read auth message")
		if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			s.logger.Info("Client disconnected during auth")
		} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			s.logger.Error("Authentication timeout - client did not send auth message")
		}
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

	// Reset read deadline to normal value after successful auth
	if err := client.conn.SetReadDeadline(time.Now().Add(s.config.WebSocket.PongWait)); err != nil {
		s.logger.WithError(err).Error("Failed to reset read deadline after auth")
		return
	}

	// Send success message
	client.SendMessage(models.WSMessage{
		Type: models.WSMessageTypeAuthSuccess,
		Data: map[string]interface{}{
			"server_id": client.ServerID,
		},
	})

	s.logger.WithField("server_id", client.ServerID).Info("✅ Authentication successful, starting message handling")

	// Create channels for non-blocking message handling
	messageChan := make(chan models.WSMessage, 10)
	errorChan := make(chan error, 1)

	// Start message reader in separate goroutine
	go func() {
		defer close(messageChan)
		defer close(errorChan)

		for {
			// Set read deadline before each read attempt (less aggressive)
			deadline := time.Now().Add(s.config.WebSocket.PongWait)
			if err := client.conn.SetReadDeadline(deadline); err != nil {
				s.logger.WithFields(logrus.Fields{
					"server_id":  client.ServerID,
					"error":      err.Error(),
					"error_type": "read_deadline_failed",
				}).Error("Failed to set read deadline")
				errorChan <- err
				return
			}

			s.logger.WithFields(logrus.Fields{
				"server_id": client.ServerID,
				"deadline":  deadline.Format("15:04:05"),
			}).Debug("Waiting for WebSocket message")

			msg, err := client.ReadMessage()
			if err != nil {
				// Better error classification
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					s.logger.WithFields(logrus.Fields{
						"server_id":  client.ServerID,
						"error":      err.Error(),
						"error_type": "normal_close",
					}).Info("WebSocket closed normally")
					errorChan <- err
					return
				} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Don't treat timeout as fatal error, just log and continue
					s.logger.WithFields(logrus.Fields{
						"server_id":  client.ServerID,
						"error":      err.Error(),
						"error_type": "read_timeout",
					}).Debug("WebSocket read timeout, continuing...")
					// Reset read deadline and continue
					continue
				} else {
					s.logger.WithFields(logrus.Fields{
						"server_id":  client.ServerID,
						"error":      err.Error(),
						"error_type": "read_error",
					}).Error("WebSocket read error")
					errorChan <- err
					return
				}
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
			// Check if connection is still active before sending ping
			client.mutex.RLock()
			if client.closed {
				client.mutex.RUnlock()
				s.logger.WithField("server_id", client.ServerID).Info("Client already closed, stopping ping")
				return
			}
			client.mutex.RUnlock()

			// Update write deadline before sending ping
			if err := client.conn.SetWriteDeadline(time.Now().Add(s.config.WebSocket.WriteTimeout)); err != nil {
				s.logger.WithFields(logrus.Fields{
					"server_id":  client.ServerID,
					"error":      err.Error(),
					"error_type": "write_deadline_failed",
				}).Error("Failed to set write deadline")
				return
			}

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
			// Handle incoming message in separate goroutine to avoid blocking
			go func() {
				defer func() {
					if r := recover(); r != nil {
						s.logger.WithFields(logrus.Fields{
							"server_id": client.ServerID,
							"panic":     r,
						}).Error("Recovered from panic in message handler")
					}
				}()
				s.handleMessage(ctx, client, msg)
			}()
			// Continue processing - don't return here

		case err := <-errorChan:
			// Handle read error with better classification
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.WithFields(logrus.Fields{
					"server_id":  client.ServerID,
					"error":      err.Error(),
					"error_type": "unexpected_close",
				}).Error("WebSocket closed unexpectedly")
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				s.logger.WithFields(logrus.Fields{
					"server_id":  client.ServerID,
					"error":      err.Error(),
					"error_type": "read_timeout",
				}).Warn("WebSocket read timeout, but connection may still be alive")
				// Don't return on timeout, let ping mechanism handle it
				continue
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
	}).Debug("Processing WebSocket message")

	switch msg.Type {
	case models.WSMessageTypeMetrics:
		s.logger.WithFields(logrus.Fields{
			"server_id":    client.ServerID,
			"message_type": msg.Type,
		}).Debug("📊 Processing metrics message")
		s.handleMetrics(ctx, client, msg)
	case models.WSMessageTypeHeartbeat:
		s.logger.WithFields(logrus.Fields{
			"server_id":    client.ServerID,
			"message_type": msg.Type,
		}).Debug("💓 Processing heartbeat message")
		s.handleHeartbeat(ctx, client, msg)
	case models.WSMessageTypeAuth:
		s.logger.WithField("server_id", client.ServerID).Warn("🔐 Received duplicate auth message")
		// Ignore duplicate auth messages
	default:
		s.logger.WithFields(logrus.Fields{
			"server_id":    client.ServerID,
			"unknown_type": msg.Type,
		}).Warn("❓ Unknown message type")
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

	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		s.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to marshal metrics data")
		return
	}

	// Try to parse as new format first (MetricsV2)
	// Agent sends: {"type": "metrics", "server_id": "...", "data": {"metrics": {"metrics": {...}}}}
	var nestedMsg struct {
		Metrics models.MetricsV2 `json:"metrics"`
	}

	parseErr := json.Unmarshal(dataBytes, &nestedMsg)
	var newMetricsMsg struct {
		Metrics models.MetricsV2 `json:"metrics"`
	}

	if parseErr == nil && !nestedMsg.Metrics.Timestamp.IsZero() {
		// Successfully parsed nested structure
		newMetricsMsg.Metrics = nestedMsg.Metrics
		parseErr = nil
	} else {
		// Fallback: try direct structure (data.metrics)
		var directMsg struct {
			Metrics models.MetricsV2 `json:"metrics"`
		}
		if directErr := json.Unmarshal(dataBytes, &directMsg); directErr == nil && !directMsg.Metrics.Timestamp.IsZero() {
			newMetricsMsg = directMsg
			parseErr = nil
		}
	}

	if parseErr == nil && !newMetricsMsg.Metrics.Timestamp.IsZero() {
		s.logger.WithFields(logrus.Fields{
			"server_id":   client.ServerID,
			"format":      "v2",
			"cpu_total":   newMetricsMsg.Metrics.CPUUsage.UsageTotal,
			"memory_used": newMetricsMsg.Metrics.Memory.UsedPercent,
			"temperature": newMetricsMsg.Metrics.Temperature.Highest,
		}).Info("📊 Using new metrics format (V2)")

		// Convert V2 to old format for storage compatibility
		oldMetrics := s.convertV2ToOldFormat(&newMetricsMsg.Metrics)
		oldMetrics.Time = newMetricsMsg.Metrics.Timestamp

		s.logger.WithFields(logrus.Fields{
			"server_id":   client.ServerID,
			"cpu":         oldMetrics.CPU,
			"memory":      oldMetrics.Memory,
			"temperature": oldMetrics.TemperatureDetails.HighestTemperature,
		}).Info("Storing V2 metrics")

		// Store converted metrics
		if err := s.storage.StoreMetric(ctx, client.ServerID, oldMetrics); err != nil {
			s.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to store V2 metrics")
			return
		}

		s.logger.WithField("server_id", client.ServerID).Info("✅ Successfully stored V2 metrics")
		return
	}

	// Fallback to old format
	var metricsMsg models.MetricsMessage
	if err := json.Unmarshal(dataBytes, &metricsMsg); err != nil {
		s.logger.WithError(err).WithField("server_id", client.ServerID).Error("Invalid metrics message format")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": metricsMsg.ServerID,
		"cpu":       metricsMsg.Metrics.CPU,
		"memory":    metricsMsg.Metrics.Memory,
		"disk":      metricsMsg.Metrics.Disk,
		"format":    "v1",
	}).Info("📊 Using old metrics format (V1)")

	// Store metrics
	if err := s.storage.StoreMetric(ctx, client.ServerID, &metricsMsg.Metrics); err != nil {
		s.logger.WithError(err).WithField("server_id", client.ServerID).Error("Failed to store metrics")
		return
	}

	s.logger.WithField("server_id", client.ServerID).Info("✅ Successfully stored V1 metrics")
}

// handleHeartbeat handles heartbeat messages
func (s *Server) handleHeartbeat(ctx context.Context, client *Client, msg models.WSMessage) {
	// Update server status
	if err := s.storage.SetServerStatus(ctx, client.ServerID, "online"); err != nil {
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

// convertV2ToOldFormat converts new MetricsV2 format to old ServerMetrics format
func (s *Server) convertV2ToOldFormat(v2 *models.MetricsV2) *models.ServerMetrics {
	old := &models.ServerMetrics{}

	// Aggregated values for backward compatibility
	old.CPU = v2.CPUUsage.UsageTotal
	old.Memory = v2.Memory.UsedPercent

	// Calculate average disk usage
	if len(v2.Disks) > 0 {
		var totalDiskUsage float64
		for _, disk := range v2.Disks {
			totalDiskUsage += disk.UsedPercent
		}
		old.Disk = totalDiskUsage / float64(len(v2.Disks))
	}

	// Calculate total network traffic in MB
	var totalRxMB, totalTxMB float64
	for _, iface := range v2.Network.Interfaces {
		totalRxMB += float64(iface.RxBytes) / 1024 / 1024
		totalTxMB += float64(iface.TxBytes) / 1024 / 1024
	}
	old.Network = totalRxMB + totalTxMB

	// CPU detailed metrics
	old.CPUUsage.UsageTotal = v2.CPUUsage.UsageTotal
	old.CPUUsage.UsageUser = v2.CPUUsage.UsageUser
	old.CPUUsage.UsageSystem = v2.CPUUsage.UsageSystem
	old.CPUUsage.UsageIdle = v2.CPUUsage.UsageIdle
	old.CPUUsage.LoadAverage.Load1 = v2.CPUUsage.LoadAverage.Load1Min
	old.CPUUsage.LoadAverage.Load5 = v2.CPUUsage.LoadAverage.Load5Min
	old.CPUUsage.LoadAverage.Load15 = v2.CPUUsage.LoadAverage.Load15Min
	old.CPUUsage.Frequency = v2.CPUUsage.FrequencyMHz

	// Memory detailed metrics
	old.MemoryDetails.TotalGB = v2.Memory.TotalGB
	old.MemoryDetails.UsedGB = v2.Memory.UsedGB
	old.MemoryDetails.AvailableGB = v2.Memory.AvailableGB
	old.MemoryDetails.FreeGB = v2.Memory.FreeGB
	old.MemoryDetails.BuffersGB = v2.Memory.BuffersGB
	old.MemoryDetails.CachedGB = v2.Memory.CachedGB
	old.MemoryDetails.UsedPercent = v2.Memory.UsedPercent

	// Disk detailed metrics
	if len(v2.Disks) > 0 {
		old.DiskDetails = make([]struct {
			Path        string  `json:"path"`
			TotalGB     float64 `json:"total_gb"`
			UsedGB      float64 `json:"used_gb"`
			FreeGB      float64 `json:"free_gb"`
			UsedPercent float64 `json:"used_percent"`
			Filesystem  string  `json:"filesystem"`
		}, len(v2.Disks))
		for i, disk := range v2.Disks {
			old.DiskDetails[i].Path = disk.MountPoint
			old.DiskDetails[i].UsedGB = disk.UsedGB
			old.DiskDetails[i].FreeGB = disk.FreeGB
			old.DiskDetails[i].UsedPercent = disk.UsedPercent
			old.DiskDetails[i].TotalGB = disk.UsedGB + disk.FreeGB
		}
	}

	// Network detailed metrics
	if len(v2.Network.Interfaces) > 0 {
		old.NetworkDetails.Interfaces = make([]struct {
			Name        string  `json:"name"`
			RxBytes     int64   `json:"rx_bytes"`
			TxBytes     int64   `json:"tx_bytes"`
			RxPackets   int64   `json:"rx_packets"`
			TxPackets   int64   `json:"tx_packets"`
			RxSpeedMbps float64 `json:"rx_speed_mbps"`
			TxSpeedMbps float64 `json:"tx_speed_mbps"`
			Status      string  `json:"status"`
		}, len(v2.Network.Interfaces))
		for i, iface := range v2.Network.Interfaces {
			old.NetworkDetails.Interfaces[i].Name = iface.Name
			old.NetworkDetails.Interfaces[i].RxBytes = iface.RxBytes
			old.NetworkDetails.Interfaces[i].TxBytes = iface.TxBytes
			old.NetworkDetails.Interfaces[i].RxPackets = iface.RxPackets
			old.NetworkDetails.Interfaces[i].TxPackets = iface.TxPackets
			old.NetworkDetails.Interfaces[i].RxSpeedMbps = iface.RxSpeedMbps
			old.NetworkDetails.Interfaces[i].TxSpeedMbps = iface.TxSpeedMbps
			old.NetworkDetails.Interfaces[i].Status = iface.Status
		}
		old.NetworkDetails.TotalRxMbps = v2.Network.TotalRxMbps
		old.NetworkDetails.TotalTxMbps = v2.Network.TotalTxMbps
		old.Network = v2.Network.TotalRxMbps + v2.Network.TotalTxMbps // Use total as aggregate
	}

	// Temperature metrics
	old.TemperatureDetails.CPUTemperature = v2.Temperature.CPU
	old.TemperatureDetails.GPUTemperature = v2.Temperature.GPU
	old.TemperatureDetails.HighestTemperature = v2.Temperature.Highest

	if len(v2.Temperature.Storage) > 0 {
		old.TemperatureDetails.StorageTemperatures = make([]struct {
			Device      string  `json:"device"`
			Type        string  `json:"type"`
			Temperature float64 `json:"temperature"`
		}, len(v2.Temperature.Storage))

		for i, storage := range v2.Temperature.Storage {
			old.TemperatureDetails.StorageTemperatures[i].Device = storage.Device
			old.TemperatureDetails.StorageTemperatures[i].Temperature = storage.Temperature
		}
	}

	// System metrics
	old.SystemDetails.ProcessesTotal = v2.System.ProcessesTotal
	old.SystemDetails.ProcessesRunning = v2.System.ProcessesRunning
	old.SystemDetails.ProcessesSleeping = v2.System.ProcessesSleeping
	old.SystemDetails.UptimeSeconds = v2.System.UptimeSeconds

	return old
}
