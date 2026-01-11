package services

import (
	"context"

	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/sirupsen/logrus"
)

// MetricsService handles metrics operations
type MetricsService struct {
	storage storage.Storage
	logger  *logrus.Logger
}

// NewMetricsService creates a new metrics service
func NewMetricsService(storage storage.Storage, logger *logrus.Logger) *MetricsService {
	return &MetricsService{
		storage: storage,
		logger:  logger,
	}
}

// GetServerMetrics retrieves metrics for a specific server
func (s *MetricsService) GetServerMetrics(ctx context.Context, serverID string) (map[string]interface{}, error) {
	// Get metrics from Redis (real-time data)
	metrics, err := s.storage.GetMetric(ctx, serverID)
	if err != nil {
		s.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get server metrics from Redis")
		// Fallback to PostgreSQL if Redis fails
		return s.storage.GetServerMetrics(ctx, serverID)
	}

	s.logger.WithField("server_id", serverID).Debug("Retrieved server metrics from Redis")
	return metrics, nil
}

// StoreServerMetrics stores metrics for a server
func (s *MetricsService) StoreServerMetrics(ctx context.Context, serverID string, data map[string]interface{}) error {
	// Store in Redis for real-time access
	if err := s.storage.StoreMetric(ctx, serverID, data); err != nil {
		s.logger.WithError(err).WithField("server_id", serverID).Error("Failed to store metrics in Redis")
		return err
	}

	s.logger.WithField("server_id", serverID).Debug("Stored server metrics in Redis")
	return nil
}
