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
	"github.com/godofphonk/ServerEyeAPI/internal/storage/postgres"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/repositories"
	"github.com/sirupsen/logrus"
)

// PostgresServerRepository implements ServerRepository for PostgreSQL
type PostgresServerRepository struct {
	client *postgres.Client
	logger *logrus.Logger
}

// NewPostgresServerRepository creates a new PostgreSQL server repository
func NewPostgresServerRepository(client *postgres.Client, logger *logrus.Logger) repositories.ServerRepository {
	return &PostgresServerRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new server
func (r *PostgresServerRepository) Create(ctx context.Context, server *models.Server) error {
	query := `
		INSERT INTO servers (server_id, server_key, secret_key, hostname, os_info, agent_version, status, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (server_id) DO UPDATE SET
			server_key = EXCLUDED.server_key,
			secret_key = EXCLUDED.secret_key,
			hostname = EXCLUDED.hostname,
			os_info = EXCLUDED.os_info,
			agent_version = EXCLUDED.agent_version,
			status = EXCLUDED.status,
			last_seen = EXCLUDED.last_seen,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.client.GetDB().ExecContext(ctx, query,
		server.ID, server.ServerKey, server.SecretKey, server.Hostname,
		server.OSInfo, server.AgentVersion, server.Status,
		server.LastSeen, server.CreatedAt, server.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("server_id", server.ID).Error("Failed to create server")
		return fmt.Errorf("failed to create server: %w", err)
	}

	r.logger.WithField("server_id", server.ID).Debug("Server created successfully")
	return nil
}

// GetByID retrieves a server by ID
func (r *PostgresServerRepository) GetByID(ctx context.Context, id string) (*models.Server, error) {
	query := `
		SELECT server_id, server_key, secret_key, hostname, os_info, agent_version, 
			   status, last_seen, created_at, updated_at
		FROM servers 
		WHERE server_id = $1
	`

	server := &models.Server{}
	err := r.client.GetDB().QueryRowContext(ctx, query, id).Scan(
		&server.ID, &server.ServerKey, &server.SecretKey, &server.Hostname,
		&server.OSInfo, &server.AgentVersion, &server.Status,
		&server.LastSeen, &server.CreatedAt, &server.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server not found: %s", id)
		}
		r.logger.WithError(err).WithField("server_id", id).Error("Failed to get server by ID")
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	r.logger.WithField("server_id", id).Debug("Server retrieved successfully")
	return server, nil
}

// GetByKey retrieves a server by server key
func (r *PostgresServerRepository) GetByKey(ctx context.Context, serverKey string) (*models.Server, error) {
	query := `
		SELECT server_id, server_key, secret_key, hostname, os_info, agent_version, 
			   status, last_seen, created_at, updated_at
		FROM servers 
		WHERE server_key = $1
	`

	server := &models.Server{}
	err := r.client.DB().QueryRowContext(ctx, query, serverKey).Scan(
		&server.ID, &server.ServerKey, &server.SecretKey, &server.Hostname,
		&server.OSInfo, &server.AgentVersion, &server.Status,
		&server.LastSeen, &server.CreatedAt, &server.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server not found for key: %s", serverKey)
		}
		r.logger.WithError(err).WithField("server_key", serverKey).Error("Failed to get server by key")
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	r.logger.WithField("server_key", serverKey).Debug("Server retrieved successfully")
	return server, nil
}

// Update updates a server
func (r *PostgresServerRepository) Update(ctx context.Context, server *models.Server) error {
	query := `
		UPDATE servers 
		SET hostname = $2, os_info = $3, agent_version = $4, status = $5, 
			last_seen = $6, updated_at = $7
		WHERE server_id = $1
	`

	server.UpdatedAt = time.Now()

	result, err := r.client.GetDB().ExecContext(ctx, query,
		server.ID, server.Hostname, server.OSInfo, server.AgentVersion,
		server.Status, server.LastSeen, server.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("server_id", server.ID).Error("Failed to update server")
		return fmt.Errorf("failed to update server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server not found: %s", server.ID)
	}

	r.logger.WithField("server_id", server.ID).Debug("Server updated successfully")
	return nil
}

// Delete deletes a server
func (r *PostgresServerRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM servers WHERE server_id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).WithField("server_id", id).Error("Failed to delete server")
		return fmt.Errorf("failed to delete server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server not found: %s", id)
	}

	r.logger.WithField("server_id", id).Debug("Server deleted successfully")
	return nil
}

// List retrieves all servers
func (r *PostgresServerRepository) List(ctx context.Context) ([]*models.Server, error) {
	query := `
		SELECT server_id, server_key, secret_key, hostname, os_info, agent_version, 
			   status, last_seen, created_at, updated_at
		FROM servers 
		ORDER BY created_at DESC
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list servers")
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}
	defer rows.Close()

	var servers []*models.Server
	for rows.Next() {
		server := &models.Server{}
		err := rows.Scan(
			&server.ID, &server.ServerKey, &server.SecretKey, &server.Hostname,
			&server.OSInfo, &server.AgentVersion, &server.Status,
			&server.LastSeen, &server.CreatedAt, &server.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan server row")
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, server)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating server rows")
		return nil, fmt.Errorf("error iterating servers: %w", err)
	}

	r.logger.WithField("count", len(servers)).Debug("Servers listed successfully")
	return servers, nil
}

// ListByStatus retrieves servers by status
func (r *PostgresServerRepository) ListByStatus(ctx context.Context, status string) ([]*models.Server, error) {
	query := `
		SELECT server_id, server_key, secret_key, hostname, os_info, agent_version, 
			   status, last_seen, created_at, updated_at
		FROM servers 
		WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.client.DB().QueryContext(ctx, query, status)
	if err != nil {
		r.logger.WithError(err).WithField("status", status).Error("Failed to list servers by status")
		return nil, fmt.Errorf("failed to list servers by status: %w", err)
	}
	defer rows.Close()

	var servers []*models.Server
	for rows.Next() {
		server := &models.Server{}
		err := rows.Scan(
			&server.ID, &server.ServerKey, &server.SecretKey, &server.Hostname,
			&server.OSInfo, &server.AgentVersion, &server.Status,
			&server.LastSeen, &server.CreatedAt, &server.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan server row")
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, server)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating server rows")
		return nil, fmt.Errorf("error iterating servers: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"status": status,
		"count":  len(servers),
	}).Debug("Servers listed by status successfully")
	return servers, nil
}

// ListByHostname retrieves servers by hostname
func (r *PostgresServerRepository) ListByHostname(ctx context.Context, hostname string) ([]*models.Server, error) {
	query := `
		SELECT server_id, server_key, secret_key, hostname, os_info, agent_version, 
			   status, last_seen, created_at, updated_at
		FROM servers 
		WHERE hostname = $1
		ORDER BY created_at DESC
	`

	rows, err := r.client.DB().QueryContext(ctx, query, hostname)
	if err != nil {
		r.logger.WithError(err).WithField("hostname", hostname).Error("Failed to list servers by hostname")
		return nil, fmt.Errorf("failed to list servers by hostname: %w", err)
	}
	defer rows.Close()

	var servers []*models.Server
	for rows.Next() {
		server := &models.Server{}
		err := rows.Scan(
			&server.ID, &server.ServerKey, &server.SecretKey, &server.Hostname,
			&server.OSInfo, &server.AgentVersion, &server.Status,
			&server.LastSeen, &server.CreatedAt, &server.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan server row")
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, server)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating server rows")
		return nil, fmt.Errorf("error iterating servers: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"hostname": hostname,
		"count":    len(servers),
	}).Debug("Servers listed by hostname successfully")
	return servers, nil
}

// UpdateStatus updates server status
func (r *PostgresServerRepository) UpdateStatus(ctx context.Context, serverID string, status string) error {
	query := `UPDATE servers SET status = $2, updated_at = $3 WHERE server_id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, serverID, status, time.Now())
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"server_id": serverID,
			"status":    status,
		}).Error("Failed to update server status")
		return fmt.Errorf("failed to update server status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server not found: %s", serverID)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"status":    status,
	}).Debug("Server status updated successfully")
	return nil
}

// UpdateLastSeen updates server last seen timestamp
func (r *PostgresServerRepository) UpdateLastSeen(ctx context.Context, serverID string, lastSeen time.Time) error {
	query := `UPDATE servers SET last_seen = $2, updated_at = $3 WHERE server_id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, serverID, lastSeen, time.Now())
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to update server last seen")
		return fmt.Errorf("failed to update server last seen: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server not found: %s", serverID)
	}

	r.logger.WithField("server_id", serverID).Debug("Server last seen updated successfully")
	return nil
}

// Ping checks database connectivity
func (r *PostgresServerRepository) Ping(ctx context.Context) error {
	return r.client.Ping()
}
