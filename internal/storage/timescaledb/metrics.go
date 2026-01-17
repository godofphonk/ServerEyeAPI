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

// StoreMetric stores server metrics in TimescaleDB
func (c *Client) StoreMetric(ctx context.Context, serverID string, metrics *models.ServerMetrics) error {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	INSERT INTO server_metrics (
		time, server_id, cpu_usage, memory_usage, disk_usage, network_usage,
		cpu_usage_total, cpu_cores, memory_total_gb, memory_used_gb,
		memory_available_gb, memory_free_gb, memory_buffers_gb, memory_cached_gb,
		cpu_temperature, highest_temperature, hostname, os_info
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
	)`

	_, err := c.pool.Exec(ctx, query,
		metrics.Time,
		serverID,
		metrics.CPU,
		metrics.Memory,
		metrics.Disk,
		metrics.Network,
		metrics.CPUUsage.UsageTotal,
		metrics.CPUUsage.Cores,
		metrics.MemoryDetails.TotalGB,
		metrics.MemoryDetails.UsedGB,
		metrics.MemoryDetails.AvailableGB,
		metrics.MemoryDetails.FreeGB,
		metrics.MemoryDetails.BuffersGB,
		metrics.MemoryDetails.CachedGB,
		metrics.TemperatureDetails.CPUTemperature,
		metrics.TemperatureDetails.HighestTemperature,
		metrics.SystemDetails.Hostname,
		metrics.SystemDetails.OS,
	)

	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"server_id": serverID,
			"cpu":       metrics.CPU,
			"memory":    metrics.Memory,
		}).Error("Failed to store metrics in TimescaleDB")
		return fmt.Errorf("failed to store metrics: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"server_id": serverID,
		"cpu":       metrics.CPU,
		"memory":    metrics.Memory,
		"disk":      metrics.Disk,
	}).Debug("Metrics stored in TimescaleDB")

	return nil
}

// GetLatestMetric retrieves the most recent metrics for a server
func (c *Client) GetLatestMetric(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT 
		cpu_usage, memory_usage, disk_usage, network_usage,
		cpu_usage_total, cpu_cores, memory_total_gb, memory_used_gb,
		memory_available_gb, memory_free_gb, memory_buffers_gb, memory_cached_gb,
		cpu_temperature, highest_temperature, hostname, os_info, time
	FROM server_metrics 
	WHERE server_id = $1 
	ORDER BY time DESC 
	LIMIT 1`

	var metrics models.ServerMetrics

	var availableGB, freeGB, buffersGB, cachedGB sql.NullFloat64

	err := c.pool.QueryRow(ctx, query, serverID).Scan(
		&metrics.CPU,
		&metrics.Memory,
		&metrics.Disk,
		&metrics.Network,
		&metrics.CPUUsage.UsageTotal,
		&metrics.CPUUsage.Cores,
		&metrics.MemoryDetails.TotalGB,
		&metrics.MemoryDetails.UsedGB,
		&availableGB,
		&freeGB,
		&buffersGB,
		&cachedGB,
		&metrics.TemperatureDetails.CPUTemperature,
		&metrics.TemperatureDetails.HighestTemperature,
		&metrics.SystemDetails.Hostname,
		&metrics.SystemDetails.OS,
		&metrics.Time,
	)

	// Convert NullFloat64 to regular float64
	if availableGB.Valid {
		metrics.MemoryDetails.AvailableGB = availableGB.Float64
	}
	if freeGB.Valid {
		metrics.MemoryDetails.FreeGB = freeGB.Float64
	}
	if buffersGB.Valid {
		metrics.MemoryDetails.BuffersGB = buffersGB.Float64
	}
	if cachedGB.Valid {
		metrics.MemoryDetails.CachedGB = cachedGB.Float64
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no metrics found for server %s", serverID)
		}
		return nil, fmt.Errorf("failed to retrieve metrics: %w", err)
	}

	c.logger.WithField("server_id", serverID).Debug("Retrieved latest metrics from TimescaleDB")
	return &metrics, nil
}

// GetMetricsHistory retrieves historical metrics for a server with aggregation
func (c *Client) GetMetricsHistory(ctx context.Context, serverID string, start, end time.Time, interval time.Duration) ([]*models.ServerMetrics, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT 
		time_bucket($1, time) as bucket,
		AVG(cpu_usage) as cpu_usage,
		MAX(cpu_usage) as max_cpu,
		AVG(memory_usage) as memory_usage,
		MAX(memory_usage) as max_memory,
		AVG(disk_usage) as disk_usage,
		MAX(disk_usage) as max_disk,
		AVG(network_usage) as network_usage,
		MAX(network_usage) as max_network,
		AVG(highest_temperature) as avg_temperature,
		MAX(highest_temperature) as max_temperature,
		COUNT(*) as sample_count
	FROM server_metrics 
	WHERE server_id = $2 AND time BETWEEN $3 AND $4
	GROUP BY bucket
	ORDER BY bucket`

	rows, err := c.pool.Query(ctx, query, interval, serverID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics history: %w", err)
	}
	defer rows.Close()

	var metrics []*models.ServerMetrics
	for rows.Next() {
		var metric models.ServerMetrics
		var sampleCount int64
		var maxCPU, maxMemory, maxDisk, maxNetwork, maxTemp sql.NullFloat64

		if err := rows.Scan(
			&metric.Time,
			&metric.CPU,
			&maxCPU,
			&metric.Memory,
			&maxMemory,
			&metric.Disk,
			&maxDisk,
			&metric.Network,
			&maxNetwork,
			&metric.TemperatureDetails.HighestTemperature,
			&maxTemp,
			&sampleCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan metrics row: %w", err)
		}

		// Set max values if available
		if maxCPU.Valid {
			metric.CPUUsage.UsageTotal = maxCPU.Float64
		}
		if maxMemory.Valid {
			metric.MemoryDetails.UsedPercent = maxMemory.Float64
		}
		if maxDisk.Valid {
			// Use max_disk as additional info
		}
		if maxNetwork.Valid {
			// Use max_network as additional info
		}
		if maxTemp.Valid {
			metric.TemperatureDetails.HighestTemperature = maxTemp.Float64
		}

		metrics = append(metrics, &metric)
	}

	return metrics, nil
}

// GetAggregatedMetrics retrieves pre-aggregated metrics from continuous aggregates
func (c *Client) GetAggregatedMetrics(ctx context.Context, serverID string, period string) (*AggregatedMetrics, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var viewName string
	var timeFilter string

	switch period {
	case "1h":
		viewName = "metrics_1h_avg"
		timeFilter = "hour >= NOW() - INTERVAL '24 hours'"
	case "5m":
		viewName = "metrics_5m_avg"
		timeFilter = "five_min >= NOW() - INTERVAL '7 days'"
	default:
		return nil, fmt.Errorf("unsupported period: %s", period)
	}

	query := fmt.Sprintf(`
	SELECT 
		AVG(avg_cpu) as overall_avg_cpu,
		MAX(max_cpu) as overall_max_cpu,
		AVG(avg_memory) as overall_avg_memory,
		MAX(max_memory) as overall_max_memory,
		AVG(avg_disk) as overall_avg_disk,
		MAX(max_disk) as overall_max_disk,
		AVG(avg_network) as overall_avg_network,
		MAX(max_network) as overall_max_network,
		AVG(avg_temperature) as overall_avg_temperature,
		MAX(max_temperature) as overall_max_temperature,
		COUNT(*) as data_points
	FROM %s
	WHERE server_id = $1 AND %s`, viewName, timeFilter)

	var aggregated AggregatedMetrics
	err := c.pool.QueryRow(ctx, query, serverID).Scan(
		&aggregated.AvgCPU,
		&aggregated.MaxCPU,
		&aggregated.AvgMemory,
		&aggregated.MaxMemory,
		&aggregated.AvgDisk,
		&aggregated.MaxDisk,
		&aggregated.AvgNetwork,
		&aggregated.MaxNetwork,
		&aggregated.AvgTemperature,
		&aggregated.MaxTemperature,
		&aggregated.DataPoints,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return &AggregatedMetrics{}, nil
		}
		return nil, fmt.Errorf("failed to retrieve aggregated metrics: %w", err)
	}

	return &aggregated, nil
}

// GetAllServersMetrics retrieves latest metrics for all servers
func (c *Client) GetAllServersMetrics(ctx context.Context) (map[string]*models.ServerMetrics, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT DISTINCT ON (server_id)
		server_id,
		cpu_usage, memory_usage, disk_usage, network_usage,
		cpu_temperature, highest_temperature,
		hostname, os_info, time
	FROM server_metrics 
	ORDER BY server_id, time DESC`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all servers metrics: %w", err)
	}
	defer rows.Close()

	metrics := make(map[string]*models.ServerMetrics)

	for rows.Next() {
		var serverID string
		var metric models.ServerMetrics

		if err := rows.Scan(
			&serverID,
			&metric.CPU,
			&metric.Memory,
			&metric.Disk,
			&metric.Network,
			&metric.TemperatureDetails.CPUTemperature,
			&metric.TemperatureDetails.HighestTemperature,
			&metric.SystemDetails.Hostname,
			&metric.SystemDetails.OS,
			&metric.Time,
		); err != nil {
			return nil, fmt.Errorf("failed to scan server metrics row: %w", err)
		}

		metrics[serverID] = &metric
	}

	return metrics, nil
}

// GetMetricsByTimeRange retrieves metrics within a time range for multiple servers
func (c *Client) GetMetricsByTimeRange(ctx context.Context, serverIDs []string, start, end time.Time) (map[string][]*models.ServerMetrics, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(serverIDs) == 0 {
		return make(map[string][]*models.ServerMetrics), nil
	}

	// Build IN clause
	placeholders := make([]string, len(serverIDs))
	args := make([]interface{}, len(serverIDs)+2)
	for i, id := range serverIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args[i+2] = id
	}
	args[0] = start
	args[1] = end

	query := fmt.Sprintf(`
	SELECT 
		server_id,
		cpu_usage, memory_usage, disk_usage, network_usage,
		cpu_temperature, highest_temperature,
		hostname, os_info, time
	FROM server_metrics 
	WHERE server_id IN (%s) AND time BETWEEN $1 AND $2
	ORDER BY server_id, time DESC`,
		fmt.Sprintf("%s", placeholders),
	)

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics by time range: %w", err)
	}
	defer rows.Close()

	metrics := make(map[string][]*models.ServerMetrics)

	for rows.Next() {
		var serverID string
		var metric models.ServerMetrics

		if err := rows.Scan(
			&serverID,
			&metric.CPU,
			&metric.Memory,
			&metric.Disk,
			&metric.Network,
			&metric.TemperatureDetails.CPUTemperature,
			&metric.TemperatureDetails.HighestTemperature,
			&metric.SystemDetails.Hostname,
			&metric.SystemDetails.OS,
			&metric.Time,
		); err != nil {
			return nil, fmt.Errorf("failed to scan time range metrics row: %w", err)
		}

		metrics[serverID] = append(metrics[serverID], &metric)
	}

	return metrics, nil
}

// DeleteOldMetrics removes metrics older than the specified duration
func (c *Client) DeleteOldMetrics(ctx context.Context, olderThan time.Duration) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `DELETE FROM server_metrics WHERE time < NOW() - INTERVAL $1`

	result, err := c.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old metrics: %w", err)
	}

	deletedCount := result.RowsAffected()
	c.logger.WithField("deleted_count", deletedCount).Info("Old metrics deleted")

	return deletedCount, nil
}

// GetMetricsStats returns statistics about metrics storage
func (c *Client) GetMetricsStats(ctx context.Context) (*MetricsStats, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
	SELECT 
		COUNT(*) as total_records,
		COUNT(DISTINCT server_id) as unique_servers,
		MIN(time) as earliest_record,
		MAX(time) as latest_record,
		pg_size_pretty(pg_total_relation_size('server_metrics')) as table_size
	FROM server_metrics`

	var stats MetricsStats
	var earliestRecord, latestRecord sql.NullTime
	var tableSize string

	err := c.pool.QueryRow(ctx, query).Scan(
		&stats.TotalRecords,
		&stats.UniqueServers,
		&earliestRecord,
		&latestRecord,
		&tableSize,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get metrics stats: %w", err)
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

// AggregatedMetrics represents aggregated metric data
type AggregatedMetrics struct {
	AvgCPU         float64 `json:"avg_cpu"`
	MaxCPU         float64 `json:"max_cpu"`
	AvgMemory      float64 `json:"avg_memory"`
	MaxMemory      float64 `json:"max_memory"`
	AvgDisk        float64 `json:"avg_disk"`
	MaxDisk        float64 `json:"max_disk"`
	AvgNetwork     float64 `json:"avg_network"`
	MaxNetwork     float64 `json:"max_network"`
	AvgTemperature float64 `json:"avg_temperature"`
	MaxTemperature float64 `json:"max_temperature"`
	DataPoints     int64   `json:"data_points"`
}

// MetricsStats represents metrics storage statistics
type MetricsStats struct {
	TotalRecords   int64     `json:"total_records"`
	UniqueServers  int64     `json:"unique_servers"`
	EarliestRecord time.Time `json:"earliest_record"`
	LatestRecord   time.Time `json:"latest_record"`
	TableSize      string    `json:"table_size"`
}
