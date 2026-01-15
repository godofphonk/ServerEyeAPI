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
