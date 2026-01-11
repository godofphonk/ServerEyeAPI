package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// RegisterKeyRequest represents request to register a key
type RegisterKeyRequest struct {
	SecretKey       string `json:"secret_key"`
	AgentVersion    string `json:"agent_version"`
	OperatingSystem string `json:"operating_system"`
	Hostname        string `json:"hostname"`
}

// RegisterKeyResponse represents response after successful registration
type RegisterKeyResponse struct {
	ServerID  string `json:"server_id"`
	ServerKey string `json:"server_key"`
	Status    string `json:"status"`
}

// handleRegisterKey handles agent registration
func (s *Server) handleRegisterKey(w http.ResponseWriter, r *http.Request) {
	var req RegisterKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SecretKey == "" {
		s.writeError(w, "secret_key is required", http.StatusBadRequest)
		return
	}

	// Generate server ID and key
	serverID := generateServerID()
	serverKey := generateServerKey()

	s.logger.WithFields(logrus.Fields{
		"secret_key":       req.SecretKey,
		"hostname":         req.Hostname,
		"operating_system": req.OperatingSystem,
		"agent_version":    req.AgentVersion,
	}).Info("Attempting to register key")

	// Store in database
	if err := s.storage.InsertGeneratedKey(r.Context(), req.SecretKey, req.AgentVersion, req.OperatingSystem, req.Hostname); err != nil {
		s.logger.WithError(err).WithField("secret_key", req.SecretKey).Error("Failed to insert generated key")
		s.writeError(w, "Failed to register key", http.StatusInternalServerError)
		return
	}

	// Store mapping for authentication
	// In a real implementation, you'd store this mapping in Redis or database
	// For now, we'll generate and return

	response := RegisterKeyResponse{
		ServerID:  serverID,
		ServerKey: serverKey,
		Status:    "registered",
	}

	s.logger.WithFields(logrus.Fields{
		"server_id":        serverID,
		"agent_version":    req.AgentVersion,
		"operating_system": req.OperatingSystem,
		"hostname":         req.Hostname,
	}).Info("Agent registered")

	s.writeJSON(w, http.StatusCreated, response)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := s.storage.Ping(); err != nil {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"clients":   s.wsServer.GetClientCount(),
	})
}

// Helper functions

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, message string, status int) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

func generateServerID() string {
	return "srv_" + randomString(12)
}

func generateServerKey() string {
	return "key_" + randomString(24)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// handleGetServerMetrics returns latest metrics for a specific server
func (s *Server) handleGetServerMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		s.writeError(w, "server_id is required", http.StatusBadRequest)
		return
	}

	// Try to get from Redis first
	if s.redisStorage != nil {
		metrics, err := s.redisStorage.GetMetric(r.Context(), serverID)
		if err != nil {
			s.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get metrics from Redis")
		} else {
			s.writeJSON(w, http.StatusOK, map[string]interface{}{
				"server_id": serverID,
				"metrics":   metrics,
				"timestamp": time.Now(),
			})
			return
		}
	}

	// Fallback to PostgreSQL
	metrics, err := s.storage.GetLatestMetrics(r.Context(), serverID)
	if err != nil {
		s.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get metrics")
		s.writeError(w, "Failed to get server metrics", http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"server_id": serverID,
		"metrics":   metrics,
		"timestamp": time.Now(),
	})
}

// handleSendCommand sends a command to a specific server
func (s *Server) handleSendCommand(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServerID string                 `json:"server_id"`
		Command  map[string]interface{} `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ServerID == "" {
		s.writeError(w, "server_id is required", http.StatusBadRequest)
		return
	}

	if len(req.Command) == 0 {
		s.writeError(w, "command is required", http.StatusBadRequest)
		return
	}

	// Store command in Redis
	if s.redisStorage != nil {
		err := s.redisStorage.StoreCommand(r.Context(), req.ServerID, req.Command)
		if err != nil {
			s.logger.WithError(err).WithField("server_id", req.ServerID).Error("Failed to store command in Redis")
		}
	}

	// Send command via WebSocket to agent
	s.wsServer.BroadcastCommand(req.ServerID, req.Command)

	s.logger.WithFields(logrus.Fields{
		"server_id": req.ServerID,
		"command":   req.Command,
	}).Info("Command sent to server")

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Command sent successfully",
		"server_id": req.ServerID,
		"command":   req.Command,
		"timestamp": time.Now(),
	})
}

// handleListServers returns all registered servers with their status
func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	var servers []map[string]interface{}

	// Try to get from Redis first
	if s.redisStorage != nil {
		redisServers, err := s.redisStorage.GetAllServers(r.Context())
		if err != nil {
			s.logger.WithError(err).Error("Failed to get servers from Redis")
		} else {
			servers = redisServers
		}
	}

	// If no servers in Redis, try to get from database
	if len(servers) == 0 {
		dbServers, err := s.storage.GetServers(r.Context())
		if err != nil {
			s.logger.WithError(err).Error("Failed to get servers from database")
			s.writeError(w, "Failed to get servers", http.StatusInternalServerError)
			return
		}

		for _, serverID := range dbServers {
			servers = append(servers, map[string]interface{}{
				"server_id": serverID,
				"status":    map[string]interface{}{"online": false},
			})
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"servers":   servers,
		"count":     len(servers),
		"timestamp": time.Now(),
	})
}

// handleGetServerStatus returns status of a specific server
func (s *Server) handleGetServerStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		s.writeError(w, "server_id is required", http.StatusBadRequest)
		return
	}

	// Try to get from Redis first
	if s.redisStorage != nil {
		status, err := s.redisStorage.GetServerStatus(r.Context(), serverID)
		if err != nil {
			s.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get server status from Redis")
		} else {
			s.writeJSON(w, http.StatusOK, map[string]interface{}{
				"server_id": serverID,
				"status":    status,
				"timestamp": time.Now(),
			})
			return
		}
	}

	// Default status if not found
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"server_id": serverID,
		"status":    map[string]interface{}{"online": false},
		"timestamp": time.Now(),
	})
}
