package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/godofphonk/ServerEyeAPI/internal/version"
	"github.com/sirupsen/logrus"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	storage storage.Storage
	logger  *logrus.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(storage storage.Storage, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		storage: storage,
		logger:  logger,
	}
}

// Health handles health check requests
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := &models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   version.Version,
		Clients:   0, // TODO: Get actual WebSocket client count
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes JSON response
func (h *HealthHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
