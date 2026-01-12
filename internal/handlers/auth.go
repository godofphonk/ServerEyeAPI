package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/sirupsen/logrus"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService *services.AuthService
	logger      *logrus.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// RegisterKey handles POST /RegisterKey
func (h *AuthHandler) RegisterKey(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Hostname == "" {
		h.writeError(w, "hostname is required", http.StatusBadRequest)
		return
	}
	if req.OperatingSystem == "" {
		h.writeError(w, "operating_system is required", http.StatusBadRequest)
		return
	}
	if req.AgentVersion == "" {
		h.writeError(w, "agent_version is required", http.StatusBadRequest)
		return
	}

	// Register key
	response, err := h.authService.RegisterKey(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to register key")
		h.writeError(w, "Failed to register key", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// writeJSON writes JSON response
func (h *AuthHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func (h *AuthHandler) writeError(w http.ResponseWriter, message string, status int) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
