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

// ServerRepository implements interfaces.ServerRepository for PostgreSQL
type ServerRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewServerRepository creates a new PostgreSQL server repository
func NewServerRepository(db *sql.DB, logger *logrus.Logger) interfaces.ServerRepository {
	return &ServerRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new server
func (r *ServerRepository) Create(ctx context.Context, server *models.Server) error {
	query := `
		INSERT INTO servers (server_id, server_key, hostname, os_info, agent_version, status, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		server.ID,
		server.ServerKey,
		server.Hostname,
		server.OSInfo,
		server.AgentVersion,
		server.Status,
		server.LastSeen,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": server.ID,
		"hostname":  server.Hostname,
	}).Info("Server created successfully")

	return nil
}

// GetByID retrieves a server by ID
func (r *ServerRepository) GetByID(ctx context.Context, id string) (*models.Server, error) {
	query := `
		SELECT server_id, server_key, hostname, os_info, agent_version, status, sources, last_seen, created_at, updated_at
		FROM servers
		WHERE server_id = $1
	`

	var server models.Server
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&server.ID,
		&server.ServerKey,
		&server.Hostname,
		&server.OSInfo,
		&server.AgentVersion,
		&server.Status,
		&server.Sources,
		&server.LastSeen,
		&server.CreatedAt,
		&server.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server not found for id: %s", id)
		}
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	return &server, nil
}

// GetByKey retrieves a server by server key
func (r *ServerRepository) GetByKey(ctx context.Context, serverKey string) (*models.Server, error) {
	query := `
		SELECT server_id, server_key, hostname, os_info, agent_version, status, sources, last_seen, created_at, updated_at
		FROM servers
		WHERE server_key = $1
	`

	var server models.Server
	err := r.db.QueryRowContext(ctx, query, serverKey).Scan(
		&server.ID,
		&server.ServerKey,
		&server.Hostname,
		&server.OSInfo,
		&server.AgentVersion,
		&server.Status,
		&server.Sources,
		&server.LastSeen,
		&server.CreatedAt,
		&server.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server not found for server_key: %s", serverKey)
		}
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	return &server, nil
}

// Update updates a server
func (r *ServerRepository) Update(ctx context.Context, server *models.Server) error {
	query := `
		UPDATE servers
		SET hostname = $2, os_info = $3, agent_version = $4, status = $5, last_seen = $6, updated_at = $7
		WHERE server_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		server.ID,
		server.Hostname,
		server.OSInfo,
		server.AgentVersion,
		server.Status,
		server.LastSeen,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when updating server with id: %s", server.ID)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": server.ID,
	}).Info("Server updated successfully")

	return nil
}

// Delete deletes a server by ID
func (r *ServerRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM servers WHERE server_id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when deleting server with id: %s", id)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": id,
	}).Info("Server deleted successfully")

	return nil
}

// List retrieves servers with optional filters
func (r *ServerRepository) List(ctx context.Context, opts ...interfaces.ListOption) ([]*models.Server, error) {
	options := &interfaces.ListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	query := `
		SELECT server_id, server_key, hostname, os_info, agent_version, status, sources, last_seen, created_at, updated_at
		FROM servers
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
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}
	defer rows.Close()

	var servers []*models.Server
	for rows.Next() {
		var server models.Server
		err := rows.Scan(
			&server.ID,
			&server.ServerKey,
			&server.Hostname,
			&server.OSInfo,
			&server.AgentVersion,
			&server.Status,
			&server.Sources,
			&server.LastSeen,
			&server.CreatedAt,
			&server.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, &server)
	}

	return servers, nil
}

// ListByStatus retrieves servers by status
func (r *ServerRepository) ListByStatus(ctx context.Context, status string) ([]*models.Server, error) {
	return r.List(ctx, interfaces.WithStatus(status))
}

// ListByHostname retrieves servers by hostname
func (r *ServerRepository) ListByHostname(ctx context.Context, hostname string) ([]*models.Server, error) {
	query := `
		SELECT server_id, server_key, hostname, os_info, agent_version, status, sources, last_seen, created_at, updated_at
		FROM servers
		WHERE hostname = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers by hostname: %w", err)
	}
	defer rows.Close()

	var servers []*models.Server
	for rows.Next() {
		var server models.Server
		err := rows.Scan(
			&server.ID,
			&server.ServerKey,
			&server.Hostname,
			&server.OSInfo,
			&server.AgentVersion,
			&server.Status,
			&server.Sources,
			&server.LastSeen,
			&server.CreatedAt,
			&server.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, &server)
	}

	return servers, nil
}

// UpdateStatus updates server status
func (r *ServerRepository) UpdateStatus(ctx context.Context, serverID string, status string) error {
	query := `
		UPDATE servers
		SET status = $2, updated_at = $3
		WHERE server_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, serverID, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update server status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when updating status for server_id: %s", serverID)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"status":    status,
	}).Info("Server status updated successfully")

	return nil
}

// UpdateLastSeen updates server last seen timestamp
func (r *ServerRepository) UpdateLastSeen(ctx context.Context, serverID string, lastSeen time.Time) error {
	query := `
		UPDATE servers
		SET last_seen = $2, updated_at = $3
		WHERE server_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, serverID, lastSeen, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update server last seen: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when updating last seen for server_id: %s", serverID)
	}

	return nil
}

// UpdateSources updates server sources
func (r *ServerRepository) UpdateSources(ctx context.Context, serverID string, sources string) error {
	query := `
		UPDATE servers
		SET sources = $2, updated_at = $3
		WHERE server_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, serverID, sources, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update server sources: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when updating sources for server_id: %s", serverID)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"sources":   sources,
	}).Info("Server sources updated successfully")

	return nil
}

// Ping checks database connectivity
func (r *ServerRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
