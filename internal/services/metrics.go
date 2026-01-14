package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
)

// MetricsService handles metrics-related business logic
type MetricsService struct {
	keyRepo interfaces.GeneratedKeyRepository
	storage storage.Storage
	logger  *logrus.Logger
}

// NewMetricsService creates a new metrics service
func NewMetricsService(keyRepo interfaces.GeneratedKeyRepository, storage storage.Storage, logger *logrus.Logger) *MetricsService {
	return &MetricsService{
		keyRepo: keyRepo,
		storage: storage,
		logger:  logger,
	}
}

// StoreMetricsRequest represents a metrics storage request
type StoreMetricsRequest struct {
	ServerID string        `json:"server_id" validate:"required"`
	Metrics  ServerMetrics `json:"metrics" validate:"required"`
	System   SystemInfo    `json:"system" validate:"required"`
}

// ServerMetrics represents server performance metrics
type ServerMetrics struct {
	CPU     float64   `json:"cpu"`     // CPU usage percentage (0-100)
	Memory  float64   `json:"memory"`  // Memory usage percentage (0-100)
	Disk    float64   `json:"disk"`    // Disk usage percentage (0-100)
	Network float64   `json:"network"` // Network usage in MB/s
	Time    time.Time `json:"time"`    // Timestamp when metrics were collected
}

// SystemInfo represents system information
type SystemInfo struct {
	OS           string `json:"os"`           // Operating system name
	Architecture string `json:"architecture"` // System architecture
	Kernel       string `json:"kernel"`       // Kernel version
	Uptime       int64  `json:"uptime"`       // System uptime in seconds
	Hostname     string `json:"hostname"`     // Server hostname
}

// MetricsMessage represents a complete metrics message from agent
type MetricsMessage struct {
	ServerID string        `json:"server_id"`
	Metrics  ServerMetrics `json:"metrics"`
	System   SystemInfo    `json:"system"`
}

// ServerStatus represents server status information
type ServerStatus struct {
	Online       bool      `json:"online"`
	LastSeen     time.Time `json:"last_seen"`
	Version      string    `json:"version"`
	OSInfo       string    `json:"os_info"`
	AgentVersion string    `json:"agent_version"`
	Hostname     string    `json:"hostname"`
}

// StoreMetrics stores server metrics with validation and business logic
func (s *MetricsService) StoreMetrics(ctx context.Context, req *StoreMetricsRequest) error {
	// Validate request
	if err := s.validateMetricsRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Verify server exists
	_, err := s.keyRepo.GetByServerID(ctx, req.ServerID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Validate metrics values
	if err := s.validateMetricsValues(&req.Metrics); err != nil {
		return fmt.Errorf("invalid metrics values: %w", err)
	}

	// Store metrics in storage
	serverMetrics := &models.ServerMetrics{
		CPU:     req.Metrics.CPU,
		Memory:  req.Metrics.Memory,
		Disk:    req.Metrics.Disk,
		Network: req.Metrics.Network,
		Time:    req.Metrics.Time,
	}

	if err := s.storage.StoreMetric(ctx, req.ServerID, serverMetrics); err != nil {
		s.logger.WithFields(logrus.Fields{
			"server_id": req.ServerID,
			"error":     err.Error(),
		}).Error("Failed to store metrics")
		return fmt.Errorf("failed to store metrics: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": req.ServerID,
		"cpu":       req.Metrics.CPU,
		"memory":    req.Metrics.Memory,
		"disk":      req.Metrics.Disk,
	}).Info("Metrics stored successfully")

	return nil
}

// GetServerMetrics retrieves latest metrics for a server
func (s *MetricsService) GetServerMetrics(ctx context.Context, serverID string) (*ServerStatus, error) {
	// Verify server exists
	key, err := s.keyRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Return server status (TODO: implement proper metrics retrieval)
	status := &ServerStatus{
		Online:       true,
		LastSeen:     time.Now(),
		Version:      "",
		OSInfo:       key.OSInfo,
		AgentVersion: key.AgentVersion,
		Hostname:     key.Hostname,
	}

	return status, nil
}

// GetAllServerMetrics retrieves complete metrics for a server
func (s *MetricsService) GetAllServerMetrics(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	// Verify server exists
	_, err := s.keyRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Get metrics from storage (Redis with PostgreSQL fallback)
	metrics, err := s.storage.GetMetric(ctx, serverID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"server_id": serverID,
			"error":     err.Error(),
		}).Error("Failed to retrieve metrics")
		return nil, fmt.Errorf("failed to retrieve metrics: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"cpu":       metrics.CPU,
		"memory":    metrics.Memory,
		"disk":      metrics.Disk,
	}).Debug("Retrieved server metrics")

	return metrics, nil
}

// GetServerMetricsWithStatus retrieves both metrics and status for a server
func (s *MetricsService) GetServerMetricsWithStatus(ctx context.Context, serverID string) (map[string]interface{}, error) {
	// Verify server exists
	key, err := s.keyRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Get metrics
	metrics, err := s.storage.GetMetric(ctx, serverID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"server_id": serverID,
			"error":     err.Error(),
		}).Warn("Failed to retrieve metrics, returning status only")

		// Return status only if metrics not available
		return map[string]interface{}{
			"server_id": serverID,
			"timestamp": time.Now(),
			"status": map[string]interface{}{
				"online":        false,
				"last_seen":     key.CreatedAt,
				"os_info":       key.OSInfo,
				"agent_version": key.AgentVersion,
				"hostname":      key.Hostname,
			},
			"metrics": nil,
			"alerts":  []string{"Metrics not available"},
		}, nil
	}

	// Combine metrics and status
	response := map[string]interface{}{
		"server_id": serverID,
		"timestamp": metrics.Time,
		"status": map[string]interface{}{
			"online":        true,
			"last_seen":     metrics.Time,
			"os_info":       key.OSInfo,
			"agent_version": key.AgentVersion,
			"hostname":      key.Hostname,
		},
		"metrics": metrics,
		"alerts":  s.generateAlerts(metrics),
	}

	return response, nil
}

// GetMetricsHistory retrieves historical metrics for a server
func (s *MetricsService) GetMetricsHistory(ctx context.Context, serverID string, limit int) ([]*MetricsMessage, error) {
	// Verify server exists
	_, err := s.keyRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// TODO: Implement metrics history retrieval from database
	s.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"limit":     limit,
	}).Info("Retrieving metrics history")

	// Return empty for now
	return []*MetricsMessage{}, nil
}

// ProcessWebSocketMetrics processes metrics received via WebSocket
func (s *MetricsService) ProcessWebSocketMetrics(ctx context.Context, msg *MetricsMessage) error {
	// Validate message
	if msg.ServerID == "" {
		return fmt.Errorf("server_id is required")
	}

	// Verify server exists
	_, err := s.keyRepo.GetByServerID(ctx, msg.ServerID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Validate metrics
	if err := s.validateMetricsValues(&msg.Metrics); err != nil {
		return fmt.Errorf("invalid metrics values: %w", err)
	}

	// Store metrics in storage
	serverMetrics := &models.ServerMetrics{
		CPU:     msg.Metrics.CPU,
		Memory:  msg.Metrics.Memory,
		Disk:    msg.Metrics.Disk,
		Network: msg.Metrics.Network,
		Time:    msg.Metrics.Time,
	}

	if err := s.storage.StoreMetric(ctx, msg.ServerID, serverMetrics); err != nil {
		s.logger.WithFields(logrus.Fields{
			"server_id": msg.ServerID,
			"error":     err.Error(),
		}).Error("Failed to store WebSocket metrics")
		return fmt.Errorf("failed to store WebSocket metrics: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"server_id": msg.ServerID,
		"cpu":       msg.Metrics.CPU,
		"memory":    msg.Metrics.Memory,
		"disk":      msg.Metrics.Disk,
		"network":   msg.Metrics.Network,
	}).Info("WebSocket metrics processed successfully")

	return nil
}

// validateMetricsRequest validates metrics storage request
func (s *MetricsService) validateMetricsRequest(req *StoreMetricsRequest) error {
	if req.ServerID == "" {
		return fmt.Errorf("server_id is required")
	}

	if req.Metrics.Time.IsZero() {
		return fmt.Errorf("metrics timestamp is required")
	}

	return s.validateMetricsValues(&req.Metrics)
}

// validateMetricsValues validates metric values
func (s *MetricsService) validateMetricsValues(metrics *ServerMetrics) error {
	// CPU usage should be between 0-100
	if metrics.CPU < 0 || metrics.CPU > 100 {
		return fmt.Errorf("invalid CPU usage: %f (must be 0-100)", metrics.CPU)
	}

	// Memory usage should be between 0-100
	if metrics.Memory < 0 || metrics.Memory > 100 {
		return fmt.Errorf("invalid memory usage: %f (must be 0-100)", metrics.Memory)
	}

	// Disk usage should be between 0-100
	if metrics.Disk < 0 || metrics.Disk > 100 {
		return fmt.Errorf("invalid disk usage: %f (must be 0-100)", metrics.Disk)
	}

	// Network usage should be non-negative
	if metrics.Network < 0 {
		return fmt.Errorf("invalid network usage: %f (must be >= 0)", metrics.Network)
	}

	return nil
}

// generateAlerts creates alerts based on metrics thresholds
func (s *MetricsService) generateAlerts(metrics *models.ServerMetrics) []string {
	var alerts []string

	// CPU alerts
	if metrics.CPU > 80 {
		alerts = append(alerts, fmt.Sprintf("High CPU usage: %.1f%%", metrics.CPU))
	} else if metrics.CPU > 60 {
		alerts = append(alerts, fmt.Sprintf("Moderate CPU usage: %.1f%%", metrics.CPU))
	}

	// Memory alerts
	if metrics.Memory > 85 {
		alerts = append(alerts, fmt.Sprintf("High memory usage: %.1f%%", metrics.Memory))
	} else if metrics.Memory > 70 {
		alerts = append(alerts, fmt.Sprintf("Moderate memory usage: %.1f%%", metrics.Memory))
	}

	// Disk alerts
	if metrics.Disk > 90 {
		alerts = append(alerts, fmt.Sprintf("Critical disk usage: %.1f%%", metrics.Disk))
	} else if metrics.Disk > 80 {
		alerts = append(alerts, fmt.Sprintf("High disk usage: %.1f%%", metrics.Disk))
	}

	// Temperature alerts
	if metrics.TemperatureDetails.CPUTemperature > 80 {
		alerts = append(alerts, fmt.Sprintf("High CPU temperature: %.1f°C", metrics.TemperatureDetails.CPUTemperature))
	}

	if metrics.TemperatureDetails.HighestTemperature > 85 {
		alerts = append(alerts, fmt.Sprintf("High system temperature: %.1f°C", metrics.TemperatureDetails.HighestTemperature))
	}

	// Load average alerts
	if metrics.CPUUsage.LoadAverage.Load1 > 2.0 {
		alerts = append(alerts, fmt.Sprintf("High load average (1m): %.2f", metrics.CPUUsage.LoadAverage.Load1))
	}

	// Network alerts (if unusually high)
	if metrics.Network > 1000 { // > 1GB/s
		alerts = append(alerts, fmt.Sprintf("High network usage: %.1f MB/s", metrics.Network))
	}

	return alerts
}

// Ping checks repository connectivity
func (s *MetricsService) Ping(ctx context.Context) error {
	if err := s.keyRepo.Ping(ctx); err != nil {
		return fmt.Errorf("key repository ping failed: %w", err)
	}

	return nil
}
