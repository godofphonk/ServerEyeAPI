package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
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
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check dependencies
	deps := h.CheckDependencies(ctx)

	// Determine overall status
	status := "healthy"
	for _, healthy := range deps {
		if !healthy {
			status = "unhealthy"
			break
		}
	}

	response := &models.HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   version.Version,
		Clients:   0, // TODO: Get actual WebSocket client count
	}

	// Set HTTP status based on health
	httpStatus := http.StatusOK
	if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	h.writeJSON(w, httpStatus, response)
}

// CheckDependencies checks the health of all dependencies
func (h *HealthHandler) CheckDependencies(ctx context.Context) map[string]bool {
	deps := make(map[string]bool)

	// Check database connectivity
	if h.storage != nil {
		err := h.storage.Ping()
		deps["database"] = err == nil
		if err != nil {
			h.logger.WithError(err).Error("Database health check failed")
		}
	} else {
		deps["database"] = false
		h.logger.Error("Storage is nil - cannot check database")
	}

	// TODO: Add Redis check when Redis client is available
	// deps["redis"] = h.redis.Ping(ctx) == nil

	// TODO: Add WebSocket server check
	// deps["websocket"] = h.wsServer.IsHealthy()

	// Add system resource checks
	deps["memory"] = h.checkMemoryUsage()
	deps["disk"] = h.checkDiskSpace()

	return deps
}

// checkMemoryUsage checks memory usage (basic implementation)
func (h *HealthHandler) checkMemoryUsage() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Consider unhealthy if using more than 80% of available memory
	// This is a simple check - in production you might want more sophisticated monitoring
	return m.Alloc < m.Sys*4/5 // Less than 80% of system memory
}

// checkDiskSpace checks available disk space
func (h *HealthHandler) checkDiskSpace() bool {
	// Basic disk space check
	// In production, you might want to check specific directories
	return true // TODO: Implement actual disk space checking
}

// writeJSON writes JSON response
func (h *HealthHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
