package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ServerSourcesHandler handles server source management requests
type ServerSourcesHandler struct {
	serverService *services.ServerService
	logger        *logrus.Logger
}

// NewServerSourcesHandler creates a new server sources handler
func NewServerSourcesHandler(serverService *services.ServerService, logger *logrus.Logger) *ServerSourcesHandler {
	return &ServerSourcesHandler{
		serverService: serverService,
		logger:        logger,
	}
}

// AddServerSource handles POST /api/servers/{server_id}/sources
func (h *ServerSourcesHandler) AddServerSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		h.writeError(w, "server_id is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Source string `json:"source"` // "TGBot" or "Web"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate source
	if req.Source != "TGBot" && req.Source != "Web" {
		h.writeError(w, "Source must be 'TGBot' or 'Web'", http.StatusBadRequest)
		return
	}

	// Add source
	err := h.serverService.AddServerSource(r.Context(), serverID, req.Source)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"server_id": serverID,
			"source":    req.Source,
		}).Error("Failed to add server source")
		h.writeError(w, "Failed to add server source", http.StatusInternalServerError)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"source":    req.Source,
	}).Info("Server source added successfully")

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"server_id": serverID,
		"source":    req.Source,
		"message":   "Source added successfully",
	})
}

// GetServerSources handles GET /api/servers/{server_id}/sources
func (h *ServerSourcesHandler) GetServerSources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		h.writeError(w, "server_id is required", http.StatusBadRequest)
		return
	}

	sources, err := h.serverService.GetServerSources(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get server sources")
		h.writeError(w, "Failed to get server sources", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"server_id": serverID,
		"sources":   sources,
	})
}

// RemoveServerSource handles DELETE /api/servers/{server_id}/sources/{source}
func (h *ServerSourcesHandler) RemoveServerSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	source := vars["source"]

	if serverID == "" {
		h.writeError(w, "server_id is required", http.StatusBadRequest)
		return
	}

	if source == "" {
		h.writeError(w, "source is required", http.StatusBadRequest)
		return
	}

	// Validate source
	if source != "TGBot" && source != "Web" {
		h.writeError(w, "Source must be 'TGBot' or 'Web'", http.StatusBadRequest)
		return
	}

	err := h.serverService.RemoveServerSource(r.Context(), serverID, source)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"server_id": serverID,
			"source":    source,
		}).Error("Failed to remove server source")
		h.writeError(w, "Failed to remove server source", http.StatusInternalServerError)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"source":    source,
	}).Info("Server source removed successfully")

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"server_id": serverID,
		"source":    source,
		"message":   "Source removed successfully",
	})
}

// writeJSON writes JSON response
func (h *ServerSourcesHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func (h *ServerSourcesHandler) writeError(w http.ResponseWriter, message string, status int) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
