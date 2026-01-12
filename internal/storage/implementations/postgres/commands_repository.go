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

// PostgresCommandsRepository implements CommandsRepository for PostgreSQL
type PostgresCommandsRepository struct {
	client *postgres.Client
	logger *logrus.Logger
}

// NewPostgresCommandsRepository creates a new PostgreSQL commands repository
func NewPostgresCommandsRepository(client *postgres.Client, logger *logrus.Logger) repositories.CommandsRepository {
	return &PostgresCommandsRepository{
		client: client,
		logger: logger,
	}
}

// Store stores a new command
func (r *PostgresCommandsRepository) Store(ctx context.Context, serverID string, command *models.Command) error {
	if command.ID == "" {
		command.ID = uuid.New().String()
	}
	if command.CreatedAt.IsZero() {
		command.CreatedAt = time.Now()
	}
	if command.Status == "" {
		command.Status = "pending"
	}

	query := `
		INSERT INTO commands (id, server_id, type, payload, status, created_at, processed_at, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.client.GetDB().ExecContext(ctx, query,
		command.ID, serverID, command.Type, command.Payload,
		command.Status, command.CreatedAt, command.ProcessedAt, command.Error,
	)

	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"command_id": command.ID,
			"server_id":  serverID,
		}).Error("Failed to store command")
		return fmt.Errorf("failed to store command: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"command_id": command.ID,
		"server_id":  serverID,
		"type":       command.Type,
	}).Debug("Command stored successfully")
	return nil
}

// GetPending retrieves pending commands for a server
func (r *PostgresCommandsRepository) GetPending(ctx context.Context, serverID string) ([]*models.Command, error) {
	query := `
		SELECT id, server_id, type, payload, status, created_at, processed_at, error
		FROM commands 
		WHERE server_id = $1 AND status = 'pending'
		ORDER BY created_at ASC
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query, serverID)
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get pending commands")
		return nil, fmt.Errorf("failed to get pending commands: %w", err)
	}
	defer rows.Close()

	var commands []*models.Command
	for rows.Next() {
		command := &models.Command{}
		err := rows.Scan(
			&command.ID, &command.ServerID, &command.Type, &command.Payload,
			&command.Status, &command.CreatedAt, &command.ProcessedAt, &command.Error,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan command row")
			return nil, fmt.Errorf("failed to scan command: %w", err)
		}
		commands = append(commands, command)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating command rows")
		return nil, fmt.Errorf("error iterating commands: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"count":     len(commands),
	}).Debug("Pending commands retrieved successfully")
	return commands, nil
}

// GetHistory retrieves command history for a server
func (r *PostgresCommandsRepository) GetHistory(ctx context.Context, serverID string, limit int) ([]*models.Command, error) {
	query := `
		SELECT id, server_id, type, payload, status, created_at, processed_at, error
		FROM commands 
		WHERE server_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query, serverID, limit)
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to get command history")
		return nil, fmt.Errorf("failed to get command history: %w", err)
	}
	defer rows.Close()

	var commands []*models.Command
	for rows.Next() {
		command := &models.Command{}
		err := rows.Scan(
			&command.ID, &command.ServerID, &command.Type, &command.Payload,
			&command.Status, &command.CreatedAt, &command.ProcessedAt, &command.Error,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan command row")
			return nil, fmt.Errorf("failed to scan command: %w", err)
		}
		commands = append(commands, command)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating command rows")
		return nil, fmt.Errorf("error iterating commands: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"limit":     limit,
		"count":     len(commands),
	}).Debug("Command history retrieved successfully")
	return commands, nil
}

// MarkProcessed marks a command as processed
func (r *PostgresCommandsRepository) MarkProcessed(ctx context.Context, commandID string) error {
	query := `UPDATE commands SET status = 'processed', processed_at = $2 WHERE id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, commandID, time.Now())
	if err != nil {
		r.logger.WithError(err).WithField("command_id", commandID).Error("Failed to mark command as processed")
		return fmt.Errorf("failed to mark command as processed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("command not found: %s", commandID)
	}

	r.logger.WithField("command_id", commandID).Debug("Command marked as processed successfully")
	return nil
}

// MarkFailed marks a command as failed
func (r *PostgresCommandsRepository) MarkFailed(ctx context.Context, commandID string, errorMsg string) error {
	query := `UPDATE commands SET status = 'failed', processed_at = $2, error = $3 WHERE id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, commandID, time.Now(), errorMsg)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"command_id": commandID,
			"error":      errorMsg,
		}).Error("Failed to mark command as failed")
		return fmt.Errorf("failed to mark command as failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("command not found: %s", commandID)
	}

	r.logger.WithFields(logrus.Fields{
		"command_id": commandID,
		"error":      errorMsg,
	}).Debug("Command marked as failed successfully")
	return nil
}

// GetByID retrieves a command by ID
func (r *PostgresCommandsRepository) GetByID(ctx context.Context, commandID string) (*models.Command, error) {
	query := `
		SELECT id, server_id, type, payload, status, created_at, processed_at, error
		FROM commands 
		WHERE id = $1
	`

	command := &models.Command{}
	err := r.client.GetDB().QueryRowContext(ctx, query, commandID).Scan(
		&command.ID, &command.ServerID, &command.Type, &command.Payload,
		&command.Status, &command.CreatedAt, &command.ProcessedAt, &command.Error,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("command not found: %s", commandID)
		}
		r.logger.WithError(err).WithField("command_id", commandID).Error("Failed to get command by ID")
		return nil, fmt.Errorf("failed to get command: %w", err)
	}

	r.logger.WithField("command_id", commandID).Debug("Command retrieved successfully")
	return command, nil
}

// GetByType retrieves commands by type for a server
func (r *PostgresCommandsRepository) GetByType(ctx context.Context, serverID string, commandType string) ([]*models.Command, error) {
	query := `
		SELECT id, server_id, type, payload, status, created_at, processed_at, error
		FROM commands 
		WHERE server_id = $1 AND type = $2
		ORDER BY created_at DESC
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query, serverID, commandType)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"server_id": serverID,
			"type":      commandType,
		}).Error("Failed to get commands by type")
		return nil, fmt.Errorf("failed to get commands by type: %w", err)
	}
	defer rows.Close()

	var commands []*models.Command
	for rows.Next() {
		command := &models.Command{}
		err := rows.Scan(
			&command.ID, &command.ServerID, &command.Type, &command.Payload,
			&command.Status, &command.CreatedAt, &command.ProcessedAt, &command.Error,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan command row")
			return nil, fmt.Errorf("failed to scan command: %w", err)
		}
		commands = append(commands, command)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating command rows")
		return nil, fmt.Errorf("error iterating commands: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"type":      commandType,
		"count":     len(commands),
	}).Debug("Commands retrieved by type successfully")
	return commands, nil
}

// DeleteProcessed deletes processed commands older than specified duration
func (r *PostgresCommandsRepository) DeleteProcessed(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM commands WHERE status = 'processed' AND processed_at < $1`

	cutoffTime := time.Now().Add(-olderThan)
	result, err := r.client.GetDB().ExecContext(ctx, query, cutoffTime)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete processed commands")
		return fmt.Errorf("failed to delete processed commands: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"older_than":    olderThan,
		"rows_affected": rowsAffected,
	}).Debug("Processed commands deleted successfully")
	return nil
}

// DeleteByServer deletes all commands for a server
func (r *PostgresCommandsRepository) DeleteByServer(ctx context.Context, serverID string) error {
	query := `DELETE FROM commands WHERE server_id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, serverID)
	if err != nil {
		r.logger.WithError(err).WithField("server_id", serverID).Error("Failed to delete commands by server")
		return fmt.Errorf("failed to delete commands by server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"server_id":     serverID,
		"rows_affected": rowsAffected,
	}).Debug("Commands deleted by server successfully")
	return nil
}

// Ping checks database connectivity
func (r *PostgresCommandsRepository) Ping(ctx context.Context) error {
	return r.client.Ping()
}
