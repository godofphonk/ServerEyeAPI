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
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type AlertRepository struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

func NewAlertRepository(pool *pgxpool.Pool, logger *logrus.Logger) interfaces.AlertRepository {
	return &AlertRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *AlertRepository) Create(ctx context.Context, alert *models.Alert) error {
	query := `
		INSERT INTO alerts (
			id, type, server_id, severity, title, message, 
			device, temperature, threshold, value, status, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.pool.Exec(ctx, query,
		alert.ID,
		alert.Type,
		alert.ServerID,
		alert.Severity,
		alert.Title,
		alert.Message,
		alert.Device,
		alert.Temperature,
		alert.Threshold,
		alert.Value,
		alert.Status,
		alert.CreatedAt,
		alert.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create alert")
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

func (r *AlertRepository) GetByID(ctx context.Context, alertID string) (*models.Alert, error) {
	query := `
		SELECT id, type, server_id, severity, title, message, 
		       device, temperature, threshold, value, status, 
		       created_at, updated_at, resolved_at
		FROM alerts
		WHERE id = $1
	`

	alert := &models.Alert{}
	var resolvedAt sql.NullTime

	err := r.pool.QueryRow(ctx, query, alertID).Scan(
		&alert.ID,
		&alert.Type,
		&alert.ServerID,
		&alert.Severity,
		&alert.Title,
		&alert.Message,
		&alert.Device,
		&alert.Temperature,
		&alert.Threshold,
		&alert.Value,
		&alert.Status,
		&alert.CreatedAt,
		&alert.UpdatedAt,
		&resolvedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert not found")
		}
		r.logger.WithError(err).Error("Failed to get alert by ID")
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	if resolvedAt.Valid {
		alert.ResolvedAt = &resolvedAt.Time
	}

	return alert, nil
}

func (r *AlertRepository) GetByServerID(ctx context.Context, serverID string, limit int) ([]*models.Alert, error) {
	query := `
		SELECT id, type, server_id, severity, title, message, 
		       device, temperature, threshold, value, status, 
		       created_at, updated_at, resolved_at
		FROM alerts
		WHERE server_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, serverID, limit)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get alerts by server ID")
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&alert.ID,
			&alert.Type,
			&alert.ServerID,
			&alert.Severity,
			&alert.Title,
			&alert.Message,
			&alert.Device,
			&alert.Temperature,
			&alert.Threshold,
			&alert.Value,
			&alert.Status,
			&alert.CreatedAt,
			&alert.UpdatedAt,
			&resolvedAt,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan alert")
			continue
		}

		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *AlertRepository) GetActiveByServerID(ctx context.Context, serverID string) ([]*models.Alert, error) {
	query := `
		SELECT id, type, server_id, severity, title, message, 
		       device, temperature, threshold, value, status, 
		       created_at, updated_at, resolved_at
		FROM alerts
		WHERE server_id = $1 AND status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, serverID)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get active alerts")
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&alert.ID,
			&alert.Type,
			&alert.ServerID,
			&alert.Severity,
			&alert.Title,
			&alert.Message,
			&alert.Device,
			&alert.Temperature,
			&alert.Threshold,
			&alert.Value,
			&alert.Status,
			&alert.CreatedAt,
			&alert.UpdatedAt,
			&resolvedAt,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan alert")
			continue
		}

		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *AlertRepository) GetByServerIDAndType(ctx context.Context, serverID string, alertType models.AlertType) ([]*models.Alert, error) {
	query := `
		SELECT id, type, server_id, severity, title, message, 
		       device, temperature, threshold, value, status, 
		       created_at, updated_at, resolved_at
		FROM alerts
		WHERE server_id = $1 AND type = $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, serverID, alertType)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get alerts by type")
		return nil, fmt.Errorf("failed to get alerts by type: %w", err)
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&alert.ID,
			&alert.Type,
			&alert.ServerID,
			&alert.Severity,
			&alert.Title,
			&alert.Message,
			&alert.Device,
			&alert.Temperature,
			&alert.Threshold,
			&alert.Value,
			&alert.Status,
			&alert.CreatedAt,
			&alert.UpdatedAt,
			&resolvedAt,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan alert")
			continue
		}

		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *AlertRepository) GetByTimeRange(ctx context.Context, serverID string, start, end time.Time) ([]*models.Alert, error) {
	query := `
		SELECT id, type, server_id, severity, title, message, 
		       device, temperature, threshold, value, status, 
		       created_at, updated_at, resolved_at
		FROM alerts
		WHERE server_id = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, serverID, start, end)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get alerts by time range")
		return nil, fmt.Errorf("failed to get alerts by time range: %w", err)
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&alert.ID,
			&alert.Type,
			&alert.ServerID,
			&alert.Severity,
			&alert.Title,
			&alert.Message,
			&alert.Device,
			&alert.Temperature,
			&alert.Threshold,
			&alert.Value,
			&alert.Status,
			&alert.CreatedAt,
			&alert.UpdatedAt,
			&resolvedAt,
		)

		if err != nil {
			r.logger.WithError(err).Error("Failed to scan alert")
			continue
		}

		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *AlertRepository) Update(ctx context.Context, alert *models.Alert) error {
	query := `
		UPDATE alerts
		SET type = $2, severity = $3, title = $4, message = $5,
		    device = $6, temperature = $7, threshold = $8, value = $9,
		    status = $10, updated_at = $11, resolved_at = $12
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		alert.ID,
		alert.Type,
		alert.Severity,
		alert.Title,
		alert.Message,
		alert.Device,
		alert.Temperature,
		alert.Threshold,
		alert.Value,
		alert.Status,
		alert.UpdatedAt,
		alert.ResolvedAt,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to update alert")
		return fmt.Errorf("failed to update alert: %w", err)
	}

	return nil
}

func (r *AlertRepository) Resolve(ctx context.Context, alertID string) error {
	query := `
		UPDATE alerts
		SET status = 'resolved', resolved_at = $2, updated_at = $2
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.pool.Exec(ctx, query, alertID, now)
	if err != nil {
		r.logger.WithError(err).Error("Failed to resolve alert")
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	return nil
}

func (r *AlertRepository) ResolveByServerIDAndType(ctx context.Context, serverID string, alertType models.AlertType) error {
	query := `
		UPDATE alerts
		SET status = 'resolved', resolved_at = $3, updated_at = $3
		WHERE server_id = $1 AND type = $2 AND status = 'active'
	`

	now := time.Now()
	_, err := r.pool.Exec(ctx, query, serverID, alertType, now)
	if err != nil {
		r.logger.WithError(err).Error("Failed to resolve alerts by type")
		return fmt.Errorf("failed to resolve alerts by type: %w", err)
	}

	return nil
}

func (r *AlertRepository) Delete(ctx context.Context, alertID string) error {
	query := `DELETE FROM alerts WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, alertID)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete alert")
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	return nil
}

func (r *AlertRepository) GetStats(ctx context.Context, serverID string, duration time.Duration) (*models.AlertStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_alerts,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_alerts,
			COUNT(CASE WHEN status = 'resolved' THEN 1 END) as resolved_alerts,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_count,
			COUNT(CASE WHEN severity = 'warning' THEN 1 END) as warning_count,
			COUNT(CASE WHEN severity = 'info' THEN 1 END) as info_count,
			MAX(created_at) as last_alert_time
		FROM alerts
		WHERE server_id = $1 AND created_at >= $2
	`

	startTime := time.Now().Add(-duration)
	stats := &models.AlertStats{
		ServerID: serverID,
	}

	var lastAlertTime sql.NullTime

	err := r.pool.QueryRow(ctx, query, serverID, startTime).Scan(
		&stats.TotalAlerts,
		&stats.ActiveAlerts,
		&stats.ResolvedAlerts,
		&stats.CriticalCount,
		&stats.WarningCount,
		&stats.InfoCount,
		&lastAlertTime,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to get alert stats")
		return nil, fmt.Errorf("failed to get alert stats: %w", err)
	}

	if lastAlertTime.Valid {
		stats.LastAlertTime = lastAlertTime.Time
	}

	return stats, nil
}
