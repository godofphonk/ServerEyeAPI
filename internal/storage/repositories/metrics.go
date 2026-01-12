package repositories

import (
	"context"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
)

// MetricsRepository defines operations for server metrics
type MetricsRepository interface {
	// Basic operations
	Store(ctx context.Context, serverID string, metrics *models.ServerMetrics) error
	Get(ctx context.Context, serverID string) (*models.ServerMetrics, error)
	Delete(ctx context.Context, serverID string) error

	// Query operations
	GetLatest(ctx context.Context, serverID string, limit int) ([]*models.ServerMetrics, error)
	GetAll(ctx context.Context) (map[string]*models.ServerMetrics, error)
	GetByTimeRange(ctx context.Context, serverID string, start, end time.Time) ([]*models.ServerMetrics, error)

	// Cleanup operations
	DeleteOlderThan(ctx context.Context, olderThan time.Duration) error
	DeleteByTimeRange(ctx context.Context, serverID string, start, end time.Time) error

	// Health check
	Ping(ctx context.Context) error
}
