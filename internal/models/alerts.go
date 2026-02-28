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

package models

import "time"

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeStorageTemperature AlertType = "storage_temperature"
	AlertTypeCPUTemperature     AlertType = "cpu_temperature"
	AlertTypeMemoryUsage        AlertType = "memory_usage"
	AlertTypeDiskUsage          AlertType = "disk_usage"
)

// Alert represents a system alert
type Alert struct {
	ID          string        `json:"id"`
	Type        AlertType     `json:"type"`
	ServerID    string        `json:"server_id"`
	Severity    AlertSeverity `json:"severity"`
	Title       string        `json:"title"`
	Message     string        `json:"message"`
	Device      string        `json:"device,omitempty"`      // For storage devices
	Temperature float64       `json:"temperature,omitempty"` // For temperature alerts
	Threshold   float64       `json:"threshold,omitempty"`   // Threshold value
	Value       float64       `json:"value,omitempty"`       // Current value
	Status      string        `json:"status"`                // active, resolved
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`
}

// StorageTemperatureAlert represents a storage temperature specific alert
type StorageTemperatureAlert struct {
	Device      string        `json:"device"`      // e.g., /dev/nvme0n1
	Type        string        `json:"type"`        // NVMe, SSD, HDD
	Temperature float64       `json:"temperature"` // Current temperature
	Severity    AlertSeverity `json:"severity"`    // info, warning, critical
	Threshold   float64       `json:"threshold"`   // Alert threshold
	Status      string        `json:"status"`      // normal, warning, critical
	Message     string        `json:"message"`     // Alert message
}

// TemperatureThresholds defines temperature thresholds for different storage types
type TemperatureThresholds struct {
	NVMe struct {
		Normal   float64 `json:"normal"`   // < 60°C
		Warning  float64 `json:"warning"`  // 60-75°C
		Critical float64 `json:"critical"` // > 75°C
	} `json:"nvme"`
	SSD struct {
		Normal   float64 `json:"normal"`   // < 60°C
		Warning  float64 `json:"warning"`  // 60-75°C
		Critical float64 `json:"critical"` // > 75°C
	} `json:"ssd"`
	HDD struct {
		Normal   float64 `json:"normal"`   // < 45°C
		Warning  float64 `json:"warning"`  // 45-50°C
		Critical float64 `json:"critical"` // > 50°C
	} `json:"hdd"`
}

// GetDefaultThresholds returns default temperature thresholds
func GetDefaultThresholds() *TemperatureThresholds {
	return &TemperatureThresholds{
		NVMe: struct {
			Normal   float64 `json:"normal"`
			Warning  float64 `json:"warning"`
			Critical float64 `json:"critical"`
		}{
			Normal:   60.0,
			Warning:  75.0,
			Critical: 75.0,
		},
		SSD: struct {
			Normal   float64 `json:"normal"`
			Warning  float64 `json:"warning"`
			Critical float64 `json:"critical"`
		}{
			Normal:   60.0,
			Warning:  75.0,
			Critical: 75.0,
		},
		HDD: struct {
			Normal   float64 `json:"normal"`
			Warning  float64 `json:"warning"`
			Critical float64 `json:"critical"`
		}{
			Normal:   45.0,
			Warning:  50.0,
			Critical: 50.0,
		},
	}
}

// EvaluateStorageTemperature evaluates storage temperature and returns alert status
func EvaluateStorageTemperature(deviceType string, temperature float64) StorageTemperatureAlert {
	thresholds := GetDefaultThresholds()

	var warningThreshold, criticalThreshold float64
	var alertSeverity AlertSeverity
	var status string
	var message string

	switch deviceType {
	case "NVMe":
		warningThreshold = thresholds.NVMe.Warning
		criticalThreshold = thresholds.NVMe.Critical
	case "SSD":
		warningThreshold = thresholds.SSD.Warning
		criticalThreshold = thresholds.SSD.Critical
	case "HDD":
		warningThreshold = thresholds.HDD.Warning
		criticalThreshold = thresholds.HDD.Critical
	default:
		// Default to SSD thresholds
		warningThreshold = thresholds.SSD.Warning
		criticalThreshold = thresholds.SSD.Critical
	}

	if temperature >= criticalThreshold {
		alertSeverity = AlertSeverityCritical
		status = "critical"
		message = "Storage temperature critically high"
	} else if temperature >= warningThreshold {
		alertSeverity = AlertSeverityWarning
		status = "warning"
		message = "Storage temperature above normal range"
	} else {
		alertSeverity = AlertSeverityInfo
		status = "normal"
		message = "Storage temperature normal"
	}

	return StorageTemperatureAlert{
		Type:        deviceType,
		Temperature: temperature,
		Severity:    alertSeverity,
		Threshold:   warningThreshold,
		Status:      status,
		Message:     message,
	}
}
