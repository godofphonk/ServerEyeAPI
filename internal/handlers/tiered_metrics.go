package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/timescaledb"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// TieredMetricsHandler handles tiered metrics endpoints
type TieredMetricsHandler struct {
	service *services.TieredMetricsService
	logger  *logrus.Logger
}

// NewTieredMetricsHandler creates a new tiered metrics handler
func NewTieredMetricsHandler(service *services.TieredMetricsService, logger *logrus.Logger) *TieredMetricsHandler {
	return &TieredMetricsHandler{
		service: service,
		logger:  logger,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON writes JSON response
func (h *TieredMetricsHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// GetMetrics retrieves metrics with automatic granularity selection
func (h *TieredMetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	if serverID == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "server_id is required"})
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "start and end query parameters are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid start time format, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid end time format, use RFC3339"})
		return
	}

	if endTime.Before(startTime) {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "end time must be after start time"})
		return
	}

	// Limit time range to maximum 30 days
	if endTime.Sub(startTime) > 30*24*time.Hour {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "time range cannot exceed 30 days"})
		return
	}

	response, err := h.service.GetMetricsWithAutoGranularity(r.Context(), serverID, startTime, endTime)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get tiered metrics")
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve metrics"})
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetRealTimeMetrics gets real-time metrics (last hour with 1-minute granularity)
func (h *TieredMetricsHandler) GetRealTimeMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	if serverID == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "server_id is required"})
		return
	}

	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "1h"
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid duration format"})
		return
	}

	response, err := h.service.GetRealTimeMetrics(r.Context(), serverID, duration)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get real-time metrics")
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve real-time metrics"})
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetHistoricalMetrics gets historical metrics with specified granularity
func (h *TieredMetricsHandler) GetHistoricalMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	if serverID == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "server_id is required"})
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "start and end query parameters are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid start time format"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid end time format"})
		return
	}

	granularityStr := r.URL.Query().Get("granularity")
	var granularity timescaledb.MetricsGranularity

	switch granularityStr {
	case "1m":
		granularity = timescaledb.Granularity1Min
	case "5m":
		granularity = timescaledb.Granularity5Min
	case "10m":
		granularity = timescaledb.Granularity10Min
	case "1h":
		granularity = timescaledb.Granularity1Hour
	default:
		// Auto-determine based on time range
		duration := endTime.Sub(startTime)
		if duration <= time.Hour {
			granularity = timescaledb.Granularity1Min
		} else if duration <= 3*time.Hour {
			granularity = timescaledb.Granularity5Min
		} else if duration <= 24*time.Hour {
			granularity = timescaledb.Granularity10Min
		} else {
			granularity = timescaledb.Granularity1Hour
		}
	}

	response, err := h.service.GetHistoricalMetrics(r.Context(), serverID, startTime, endTime, granularity)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get historical metrics")
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve historical metrics"})
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetDashboardMetrics gets optimized metrics for dashboard display
func (h *TieredMetricsHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	if serverID == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "server_id is required"})
		return
	}

	metrics, err := h.service.GetDashboardMetrics(r.Context(), serverID)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get dashboard metrics")
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve dashboard metrics"})
		return
	}

	h.writeJSON(w, http.StatusOK, metrics)
}

// GetMetricsComparison compares metrics between two time periods
func (h *TieredMetricsHandler) GetMetricsComparison(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	if serverID == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "server_id is required"})
		return
	}

	// Parse period 1
	p1StartStr := r.URL.Query().Get("period1_start")
	p1EndStr := r.URL.Query().Get("period1_end")
	if p1StartStr == "" || p1EndStr == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "period1_start and period1_end are required"})
		return
	}

	p1Start, err := time.Parse(time.RFC3339, p1StartStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid period1_start time format"})
		return
	}

	p1End, err := time.Parse(time.RFC3339, p1EndStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid period1_end time format"})
		return
	}

	// Parse period 2
	p2StartStr := r.URL.Query().Get("period2_start")
	p2EndStr := r.URL.Query().Get("period2_end")
	if p2StartStr == "" || p2EndStr == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "period2_start and period2_end are required"})
		return
	}

	p2Start, err := time.Parse(time.RFC3339, p2StartStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid period2_start time format"})
		return
	}

	p2End, err := time.Parse(time.RFC3339, p2EndStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid period2_end time format"})
		return
	}

	comparison, err := h.service.GetMetricsComparison(
		r.Context(),
		serverID,
		p1Start, p1End,
		p2Start, p2End,
	)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get metrics comparison")
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve metrics comparison"})
		return
	}

	h.writeJSON(w, http.StatusOK, comparison)
}

// GetMetricsSummary returns a summary of metrics statistics
func (h *TieredMetricsHandler) GetMetricsSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.GetMetricsSummary(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get metrics summary")
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve metrics summary"})
		return
	}

	h.writeJSON(w, http.StatusOK, summary)
}

// GetMetricsHeatmap returns metrics data for heatmap visualization
func (h *TieredMetricsHandler) GetMetricsHeatmap(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	if serverID == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "server_id is required"})
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "start and end query parameters are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid start time format"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid end time format"})
		return
	}

	heatmapData, err := h.service.GetMetricsHeatmap(r.Context(), serverID, startTime, endTime)
	if err != nil {
		h.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get heatmap data")
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve heatmap data"})
		return
	}

	h.writeJSON(w, http.StatusOK, heatmapData)
}
