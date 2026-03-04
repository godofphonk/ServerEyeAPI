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

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type AlertService struct {
	alertRepo interfaces.AlertRepository
	logger    *logrus.Logger
}

func NewAlertService(alertRepo interfaces.AlertRepository, logger *logrus.Logger) *AlertService {
	return &AlertService{
		alertRepo: alertRepo,
		logger:    logger,
	}
}

func (s *AlertService) EvaluateMetrics(ctx context.Context, serverID string, metrics *models.ServerMetrics) ([]*models.Alert, error) {
	var alerts []*models.Alert

	cpuAlerts := s.evaluateCPU(serverID, metrics)
	alerts = append(alerts, cpuAlerts...)

	memoryAlerts := s.evaluateMemory(serverID, metrics)
	alerts = append(alerts, memoryAlerts...)

	diskAlerts := s.evaluateDisk(serverID, metrics)
	alerts = append(alerts, diskAlerts...)

	tempAlerts := s.evaluateTemperature(serverID, metrics)
	alerts = append(alerts, tempAlerts...)

	loadAlerts := s.evaluateLoadAverage(serverID, metrics)
	alerts = append(alerts, loadAlerts...)

	networkAlerts := s.evaluateNetwork(serverID, metrics)
	alerts = append(alerts, networkAlerts...)

	storageAlerts := s.evaluateStorageTemperatures(serverID, metrics)
	alerts = append(alerts, storageAlerts...)

	for _, alert := range alerts {
		if err := s.alertRepo.Create(ctx, alert); err != nil {
			s.logger.WithError(err).WithField("alert_id", alert.ID).Error("Failed to create alert")
		}
	}

	return alerts, nil
}

func (s *AlertService) evaluateCPU(serverID string, metrics *models.ServerMetrics) []*models.Alert {
	var alerts []*models.Alert

	if metrics.CPU > 80 {
		alerts = append(alerts, &models.Alert{
			ID:        uuid.New().String(),
			Type:      models.AlertTypeCPUTemperature,
			ServerID:  serverID,
			Severity:  models.AlertSeverityCritical,
			Title:     "High CPU Usage",
			Message:   fmt.Sprintf("CPU usage is critically high: %.1f%%", metrics.CPU),
			Value:     metrics.CPU,
			Threshold: 80.0,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	} else if metrics.CPU > 60 {
		alerts = append(alerts, &models.Alert{
			ID:        uuid.New().String(),
			Type:      models.AlertTypeCPUTemperature,
			ServerID:  serverID,
			Severity:  models.AlertSeverityWarning,
			Title:     "Moderate CPU Usage",
			Message:   fmt.Sprintf("CPU usage is elevated: %.1f%%", metrics.CPU),
			Value:     metrics.CPU,
			Threshold: 60.0,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	return alerts
}

func (s *AlertService) evaluateMemory(serverID string, metrics *models.ServerMetrics) []*models.Alert {
	var alerts []*models.Alert

	if metrics.Memory > 85 {
		alerts = append(alerts, &models.Alert{
			ID:        uuid.New().String(),
			Type:      models.AlertTypeMemoryUsage,
			ServerID:  serverID,
			Severity:  models.AlertSeverityCritical,
			Title:     "High Memory Usage",
			Message:   fmt.Sprintf("Memory usage is critically high: %.1f%%", metrics.Memory),
			Value:     metrics.Memory,
			Threshold: 85.0,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	} else if metrics.Memory > 70 {
		alerts = append(alerts, &models.Alert{
			ID:        uuid.New().String(),
			Type:      models.AlertTypeMemoryUsage,
			ServerID:  serverID,
			Severity:  models.AlertSeverityWarning,
			Title:     "Moderate Memory Usage",
			Message:   fmt.Sprintf("Memory usage is elevated: %.1f%%", metrics.Memory),
			Value:     metrics.Memory,
			Threshold: 70.0,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	return alerts
}

func (s *AlertService) evaluateDisk(serverID string, metrics *models.ServerMetrics) []*models.Alert {
	var alerts []*models.Alert

	if metrics.Disk > 90 {
		alerts = append(alerts, &models.Alert{
			ID:        uuid.New().String(),
			Type:      models.AlertTypeDiskUsage,
			ServerID:  serverID,
			Severity:  models.AlertSeverityCritical,
			Title:     "Critical Disk Usage",
			Message:   fmt.Sprintf("Disk usage is critically high: %.1f%%", metrics.Disk),
			Value:     metrics.Disk,
			Threshold: 90.0,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	} else if metrics.Disk > 80 {
		alerts = append(alerts, &models.Alert{
			ID:        uuid.New().String(),
			Type:      models.AlertTypeDiskUsage,
			ServerID:  serverID,
			Severity:  models.AlertSeverityWarning,
			Title:     "High Disk Usage",
			Message:   fmt.Sprintf("Disk usage is high: %.1f%%", metrics.Disk),
			Value:     metrics.Disk,
			Threshold: 80.0,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	return alerts
}

func (s *AlertService) evaluateTemperature(serverID string, metrics *models.ServerMetrics) []*models.Alert {
	var alerts []*models.Alert

	if metrics.TemperatureDetails.CPUTemperature > 80 {
		alerts = append(alerts, &models.Alert{
			ID:          uuid.New().String(),
			Type:        models.AlertTypeCPUTemperature,
			ServerID:    serverID,
			Severity:    models.AlertSeverityCritical,
			Title:       "High CPU Temperature",
			Message:     fmt.Sprintf("CPU temperature is critically high: %.1f°C", metrics.TemperatureDetails.CPUTemperature),
			Temperature: metrics.TemperatureDetails.CPUTemperature,
			Threshold:   80.0,
			Status:      "active",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	if metrics.TemperatureDetails.HighestTemperature > 85 {
		alerts = append(alerts, &models.Alert{
			ID:          uuid.New().String(),
			Type:        models.AlertTypeSystemTemperature,
			ServerID:    serverID,
			Severity:    models.AlertSeverityCritical,
			Title:       "High System Temperature",
			Message:     fmt.Sprintf("System temperature is critically high: %.1f°C", metrics.TemperatureDetails.HighestTemperature),
			Temperature: metrics.TemperatureDetails.HighestTemperature,
			Threshold:   85.0,
			Status:      "active",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	return alerts
}

func (s *AlertService) evaluateLoadAverage(serverID string, metrics *models.ServerMetrics) []*models.Alert {
	var alerts []*models.Alert

	if metrics.CPUUsage.LoadAverage.Load1 > 2.0 {
		alerts = append(alerts, &models.Alert{
			ID:        uuid.New().String(),
			Type:      models.AlertTypeLoadAverage,
			ServerID:  serverID,
			Severity:  models.AlertSeverityWarning,
			Title:     "High Load Average",
			Message:   fmt.Sprintf("Load average (1m) is high: %.2f", metrics.CPUUsage.LoadAverage.Load1),
			Value:     metrics.CPUUsage.LoadAverage.Load1,
			Threshold: 2.0,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	return alerts
}

func (s *AlertService) evaluateNetwork(serverID string, metrics *models.ServerMetrics) []*models.Alert {
	var alerts []*models.Alert

	if metrics.Network > 1000 {
		alerts = append(alerts, &models.Alert{
			ID:        uuid.New().String(),
			Type:      models.AlertTypeNetworkUsage,
			ServerID:  serverID,
			Severity:  models.AlertSeverityWarning,
			Title:     "High Network Usage",
			Message:   fmt.Sprintf("Network usage is unusually high: %.1f MB/s", metrics.Network),
			Value:     metrics.Network,
			Threshold: 1000.0,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	return alerts
}

func (s *AlertService) evaluateStorageTemperatures(serverID string, metrics *models.ServerMetrics) []*models.Alert {
	var alerts []*models.Alert

	for _, storage := range metrics.TemperatureDetails.StorageTemperatures {
		evaluation := models.EvaluateStorageTemperature(storage.Type, storage.Temperature)

		if evaluation.Severity == models.AlertSeverityWarning || evaluation.Severity == models.AlertSeverityCritical {
			alerts = append(alerts, &models.Alert{
				ID:          uuid.New().String(),
				Type:        models.AlertTypeStorageTemperature,
				ServerID:    serverID,
				Severity:    evaluation.Severity,
				Title:       "Storage Temperature Alert",
				Message:     evaluation.Message,
				Device:      storage.Device,
				Temperature: storage.Temperature,
				Threshold:   evaluation.Threshold,
				Value:       storage.Temperature,
				Status:      "active",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			})
		}
	}

	return alerts
}

func (s *AlertService) GetActiveAlerts(ctx context.Context, serverID string) ([]*models.Alert, error) {
	return s.alertRepo.GetActiveByServerID(ctx, serverID)
}

func (s *AlertService) GetAlertsByType(ctx context.Context, serverID string, alertType models.AlertType) ([]*models.Alert, error) {
	return s.alertRepo.GetByServerIDAndType(ctx, serverID, alertType)
}

func (s *AlertService) GetAlertsByTimeRange(ctx context.Context, serverID string, start, end time.Time) ([]*models.Alert, error) {
	return s.alertRepo.GetByTimeRange(ctx, serverID, start, end)
}

func (s *AlertService) ResolveAlert(ctx context.Context, alertID string) error {
	return s.alertRepo.Resolve(ctx, alertID)
}

func (s *AlertService) ResolveAlertsByType(ctx context.Context, serverID string, alertType models.AlertType) error {
	return s.alertRepo.ResolveByServerIDAndType(ctx, serverID, alertType)
}

func (s *AlertService) GetAlertStats(ctx context.Context, serverID string, duration time.Duration) (*models.AlertStats, error) {
	return s.alertRepo.GetStats(ctx, serverID, duration)
}
