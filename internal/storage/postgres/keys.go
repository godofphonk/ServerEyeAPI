package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/sirupsen/logrus"
)

// InsertGeneratedKey inserts a new generated key
func (c *Client) InsertGeneratedKey(ctx context.Context, secretKey, agentVersion, operatingSystem, hostname string) error {
	query := `
		INSERT INTO generated_keys (secret_key, agent_version, os_info, hostname, status)
		VALUES ($1, $2, $3, $4, 'generated')
		ON CONFLICT (secret_key) DO NOTHING
	`

	_, err := c.db.ExecContext(ctx, query, secretKey, agentVersion, operatingSystem, hostname)
	if err != nil {
		return fmt.Errorf("failed to insert generated key: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"secret_key":       secretKey,
		"agent_version":    agentVersion,
		"operating_system": operatingSystem,
		"hostname":         hostname,
	}).Info("Generated key inserted successfully")

	return nil
}

// InsertGeneratedKeyWithIDs inserts a new generated key with server_id and server_key
func (c *Client) InsertGeneratedKeyWithIDs(ctx context.Context, secretKey, serverID, serverKey, agentVersion, operatingSystem, hostname string) error {
	query := `
		INSERT INTO generated_keys (secret_key, server_id, server_key, agent_version, os_info, hostname, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'generated')
		ON CONFLICT (secret_key) DO UPDATE SET
			server_id = EXCLUDED.server_id,
			server_key = EXCLUDED.server_key,
			agent_version = EXCLUDED.agent_version,
			os_info = EXCLUDED.os_info,
			hostname = EXCLUDED.hostname,
			status = EXCLUDED.status
	`

	_, err := c.db.ExecContext(ctx, query, secretKey, serverID, serverKey, agentVersion, operatingSystem, hostname)
	if err != nil {
		return fmt.Errorf("failed to insert generated key with IDs: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"secret_key":       secretKey,
		"server_id":        serverID,
		"server_key":       serverKey,
		"agent_version":    agentVersion,
		"operating_system": operatingSystem,
		"hostname":         hostname,
	}).Info("Generated key with IDs inserted successfully")

	return nil
}

// GetServerByKey retrieves server information by server key
func (c *Client) GetServerByKey(ctx context.Context, serverKey string) (*models.ServerInfo, error) {
	query := `
		SELECT server_id, secret_key, hostname
		FROM generated_keys 
		WHERE server_key = $1 AND status = 'generated'
	`

	var info models.ServerInfo
	err := c.db.QueryRowContext(ctx, query, serverKey).Scan(
		&info.ServerID,
		&info.SecretKey,
		&info.Hostname,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("server key not found")
		}
		return nil, fmt.Errorf("failed to get server by key: %w", err)
	}

	return &info, nil
}

// GetServers retrieves all server IDs
func (c *Client) GetServers(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT server_id FROM servers WHERE status != 'deleted'`

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query servers: %w", err)
	}
	defer rows.Close()

	var servers []string
	for rows.Next() {
		var serverID string
		if err := rows.Scan(&serverID); err != nil {
			return nil, fmt.Errorf("failed to scan server ID: %w", err)
		}
		servers = append(servers, serverID)
	}

	return servers, nil
}

// GetServerMetrics retrieves metrics for a server from PostgreSQL
func (c *Client) GetServerMetrics(ctx context.Context, serverID string) (map[string]interface{}, error) {
	query := `
		SELECT hostname, os_info, agent_version, status, last_seen
		FROM servers
		WHERE server_id = $1
	`

	var hostname, osInfo, agentVersion, status string
	var lastSeen interface{}

	err := c.db.QueryRowContext(ctx, query, serverID).Scan(
		&hostname, &osInfo, &agentVersion, &status, &lastSeen,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get server metrics: %w", err)
	}

	return map[string]interface{}{
		"hostname":      hostname,
		"os_info":       osInfo,
		"agent_version": agentVersion,
		"status":        status,
		"last_seen":     lastSeen,
	}, nil
}

// GetPendingCommands retrieves pending commands for a server
func (c *Client) GetPendingCommands(ctx context.Context, serverID string) ([]string, error) {
	// This would typically query a commands table
	// For now, return empty as commands are handled in Redis
	return []string{}, nil
}

// StoreDLQMessage stores a message in the dead letter queue
func (c *Client) StoreDLQMessage(ctx context.Context, topic string, partition int, offset int64, message []byte, errorMsg string) error {
	query := `
		INSERT INTO dead_letter_queue (topic, partition, message_offset, message, error, attempts)
		VALUES ($1, $2, $3, $4, $5, 1)
	`

	_, err := c.db.ExecContext(ctx, query, topic, partition, offset, message, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to store DLQ message: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"topic": topic,
		"error": errorMsg,
	}).Warn("Message stored in dead letter queue")

	return nil
}
