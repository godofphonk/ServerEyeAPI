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
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PostgresConnectionsRepository implements ConnectionsRepository for PostgreSQL
type PostgresConnectionsRepository struct {
	client *postgres.Client
	logger *logrus.Logger
}

// NewPostgresConnectionsRepository creates a new PostgreSQL connections repository
func NewPostgresConnectionsRepository(client *postgres.Client, logger *logrus.Logger) repositories.ConnectionsRepository {
	return &PostgresConnectionsRepository{
		client: client,
		logger: logger,
	}
}

// Store stores a new connection
func (r *PostgresConnectionsRepository) Store(ctx context.Context, serverID string, conn *models.Connection) error {
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	if conn.ConnectedAt.IsZero() {
		conn.ConnectedAt = time.Now()
	}
	if conn.LastActivity.IsZero() {
		conn.LastActivity = time.Now()
	}
	if conn.Status == "" {
		conn.Status = "active"
	}

	query := `
		INSERT INTO connections (id, server_id, type, remote_addr, user_agent, status, metadata, connected_at, disconnected_at, last_activity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.client.GetDB().ExecContext(ctx, query,
		conn.ID, serverID, conn.Type, conn.RemoteAddr, conn.UserAgent,
		conn.Status, conn.Metadata, conn.ConnectedAt, conn.DisconnectedAt, conn.LastActivity,
	)

	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"connection_id": conn.ID,
			"server_id":     serverID,
		}).Error("Failed to store connection")
		return fmt.Errorf("failed to store connection: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"connection_id": conn.ID,
		"server_id":     serverID,
		"type":          conn.Type,
	}).Debug("Connection stored successfully")
	return nil
}

// GetActive retrieves active connections for a server
func (r *PostgresConnectionsRepository) GetActive(ctx context.Context, serverID string) ([]*models.Connection, error) {
	query := `
		SELECT id, server_id, type, remote_addr, user_agent, status, metadata, connected_at, disconnected_at, last_activity
		FROM connections 
		WHERE server_id = $1 AND status = 'active'
		ORDER BY connected_at DESC
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query, serverID)
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get active connections")
		return nil, fmt.Errorf("failed to get active connections: %w", err)
	}
	defer rows.Close()

	var connections []*models.Connection
	for rows.Next() {
		conn := &models.Connection{}
		err := rows.Scan(
			&conn.ID, &conn.ServerID, &conn.Type, &conn.RemoteAddr, &conn.UserAgent,
			&conn.Status, &conn.Metadata, &conn.ConnectedAt, &conn.DisconnectedAt, &conn.LastActivity,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan connection row")
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}
		connections = append(connections, conn)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating connection rows")
		return nil, fmt.Errorf("error iterating connections: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"count":     len(connections),
	}).Debug("Active connections retrieved successfully")
	return connections, nil
}

// GetHistory retrieves connection history for a server
func (r *PostgresConnectionsRepository) GetHistory(ctx context.Context, serverID string, limit int) ([]*models.Connection, error) {
	query := `
		SELECT id, server_id, type, remote_addr, user_agent, status, metadata, connected_at, disconnected_at, last_activity
		FROM connections 
		WHERE server_id = $1
		ORDER BY connected_at DESC
		LIMIT $2
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query, serverID, limit)
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get connection history")
		return nil, fmt.Errorf("failed to get connection history: %w", err)
	}
	defer rows.Close()

	var connections []*models.Connection
	for rows.Next() {
		conn := &models.Connection{}
		err := rows.Scan(
			&conn.ID, &conn.ServerID, &conn.Type, &conn.RemoteAddr, &conn.UserAgent,
			&conn.Status, &conn.Metadata, &conn.ConnectedAt, &conn.DisconnectedAt, &conn.LastActivity,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan connection row")
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}
		connections = append(connections, conn)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating connection rows")
		return nil, fmt.Errorf("error iterating connections: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"limit":     limit,
		"count":     len(connections),
	}).Debug("Connection history retrieved successfully")
	return connections, nil
}

// GetByID retrieves a connection by ID
func (r *PostgresConnectionsRepository) GetByID(ctx context.Context, connectionID string) (*models.Connection, error) {
	query := `
		SELECT id, server_id, type, remote_addr, user_agent, status, metadata, connected_at, disconnected_at, last_activity
		FROM connections 
		WHERE id = $1
	`

	conn := &models.Connection{}
	err := r.client.GetDB().QueryRowContext(ctx, query, connectionID).Scan(
		&conn.ID, &conn.ServerID, &conn.Type, &conn.RemoteAddr, &conn.UserAgent,
		&conn.Status, &conn.Metadata, &conn.ConnectedAt, &conn.DisconnectedAt, &conn.LastActivity,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("connection not found: %s", connectionID)
		}
		r.logger.WithError(err).WithField("connection_id", connectionID).Error("Failed to get connection by ID")
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	r.logger.WithField("connection_id", connectionID).Debug("Connection retrieved successfully")
	return conn, nil
}

// GetByType retrieves connections by type for a server
func (r *PostgresConnectionsRepository) GetByType(ctx context.Context, serverID string, connType string) ([]*models.Connection, error) {
	query := `
		SELECT id, server_id, type, remote_addr, user_agent, status, metadata, connected_at, disconnected_at, last_activity
		FROM connections 
		WHERE server_id = $1 AND type = $2
		ORDER BY connected_at DESC
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query, serverID, connType)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"server_id": serverID,
			"type":      connType,
		}).Error("Failed to get connections by type")
		return nil, fmt.Errorf("failed to get connections by type: %w", err)
	}
	defer rows.Close()

	var connections []*models.Connection
	for rows.Next() {
		conn := &models.Connection{}
		err := rows.Scan(
			&conn.ID, &conn.ServerID, &conn.Type, &conn.RemoteAddr, &conn.UserAgent,
			&conn.Status, &conn.Metadata, &conn.ConnectedAt, &conn.DisconnectedAt, &conn.LastActivity,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan connection row")
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}
		connections = append(connections, conn)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating connection rows")
		return nil, fmt.Errorf("error iterating connections: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"type":      connType,
		"count":     len(connections),
	}).Debug("Connections retrieved by type successfully")
	return connections, nil
}

// GetAll retrieves all connections
func (r *PostgresConnectionsRepository) GetAll(ctx context.Context) ([]*models.Connection, error) {
	query := `
		SELECT id, server_id, type, remote_addr, user_agent, status, metadata, connected_at, disconnected_at, last_activity
		FROM connections 
		ORDER BY connected_at DESC
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get all connections")
		return nil, fmt.Errorf("failed to get all connections: %w", err)
	}
	defer rows.Close()

	var connections []*models.Connection
	for rows.Next() {
		conn := &models.Connection{}
		err := rows.Scan(
			&conn.ID, &conn.ServerID, &conn.Type, &conn.RemoteAddr, &conn.UserAgent,
			&conn.Status, &conn.Metadata, &conn.ConnectedAt, &conn.DisconnectedAt, &conn.LastActivity,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan connection row")
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}
		connections = append(connections, conn)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating connection rows")
		return nil, fmt.Errorf("error iterating connections: %w", err)
	}

	r.logger.WithField("count", len(connections)).Debug("All connections retrieved successfully")
	return connections, nil
}

// Close closes a connection
func (r *PostgresConnectionsRepository) Close(ctx context.Context, connectionID string) error {
	query := `UPDATE connections SET status = 'disconnected', disconnected_at = $2 WHERE id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, connectionID, time.Now())
	if err != nil {
		r.logger.WithError(err).WithField("connection_id", connectionID).Error("Failed to close connection")
		return fmt.Errorf("failed to close connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("connection not found: %s", connectionID)
	}

	r.logger.WithField("connection_id", connectionID).Debug("Connection closed successfully")
	return nil
}

// CloseByServer closes all connections for a server
func (r *PostgresConnectionsRepository) CloseByServer(ctx context.Context, serverID string) error {
	query := `UPDATE connections SET status = 'disconnected', disconnected_at = $2 WHERE server_id = $1 AND status = 'active'`

	result, err := r.client.GetDB().ExecContext(ctx, query, serverID, time.Now())
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to close connections by server")
		return fmt.Errorf("failed to close connections by server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id":     serverID,
		"rows_affected": rowsAffected,
	}).Debug("Connections closed by server successfully")
	return nil
}

// MarkDisconnected marks a connection as disconnected
func (r *PostgresConnectionsRepository) MarkDisconnected(ctx context.Context, connectionID string) error {
	query := `UPDATE connections SET status = 'disconnected', disconnected_at = $2 WHERE id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, connectionID, time.Now())
	if err != nil {
		r.logger.WithError(err).WithField("connection_id", connectionID).Error("Failed to mark connection as disconnected")
		return fmt.Errorf("failed to mark connection as disconnected: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("connection not found: %s", connectionID)
	}

	r.logger.WithField("connection_id", connectionID).Debug("Connection marked as disconnected successfully")
	return nil
}

// DeleteOlderThan deletes connections older than specified duration
func (r *PostgresConnectionsRepository) DeleteOlderThan(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM connections WHERE connected_at < $1`

	cutoffTime := time.Now().Add(-olderThan)
	result, err := r.client.GetDB().ExecContext(ctx, query, cutoffTime)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete older connections")
		return fmt.Errorf("failed to delete older connections: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"older_than":    olderThan,
		"rows_affected": rowsAffected,
	}).Debug("Older connections deleted successfully")
	return nil
}

// DeleteDisconnected deletes disconnected connections older than specified duration
func (r *PostgresConnectionsRepository) DeleteDisconnected(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM connections WHERE status = 'disconnected' AND disconnected_at < $1`

	cutoffTime := time.Now().Add(-olderThan)
	result, err := r.client.GetDB().ExecContext(ctx, query, cutoffTime)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete disconnected connections")
		return fmt.Errorf("failed to delete disconnected connections: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"older_than":    olderThan,
		"rows_affected": rowsAffected,
	}).Debug("Disconnected connections deleted successfully")
	return nil
}

// DeleteByServer deletes all connections for a server
func (r *PostgresConnectionsRepository) DeleteByServer(ctx context.Context, serverID string) error {
	query := `DELETE FROM connections WHERE server_id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, serverID)
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to delete connections by server")
		return fmt.Errorf("failed to delete connections by server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id":     serverID,
		"rows_affected": rowsAffected,
	}).Debug("Connections deleted by server successfully")
	return nil
}

// Ping checks database connectivity
func (r *PostgresConnectionsRepository) Ping(ctx context.Context) error {
	return r.client.Ping()
}
