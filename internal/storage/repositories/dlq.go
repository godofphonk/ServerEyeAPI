package repositories

import (
	"context"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
)

// DLQRepository defines operations for dead letter queue
type DLQRepository interface {
	// Basic operations
	Store(ctx context.Context, dlq *models.DLQMessage) error
	GetByTopic(ctx context.Context, topic string, limit int) ([]*models.DLQMessage, error)
	GetByID(ctx context.Context, id string) (*models.DLQMessage, error)

	// Query operations
	GetAll(ctx context.Context) ([]*models.DLQMessage, error)
	GetByStatus(ctx context.Context, status string, limit int) ([]*models.DLQMessage, error)
	GetOlderThan(ctx context.Context, olderThan time.Duration) ([]*models.DLQMessage, error)

	// Management operations
	Delete(ctx context.Context, id string) error
	Requeue(ctx context.Context, id string) error
	MarkProcessed(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, error string) error

	// Cleanup operations
	DeleteProcessed(ctx context.Context, olderThan time.Duration) error
	DeleteByTopic(ctx context.Context, topic string) error

	// Health check
	Ping(ctx context.Context) error
}
