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
	NetworkDetails struct {
		Interfaces []struct {
			Name        string  `json:"name"`          // Interface name (e.g., eth0, enp111s0)
			RxBytes     int64   `json:"rx_bytes"`      // Total received bytes
			TxBytes     int64   `json:"tx_bytes"`      // Total transmitted bytes
			RxPackets   int64   `json:"rx_packets"`    // Total received packets
			TxPackets   int64   `json:"tx_packets"`    // Total transmitted packets
			RxSpeedMbps float64 `json:"rx_speed_mbps"` // Current receive speed in Mbps
			TxSpeedMbps float64 `json:"tx_speed_mbps"` // Current transmit speed in Mbps
			Status      string  `json:"status"`        // Interface status (up/down)
		} `json:"interfaces"` // Network interfaces list
		TotalRxMbps float64 `json:"total_rx_mbps"` // Total receive speed across all interfaces
		TotalTxMbps float64 `json:"total_tx_mbps"` // Total transmit speed across all interfaces
	} `json:"network_details"` // Detailed network statistics
	TemperatureDetails struct {
		CPUTemperature      float64 `json:"cpu_temperature"`    // CPU temperature in Celsius
		GPUTemperature      float64 `json:"gpu_temperature"`    // GPU temperature in Celsius (0 if not available)
		SystemTemperature   float64 `json:"system_temperature"` // System/board temperature in Celsius
		StorageTemperatures []struct {
			Device      string  `json:"device"`      // Storage device name (e.g., /dev/sda, nvme0n1)
			Type        string  `json:"type"`        // Storage type (HDD, SSD, NVMe)
			Temperature float64 `json:"temperature"` // Storage temperature in Celsius
		} `json:"storage_temperatures"` // Individual storage device temperatures
		HighestTemperature float64 `json:"highest_temperature"` // Highest temperature across all sensors
		TemperatureUnit    string  `json:"temperature_unit"`    // Temperature unit (celsius)
	} `json:"temperature_details"` // Detailed temperature monitoring
	SystemDetails struct {
		Hostname          string `json:"hostname"`           // System hostname
		OS                string `json:"os"`                 // Operating system name and version
		Kernel            string `json:"kernel"`             // Kernel version
		Architecture      string `json:"architecture"`       // System architecture (x86_64, arm64, etc.)
		UptimeSeconds     int64  `json:"uptime_seconds"`     // System uptime in seconds
		UptimeHuman       string `json:"uptime_human"`       // Human-readable uptime format
		BootTime          string `json:"boot_time"`          // System boot timestamp
		ProcessesTotal    int    `json:"processes_total"`    // Total number of processes
		ProcessesRunning  int    `json:"processes_running"`  // Number of running processes
		ProcessesSleeping int    `json:"processes_sleeping"` // Number of sleeping processes
	} `json:"system_details"` // Detailed system information
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
