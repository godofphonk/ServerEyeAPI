package models

import "time"

// ServerMetrics represents server performance metrics
type ServerMetrics struct {
	CPU      float64 `json:"cpu"`     // CPU usage percentage (0-100)
	Memory   float64 `json:"memory"`  // Memory usage percentage (0-100)
	Disk     float64 `json:"disk"`    // Disk usage percentage (0-100)
	Network  float64 `json:"network"` // Network usage in MB/s
	CPUUsage struct {
		UsageTotal  float64 `json:"usage_total"`  // Total CPU usage percentage (0-100)
		UsageUser   float64 `json:"usage_user"`   // User space CPU usage percentage
		UsageSystem float64 `json:"usage_system"` // System space CPU usage percentage
		UsageIdle   float64 `json:"usage_idle"`   // Idle CPU usage percentage
		LoadAverage struct {
			Load1  float64 `json:"load_1min"`  // 1-minute load average
			Load5  float64 `json:"load_5min"`  // 5-minute load average
			Load15 float64 `json:"load_15min"` // 15-minute load average
		} `json:"load_average"` // System load averages
		Cores     int     `json:"cores"`     // Number of CPU cores
		Frequency float64 `json:"frequency"` // CPU frequency in MHz
	} `json:"cpu_usage"` // Detailed CPU usage statistics
	MemoryDetails struct {
		TotalGB     float64 `json:"total_gb"`     // Total memory in GB
		UsedGB      float64 `json:"used_gb"`      // Used memory in GB
		AvailableGB float64 `json:"available_gb"` // Available memory in GB
		FreeGB      float64 `json:"free_gb"`      // Free memory in GB
		BuffersGB   float64 `json:"buffers_gb"`   // Buffers in GB
		CachedGB    float64 `json:"cached_gb"`    // Cached memory in GB
		UsedPercent float64 `json:"used_percent"` // Memory usage percentage (0-100)
	} `json:"memory_details"` // Detailed memory statistics
	DiskDetails []struct {
		Path        string  `json:"path"`         // Mount path
		TotalGB     float64 `json:"total_gb"`     // Total disk space in GB
		UsedGB      float64 `json:"used_gb"`      // Used disk space in GB
		FreeGB      float64 `json:"free_gb"`      // Free disk space in GB
		UsedPercent float64 `json:"used_percent"` // Disk usage percentage (0-100)
		Filesystem  string  `json:"filesystem"`   // Filesystem type
	} `json:"disk_details"` // Detailed disk statistics for all mounts
	Time time.Time `json:"time"` // Timestamp when metrics were collected
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
