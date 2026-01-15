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
	"github.com/sirupsen/logrus"
)

// MetricsRepository handles metrics operations in PostgreSQL
type MetricsRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewMetricsRepository creates a new metrics repository
func NewMetricsRepository(db *sql.DB, logger *logrus.Logger) *MetricsRepository {
	return &MetricsRepository{
		db:     db,
		logger: logger,
	}
}

// StoreMetric stores a server metric
func (r *MetricsRepository) StoreMetric(ctx context.Context, serverID string, metric *models.ServerMetrics) error {
	query := `
		INSERT INTO metrics (server_id, cpu, memory, disk, network, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (server_id, timestamp) DO UPDATE SET
			cpu = EXCLUDED.cpu,
			memory = EXCLUDED.memory,
			disk = EXCLUDED.disk,
			network = EXCLUDED.network
	`

	_, err := r.db.ExecContext(ctx, query,
		serverID,
		metric.CPU,
		metric.Memory,
		metric.Disk,
		metric.Network,
		metric.Time,
	)

	if err != nil {
		return fmt.Errorf("failed to store metric: %w", err)
	}

	return nil
}

// GetLatestMetrics retrieves latest metrics for a server
func (r *MetricsRepository) GetLatestMetrics(ctx context.Context, serverID string) (*models.ServerMetrics, error) {
	query := `
		SELECT cpu, memory, disk, network, timestamp
		FROM metrics
		WHERE server_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var metric models.ServerMetrics
	err := r.db.QueryRowContext(ctx, query, serverID).Scan(
		&metric.CPU,
		&metric.Memory,
		&metric.Disk,
		&metric.Network,
		&metric.Time,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no metrics found for server %s", serverID)
		}
		return nil, fmt.Errorf("failed to get latest metrics: %w", err)
	}

	return &metric, nil
}

// GetMetricsHistory retrieves metrics history for a server
func (r *MetricsRepository) GetMetricsHistory(ctx context.Context, serverID string, limit int) ([]*models.ServerMetrics, error) {
	query := `
		SELECT cpu, memory, disk, network, timestamp
		FROM metrics
		WHERE server_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, serverID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics history: %w", err)
	}
	defer rows.Close()

	var metrics []*models.ServerMetrics
	for rows.Next() {
		var metric models.ServerMetrics
		err := rows.Scan(
			&metric.CPU,
			&metric.Memory,
			&metric.Disk,
			&metric.Network,
			&metric.Time,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, &metric)
	}

	return metrics, nil
}

// DeleteOldMetrics deletes metrics older than specified duration
func (r *MetricsRepository) DeleteOldMetrics(ctx context.Context, olderThan time.Duration) error {
	query := `
		DELETE FROM metrics
		WHERE timestamp < $1
	`

	_, err := r.db.ExecContext(ctx, query, time.Now().Add(-olderThan))
	if err != nil {
		return fmt.Errorf("failed to delete old metrics: %w", err)
	}

	return nil
}
