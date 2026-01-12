package models

import "time"

// ServerMetrics represents server performance metrics
type ServerMetrics struct {
	CPU     float64   `json:"cpu"`     // CPU usage percentage (0-100)
	Memory  float64   `json:"memory"`  // Memory usage percentage (0-100)
	Disk    float64   `json:"disk"`    // Disk usage percentage (0-100)
	Network float64   `json:"network"` // Network usage in MB/s
	Time    time.Time `json:"time"`    // Timestamp when metrics were collected
}

// SystemInfo represents system information
type SystemInfo struct {
	OS           string `json:"os"`           // Operating system name
	Architecture string `json:"architecture"` // System architecture
	Kernel       string `json:"kernel"`       // Kernel version
	Uptime       int64  `json:"uptime"`       // System uptime in seconds
	Hostname     string `json:"hostname"`     // Server hostname
}

// MetricsMessage represents a complete metrics message from agent
type MetricsMessage struct {
	ServerID string        `json:"server_id"`
	Metrics  ServerMetrics `json:"metrics"`
	System   SystemInfo    `json:"system"`
}

// ServerStatus represents server status information
type ServerStatus struct {
	Online       bool      `json:"online"`
	LastSeen     time.Time `json:"last_seen"`
	Version      string    `json:"version"`
	OSInfo       string    `json:"os_info"`
	AgentVersion string    `json:"agent_version"`
	Hostname     string    `json:"hostname"`
}
