package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type MetricsPushHandler struct {
	storage storage.Storage
	logger  *logrus.Logger
}

func NewMetricsPushHandler(storage storage.Storage, logger *logrus.Logger) *MetricsPushHandler {
	return &MetricsPushHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *MetricsPushHandler) PushMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		h.logger.Warn("Missing server_key in request")
		http.Error(w, "server_key is required", http.StatusBadRequest)
		return
	}

	serverInfo, err := h.storage.GetServerByKey(r.Context(), serverKey)
	if err != nil {
		h.logger.WithError(err).WithField("server_key", serverKey).Error("Failed to get server by key")
		http.Error(w, "Invalid server key", http.StatusUnauthorized)
		return
	}

	var metricsMsg models.MetricsMessage
	if err := json.NewDecoder(r.Body).Decode(&metricsMsg); err != nil {
		h.logger.WithError(err).Error("Failed to decode metrics message")
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	metricsMsg.ServerID = serverInfo.ServerID

	// Always set timestamp to current time
	metricsMsg.Metrics.Time = time.Now()

	h.logger.WithFields(logrus.Fields{
		"server_id": serverInfo.ServerID,
		"timestamp": metricsMsg.Metrics.Time,
		"is_zero":   metricsMsg.Metrics.Time.IsZero(),
	}).Info("About to store metrics with timestamp")

	if err := h.storage.StoreMetric(r.Context(), serverInfo.ServerID, &metricsMsg.Metrics); err != nil {
		h.logger.WithError(err).WithField("server_id", serverInfo.ServerID).Error("Failed to store metrics")
		http.Error(w, "Failed to store metrics", http.StatusInternalServerError)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"server_id": serverInfo.ServerID,
		"hostname":  serverInfo.Hostname,
		"cpu":       metricsMsg.Metrics.CPU,
		"memory":    metricsMsg.Metrics.Memory,
	}).Info("Metrics stored successfully via HTTP")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"server_id": serverInfo.ServerID,
		"timestamp": time.Now().Unix(),
	})
}

func (h *MetricsPushHandler) PushHeartbeat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		h.logger.Warn("Missing server_key in request")
		http.Error(w, "server_key is required", http.StatusBadRequest)
		return
	}

	serverInfo, err := h.storage.GetServerByKey(r.Context(), serverKey)
	if err != nil {
		h.logger.WithError(err).WithField("server_key", serverKey).Error("Failed to get server by key")
		http.Error(w, "Invalid server key", http.StatusUnauthorized)
		return
	}

	if err := h.storage.SetServerStatus(r.Context(), serverInfo.ServerID, "online"); err != nil {
		h.logger.WithError(err).WithField("server_id", serverInfo.ServerID).Error("Failed to update server status")
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"server_id": serverInfo.ServerID,
		"hostname":  serverInfo.Hostname,
	}).Debug("Heartbeat received via HTTP")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"server_id": serverInfo.ServerID,
		"status":    "online",
		"timestamp": time.Now().Unix(),
	})
}

func (h *MetricsPushHandler) PushMetricsByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		h.logger.Warn("Missing server_id in request")
		http.Error(w, "server_id is required", http.StatusBadRequest)
		return
	}

	var metricsMsg models.MetricsMessage
	if err := json.NewDecoder(r.Body).Decode(&metricsMsg); err != nil {
		h.logger.WithError(err).Error("Failed to decode metrics message")
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	metricsMsg.ServerID = serverID

	// Always set timestamp to current time
	metricsMsg.Metrics.Time = time.Now()

	if err := h.storage.StoreMetric(r.Context(), serverID, &metricsMsg.Metrics); err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to store metrics")
		http.Error(w, "Failed to store metrics", http.StatusInternalServerError)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"cpu":       metricsMsg.Metrics.CPU,
		"memory":    metricsMsg.Metrics.Memory,
	}).Info("Metrics stored successfully via HTTP (by ID)")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"server_id": serverID,
		"timestamp": time.Now().Unix(),
	})
}

func (h *MetricsPushHandler) PushHeartbeatByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]

	if serverID == "" {
		h.logger.Warn("Missing server_id in request")
		http.Error(w, "server_id is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.storage.SetServerStatus(ctx, serverID, "online"); err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to update server status")
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	h.logger.WithField("server_id", serverID).Debug("Heartbeat received via HTTP (by ID)")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"server_id": serverID,
		"status":    "online",
		"timestamp": time.Now().Unix(),
	})
}
