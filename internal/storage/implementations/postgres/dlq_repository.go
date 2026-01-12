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

// PostgresDLQRepository implements DLQRepository for PostgreSQL
type PostgresDLQRepository struct {
	client *postgres.Client
	logger *logrus.Logger
}

// NewPostgresDLQRepository creates a new PostgreSQL DLQ repository
func NewPostgresDLQRepository(client *postgres.Client, logger *logrus.Logger) repositories.DLQRepository {
	return &PostgresDLQRepository{
		client: client,
		logger: logger,
	}
}

// Store stores a new DLQ message
func (r *PostgresDLQRepository) Store(ctx context.Context, dlq *models.DLQMessage) error {
	if dlq.ID == "" {
		dlq.ID = uuid.New().String()
	}
	if dlq.CreatedAt.IsZero() {
		dlq.CreatedAt = time.Now()
	}
	if dlq.UpdatedAt.IsZero() {
		dlq.UpdatedAt = time.Now()
	}
	if dlq.Status == "" {
		dlq.Status = "pending"
	}

	query := `
		INSERT INTO dlq_messages (id, topic, message, metadata, error, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.client.GetDB().ExecContext(ctx, query,
		dlq.ID, dlq.Topic, dlq.Message, dlq.Metadata,
		dlq.Error, dlq.Status, dlq.CreatedAt, dlq.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"dlq_id": dlq.ID,
			"topic":  dlq.Topic,
		}).Error("Failed to store DLQ message")
		return fmt.Errorf("failed to store DLQ message: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"dlq_id": dlq.ID,
		"topic":  dlq.Topic,
	}).Debug("DLQ message stored successfully")
	return nil
}

// GetByTopic retrieves DLQ messages by topic
func (r *PostgresDLQRepository) GetByTopic(ctx context.Context, topic string, limit int) ([]*models.DLQMessage, error) {
	query := `
		SELECT id, topic, message, metadata, error, status, created_at, updated_at
		FROM dlq_messages 
		WHERE topic = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query, topic, limit)
	if err != nil {
		r.logger.WithError(err).WithField("topic", topic).Error("Failed to get DLQ messages by topic")
		return nil, fmt.Errorf("failed to get DLQ messages by topic: %w", err)
	}
	defer rows.Close()

	var messages []*models.DLQMessage
	for rows.Next() {
		message := &models.DLQMessage{}
		err := rows.Scan(
			&message.ID, &message.Topic, &message.Message, &message.Metadata,
			&message.Error, &message.Status, &message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan DLQ message row")
			return nil, fmt.Errorf("failed to scan DLQ message: %w", err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating DLQ message rows")
		return nil, fmt.Errorf("error iterating DLQ messages: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"topic": topic,
		"limit": limit,
		"count": len(messages),
	}).Debug("DLQ messages retrieved by topic successfully")
	return messages, nil
}

// GetByID retrieves a DLQ message by ID
func (r *PostgresDLQRepository) GetByID(ctx context.Context, id string) (*models.DLQMessage, error) {
	query := `
		SELECT id, topic, message, metadata, error, status, created_at, updated_at
		FROM dlq_messages 
		WHERE id = $1
	`

	message := &models.DLQMessage{}
	err := r.client.GetDB().QueryRowContext(ctx, query, id).Scan(
		&message.ID, &message.Topic, &message.Message, &message.Metadata,
		&message.Error, &message.Status, &message.CreatedAt, &message.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("DLQ message not found: %s", id)
		}
		r.logger.WithError(err).WithField("dlq_id", id).Error("Failed to get DLQ message by ID")
		return nil, fmt.Errorf("failed to get DLQ message: %w", err)
	}

	r.logger.WithField("dlq_id", id).Debug("DLQ message retrieved successfully")
	return message, nil
}

// GetAll retrieves all DLQ messages
func (r *PostgresDLQRepository) GetAll(ctx context.Context) ([]*models.DLQMessage, error) {
	query := `
		SELECT id, topic, message, metadata, error, status, created_at, updated_at
		FROM dlq_messages 
		ORDER BY created_at DESC
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get all DLQ messages")
		return nil, fmt.Errorf("failed to get all DLQ messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.DLQMessage
	for rows.Next() {
		message := &models.DLQMessage{}
		err := rows.Scan(
			&message.ID, &message.Topic, &message.Message, &message.Metadata,
			&message.Error, &message.Status, &message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan DLQ message row")
			return nil, fmt.Errorf("failed to scan DLQ message: %w", err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating DLQ message rows")
		return nil, fmt.Errorf("error iterating DLQ messages: %w", err)
	}

	r.logger.WithField("count", len(messages)).Debug("All DLQ messages retrieved successfully")
	return messages, nil
}

// GetByStatus retrieves DLQ messages by status
func (r *PostgresDLQRepository) GetByStatus(ctx context.Context, status string, limit int) ([]*models.DLQMessage, error) {
	query := `
		SELECT id, topic, message, metadata, error, status, created_at, updated_at
		FROM dlq_messages 
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.client.GetDB().QueryContext(ctx, query, status, limit)
	if err != nil {
		r.logger.WithError(err).WithField("status", status).Error("Failed to get DLQ messages by status")
		return nil, fmt.Errorf("failed to get DLQ messages by status: %w", err)
	}
	defer rows.Close()

	var messages []*models.DLQMessage
	for rows.Next() {
		message := &models.DLQMessage{}
		err := rows.Scan(
			&message.ID, &message.Topic, &message.Message, &message.Metadata,
			&message.Error, &message.Status, &message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan DLQ message row")
			return nil, fmt.Errorf("failed to scan DLQ message: %w", err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating DLQ message rows")
		return nil, fmt.Errorf("error iterating DLQ messages: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"status": status,
		"limit":  limit,
		"count":  len(messages),
	}).Debug("DLQ messages retrieved by status successfully")
	return messages, nil
}

// GetOlderThan retrieves DLQ messages older than specified duration
func (r *PostgresDLQRepository) GetOlderThan(ctx context.Context, olderThan time.Duration) ([]*models.DLQMessage, error) {
	query := `
		SELECT id, topic, message, metadata, error, status, created_at, updated_at
		FROM dlq_messages 
		WHERE created_at < $1
		ORDER BY created_at ASC
	`

	cutoffTime := time.Now().Add(-olderThan)
	rows, err := r.client.GetDB().QueryContext(ctx, query, cutoffTime)
	if err != nil {
		r.logger.WithError(err).WithField("older_than", olderThan).Error("Failed to get older DLQ messages")
		return nil, fmt.Errorf("failed to get older DLQ messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.DLQMessage
	for rows.Next() {
		message := &models.DLQMessage{}
		err := rows.Scan(
			&message.ID, &message.Topic, &message.Message, &message.Metadata,
			&message.Error, &message.Status, &message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan DLQ message row")
			return nil, fmt.Errorf("failed to scan DLQ message: %w", err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating DLQ message rows")
		return nil, fmt.Errorf("error iterating DLQ messages: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"older_than": olderThan,
		"count":      len(messages),
	}).Debug("Older DLQ messages retrieved successfully")
	return messages, nil
}

// Delete deletes a DLQ message
func (r *PostgresDLQRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM dlq_messages WHERE id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).WithField("dlq_id", id).Error("Failed to delete DLQ message")
		return fmt.Errorf("failed to delete DLQ message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("DLQ message not found: %s", id)
	}

	r.logger.WithField("dlq_id", id).Debug("DLQ message deleted successfully")
	return nil
}

// Requeue marks a DLQ message as requeued
func (r *PostgresDLQRepository) Requeue(ctx context.Context, id string) error {
	query := `UPDATE dlq_messages SET status = 'requeued', updated_at = $2 WHERE id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, id, time.Now())
	if err != nil {
		r.logger.WithError(err).WithField("dlq_id", id).Error("Failed to requeue DLQ message")
		return fmt.Errorf("failed to requeue DLQ message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("DLQ message not found: %s", id)
	}

	r.logger.WithField("dlq_id", id).Debug("DLQ message requeued successfully")
	return nil
}

// MarkProcessed marks a DLQ message as processed
func (r *PostgresDLQRepository) MarkProcessed(ctx context.Context, id string) error {
	query := `UPDATE dlq_messages SET status = 'processed', updated_at = $2 WHERE id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, id, time.Now())
	if err != nil {
		r.logger.WithError(err).WithField("dlq_id", id).Error("Failed to mark DLQ message as processed")
		return fmt.Errorf("failed to mark DLQ message as processed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("DLQ message not found: %s", id)
	}

	r.logger.WithField("dlq_id", id).Debug("DLQ message marked as processed successfully")
	return nil
}

// MarkFailed marks a DLQ message as failed
func (r *PostgresDLQRepository) MarkFailed(ctx context.Context, id string, errorMsg string) error {
	query := `UPDATE dlq_messages SET status = 'failed', error = $2, updated_at = $3 WHERE id = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, id, errorMsg, time.Now())
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"dlq_id": id,
			"error":  errorMsg,
		}).Error("Failed to mark DLQ message as failed")
		return fmt.Errorf("failed to mark DLQ message as failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("DLQ message not found: %s", id)
	}

	r.logger.WithFields(logrus.Fields{
		"dlq_id": id,
		"error":  errorMsg,
	}).Debug("DLQ message marked as failed successfully")
	return nil
}

// DeleteProcessed deletes processed DLQ messages older than specified duration
func (r *PostgresDLQRepository) DeleteProcessed(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM dlq_messages WHERE status = 'processed' AND updated_at < $1`

	cutoffTime := time.Now().Add(-olderThan)
	result, err := r.client.GetDB().ExecContext(ctx, query, cutoffTime)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete processed DLQ messages")
		return fmt.Errorf("failed to delete processed DLQ messages: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"older_than":    olderThan,
		"rows_affected": rowsAffected,
	}).Debug("Processed DLQ messages deleted successfully")
	return nil
}

// DeleteByTopic deletes all DLQ messages for a topic
func (r *PostgresDLQRepository) DeleteByTopic(ctx context.Context, topic string) error {
	query := `DELETE FROM dlq_messages WHERE topic = $1`

	result, err := r.client.GetDB().ExecContext(ctx, query, topic)
	if err != nil {
		r.logger.WithError(err).WithField("topic", topic).Error("Failed to delete DLQ messages by topic")
		return fmt.Errorf("failed to delete DLQ messages by topic: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"topic":         topic,
		"rows_affected": rowsAffected,
	}).Debug("DLQ messages deleted by topic successfully")
	return nil
}

// Ping checks database connectivity
func (r *PostgresDLQRepository) Ping(ctx context.Context) error {
	return r.client.Ping()
}
