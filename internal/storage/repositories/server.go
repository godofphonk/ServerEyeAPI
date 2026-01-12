package repositories

import (
	"context"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
)

// ServerRepository defines operations for server management
type ServerRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, server *models.Server) error
	GetByID(ctx context.Context, id string) (*models.Server, error)
	GetByKey(ctx context.Context, serverKey string) (*models.Server, error)
	Update(ctx context.Context, server *models.Server) error
	Delete(ctx context.Context, id string) error
	
	// Query operations
	List(ctx context.Context) ([]*models.Server, error)
	ListByStatus(ctx context.Context, status string) ([]*models.Server, error)
	ListByHostname(ctx context.Context, hostname string) ([]*models.Server, error)
	
	// Status operations
	UpdateStatus(ctx context.Context, serverID string, status string) error
	UpdateLastSeen(ctx context.Context, serverID string, lastSeen time.Time) error
	
	// Health check
	Ping(ctx context.Context) error
}
