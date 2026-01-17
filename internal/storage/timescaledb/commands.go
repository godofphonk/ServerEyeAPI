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

package timescaledb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// StoreCommand stores a command for a server in TimescaleDB
func (c *Client) StoreCommand(ctx context.Context, serverID string, command map[string]interface{}) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Extract command details
	commandType, _ := command["type"].(string)
	if commandType == "" {
		commandType = "unknown"
	}

	// Convert command data to JSON
	commandDataJSON, err := json.Marshal(command)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal command data")
		return fmt.Errorf("failed to marshal command data: %w", err)
	}

	// Set expiration time (default 1 hour)
	expiresAt := time.Now().Add(time.Hour)

	query := `
	INSERT INTO server_commands (
		server_id, command_type, command_data, status, expires_at
	) VALUES (
		$1, $2, $3, 'pending', $4
	)`

	_, err = c.pool.Exec(ctx, query, serverID, commandType, commandDataJSON, expiresAt)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"server_id": serverID,
			"type":      commandType,
		}).Error("Failed to store command in TimescaleDB")
		return fmt.Errorf("failed to store command: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"type":      commandType,
	}).Debug("Command stored in TimescaleDB")

	return nil
}

// GetPendingCommands retrieves pending commands for a server
func (c *Client) GetPendingCommands(ctx context.Context, serverID string) ([]map[string]interface{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT 
		command_id, command_type, command_data, status, created_at, expires_at
	FROM server_commands 
	WHERE server_id = $1 AND status = 'pending' AND expires_at > NOW()
	ORDER BY created_at ASC`

	rows, err := c.pool.Query(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending commands: %w", err)
	}
	defer rows.Close()

	var commands []map[string]interface{}

	for rows.Next() {
		var commandID uuid.UUID
		var commandType, status string
		var commandDataJSON []byte
		var createdAt, expiresAt time.Time

		if err := rows.Scan(
			&commandID,
			&commandType,
			&commandDataJSON,
			&status,
			&createdAt,
			&expiresAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan command row: %w", err)
		}

		// Parse command data
		var commandData map[string]interface{}
		if len(commandDataJSON) > 0 {
			if err := json.Unmarshal(commandDataJSON, &commandData); err != nil {
				c.logger.WithError(err).Warn("Failed to unmarshal command data")
				commandData = make(map[string]interface{})
			}
		}

		command := map[string]interface{}{
			"command_id": commandID.String(),
			"type":       commandType,
			"data":       commandData,
			"status":     status,
			"created_at": createdAt,
			"expires_at": expiresAt,
		}

		commands = append(commands, command)
	}

	return commands, nil
}

// GetCommands retrieves command history for a server
func (c *Client) GetCommands(ctx context.Context, serverID string, limit int) ([]map[string]interface{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if limit <= 0 {
		limit = 50 // default limit
	}

	query := `
	SELECT 
		command_id, command_type, command_data, status, response, error_message,
		created_at, sent_at, executed_at, expires_at, retry_count
	FROM server_commands 
	WHERE server_id = $1
	ORDER BY created_at DESC
	LIMIT $2`

	rows, err := c.pool.Query(ctx, query, serverID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query commands: %w", err)
	}
	defer rows.Close()

	var commands []map[string]interface{}

	for rows.Next() {
		var commandID uuid.UUID
		var commandType, status, errorMessage string
		var commandDataJSON, responseJSON []byte
		var createdAt, sentAt, executedAt, expiresAt sql.NullTime
		var retryCount int

		if err := rows.Scan(
			&commandID,
			&commandType,
			&commandDataJSON,
			&status,
			&responseJSON,
			&errorMessage,
			&createdAt,
			&sentAt,
			&executedAt,
			&expiresAt,
			&retryCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan command history row: %w", err)
		}

		// Parse command data
		var commandData map[string]interface{}
		if len(commandDataJSON) > 0 {
			if err := json.Unmarshal(commandDataJSON, &commandData); err != nil {
				c.logger.WithError(err).Warn("Failed to unmarshal command data")
				commandData = make(map[string]interface{})
			}
		}

		// Parse response if available
		var response map[string]interface{}
		if len(responseJSON) > 0 {
			if err := json.Unmarshal(responseJSON, &response); err != nil {
				c.logger.WithError(err).Warn("Failed to unmarshal command response")
				response = make(map[string]interface{})
			}
		}

		command := map[string]interface{}{
			"command_id":    commandID.String(),
			"type":          commandType,
			"data":          commandData,
			"status":        status,
			"response":      response,
			"error_message": errorMessage,
			"retry_count":   retryCount,
			"created_at":    createdAt.Time,
			"sent_at":       nil,
			"executed_at":   nil,
			"expires_at":    nil,
		}

		if sentAt.Valid {
			command["sent_at"] = sentAt.Time
		}
		if executedAt.Valid {
			command["executed_at"] = executedAt.Time
		}
		if expiresAt.Valid {
			command["expires_at"] = expiresAt.Time
		}

		commands = append(commands, command)
	}

	return commands, nil
}

// UpdateCommandStatus updates the status of a command
func (c *Client) UpdateCommandStatus(ctx context.Context, commandID string, status string, response map[string]interface{}, errorMessage string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Parse command ID
	uuid, err := uuid.Parse(commandID)
	if err != nil {
		return fmt.Errorf("invalid command ID: %w", err)
	}

	// Convert response to JSON
	var responseJSON []byte
	if response != nil {
		responseJSON, err = json.Marshal(response)
		if err != nil {
			c.logger.WithError(err).Warn("Failed to marshal command response")
			responseJSON = []byte("{}")
		}
	} else {
		responseJSON = []byte("{}")
	}

	query := `
	UPDATE server_commands 
	SET status = $1, response = $2, error_message = $3, executed_at = NOW()
	WHERE command_id = $4`

	_, err = c.pool.Exec(ctx, query, status, responseJSON, errorMessage, uuid)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"command_id": commandID,
			"status":     status,
		}).Error("Failed to update command status")
		return fmt.Errorf("failed to update command status: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"command_id": commandID,
		"status":     status,
	}).Debug("Command status updated in TimescaleDB")

	return nil
}

// MarkCommandAsSent marks a command as sent to the server
func (c *Client) MarkCommandAsSent(ctx context.Context, commandID string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	uuid, err := uuid.Parse(commandID)
	if err != nil {
		return fmt.Errorf("invalid command ID: %w", err)
	}

	query := `
	UPDATE server_commands 
	SET status = 'sent', sent_at = NOW()
	WHERE command_id = $1 AND status = 'pending'`

	_, err = c.pool.Exec(ctx, query, uuid)
	if err != nil {
		c.logger.WithError(err).WithField("command_id", commandID).Error("Failed to mark command as sent")
		return fmt.Errorf("failed to mark command as sent: %w", err)
	}

	c.logger.WithField("command_id", commandID).Debug("Command marked as sent")
	return nil
}

// ExpirePendingCommands marks expired pending commands as failed
func (c *Client) ExpirePendingCommands(ctx context.Context) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	UPDATE server_commands 
	SET status = 'failed', error_message = 'Command expired', executed_at = NOW()
	WHERE status = 'pending' AND expires_at < NOW()`

	result, err := c.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to expire pending commands: %w", err)
	}

	expiredCount := result.RowsAffected()
	if expiredCount > 0 {
		c.logger.WithField("expired_count", expiredCount).Info("Expired pending commands")
	}

	return expiredCount, nil
}

// DeleteOldCommands removes command records older than the specified duration
func (c *Client) DeleteOldCommands(ctx context.Context, olderThan time.Duration) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `DELETE FROM server_commands WHERE time < NOW() - INTERVAL $1`

	result, err := c.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old commands: %w", err)
	}

	deletedCount := result.RowsAffected()
	c.logger.WithField("deleted_count", deletedCount).Info("Old commands deleted")

	return deletedCount, nil
}

// GetCommandStats returns statistics about command storage
func (c *Client) GetCommandStats(ctx context.Context) (*CommandStats, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT 
		COUNT(*) as total_commands,
		COUNT(DISTINCT server_id) as unique_servers,
		COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_commands,
		COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent_commands,
		COUNT(CASE WHEN status = 'executed' THEN 1 END) as executed_commands,
		COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_commands,
		COUNT(CASE WHEN expires_at > NOW() AND status = 'pending' THEN 1 END) as active_commands,
		MIN(time) as earliest_command,
		MAX(time) as latest_command,
		pg_size_pretty(pg_total_relation_size('server_commands')) as table_size
	FROM server_commands`

	var stats CommandStats
	var earliestCommand, latestCommand sql.NullTime
	var tableSize string

	err := c.pool.QueryRow(ctx, query).Scan(
		&stats.TotalCommands,
		&stats.UniqueServers,
		&stats.PendingCommands,
		&stats.SentCommands,
		&stats.ExecutedCommands,
		&stats.FailedCommands,
		&stats.ActiveCommands,
		&earliestCommand,
		&latestCommand,
		&tableSize,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get command stats: %w", err)
	}

	stats.TableSize = tableSize

	if earliestCommand.Valid {
		stats.EarliestCommand = earliestCommand.Time
	}
	if latestCommand.Valid {
		stats.LatestCommand = latestCommand.Time
	}

	return &stats, nil
}

// CommandStats represents command storage statistics
type CommandStats struct {
	TotalCommands    int64     `json:"total_commands"`
	UniqueServers    int64     `json:"unique_servers"`
	PendingCommands  int64     `json:"pending_commands"`
	SentCommands     int64     `json:"sent_commands"`
	ExecutedCommands int64     `json:"executed_commands"`
	FailedCommands   int64     `json:"failed_commands"`
	ActiveCommands   int64     `json:"active_commands"`
	EarliestCommand  time.Time `json:"earliest_command"`
	LatestCommand    time.Time `json:"latest_command"`
	TableSize        string    `json:"table_size"`
}
