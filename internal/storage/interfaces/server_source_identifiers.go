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

	"github.com/godofphonk/ServerEyeAPI/internal/models"
)

// ServerSourceIdentifierRepository defines operations for server source identifiers
type ServerSourceIdentifierRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, identifier *models.ServerSourceIdentifier) error
	GetByID(ctx context.Context, id int64) (*models.ServerSourceIdentifier, error)
	GetByServerID(ctx context.Context, serverID string) ([]*models.ServerSourceIdentifier, error)
	GetByServerIDAndSourceType(ctx context.Context, serverID, sourceType string) ([]*models.ServerSourceIdentifier, error)
	GetByServerIDAndIdentifier(ctx context.Context, serverID, sourceType, identifier string) (*models.ServerSourceIdentifier, error)
	Update(ctx context.Context, identifier *models.ServerSourceIdentifier) error
	Delete(ctx context.Context, id int64) error
	DeleteByServerIDAndSourceType(ctx context.Context, serverID, sourceType string) error
	DeleteByServerIDSourceTypeAndIdentifier(ctx context.Context, serverID, sourceType, identifier string) error

	// Batch operations
	CreateBatch(ctx context.Context, identifiers []*models.ServerSourceIdentifier) error
	DeleteBatch(ctx context.Context, ids []int64) error

	// Query operations
	GetAllByServerID(ctx context.Context, serverID string) (map[string][]*models.ServerSourceIdentifier, error)
	GetByIdentifier(ctx context.Context, identifierType, identifier string) ([]*models.ServerSourceIdentifier, error)
	GetByTelegramID(ctx context.Context, telegramID int64) ([]*models.ServerSourceIdentifier, error)

	// Health check
	Ping(ctx context.Context) error
}

// SourceIdentifierListOption defines options for list operations
type SourceIdentifierListOption func(*SourceIdentifierListOptions)

type SourceIdentifierListOptions struct {
	SourceType     string
	IdentifierType string
	Limit          int
	Offset         int
}

// WithSourceType filters by source type
func WithSourceType(sourceType string) SourceIdentifierListOption {
	return func(opts *SourceIdentifierListOptions) {
		opts.SourceType = sourceType
	}
}

// WithIdentifierType filters by identifier type
func WithIdentifierType(identifierType string) SourceIdentifierListOption {
	return func(opts *SourceIdentifierListOptions) {
		opts.IdentifierType = identifierType
	}
}

// WithIdentifierLimit sets the limit for list operations
func WithIdentifierLimit(limit int) SourceIdentifierListOption {
	return func(opts *SourceIdentifierListOptions) {
		opts.Limit = limit
	}
}

// WithIdentifierOffset sets the offset for list operations
func WithIdentifierOffset(offset int) SourceIdentifierListOption {
	return func(opts *SourceIdentifierListOptions) {
		opts.Offset = offset
	}
}
