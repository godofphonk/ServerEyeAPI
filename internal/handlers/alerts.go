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
	"strconv"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AlertHandler struct {
	alertService *services.AlertService
	logger       *logrus.Logger
}

func NewAlertHandler(alertService *services.AlertService, logger *logrus.Logger) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
		logger:       logger,
	}
}

func (h *AlertHandler) GetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	alerts, err := h.alertService.GetActiveAlerts(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get active alerts")
		http.Error(w, "Failed to get active alerts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server_id": serverID,
		"alerts":    alerts,
		"count":     len(alerts),
	})
}

func (h *AlertHandler) GetAlertsByType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	alertTypeStr := vars["type"]

	alertType := models.AlertType(alertTypeStr)

	alerts, err := h.alertService.GetAlertsByType(r.Context(), serverID, alertType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get alerts by type")
		http.Error(w, "Failed to get alerts by type", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server_id": serverID,
		"type":      alertType,
		"alerts":    alerts,
		"count":     len(alerts),
	})
}

func (h *AlertHandler) GetAlertsByTimeRange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		http.Error(w, "start and end parameters are required", http.StatusBadRequest)
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	alerts, err := h.alertService.GetAlertsByTimeRange(r.Context(), serverID, start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get alerts by time range")
		http.Error(w, "Failed to get alerts by time range", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server_id": serverID,
		"start":     start,
		"end":       end,
		"alerts":    alerts,
		"count":     len(alerts),
	})
}

func (h *AlertHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID := vars["alert_id"]

	err := h.alertService.ResolveAlert(r.Context(), alertID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve alert")
		http.Error(w, "Failed to resolve alert", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Alert resolved successfully",
		"alert_id": alertID,
	})
}

func (h *AlertHandler) ResolveAlertsByType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	alertTypeStr := vars["type"]

	alertType := models.AlertType(alertTypeStr)

	err := h.alertService.ResolveAlertsByType(r.Context(), serverID, alertType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to resolve alerts by type")
		http.Error(w, "Failed to resolve alerts by type", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Alerts resolved successfully",
		"server_id": serverID,
		"type":      alertType,
	})
}

func (h *AlertHandler) GetAlertStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "24h"
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		http.Error(w, "Invalid duration format", http.StatusBadRequest)
		return
	}

	stats, err := h.alertService.GetAlertStats(r.Context(), serverID, duration)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get alert stats")
		http.Error(w, "Failed to get alert stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *AlertHandler) GetAllAlerts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	alerts, err := h.alertService.GetActiveAlerts(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get alerts")
		http.Error(w, "Failed to get alerts", http.StatusInternalServerError)
		return
	}

	if len(alerts) > limit {
		alerts = alerts[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server_id": serverID,
		"alerts":    alerts,
		"count":     len(alerts),
		"limit":     limit,
	})
}
