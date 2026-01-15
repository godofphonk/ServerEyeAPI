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
