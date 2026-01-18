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
	"fmt"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

// SetServerStatus stores server status in TimescaleDB
func (c *Client) SetServerStatus(ctx context.Context, serverID string, status *models.ServerStatus) error {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	INSERT INTO server_status (
		time, server_id, online, last_seen, version, os_info,
		agent_version, hostname, response_time_ms
	) VALUES (
		NOW(), $1, $2, $3, $4, $5, $6, $7, $8
	)`

	var lastSeen sql.NullTime
	if !status.LastSeen.IsZero() {
		lastSeen = sql.NullTime{Time: status.LastSeen, Valid: true}
	}

	_, err := c.pool.Exec(ctx, query,
		serverID,
		status.Online,
		lastSeen,
		status.Version,
		status.OSInfo,
		status.AgentVersion,
		status.Hostname,
		0, // response_time_ms - will be calculated
	)

	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"server_id": serverID,
			"online":    status.Online,
		}).Error("Failed to store server status in TimescaleDB")
		return fmt.Errorf("failed to store server status: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"online":    status.Online,
	}).Debug("Server status stored in TimescaleDB")

	return nil
}

// GetServerStatus retrieves the latest status for a server
func (c *Client) GetServerStatus(ctx context.Context, serverID string) (*models.ServerStatus, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT 
		online, last_seen, version, os_info, agent_version, hostname
	FROM server_status 
	WHERE server_id = $1 
	ORDER BY time DESC 
	LIMIT 1`

	var status models.ServerStatus
	var lastSeen sql.NullTime

	err := c.pool.QueryRow(ctx, query, serverID).Scan(
		&status.Online,
		&lastSeen,
		&status.Version,
		&status.OSInfo,
		&status.AgentVersion,
		&status.Hostname,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Return default offline status if no records found
			return &models.ServerStatus{
				Online:   false,
				LastSeen: time.Time{},
			}, nil
		}
		return nil, fmt.Errorf("failed to retrieve server status: %w", err)
	}

	if lastSeen.Valid {
		status.LastSeen = lastSeen.Time
	} else {
		status.LastSeen = time.Time{}
	}

	c.logger.WithField("server_id", serverID).Debug("Retrieved server status from TimescaleDB")
	return &status, nil
}

// GetAllServersStatus retrieves latest status for all servers
func (c *Client) GetAllServersStatus(ctx context.Context) ([]map[string]interface{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT DISTINCT ON (server_id)
		server_id,
		online, last_seen, version, os_info, agent_version, hostname
	FROM server_status 
	ORDER BY server_id, time DESC`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all servers status: %w", err)
	}
	defer rows.Close()

	var servers []map[string]interface{}

	for rows.Next() {
		var serverID, version, osInfo, agentVersion, hostname string
		var online bool
		var lastSeen sql.NullTime

		if err := rows.Scan(
			&serverID,
			&online,
			&lastSeen,
			&version,
			&osInfo,
			&agentVersion,
			&hostname,
		); err != nil {
			return nil, fmt.Errorf("failed to scan server status row: %w", err)
		}

		status := map[string]interface{}{
			"server_id":     serverID,
			"online":        online,
			"last_seen":     nil,
			"version":       version,
			"os_info":       osInfo,
			"agent_version": agentVersion,
			"hostname":      hostname,
		}

		if lastSeen.Valid {
			status["last_seen"] = lastSeen.Time
		}

		servers = append(servers, status)
	}

	return servers, nil
}

// GetActiveServers retrieves servers that are currently online
func (c *Client) GetActiveServers(ctx context.Context) ([]map[string]interface{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Use the active_servers view we created in the schema
	query := `
	SELECT server_id, hostname, os_info, agent_version, last_seen, online
	FROM active_servers
	WHERE online = TRUE`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active servers: %w", err)
	}
	defer rows.Close()

	var servers []map[string]interface{}

	for rows.Next() {
		var serverID, hostname, osInfo, agentVersion string
		var online bool
		var lastSeen sql.NullTime

		if err := rows.Scan(
			&serverID,
			&hostname,
			&osInfo,
			&agentVersion,
			&lastSeen,
			&online,
		); err != nil {
			return nil, fmt.Errorf("failed to scan active server row: %w", err)
		}

		server := map[string]interface{}{
			"server_id":     serverID,
			"hostname":      hostname,
			"os_info":       osInfo,
			"agent_version": agentVersion,
			"online":        online,
			"last_seen":     nil,
		}

		if lastSeen.Valid {
			server["last_seen"] = lastSeen.Time
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// GetServerUptime calculates uptime percentage for a server in the last 24 hours
func (c *Client) GetServerUptime(ctx context.Context, serverID string) (float64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Use the function we created in the schema
	query := `SELECT get_uptime_last_24h($1)`

	var uptime float64
	err := c.pool.QueryRow(ctx, query, serverID).Scan(&uptime)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get server uptime: %w", err)
	}

	return uptime, nil
}

// GetStatusHistory retrieves status history for a server
func (c *Client) GetStatusHistory(ctx context.Context, serverID string, start, end time.Time) ([]*models.ServerStatus, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT 
		online, last_seen, version, os_info, agent_version, hostname, time
	FROM server_status 
	WHERE server_id = $1 AND time BETWEEN $2 AND $3
	ORDER BY time DESC`

	rows, err := c.pool.Query(ctx, query, serverID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query status history: %w", err)
	}
	defer rows.Close()

	var statuses []*models.ServerStatus

	for rows.Next() {
		var status models.ServerStatus
		var lastSeen sql.NullTime

		if err := rows.Scan(
			&status.Online,
			&lastSeen,
			&status.Version,
			&status.OSInfo,
			&status.AgentVersion,
			&status.Hostname,
			&status.LastSeen, // Use time as last_seen for history
		); err != nil {
			return nil, fmt.Errorf("failed to scan status history row: %w", err)
		}

		if lastSeen.Valid {
			status.LastSeen = lastSeen.Time
		}

		statuses = append(statuses, &status)
	}

	return statuses, nil
}

// UpdateServerHeartbeat updates server heartbeat (convenience method)
func (c *Client) UpdateServerHeartbeat(ctx context.Context, serverID string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	status := &models.ServerStatus{
		Online:   true,
		LastSeen: time.Now(),
	}

	return c.SetServerStatus(ctx, serverID, status)
}

// GetOfflineServers retrieves servers that haven't been seen recently
func (c *Client) GetOfflineServers(ctx context.Context, offlineThreshold time.Duration) ([]map[string]interface{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT DISTINCT ON (server_id)
		server_id, hostname, os_info, agent_version, last_seen
	FROM server_status 
	WHERE last_seen < NOW() - INTERVAL $1 OR last_seen IS NULL
	ORDER BY server_id, time DESC`

	rows, err := c.pool.Query(ctx, query, offlineThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to query offline servers: %w", err)
	}
	defer rows.Close()

	var servers []map[string]interface{}

	for rows.Next() {
		var serverID, hostname, osInfo, agentVersion string
		var lastSeen sql.NullTime

		if err := rows.Scan(
			&serverID,
			&hostname,
			&osInfo,
			&agentVersion,
			&lastSeen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan offline server row: %w", err)
		}

		server := map[string]interface{}{
			"server_id":     serverID,
			"hostname":      hostname,
			"os_info":       osInfo,
			"agent_version": agentVersion,
			"last_seen":     nil,
		}

		if lastSeen.Valid {
			server["last_seen"] = lastSeen.Time
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// DeleteOldStatus removes status records older than the specified duration
func (c *Client) DeleteOldStatus(ctx context.Context, olderThan time.Duration) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `DELETE FROM server_status WHERE time < NOW() - INTERVAL $1`

	result, err := c.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old status records: %w", err)
	}

	deletedCount := result.RowsAffected()
	c.logger.WithField("deleted_count", deletedCount).Info("Old status records deleted")

	return deletedCount, nil
}

// GetStatusStats returns statistics about status storage
func (c *Client) GetStatusStats(ctx context.Context) (*StatusStats, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT 
		COUNT(*) as total_records,
		COUNT(DISTINCT server_id) as unique_servers,
		COUNT(CASE WHEN online THEN 1 END) as online_servers,
		COUNT(CASE WHEN NOT online THEN 1 END) as offline_servers,
		MIN(time) as earliest_record,
		MAX(time) as latest_record,
		pg_size_pretty(pg_total_relation_size('server_status')) as table_size
	FROM server_status
	WHERE time = (
		SELECT MAX(time) 
		FROM server_status s2 
		WHERE s2.server_id = server_status.server_id
	)`

	var stats StatusStats
	var earliestRecord, latestRecord sql.NullTime
	var tableSize string

	err := c.pool.QueryRow(ctx, query).Scan(
		&stats.TotalRecords,
		&stats.UniqueServers,
		&stats.OnlineServers,
		&stats.OfflineServers,
		&earliestRecord,
		&latestRecord,
		&tableSize,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get status stats: %w", err)
	}

	stats.TableSize = tableSize

	if earliestRecord.Valid {
		stats.EarliestRecord = earliestRecord.Time
	}
	if latestRecord.Valid {
		stats.LatestRecord = latestRecord.Time
	}

	return &stats, nil
}

// StatusStats represents status storage statistics
type StatusStats struct {
	TotalRecords   int64     `json:"total_records"`
	UniqueServers  int64     `json:"unique_servers"`
	OnlineServers  int64     `json:"online_servers"`
	OfflineServers int64     `json:"offline_servers"`
	EarliestRecord time.Time `json:"earliest_record"`
	LatestRecord   time.Time `json:"latest_record"`
	TableSize      string    `json:"table_size"`
}
