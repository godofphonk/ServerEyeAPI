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
