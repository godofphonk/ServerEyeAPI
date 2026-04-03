package handlers

import (
	"context"
	"encoding/json"
	"io"
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

	// Try to parse as V2 format first
	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(r.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read request body")
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}

	// Try V2 format
	var v2Msg struct {
		Metrics models.MetricsV2 `json:"metrics"`
	}
	if err := json.Unmarshal(bodyBytes, &v2Msg); err == nil && !v2Msg.Metrics.Timestamp.IsZero() {
		h.logger.WithFields(logrus.Fields{
			"server_id":   serverInfo.ServerID,
			"format":      "v2",
			"cpu_total":   v2Msg.Metrics.CPUUsage.UsageTotal,
			"memory_used": v2Msg.Metrics.Memory.UsedPercent,
			"temperature": v2Msg.Metrics.Temperature.Highest,
		}).Info("📊 HTTP: Received V2 metrics format")

		// Convert V2 to old format
		oldMetrics := h.convertV2ToOldFormat(&v2Msg.Metrics)
		oldMetrics.Time = v2Msg.Metrics.Timestamp

		if err := h.storage.StoreMetric(r.Context(), serverInfo.ServerID, oldMetrics); err != nil {
			h.logger.WithError(err).WithField("server_id", serverInfo.ServerID).Error("Failed to store V2 metrics")
			http.Error(w, "Failed to store metrics", http.StatusInternalServerError)
			return
		}

		h.logger.WithFields(logrus.Fields{
			"server_id":   serverInfo.ServerID,
			"cpu":         oldMetrics.CPU,
			"memory":      oldMetrics.Memory,
			"temperature": oldMetrics.TemperatureDetails.HighestTemperature,
		}).Info("✅ V2 metrics stored successfully via HTTP")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"server_id": serverInfo.ServerID,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// Fallback to V1 format
	var metricsMsg models.MetricsMessage
	if err := json.Unmarshal(bodyBytes, &metricsMsg); err != nil {
		h.logger.WithError(err).Error("Failed to decode metrics message (V1 or V2)")
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	metricsMsg.ServerID = serverInfo.ServerID
	metricsMsg.Metrics.Time = time.Now()

	h.logger.WithFields(logrus.Fields{
		"server_id": serverInfo.ServerID,
		"format":    "v1",
		"cpu":       metricsMsg.Metrics.CPU,
		"memory":    metricsMsg.Metrics.Memory,
	}).Info("📊 HTTP: Received V1 metrics format")

	if err := h.storage.StoreMetric(r.Context(), serverInfo.ServerID, &metricsMsg.Metrics); err != nil {
		h.logger.WithError(err).WithField("server_id", serverInfo.ServerID).Error("Failed to store V1 metrics")
		http.Error(w, "Failed to store metrics", http.StatusInternalServerError)
		return
	}

	h.logger.WithField("server_id", serverInfo.ServerID).Info("✅ V1 metrics stored successfully via HTTP")

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
