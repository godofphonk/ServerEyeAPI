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
	"encoding/json"
	"fmt"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// ServerSourceIdentifierRepository implements interfaces.ServerSourceIdentifierRepository for PostgreSQL
type ServerSourceIdentifierRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewServerSourceIdentifierRepository creates a new PostgreSQL server source identifier repository
func NewServerSourceIdentifierRepository(db *sql.DB, logger *logrus.Logger) interfaces.ServerSourceIdentifierRepository {
	return &ServerSourceIdentifierRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new server source identifier
func (r *ServerSourceIdentifierRepository) Create(ctx context.Context, identifier *models.ServerSourceIdentifier) error {
	query := `
		INSERT INTO server_source_identifiers (server_id, source_type, identifier, identifier_type, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var metadataJSON []byte
	var err error
	if identifier.Metadata != nil {
		metadataJSON, err = json.Marshal(identifier.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	} else {
		metadataJSON = []byte("{}")
	}

	err = r.db.QueryRowContext(ctx, query,
		identifier.ServerID,
		identifier.SourceType,
		identifier.Identifier,
		identifier.IdentifierType,
		metadataJSON,
		time.Now(),
		time.Now(),
	).Scan(&identifier.ID)

	if err != nil {
		return fmt.Errorf("failed to create server source identifier: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"id":               identifier.ID,
		"server_id":        identifier.ServerID,
		"source_type":      identifier.SourceType,
		"identifier":       identifier.Identifier,
		"identifier_type":  identifier.IdentifierType,
	}).Info("Server source identifier created successfully")

	return nil
}

// GetByID retrieves a server source identifier by ID
func (r *ServerSourceIdentifierRepository) GetByID(ctx context.Context, id int64) (*models.ServerSourceIdentifier, error) {
	query := `
		SELECT id, server_id, source_type, identifier, identifier_type, metadata, created_at, updated_at
		FROM server_source_identifiers
		WHERE id = $1
	`

	var identifier models.ServerSourceIdentifier
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&identifier.ID,
		&identifier.ServerID,
		&identifier.SourceType,
		&identifier.Identifier,
		&identifier.IdentifierType,
		&metadataJSON,
		&identifier.CreatedAt,
		&identifier.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server source identifier not found for id: %d", id)
		}
		return nil, fmt.Errorf("failed to get server source identifier: %w", err)
	}

	if len(metadataJSON) > 0 && string(metadataJSON) != "null" {
		if err := json.Unmarshal(metadataJSON, &identifier.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		identifier.Metadata = make(map[string]interface{})
	}

	return &identifier, nil
}

// GetByServerID retrieves all server source identifiers for a server
func (r *ServerSourceIdentifierRepository) GetByServerID(ctx context.Context, serverID string) ([]*models.ServerSourceIdentifier, error) {
	query := `
		SELECT id, server_id, source_type, identifier, identifier_type, metadata, created_at, updated_at
		FROM server_source_identifiers
		WHERE server_id = $1
		ORDER BY created_at DESC
	`

	return r.scanIdentifiers(ctx, query, serverID)
}

// GetByServerIDAndSourceType retrieves identifiers for a server and source type
func (r *ServerSourceIdentifierRepository) GetByServerIDAndSourceType(ctx context.Context, serverID, sourceType string) ([]*models.ServerSourceIdentifier, error) {
	query := `
		SELECT id, server_id, source_type, identifier, identifier_type, metadata, created_at, updated_at
		FROM server_source_identifiers
		WHERE server_id = $1 AND source_type = $2
		ORDER BY created_at DESC
	`

	return r.scanIdentifiers(ctx, query, serverID, sourceType)
}

// GetByServerIDAndIdentifier retrieves a specific identifier
func (r *ServerSourceIdentifierRepository) GetByServerIDAndIdentifier(ctx context.Context, serverID, sourceType, identifier string) (*models.ServerSourceIdentifier, error) {
	query := `
		SELECT id, server_id, source_type, identifier, identifier_type, metadata, created_at, updated_at
		FROM server_source_identifiers
		WHERE server_id = $1 AND source_type = $2 AND identifier = $3
	`

	var ident models.ServerSourceIdentifier
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, serverID, sourceType, identifier).Scan(
		&ident.ID,
		&ident.ServerID,
		&ident.SourceType,
		&ident.Identifier,
		&ident.IdentifierType,
		&metadataJSON,
		&ident.CreatedAt,
		&ident.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server source identifier not found")
		}
		return nil, fmt.Errorf("failed to get server source identifier: %w", err)
	}

	if len(metadataJSON) > 0 && string(metadataJSON) != "null" {
		if err := json.Unmarshal(metadataJSON, &ident.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		ident.Metadata = make(map[string]interface{})
	}

	return &ident, nil
}

// Update updates a server source identifier
func (r *ServerSourceIdentifierRepository) Update(ctx context.Context, identifier *models.ServerSourceIdentifier) error {
	query := `
		UPDATE server_source_identifiers
		SET identifier_type = $2, metadata = $3, updated_at = $4
		WHERE id = $1
	`

	var metadataJSON []byte
	var err error
	if identifier.Metadata != nil {
		metadataJSON, err = json.Marshal(identifier.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	} else {
		metadataJSON = []byte("{}")
	}

	result, err := r.db.ExecContext(ctx, query,
		identifier.ID,
		identifier.IdentifierType,
		metadataJSON,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update server source identifier: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when updating server source identifier with id: %d", identifier.ID)
	}

	r.logger.WithFields(logrus.Fields{
		"id": identifier.ID,
	}).Info("Server source identifier updated successfully")

	return nil
}

// Delete deletes a server source identifier by ID
func (r *ServerSourceIdentifierRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM server_source_identifiers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete server source identifier: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when deleting server source identifier with id: %d", id)
	}

	r.logger.WithFields(logrus.Fields{
		"id": id,
	}).Info("Server source identifier deleted successfully")

	return nil
}

// DeleteByServerIDAndSourceType deletes all identifiers for a server and source type
func (r *ServerSourceIdentifierRepository) DeleteByServerIDAndSourceType(ctx context.Context, serverID, sourceType string) error {
	query := `DELETE FROM server_source_identifiers WHERE server_id = $1 AND source_type = $2`

	result, err := r.db.ExecContext(ctx, query, serverID, sourceType)
	if err != nil {
		return fmt.Errorf("failed to delete server source identifiers: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id":   serverID,
		"source_type": sourceType,
		"rows_count":  rowsAffected,
	}).Info("Server source identifiers deleted successfully")

	return nil
}

// DeleteByServerIDSourceTypeAndIdentifier deletes a specific identifier
func (r *ServerSourceIdentifierRepository) DeleteByServerIDSourceTypeAndIdentifier(ctx context.Context, serverID, sourceType, identifier string) error {
	query := `DELETE FROM server_source_identifiers WHERE server_id = $1 AND source_type = $2 AND identifier = $3`

	result, err := r.db.ExecContext(ctx, query, serverID, sourceType, identifier)
	if err != nil {
		return fmt.Errorf("failed to delete server source identifier: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when deleting server source identifier")
	}

	r.logger.WithFields(logrus.Fields{
		"server_id":   serverID,
		"source_type": sourceType,
		"identifier":  identifier,
	}).Info("Server source identifier deleted successfully")

	return nil
}

// CreateBatch creates multiple identifiers in a single transaction
func (r *ServerSourceIdentifierRepository) CreateBatch(ctx context.Context, identifiers []*models.ServerSourceIdentifier) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO server_source_identifiers (server_id, source_type, identifier, identifier_type, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	for _, identifier := range identifiers {
		var metadataJSON []byte
		if identifier.Metadata != nil {
			metadataJSON, err = json.Marshal(identifier.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}
		} else {
			metadataJSON = []byte("{}")
		}

		err = tx.QueryRowContext(ctx, query,
			identifier.ServerID,
			identifier.SourceType,
			identifier.Identifier,
			identifier.IdentifierType,
			metadataJSON,
			time.Now(),
			time.Now(),
		).Scan(&identifier.ID)

		if err != nil {
			return fmt.Errorf("failed to create server source identifier: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"count": len(identifiers),
	}).Info("Server source identifiers created successfully in batch")

	return nil
}

// DeleteBatch deletes multiple identifiers in a single transaction
func (r *ServerSourceIdentifierRepository) DeleteBatch(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `DELETE FROM server_source_identifiers WHERE id = ANY($1)`

	result, err := tx.ExecContext(ctx, query, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("failed to delete server source identifiers: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"count":       len(ids),
		"rows_count":  rowsAffected,
	}).Info("Server source identifiers deleted successfully in batch")

	return nil
}

// GetAllByServerID retrieves all identifiers grouped by source type
func (r *ServerSourceIdentifierRepository) GetAllByServerID(ctx context.Context, serverID string) (map[string][]*models.ServerSourceIdentifier, error) {
	identifiers, err := r.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*models.ServerSourceIdentifier)
	for _, identifier := range identifiers {
		result[identifier.SourceType] = append(result[identifier.SourceType], identifier)
	}

	return result, nil
}

// GetByIdentifier finds all servers with a specific identifier
func (r *ServerSourceIdentifierRepository) GetByIdentifier(ctx context.Context, identifierType, identifier string) ([]*models.ServerSourceIdentifier, error) {
	query := `
		SELECT id, server_id, source_type, identifier, identifier_type, metadata, created_at, updated_at
		FROM server_source_identifiers
		WHERE identifier_type = $1 AND identifier = $2
		ORDER BY created_at DESC
	`

	return r.scanIdentifiers(ctx, query, identifierType, identifier)
}

// Ping checks database connectivity
func (r *ServerSourceIdentifierRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Helper method to scan identifiers from rows
func (r *ServerSourceIdentifierRepository) scanIdentifiers(ctx context.Context, query string, args ...interface{}) ([]*models.ServerSourceIdentifier, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query server source identifiers: %w", err)
	}
	defer rows.Close()

	var identifiers []*models.ServerSourceIdentifier
	for rows.Next() {
		var identifier models.ServerSourceIdentifier
		var metadataJSON []byte

		err := rows.Scan(
			&identifier.ID,
			&identifier.ServerID,
			&identifier.SourceType,
			&identifier.Identifier,
			&identifier.IdentifierType,
			&metadataJSON,
			&identifier.CreatedAt,
			&identifier.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server source identifier: %w", err)
		}

		if len(metadataJSON) > 0 && string(metadataJSON) != "null" {
			if err := json.Unmarshal(metadataJSON, &identifier.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		} else {
			identifier.Metadata = make(map[string]interface{})
		}

		identifiers = append(identifiers, &identifier)
	}

	return identifiers, nil
}
