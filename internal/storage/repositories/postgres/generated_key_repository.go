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

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	"github.com/sirupsen/logrus"
)

// GeneratedKeyRepository implements interfaces.GeneratedKeyRepository for PostgreSQL
type GeneratedKeyRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewGeneratedKeyRepository creates a new PostgreSQL generated key repository
func NewGeneratedKeyRepository(db *sql.DB, logger *logrus.Logger) interfaces.GeneratedKeyRepository {
	return &GeneratedKeyRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new generated key
func (r *GeneratedKeyRepository) Create(ctx context.Context, key *models.GeneratedKey) error {
	query := `
		INSERT INTO generated_keys (server_id, server_key, agent_version, os_info, hostname, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		key.ServerID,
		key.ServerKey,
		key.AgentVersion,
		key.OSInfo,
		key.Hostname,
		key.Status,
		time.Now(),
	).Scan(&key.ID)

	if err != nil {
		return fmt.Errorf("failed to create generated key: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id":  key.ServerID,
		"server_key": key.ServerKey,
		"id":         key.ID,
	}).Info("Generated key created successfully")

	return nil
}

// GetByKey retrieves a generated key by server key
func (r *GeneratedKeyRepository) GetByKey(ctx context.Context, serverKey string) (*models.GeneratedKey, error) {
	query := `
		SELECT id, server_id, server_key, agent_version, os_info, hostname, status, created_at
		FROM generated_keys
		WHERE server_key = $1
	`

	var key models.GeneratedKey
	err := r.db.QueryRowContext(ctx, query, serverKey).Scan(
		&key.ID,
		&key.ServerID,
		&key.ServerKey,
		&key.AgentVersion,
		&key.OSInfo,
		&key.Hostname,
		&key.Status,
		&key.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("generated key not found for server_key: %s", serverKey)
		}
		return nil, fmt.Errorf("failed to get generated key: %w", err)
	}

	return &key, nil
}

// GetByServerID retrieves a generated key by server ID
func (r *GeneratedKeyRepository) GetByServerID(ctx context.Context, serverID string) (*models.GeneratedKey, error) {
	query := `
		SELECT id, server_id, server_key, agent_version, os_info, hostname, status, created_at
		FROM generated_keys
		WHERE server_id = $1
	`

	var key models.GeneratedKey
	err := r.db.QueryRowContext(ctx, query, serverID).Scan(
		&key.ID,
		&key.ServerID,
		&key.ServerKey,
		&key.AgentVersion,
		&key.OSInfo,
		&key.Hostname,
		&key.Status,
		&key.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("generated key not found for server_id: %s", serverID)
		}
		return nil, fmt.Errorf("failed to get generated key: %w", err)
	}

	return &key, nil
}

// Update updates a generated key
func (r *GeneratedKeyRepository) Update(ctx context.Context, key *models.GeneratedKey) error {
	query := `
		UPDATE generated_keys
		SET agent_version = $2, os_info = $3, hostname = $4, status = $5
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		key.ID,
		key.AgentVersion,
		key.OSInfo,
		key.Hostname,
		key.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to update generated key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when updating generated key with id: %d", key.ID)
	}

	r.logger.WithFields(logrus.Fields{
		"id": key.ID,
	}).Info("Generated key updated successfully")

	return nil
}

// Delete deletes a generated key by ID
func (r *GeneratedKeyRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM generated_keys WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete generated key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when deleting generated key with id: %d", id)
	}

	r.logger.WithFields(logrus.Fields{
		"id": id,
	}).Info("Generated key deleted successfully")

	return nil
}

// List retrieves generated keys with optional filters
func (r *GeneratedKeyRepository) List(ctx context.Context, opts ...interfaces.ListOption) ([]*models.GeneratedKey, error) {
	options := &interfaces.ListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	query := `
		SELECT id, server_id, server_key, agent_version, os_info, hostname, status, created_at
		FROM generated_keys
	`

	args := []interface{}{}
	argIndex := 1

	if options.Status != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argIndex)
		args = append(args, options.Status)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, options.Limit)
		argIndex++
	}

	if options.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, options.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list generated keys: %w", err)
	}
	defer rows.Close()

	var keys []*models.GeneratedKey
	for rows.Next() {
		var key models.GeneratedKey
		err := rows.Scan(
			&key.ID,
			&key.ServerID,
			&key.ServerKey,
			&key.AgentVersion,
			&key.OSInfo,
			&key.Hostname,
			&key.Status,
			&key.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan generated key: %w", err)
		}
		keys = append(keys, &key)
	}

	return keys, nil
}

// ListByStatus retrieves generated keys by status
func (r *GeneratedKeyRepository) ListByStatus(ctx context.Context, status string) ([]*models.GeneratedKey, error) {
	return r.List(ctx, interfaces.WithStatus(status))
}

// Ping checks database connectivity
func (r *GeneratedKeyRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
