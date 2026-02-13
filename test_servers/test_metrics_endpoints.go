package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Mock data structures
type MockMetricsData struct {
	ServerID    string      `json:"server_id"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time"`
	Granularity string      `json:"granularity"`
	DataPoints  []DataPoint `json:"data_points"`
	TotalPoints int64       `json:"total_points"`
}

type DataPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUAvg      float64   `json:"cpu_avg"`
	CPUMax      float64   `json:"cpu_max"`
	CPUMin      float64   `json:"cpu_min"`
	MemoryAvg   float64   `json:"memory_avg"`
	MemoryMax   float64   `json:"memory_max"`
	MemoryMin   float64   `json:"memory_min"`
	DiskAvg     float64   `json:"disk_avg"`
	DiskMax     float64   `json:"disk_max"`
	NetworkAvg  float64   `json:"network_avg"`
	NetworkMax  float64   `json:"network_max"`
	SampleCount int64      `json:"sample_count"`
}

type MockDashboardData struct {
	ServerID       string           `json:"server_id"`
	CurrentStatus  CurrentStatus    `json:"current_status"`
	HourlyTrend    []DataPoint      `json:"hourly_trend"`
	DailyTrend     []DataPoint      `json:"daily_trend"`
	HeatmapData    []HeatmapPoint   `json:"heatmap_data"`
	Performance    PerformanceMetrics `json:"performance"`
	LastUpdated    time.Time        `json:"last_updated"`
}

type CurrentStatus struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	NetworkIO   float64   `json:"network_io"`
	LoadAvg     float64   `json:"load_avg"`
	Uptime      string    `json:"uptime"`
}

type HeatmapPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUAvg      float64   `json:"cpu_avg"`
	MemoryAvg   float64   `json:"memory_avg"`
	DiskAvg     float64   `json:"disk_avg"`
	SampleCount int64      `json:"sample_count"`
}

type PerformanceMetrics struct {
	AvgCPU    float64 `json:"avg_cpu"`
	MaxCPU    float64 `json:"max_cpu"`
	AvgMemory float64 `json:"avg_memory"`
	MaxMemory float64 `json:"max_memory"`
}

// Generate mock data
func generateMockMetrics(serverID string, start, end time.Time, granularity string) MockMetricsData {
	var points []DataPoint
	pointCount := 60 // Default to 60 points
	
	switch granularity {
	case "1m":
		pointCount = 60
	case "5m":
		pointCount = 12
	case "10m":
		pointCount = 6
	case "1h":
		pointCount = 1
	}
	
	interval := end.Sub(start) / time.Duration(pointCount)
	
	for i := 0; i < pointCount; i++ {
		timestamp := start.Add(time.Duration(i) * interval)
		points = append(points, DataPoint{
			Timestamp:   timestamp,
			CPUAvg:      60.0 + float64(i%20),
			CPUMax:      80.0 + float64(i%10),
			CPUMin:      40.0 + float64(i%15),
			MemoryAvg:   55.0 + float64(i%25),
			MemoryMax:   70.0 + float64(i%15),
			MemoryMin:   45.0 + float64(i%10),
			DiskAvg:     45.0,
			DiskMax:     45.0,
			NetworkAvg:  100.0 + float64(i%50),
			NetworkMax:  150.0 + float64(i%30),
			SampleCount: 1,
		})
	}
	
	return MockMetricsData{
		ServerID:    serverID,
		StartTime:   start,
		EndTime:     end,
		Granularity: granularity,
		DataPoints:  points,
		TotalPoints: int64(len(points)),
	}
}

func generateMockDashboard(serverID string) MockDashboardData {
	now := time.Now()
	
	// Current status
	current := CurrentStatus{
		Timestamp:   now,
		CPUUsage:    75.5,
		MemoryUsage: 62.3,
		DiskUsage:   45.8,
		NetworkIO:   125.4,
		LoadAvg:     1.85,
		Uptime:      "15 days, 3h 24m",
	}
	
	// Hourly trend (24 hours)
	var hourlyTrend []DataPoint
	for i := 0; i < 24; i++ {
		timestamp := now.Add(-time.Duration(23-i) * time.Hour)
		hourlyTrend = append(hourlyTrend, DataPoint{
			Timestamp:   timestamp,
			CPUAvg:      60.0 + float64(i%20),
			MemoryAvg:   55.0 + float64(i%25),
			DiskAvg:     45.0,
			NetworkAvg:  100.0 + float64(i%50),
			SampleCount: 60,
		})
	}
	
	// Daily trend (30 days)
	var dailyTrend []DataPoint
	for i := 0; i < 30; i++ {
		timestamp := now.Add(-time.Duration(29-i) * 24 * time.Hour)
		dailyTrend = append(dailyTrend, DataPoint{
			Timestamp:   timestamp,
			CPUAvg:      65.0 + float64(i%15),
			MemoryAvg:   58.0 + float64(i%20),
			DiskAvg:     45.0,
			NetworkAvg:  110.0 + float64(i%40),
			SampleCount: 1440,
		})
	}
	
	// Heatmap data (last 24 hours, hourly buckets)
	var heatmapData []HeatmapPoint
	for i := 0; i < 24; i++ {
		timestamp := now.Add(-time.Duration(23-i) * time.Hour)
		heatmapData = append(heatmapData, HeatmapPoint{
			Timestamp:   timestamp,
			CPUAvg:      60.0 + float64(i%20),
			MemoryAvg:   55.0 + float64(i%25),
			DiskAvg:     45.0,
			SampleCount: 60,
		})
	}
	
	performance := PerformanceMetrics{
		AvgCPU:    68.5,
		MaxCPU:    95.2,
		AvgMemory: 62.1,
		MaxMemory: 89.7,
	}
	
	return MockDashboardData{
		ServerID:      serverID,
		CurrentStatus: current,
		HourlyTrend:   hourlyTrend,
		DailyTrend:    dailyTrend,
		HeatmapData:   heatmapData,
		Performance:   performance,
		LastUpdated:   now,
	}
}

// Handler functions
func handleTieredMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	
	start, _ := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
	end, _ := time.Parse(time.RFC3339, r.URL.Query().Get("end"))
	
	// Determine granularity based on time range
	duration := end.Sub(start)
	var granularity string
	if duration <= time.Hour {
		granularity = "1m"
	} else if duration <= 3*time.Hour {
		granularity = "5m"
	} else if duration <= 24*time.Hour {
		granularity = "10m"
	} else {
		granularity = "1h"
	}
	
	data := generateMockMetrics(serverID, start, end, granularity)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func handleRealTimeMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	
	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "1h"
	}
	
	now := time.Now()
	var start time.Time
	switch durationStr {
	case "30m":
		start = now.Add(-30 * time.Minute)
	case "1h":
		start = now.Add(-1 * time.Hour)
	default:
		start = now.Add(-1 * time.Hour)
	}
	
	data := generateMockMetrics(serverID, start, now, "1m")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func handleDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	
	data := generateMockDashboard(serverID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func handleMetricsComparison(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	
	// Parse periods
	p1Start, _ := time.Parse(time.RFC3339, r.URL.Query().Get("period1_start"))
	p1End, _ := time.Parse(time.RFC3339, r.URL.Query().Get("period1_end"))
	p2Start, _ := time.Parse(time.RFC3339, r.URL.Query().Get("period2_start"))
	p2End, _ := time.Parse(time.RFC3339, r.URL.Query().Get("period2_end"))
	
	data1 := generateMockMetrics(serverID, p1Start, p1End, "10m")
	data2 := generateMockMetrics(serverID, p2Start, p2End, "10m")
	
	// Calculate averages
	var avg1CPU, avg2CPU float64
	for _, p := range data1.DataPoints {
		avg1CPU += p.CPUAvg
	}
	avg1CPU /= float64(len(data1.DataPoints))
	
	for _, p := range data2.DataPoints {
		avg2CPU += p.CPUAvg
	}
	avg2CPU /= float64(len(data2.DataPoints))
	
	comparison := map[string]interface{}{
		"server_id": serverID,
		"period1": map[string]interface{}{
			"start":       p1Start,
			"end":         p1End,
			"granularity": "10m",
		},
		"period2": map[string]interface{}{
			"start":       p2Start,
			"end":         p2End,
			"granularity": "10m",
		},
		"averages1": map[string]interface{}{
			"cpu_avg":    avg1CPU,
			"memory_avg": 58.5,
			"disk_avg":   45.8,
			"network_avg": 110.2,
		},
		"averages2": map[string]interface{}{
			"cpu_avg":    avg2CPU,
			"memory_avg": 62.1,
			"disk_avg":   45.8,
			"network_avg": 115.7,
		},
		"changes": map[string]interface{}{
			"cpu_change":    ((avg2CPU - avg1CPU) / avg1CPU) * 100,
			"memory_change": 6.2,
			"disk_change":   0.0,
			"network_change": 5.0,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

func handleMetricsHeatmap(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID := vars["server_id"]
	
	start, _ := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
	end, _ := time.Parse(time.RFC3339, r.URL.Query().Get("end"))
	
	duration := end.Sub(start)
	var granularity string
	if duration <= 24*time.Hour {
		granularity = "1h"
	} else {
		granularity = "1h"
	}
	
	data := generateMockMetrics(serverID, start, end, granularity)
	
	// Convert to heatmap format
	var heatmapPoints []HeatmapPoint
	for _, p := range data.DataPoints {
		heatmapPoints = append(heatmapPoints, HeatmapPoint{
			Timestamp:   p.Timestamp,
			CPUAvg:      p.CPUAvg,
			MemoryAvg:   p.MemoryAvg,
			DiskAvg:     p.DiskAvg,
			SampleCount: p.SampleCount,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(heatmapPoints)
}

func handleMetricsSummary(w http.ResponseWriter, r *http.Request) {
	summary := map[string]interface{}{
		"granularity_stats": map[string]interface{}{
			"1m": map[string]interface{}{
				"total_records":   180000,
				"unique_servers":  25,
				"earliest_record": time.Now().Add(-3 * time.Hour),
				"latest_record":   time.Now(),
				"table_size":      "245 MB",
			},
			"5m": map[string]interface{}{
				"total_records":   86400,
				"unique_servers":  25,
				"earliest_record": time.Now().Add(-24 * time.Hour),
				"latest_record":   time.Now(),
				"table_size":      "156 MB",
			},
			"10m": map[string]interface{}{
				"total_records":   129600,
				"unique_servers":  25,
				"earliest_record": time.Now().Add(-7 * 24 * time.Hour),
				"latest_record":   time.Now(),
				"table_size":      "98 MB",
			},
			"1h": map[string]interface{}{
				"total_records":   72000,
				"unique_servers":  25,
				"earliest_record": time.Now().Add(-30 * 24 * time.Hour),
				"latest_record":   time.Now(),
				"table_size":      "45 MB",
			},
		},
		"total_data_points": 467000,
		"total_servers":    25,
		"storage_size":     "544 MB",
		"last_updated":     time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func main() {
	router := mux.NewRouter()
	
	// Tiered metrics endpoints
	router.HandleFunc("/api/servers/{server_id}/metrics/tiered", handleTieredMetrics).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics/realtime", handleRealTimeMetrics).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics/dashboard", handleDashboardMetrics).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics/comparison", handleMetricsComparison).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics/heatmap", handleMetricsHeatmap).Methods("GET")
	router.HandleFunc("/api/metrics/summary", handleMetricsSummary).Methods("GET")
	
	// Health endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	}).Methods("GET")
	
	// Test endpoint
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"ServerEye API Metrics Test Server","endpoints":6}`)
	}).Methods("GET")
	
	fmt.Println("🚀 Starting ServerEye API Metrics Test Server on :8083")
	fmt.Println("📊 Available endpoints:")
	fmt.Println("  GET /health - Health check")
	fmt.Println("  GET /test - Test endpoint")
	fmt.Println("  GET /api/servers/{id}/metrics/tiered - Tiered metrics with auto-granularity")
	fmt.Println("  GET /api/servers/{id}/metrics/realtime - Real-time metrics (last hour)")
	fmt.Println("  GET /api/servers/{id}/metrics/dashboard - Dashboard metrics")
	fmt.Println("  GET /api/servers/{id}/metrics/comparison - Compare two periods")
	fmt.Println("  GET /api/servers/{id}/metrics/heatmap - Heatmap data")
	fmt.Println("  GET /api/metrics/summary - System metrics summary")
	fmt.Println()
	fmt.Println("🔧 Test commands:")
	fmt.Println(`  curl -s http://localhost:8083/api/servers/test-001/metrics/tiered?start=2026-02-13T15:00:00Z&end=2026-02-13T16:00:00Z | jq`)
	fmt.Println(`  curl -s http://localhost:8083/api/servers/test-001/metrics/realtime?duration=30m | jq`)
	fmt.Println(`  curl -s http://localhost:8083/api/servers/test-001/metrics/dashboard | jq`)
	fmt.Println(`  curl -s http://localhost:8083/api/metrics/summary | jq`)
	
	if err := http.ListenAndServe(":8083", router); err != nil {
		panic(err)
	}
}
