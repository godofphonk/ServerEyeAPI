package repositories

import (
	"context"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
)

// ConnectionsRepository defines operations for connection management
type ConnectionsRepository interface {
	// Basic operations
	Store(ctx context.Context, serverID string, conn *models.Connection) error
	GetActive(ctx context.Context, serverID string) ([]*models.Connection, error)
	GetHistory(ctx context.Context, serverID string, limit int) ([]*models.Connection, error)

	// Query operations
	GetByID(ctx context.Context, connectionID string) (*models.Connection, error)
	GetByType(ctx context.Context, serverID string, connType string) ([]*models.Connection, error)
	GetAll(ctx context.Context) ([]*models.Connection, error)

	// Management operations
	Close(ctx context.Context, connectionID string) error
	CloseByServer(ctx context.Context, serverID string) error
	MarkDisconnected(ctx context.Context, connectionID string) error

	// Cleanup operations
	DeleteOlderThan(ctx context.Context, olderThan time.Duration) error
	DeleteDisconnected(ctx context.Context, olderThan time.Duration) error
	DeleteByServer(ctx context.Context, serverID string) error

	// Health check
	Ping(ctx context.Context) error
}
