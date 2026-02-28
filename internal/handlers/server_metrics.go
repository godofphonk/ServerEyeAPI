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
	"fmt"
	"net/http"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ServerMetricsHandler handles server metrics requests
type ServerMetricsHandler struct {
	logger  *logrus.Logger
	storage storage.Storage
}

// NewServerMetricsHandler creates a new server metrics handler
func NewServerMetricsHandler(logger *logrus.Logger, storage storage.Storage) *ServerMetricsHandler {
	return &ServerMetricsHandler{
		logger:  logger,
		storage: storage,
	}
}

// GetServerMetricsWithTemperatures returns server metrics with temperature details
func (h *ServerMetricsHandler) GetServerMetricsWithTemperatures(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		http.Error(w, "Server ID is required", http.StatusBadRequest)
		return
	}

	// Get latest metrics
	metrics, err := h.storage.GetMetric(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get server metrics")
		http.Error(w, "Failed to get server metrics", http.StatusInternalServerError)
		return
	}

	if metrics == nil {
		http.Error(w, "Server metrics not found", http.StatusNotFound)
		return
	}

	// Get server info
	serverInfo, err := h.storage.GetServer(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get server info")
		http.Error(w, "Failed to get server info", http.StatusInternalServerError)
		return
	}

	// Convert ServerStatus to ServerInfo for response
	serverInfoForResponse := &models.ServerInfo{
		ServerID: serverID, // Use serverID from URL parameter
		Hostname: serverInfo.Hostname,
	}

	// Build response with temperature details and alerts
	response := h.buildMetricsResponse(serverInfoForResponse, metrics)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetServerMetricsWithTemperaturesByKey returns server metrics by server key
func (h *ServerMetricsHandler) GetServerMetricsWithTemperaturesByKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		http.Error(w, "Server key is required", http.StatusBadRequest)
		return
	}

	// Get server info by key
	serverInfo, err := h.storage.GetServerByKey(r.Context(), serverKey)
	if err != nil {
		h.logger.WithError(err).WithField("server_key", serverKey).Error("Failed to get server by key")
		http.Error(w, "Invalid server key", http.StatusNotFound)
		return
	}

	// Get latest metrics
	metrics, err := h.storage.GetMetric(r.Context(), serverInfo.ServerID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverInfo.ServerID).Error("Failed to get server metrics")
		http.Error(w, "Failed to get server metrics", http.StatusInternalServerError)
		return
	}

	if metrics == nil {
		http.Error(w, "Server metrics not found", http.StatusNotFound)
		return
	}

	// Build response with temperature details and alerts
	response := h.buildMetricsResponse(serverInfo, metrics)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStorageTemperatureAlerts returns temperature alerts for storage devices
func (h *ServerMetricsHandler) GetStorageTemperatureAlerts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		http.Error(w, "Server ID is required", http.StatusBadRequest)
		return
	}

	// Get latest metrics
	metrics, err := h.storage.GetMetric(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get server metrics")
		http.Error(w, "Failed to get server metrics", http.StatusInternalServerError)
		return
	}

	if metrics == nil {
		http.Error(w, "Server metrics not found", http.StatusNotFound)
		return
	}

	// Generate alerts for storage temperatures
	alerts := h.generateStorageAlerts(serverID, metrics.TemperatureDetails.StorageTemperatures)

	response := map[string]interface{}{
		"server_id": serverID,
		"alerts":    alerts,
		"timestamp": metrics.Time.Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// buildMetricsResponse builds the complete metrics response with temperature details
func (h *ServerMetricsHandler) buildMetricsResponse(serverInfo *models.ServerInfo, metrics *models.ServerMetrics) map[string]interface{} {
	// Build storage temperature details
	storageTemps := make([]map[string]interface{}, 0)
	for _, storage := range metrics.TemperatureDetails.StorageTemperatures {
		alert := models.EvaluateStorageTemperature(storage.Type, storage.Temperature)

		storageTemp := map[string]interface{}{
			"device":      storage.Device,
			"type":        storage.Type,
			"temperature": storage.Temperature,
			"status":      alert.Status,
			"severity":    alert.Severity,
			"threshold":   alert.Threshold,
			"message":     alert.Message,
		}
		storageTemps = append(storageTemps, storageTemp)
	}

	// Build temperature section
	temperature := map[string]interface{}{
		"cpu":     metrics.TemperatureDetails.CPUTemperature,
		"gpu":     metrics.TemperatureDetails.GPUTemperature,
		"system":  metrics.TemperatureDetails.SystemTemperature,
		"highest": metrics.TemperatureDetails.HighestTemperature,
		"storage": storageTemps,
		"unit":    metrics.TemperatureDetails.TemperatureUnit,
	}

	// Build complete response
	response := map[string]interface{}{
		"server": map[string]interface{}{
			"id":     serverInfo.ServerID,
			"name":   serverInfo.Hostname,
			"status": "online", // Could be determined from last_seen
		},
		"metrics": map[string]interface{}{
			"basic": map[string]interface{}{
				"cpu":     metrics.CPU,
				"memory":  metrics.Memory,
				"disk":    metrics.Disk,
				"network": metrics.Network,
			},
			"temperature": temperature,
			"timestamp":   metrics.Time.Unix(),
		},
	}

	return response
}

// generateStorageAlerts generates alerts for storage temperatures
func (h *ServerMetricsHandler) generateStorageAlerts(serverID string, storageTemps []struct {
	Device      string  `json:"device"`
	Type        string  `json:"type"`
	Temperature float64 `json:"temperature"`
}) []models.Alert {
	alerts := make([]models.Alert, 0)

	for _, storage := range storageTemps {
		alert := models.EvaluateStorageTemperature(storage.Type, storage.Temperature)

		// Only create alerts for warning and critical temperatures
		if alert.Severity == models.AlertSeverityWarning || alert.Severity == models.AlertSeverityCritical {
			serverAlert := models.Alert{
				ID:          serverID + "-" + storage.Device + "-" + fmt.Sprintf("%.0f", storage.Temperature),
				Type:        models.AlertTypeStorageTemperature,
				ServerID:    serverID,
				Severity:    alert.Severity,
				Title:       "Storage Temperature Alert",
				Message:     alert.Message,
				Device:      storage.Device,
				Temperature: storage.Temperature,
				Threshold:   alert.Threshold,
				Value:       storage.Temperature,
				Status:      "active",
			}
			alerts = append(alerts, serverAlert)
		}
	}

	return alerts
}
