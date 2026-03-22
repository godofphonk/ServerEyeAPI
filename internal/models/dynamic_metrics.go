package models

import "time"

// DynamicMetrics represents real-time server performance metrics
// These are time-series data that change frequently (ONLY dynamic data, NO static info)
type DynamicMetrics struct {
	// Core performance metrics (percentages)
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
	NetworkMbps   float64 `json:"network_mbps"`

	// Load averages (uses existing LoadAverage from metrics_v2.go)
	LoadAverage LoadAverage `json:"load_average"`

	// Temperature monitoring (uses existing TemperatureMetrics from metrics_v2.go)
	TemperatureCelsius float64            `json:"temperature_celsius"`
	Temperatures       TemperatureMetrics `json:"temperatures,omitempty"`

	// Process information
	ProcessesTotal    int `json:"processes_total"`
	ProcessesRunning  int `json:"processes_running"`
	ProcessesSleeping int `json:"processes_sleeping"`

	// System uptime (dynamic - changes every second)
	UptimeSeconds int64 `json:"uptime_seconds"`

	// Memory details (dynamic usage - uses existing MemoryMetrics from metrics_v2.go)
	MemoryDetails MemoryMetrics `json:"memory_details,omitempty"`

	// Disk details (dynamic usage - uses existing DiskMetrics from metrics_v2.go)
	DiskDetails []DiskMetrics `json:"disk_details,omitempty"`

	// Network details (dynamic traffic - uses existing NetworkMetrics from metrics_v2.go)
	NetworkDetails NetworkMetrics `json:"network_details,omitempty"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`
}

// DynamicMetricsResponse represents the API response for metrics endpoint
type DynamicMetricsResponse struct {
	ServerID string         `json:"server_id"`
	Metrics  DynamicMetrics `json:"metrics"`
}
