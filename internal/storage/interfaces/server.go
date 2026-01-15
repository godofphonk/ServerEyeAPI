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

package interfaces

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
	List(ctx context.Context, opts ...ListOption) ([]*models.Server, error)
	ListByStatus(ctx context.Context, status string) ([]*models.Server, error)
	ListByHostname(ctx context.Context, hostname string) ([]*models.Server, error)

	// Status operations
	UpdateStatus(ctx context.Context, serverID string, status string) error
	UpdateLastSeen(ctx context.Context, serverID string, lastSeen time.Time) error
	UpdateSources(ctx context.Context, serverID string, sources string) error

	// Health check
	Ping(ctx context.Context) error
}

// GeneratedKeyRepository defines operations for generated keys
type GeneratedKeyRepository interface {
	// Key operations
	Create(ctx context.Context, key *models.GeneratedKey) error
	GetByKey(ctx context.Context, serverKey string) (*models.GeneratedKey, error)
	GetByServerID(ctx context.Context, serverID string) (*models.GeneratedKey, error)
	Update(ctx context.Context, key *models.GeneratedKey) error
	Delete(ctx context.Context, id int64) error

	// Query operations
	List(ctx context.Context, opts ...ListOption) ([]*models.GeneratedKey, error)
	ListByStatus(ctx context.Context, status string) ([]*models.GeneratedKey, error)

	// Health check
	Ping(ctx context.Context) error
}

// ListOption defines options for list operations
type ListOption func(*ListOptions)

type ListOptions struct {
	Limit  int
	Offset int
	Status string
}

// WithLimit sets the limit for list operations
func WithLimit(limit int) ListOption {
	return func(opts *ListOptions) {
		opts.Limit = limit
	}
}

// WithOffset sets the offset for list operations
func WithOffset(offset int) ListOption {
	return func(opts *ListOptions) {
		opts.Offset = offset
	}
}

// WithStatus filters by status
func WithStatus(status string) ListOption {
	return func(opts *ListOptions) {
		opts.Status = status
	}
}
