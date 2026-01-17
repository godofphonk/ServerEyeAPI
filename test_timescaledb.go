package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/timescaledb"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Connect to TimescaleDB
	timescaleDBURL := "postgres://postgres:devpassword@localhost:5434/servereye?sslmode=disable"

	config := timescaledb.DefaultClientConfig()
	client, err := timescaledb.NewClient(timescaleDBURL, logger, config)
	if err != nil {
		log.Fatalf("Failed to connect to TimescaleDB: %v", err)
	}
	defer client.Close()

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping TimescaleDB: %v", err)
	}

	fmt.Println("✅ Connected to TimescaleDB successfully!")

	// Test storing metrics
	testMetrics := &models.ServerMetrics{
		CPU:     75.5,
		Memory:  60.2,
		Disk:    45.8,
		Network: 120.4,
		Time:    time.Now(),
		CPUUsage: struct {
			UsageTotal  float64 `json:"usage_total"`
			UsageUser   float64 `json:"usage_user"`
			UsageSystem float64 `json:"usage_system"`
			UsageIdle   float64 `json:"usage_idle"`
			LoadAverage struct {
				Load1  float64 `json:"load_1min"`
				Load5  float64 `json:"load_5min"`
				Load15 float64 `json:"load_15min"`
			} `json:"load_average"`
			Cores     int     `json:"cores"`
			Frequency float64 `json:"frequency"`
		}{
			UsageTotal:  75.5,
			UsageUser:   60.0,
			UsageSystem: 15.5,
			UsageIdle:   24.5,
			LoadAverage: struct {
				Load1  float64 `json:"load_1min"`
				Load5  float64 `json:"load_5min"`
				Load15 float64 `json:"load_15min"`
			}{Load1: 1.5, Load5: 1.2, Load15: 0.8},
			Cores:     8,
			Frequency: 2400.0,
		},
		MemoryDetails: struct {
			TotalGB     float64 `json:"total_gb"`
			UsedGB      float64 `json:"used_gb"`
			AvailableGB float64 `json:"available_gb"`
			FreeGB      float64 `json:"free_gb"`
			BuffersGB   float64 `json:"buffers_gb"`
			CachedGB    float64 `json:"cached_gb"`
			UsedPercent float64 `json:"used_percent"`
		}{
			TotalGB:     16.0,
			UsedGB:      9.6,
			AvailableGB: 6.4,
			FreeGB:      2.1,
			BuffersGB:   0.5,
			CachedGB:    3.0,
			UsedPercent: 60.2,
		},
		TemperatureDetails: struct {
			CPUTemperature      float64 `json:"cpu_temperature"`
			GPUTemperature      float64 `json:"gpu_temperature"`
			SystemTemperature   float64 `json:"system_temperature"`
			StorageTemperatures []struct {
				Device      string  `json:"device"`
				Type        string  `json:"type"`
				Temperature float64 `json:"temperature"`
			} `json:"storage_temperatures"`
			HighestTemperature float64 `json:"highest_temperature"`
			TemperatureUnit    string  `json:"temperature_unit"`
		}{
			CPUTemperature:    65.2,
			GPUTemperature:    0.0,
			SystemTemperature: 45.1,
			StorageTemperatures: []struct {
				Device      string  `json:"device"`
				Type        string  `json:"type"`
				Temperature float64 `json:"temperature"`
			}{{Device: "/dev/sda", Type: "SSD", Temperature: 42.3}},
			HighestTemperature: 65.2,
			TemperatureUnit:    "celsius",
		},
		SystemDetails: struct {
			Hostname          string `json:"hostname"`
			OS                string `json:"os"`
			Kernel            string `json:"kernel"`
			Architecture      string `json:"architecture"`
			UptimeSeconds     int64  `json:"uptime_seconds"`
			UptimeHuman       string `json:"uptime_human"`
			BootTime          string `json:"boot_time"`
			ProcessesTotal    int    `json:"processes_total"`
			ProcessesRunning  int    `json:"processes_running"`
			ProcessesSleeping int    `json:"processes_sleeping"`
		}{
			Hostname:          "test-server",
			OS:                "Ubuntu 22.04",
			Kernel:            "5.15.0-91-generic",
			Architecture:      "x86_64",
			UptimeSeconds:     86400,
			UptimeHuman:       "1 day",
			BootTime:          "2026-01-16T23:00:00Z",
			ProcessesTotal:    156,
			ProcessesRunning:  2,
			ProcessesSleeping: 154,
		},
	}

	serverID := "test-server-001"

	// Store metrics
	if err := client.StoreMetric(ctx, serverID, testMetrics); err != nil {
		log.Fatalf("Failed to store metrics: %v", err)
	}

	fmt.Println("✅ Metrics stored successfully!")

	// Retrieve metrics
	retrievedMetrics, err := client.GetLatestMetric(ctx, serverID)
	if err != nil {
		log.Fatalf("Failed to retrieve metrics: %v", err)
	}

	fmt.Printf("✅ Retrieved metrics: CPU=%.1f%%, Memory=%.1f%%, Disk=%.1f%%\n",
		retrievedMetrics.CPU, retrievedMetrics.Memory, retrievedMetrics.Disk)

	// Test server status
	testStatus := &models.ServerStatus{
		Online:       true,
		LastSeen:     time.Now(),
		Version:      "1.0.0",
		OSInfo:       "Ubuntu 22.04",
		AgentVersion: "1.0.0",
		Hostname:     "test-server",
	}

	if err := client.SetServerStatus(ctx, serverID, testStatus); err != nil {
		log.Fatalf("Failed to store status: %v", err)
	}

	fmt.Println("✅ Server status stored successfully!")

	// Retrieve status
	retrievedStatus, err := client.GetServerStatus(ctx, serverID)
	if err != nil {
		log.Fatalf("Failed to retrieve status: %v", err)
	}

	fmt.Printf("✅ Retrieved status: Online=%t, Hostname=%s\n",
		retrievedStatus.Online, retrievedStatus.Hostname)

	// Get stats
	stats, err := client.GetMetricsStats(ctx)
	if err != nil {
		log.Fatalf("Failed to get metrics stats: %v", err)
	}

	fmt.Printf("✅ Metrics stats: TotalRecords=%d, UniqueServers=%d, TableSize=%s\n",
		stats.TotalRecords, stats.UniqueServers, stats.TableSize)

	fmt.Println("\n🎉 All TimescaleDB tests passed successfully!")
}
