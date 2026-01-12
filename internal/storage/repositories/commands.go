package repositories

import (
	"context"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
)

// CommandsRepository defines operations for server commands
type CommandsRepository interface {
	// Basic operations
	Store(ctx context.Context, serverID string, command *models.Command) error
	GetPending(ctx context.Context, serverID string) ([]*models.Command, error)
	GetHistory(ctx context.Context, serverID string, limit int) ([]*models.Command, error)

	// Status operations
	MarkProcessed(ctx context.Context, commandID string) error
	MarkFailed(ctx context.Context, commandID string, error string) error

	// Query operations
	GetByID(ctx context.Context, commandID string) (*models.Command, error)
	GetByType(ctx context.Context, serverID string, commandType string) ([]*models.Command, error)

	// Cleanup operations
	DeleteProcessed(ctx context.Context, olderThan time.Duration) error
	DeleteByServer(ctx context.Context, serverID string) error

	// Health check
	Ping(ctx context.Context) error
}
