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
