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

// RegisterKey handles server registration
func (h *AuthHandler) RegisterKey(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SecretKey == "" {
		h.writeError(w, "secret_key is required", http.StatusBadRequest)
		return
	}

	response, err := h.authService.RegisterKey(r.Context(), &req)
	if err != nil {
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
