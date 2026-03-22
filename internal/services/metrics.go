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
	keyRepo      interfaces.GeneratedKeyRepository
	storage      storage.Storage
	alertService *AlertService
	logger       *logrus.Logger
}

// NewMetricsService creates a new metrics service
func NewMetricsService(keyRepo interfaces.GeneratedKeyRepository, storage storage.Storage, alertService *AlertService, logger *logrus.Logger) *MetricsService {
	return &MetricsService{
		keyRepo:      keyRepo,
		storage:      storage,
		alertService: alertService,
		logger:       logger,
	}
}

// GetServerByKey retrieves server information by server key
func (s *MetricsService) GetServerByKey(ctx context.Context, serverKey string) (*models.GeneratedKey, error) {
	return s.keyRepo.GetByKey(ctx, serverKey)
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

	// Return server status from storage
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

// GetServerMetricsWithStatus retrieves ONLY dynamic metrics for a server (NO static info, NO status)
func (s *MetricsService) GetServerMetricsWithStatus(ctx context.Context, serverID string) (map[string]interface{}, error) {
	// Verify server exists
	_, err := s.keyRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Get latest metrics
	metrics, err := s.storage.GetMetric(ctx, serverID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"server_id": serverID,
			"error":     err.Error(),
		}).Info("No current metrics available")

		// Return empty response with current timestamp instead of error
		// This prevents frontend from seeing zeros and allows proper "no data" handling
		return map[string]interface{}{
			"server_id": serverID,
			"metrics": map[string]interface{}{
				"cpu_percent":    0,
				"memory_percent": 0,
				"disk_percent":   0,
				"network_mbps":   0,
				"timestamp":      time.Now(),
				"status":         "no_data",
			},
		}, nil
	}

	// Build ONLY dynamic metrics response (NO static data)
	cleanMetrics := map[string]interface{}{
		"timestamp": metrics.Time,
	}

	// Core performance metrics (percentages)
	cleanMetrics["cpu_percent"] = metrics.CPU
	cleanMetrics["memory_percent"] = metrics.Memory
	cleanMetrics["disk_percent"] = metrics.Disk
	cleanMetrics["network_mbps"] = metrics.Network

	// Load averages (dynamic)
	if metrics.CPUUsage.LoadAverage.Load1 > 0 || metrics.CPUUsage.LoadAverage.Load5 > 0 {
		cleanMetrics["load_average"] = map[string]interface{}{
			"1m":  metrics.CPUUsage.LoadAverage.Load1,
			"5m":  metrics.CPUUsage.LoadAverage.Load5,
			"15m": metrics.CPUUsage.LoadAverage.Load15,
		}
	}

	// Temperature (dynamic)
	if metrics.TemperatureDetails.HighestTemperature > 0 {
		cleanMetrics["temperature_celsius"] = metrics.TemperatureDetails.HighestTemperature
		cleanMetrics["temperatures"] = map[string]interface{}{
			"cpu":     metrics.TemperatureDetails.CPUTemperature,
			"gpu":     metrics.TemperatureDetails.GPUTemperature,
			"storage": metrics.TemperatureDetails.StorageTemperatures,
			"highest": metrics.TemperatureDetails.HighestTemperature,
		}
	}

	// Process information (dynamic)
	if metrics.SystemDetails.ProcessesTotal > 0 {
		cleanMetrics["processes_total"] = metrics.SystemDetails.ProcessesTotal
		cleanMetrics["processes_running"] = metrics.SystemDetails.ProcessesRunning
		cleanMetrics["processes_sleeping"] = metrics.SystemDetails.ProcessesSleeping
	}

	// Uptime (dynamic)
	if metrics.SystemDetails.UptimeSeconds > 0 {
		cleanMetrics["uptime_seconds"] = metrics.SystemDetails.UptimeSeconds
	}

	// Memory details (dynamic usage, NOT total)
	if metrics.MemoryDetails.UsedGB > 0 {
		cleanMetrics["memory_details"] = map[string]interface{}{
			"used_gb":      metrics.MemoryDetails.UsedGB,
			"available_gb": metrics.MemoryDetails.AvailableGB,
			"free_gb":      metrics.MemoryDetails.FreeGB,
			"buffers_gb":   metrics.MemoryDetails.BuffersGB,
			"cached_gb":    metrics.MemoryDetails.CachedGB,
		}
	}

	// Disk details (dynamic usage)
	if len(metrics.DiskDetails) > 0 {
		diskDetails := make([]map[string]interface{}, 0, len(metrics.DiskDetails))
		for _, disk := range metrics.DiskDetails {
			diskDetails = append(diskDetails, map[string]interface{}{
				"path":         disk.Path,
				"used_gb":      disk.UsedGB,
				"free_gb":      disk.FreeGB,
				"used_percent": disk.UsedPercent,
			})
		}
		cleanMetrics["disk_details"] = diskDetails
	}

	// Network details (dynamic traffic)
	if len(metrics.NetworkDetails.Interfaces) > 0 {
		cleanMetrics["network_details"] = map[string]interface{}{
			"total_rx_mbps": metrics.NetworkDetails.TotalRxMbps,
			"total_tx_mbps": metrics.NetworkDetails.TotalTxMbps,
		}
	}

	response := map[string]interface{}{
		"server_id": serverID,
		"metrics":   cleanMetrics,
	}

	return response, nil
}

// GetServerStatus retrieves ONLY server status (online, last_seen, agent_version)
func (s *MetricsService) GetServerStatus(ctx context.Context, serverID string) (map[string]interface{}, error) {
	// Verify server exists
	key, err := s.keyRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// Get latest metrics to determine last_seen
	metrics, err := s.storage.GetMetric(ctx, serverID)

	var lastSeen time.Time
	var online bool

	if err != nil {
		// No metrics available - server offline
		lastSeen = key.CreatedAt
		online = false
	} else {
		// Metrics available - check if recent
		lastSeen = metrics.Time
		// Consider online if last seen within 5 minutes
		online = time.Since(lastSeen) < 5*time.Minute
	}

	response := map[string]interface{}{
		"server_id":     serverID,
		"online":        online,
		"last_seen":     lastSeen,
		"agent_version": key.AgentVersion,
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

	// GetMetricsHistory retrieves metrics history from storage
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

// Ping checks repository connectivity
func (s *MetricsService) Ping(ctx context.Context) error {
	if err := s.keyRepo.Ping(ctx); err != nil {
		return fmt.Errorf("key repository ping failed: %w", err)
	}

	return nil
}
