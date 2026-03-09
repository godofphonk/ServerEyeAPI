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

// MetricsV2 represents the new simplified metrics structure
type MetricsV2 struct {
	CPUUsage    CPUUsageMetrics    `json:"cpu_usage"`
	Memory      MemoryMetrics      `json:"memory"`
	Disks       []DiskMetrics      `json:"disks"`
	Network     NetworkMetrics     `json:"network"`
	Temperature TemperatureMetrics `json:"temperature"`
	System      SystemMetrics      `json:"system"`
	Timestamp   time.Time          `json:"timestamp"`
}

// CPUUsageMetrics represents CPU usage statistics
type CPUUsageMetrics struct {
	UsageTotal   float64      `json:"usage_total"`
	UsageUser    float64      `json:"usage_user"`
	UsageSystem  float64      `json:"usage_system"`
	UsageIdle    float64      `json:"usage_idle"`
	LoadAverage  LoadAverage  `json:"load_average"`
	FrequencyMHz float64      `json:"frequency_mhz"`
}

// LoadAverage represents system load averages
type LoadAverage struct {
	Load1Min  float64 `json:"load_1min"`
	Load5Min  float64 `json:"load_5min"`
	Load15Min float64 `json:"load_15min"`
}

// MemoryMetrics represents memory usage statistics
type MemoryMetrics struct {
	TotalGB     float64 `json:"total_gb"`
	UsedGB      float64 `json:"used_gb"`
	AvailableGB float64 `json:"available_gb"`
	FreeGB      float64 `json:"free_gb"`
	BuffersGB   float64 `json:"buffers_gb"`
	CachedGB    float64 `json:"cached_gb"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskMetrics represents disk usage for a single mount point
type DiskMetrics struct {
	MountPoint  string  `json:"mount_point"`
	DeviceName  string  `json:"device_name"`
	UsedGB      float64 `json:"used_gb"`
	FreeGB      float64 `json:"free_gb"`
	UsedPercent float64 `json:"used_percent"`
}

// NetworkMetrics represents network statistics
type NetworkMetrics struct {
	Interfaces  []NetworkInterface `json:"interfaces"`
	TotalRxMbps float64            `json:"total_rx_mbps"`
	TotalTxMbps float64            `json:"total_tx_mbps"`
}

// NetworkInterface represents a single network interface statistics
type NetworkInterface struct {
	Name        string  `json:"name"`
	RxBytes     int64   `json:"rx_bytes"`
	TxBytes     int64   `json:"tx_bytes"`
	RxPackets   int64   `json:"rx_packets"`
	TxPackets   int64   `json:"tx_packets"`
	RxSpeedMbps float64 `json:"rx_speed_mbps"`
	TxSpeedMbps float64 `json:"tx_speed_mbps"`
	Status      string  `json:"status"`
}

// TemperatureMetrics represents temperature readings
type TemperatureMetrics struct {
	CPU     float64               `json:"cpu"`
	GPU     float64               `json:"gpu"`
	Storage []StorageTemperature  `json:"storage"`
	Highest float64               `json:"highest"`
}

// StorageTemperature represents temperature of a storage device
type StorageTemperature struct {
	Device      string  `json:"device"`
	Temperature float64 `json:"temperature"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	ProcessesTotal    int   `json:"processes_total"`
	ProcessesRunning  int   `json:"processes_running"`
	ProcessesSleeping int   `json:"processes_sleeping"`
	UptimeSeconds     int64 `json:"uptime_seconds"`
}

// MetricsMessageV2 represents the complete metrics message from agent (new format)
type MetricsMessageV2 struct {
	Type      string     `json:"type"`
	ServerID  string     `json:"server_id"`
	Data      MetricsV2  `json:"data"`
	Timestamp int64      `json:"timestamp"`
}
