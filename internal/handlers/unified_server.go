package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
)

// UnifiedServerHandler handles unified server data requests
type UnifiedServerHandler struct {
	metricsService *services.MetricsService
	tieredService  *services.TieredMetricsService
	staticStorage  storage.StaticDataStorage
	logger         *logrus.Logger
}

// NewUnifiedServerHandler creates a new unified server handler
func NewUnifiedServerHandler(
	metricsService *services.MetricsService,
	tieredService *services.TieredMetricsService,
	staticStorage storage.StaticDataStorage,
	logger *logrus.Logger,
) *UnifiedServerHandler {
	return &UnifiedServerHandler{
		metricsService: metricsService,
		tieredService:  tieredService,
		staticStorage:  staticStorage,
		logger:         logger,
	}
}

// UnifiedResponse combines all server data in one response
type UnifiedResponse struct {
	ServerID  string `json:"server_id"`
	ServerKey string `json:"server_key,omitempty"`
	Timestamp string `json:"timestamp"`
	LastSeen  string `json:"last_seen,omitempty"`

	// Components
	Metrics    interface{} `json:"metrics,omitempty"`
	Status     interface{} `json:"status,omitempty"`
	StaticInfo interface{} `json:"static_info,omitempty"`

	// Performance metadata
	ResponseMeta ResponseMeta `json:"response_meta"`
}

type ResponseMeta struct {
	TotalResponseTimeMs int64                 `json:"total_response_time_ms"`
	ComponentsStatus    map[string]CompStatus `json:"components_status"`
}

type CompStatus struct {
	Available    bool   `json:"available"`
	ResponseTime int64  `json:"response_time_ms"`
	Error        string `json:"error,omitempty"`
}

// GetUnifiedServerData handles GET /api/servers/by-key/{server_key}/unified
func (h *UnifiedServerHandler) GetUnifiedServerData(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()
	vars := mux.Vars(r)
	serverKey := vars["server_key"]

	if serverKey == "" {
		h.writeError(w, "server_key is required", http.StatusBadRequest)
		return
	}

	// Parse optional query parameters
	includeMetrics := r.URL.Query().Get("include_metrics") != "false"
	includeStatus := r.URL.Query().Get("include_status") != "false"
	includeStatic := r.URL.Query().Get("include_static") != "false"

	// Get server info by key first
	serverInfo, err := h.metricsService.GetServerByKey(r.Context(), serverKey)
	if err != nil {
		h.logger.WithError(err).WithField("server_key", serverKey).Error("Failed to get server by key")
		h.writeError(w, "Server not found", http.StatusNotFound)
		return
	}

	// Use the same server_id conversion as static-info endpoint
	// Convert server_key to server_id (srv_ + last chars of key)
	staticServerID := "srv_" + serverKey[4:]

	// Create unified response
	response := UnifiedResponse{
		ServerID:  serverInfo.ServerID,
		ServerKey: serverKey,
		Timestamp: time.Now().Format(time.RFC3339),
		ResponseMeta: ResponseMeta{
			ComponentsStatus: make(map[string]CompStatus),
		},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Fetch metrics component
	if includeMetrics {
		wg.Add(1)
		go func() {
			defer wg.Done()
			componentStart := time.Now()

			// Use regular metrics (not tiered) for unified endpoint
			metrics, err := h.metricsService.GetServerMetricsWithStatus(
				r.Context(),
				serverInfo.ServerID,
			)

			componentDuration := time.Since(componentStart).Milliseconds()

			mu.Lock()
			if err != nil {
				response.ResponseMeta.ComponentsStatus["metrics"] = CompStatus{
					Available:    false,
					ResponseTime: componentDuration,
					Error:        err.Error(),
				}
			} else {
				response.ResponseMeta.ComponentsStatus["metrics"] = CompStatus{
					Available:    true,
					ResponseTime: componentDuration,
				}
				response.Metrics = metrics
			}
			mu.Unlock()
		}()
	}

	// Fetch status component
	if includeStatus {
		wg.Add(1)
		go func() {
			defer wg.Done()
			componentStart := time.Now()

			status, err := h.metricsService.GetServerStatus(r.Context(), serverInfo.ServerID)

			componentDuration := time.Since(componentStart).Milliseconds()

			mu.Lock()
			if err != nil {
				response.ResponseMeta.ComponentsStatus["status"] = CompStatus{
					Available:    false,
					ResponseTime: componentDuration,
					Error:        err.Error(),
				}
			} else {
				response.ResponseMeta.ComponentsStatus["status"] = CompStatus{
					Available:    true,
					ResponseTime: componentDuration,
				}
				response.Status = status

				// Extract last_seen from status and add to response
				if lastSeen, ok := status["last_seen"]; ok {
					if lastSeenTime, ok := lastSeen.(time.Time); ok {
						response.LastSeen = lastSeenTime.Format(time.RFC3339)
					}
				}
			}
			mu.Unlock()
		}()
	}

	// Fetch static info component
	if includeStatic {
		wg.Add(1)
		go func() {
			defer wg.Done()
			componentStart := time.Now()

			staticInfo, err := h.staticStorage.GetCompleteStaticInfo(r.Context(), staticServerID)

			componentDuration := time.Since(componentStart).Milliseconds()

			mu.Lock()
			if err != nil {
				response.ResponseMeta.ComponentsStatus["static_info"] = CompStatus{
					Available:    false,
					ResponseTime: componentDuration,
					Error:        err.Error(),
				}
			} else if staticInfo.ServerInfo == nil {
				response.ResponseMeta.ComponentsStatus["static_info"] = CompStatus{
					Available:    false,
					ResponseTime: componentDuration,
					Error:        "Static info not found",
				}
			} else {
				response.ResponseMeta.ComponentsStatus["static_info"] = CompStatus{
					Available:    true,
					ResponseTime: componentDuration,
				}
				response.StaticInfo = staticInfo
			}
			mu.Unlock()
		}()
	}

	// Wait for all components to complete
	wg.Wait()

	// Calculate total response time
	response.ResponseMeta.TotalResponseTimeMs = time.Since(requestStart).Milliseconds()

	// Log performance metrics
	h.logger.WithFields(logrus.Fields{
		"server_id":       serverInfo.ServerID,
		"server_key":      serverKey,
		"total_time_ms":   response.ResponseMeta.TotalResponseTimeMs,
		"components":      len(response.ResponseMeta.ComponentsStatus),
		"include_metrics": includeMetrics,
		"include_status":  includeStatus,
		"include_static":  includeStatic,
	}).Info("Unified server data request completed")

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes JSON response
func (h *UnifiedServerHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func (h *UnifiedServerHandler) writeError(w http.ResponseWriter, message string, status int) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
