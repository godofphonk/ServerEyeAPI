package services

import (
	"context"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
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
func (s *MetricsService) GetServerMetrics(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	// Get server status from PostgreSQL
	status, err := s.storage.GetServerMetrics(ctx, serverID)
	if err != nil {
		s.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get server status from PostgreSQL")
		return nil, err
	}

	s.logger.WithField("server_id", serverID).Debug("Retrieved server status from PostgreSQL")
	return status, nil
}

// StoreServerMetrics stores metrics for a server
func (s *MetricsService) StoreServerMetrics(ctx context.Context, serverID string, metrics *models.ServerMetrics) error {
	// Store in Redis for real-time access
	if err := s.storage.StoreMetric(ctx, serverID, metrics); err != nil {
		s.logger.WithError(err).WithField("server_id", serverID).Error("Failed to store metrics in Redis")
		return err
	}

	s.logger.WithField("server_id", serverID).Debug("Stored server metrics in Redis")
	return nil
}
