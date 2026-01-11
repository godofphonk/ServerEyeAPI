package api

import (
	"encoding/json"
	"net/http"
	"time"

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
