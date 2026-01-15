package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// MetricsHandler handles metrics requests
type MetricsHandler struct {
	metricsService *services.MetricsService
	logger         *logrus.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(metricsService *services.MetricsService, logger *logrus.Logger) *MetricsHandler {
	return &MetricsHandler{
		metricsService: metricsService,
		logger:         logger,
	}
}

// GetServerMetrics handles GET /api/servers/{server_id}/metrics
func (h *MetricsHandler) GetServerMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		h.writeError(w, "server_id is required", http.StatusBadRequest)
		return
	}

	// Get complete metrics with status
	response, err := h.metricsService.GetServerMetricsWithStatus(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get server metrics")
		h.writeError(w, "Failed to get server metrics", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetServerMetricsByKey handles GET /api/servers/by-key/{server_key}/metrics
func (h *MetricsHandler) GetServerMetricsByKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		h.writeError(w, "server_key is required", http.StatusBadRequest)
		return
	}

	// Get server info by key first
	serverInfo, err := h.metricsService.GetServerByKey(r.Context(), serverKey)
	if err != nil {
		h.logger.WithError(err).WithField("server_key", serverKey).Error("Failed to get server by key")
		h.writeError(w, "Server not found", http.StatusNotFound)
		return
	}

	// Get complete metrics with status using server_id
	response, err := h.metricsService.GetServerMetricsWithStatus(r.Context(), serverInfo.ServerID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverInfo.ServerID).Error("Failed to get server metrics")
		h.writeError(w, "Failed to get server metrics", http.StatusInternalServerError)
		return
	}

	// Add server_key to response for convenience
	responseMap := response
	responseMap["server_key"] = serverKey
	h.writeJSON(w, http.StatusOK, responseMap)
}

// writeJSON writes JSON response
func (h *MetricsHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func (h *MetricsHandler) writeError(w http.ResponseWriter, message string, status int) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
