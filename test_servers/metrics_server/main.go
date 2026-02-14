package main

import (
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
	SampleCount int64     `json:"sample_count"`
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
	fmt.Fprintf(w, `{"server_id":"%s","start_time":"%s","end_time":"%s","granularity":"%s","data_points":[],"total_points":%d}`,
		data.ServerID, data.StartTime.Format(time.RFC3339), data.EndTime.Format(time.RFC3339),
		data.Granularity, data.TotalPoints)
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
	fmt.Fprintf(w, `{"server_id":"%s","start_time":"%s","end_time":"%s","granularity":"1m","total_points":%d}`,
		data.ServerID, data.StartTime.Format(time.RFC3339), data.EndTime.Format(time.RFC3339), data.TotalPoints)
}

func handleMetricsSummary(w http.ResponseWriter, r *http.Request) {
	summary := `{
		"total_data_points": 467000,
		"total_servers": 25,
		"storage_size": "544 MB",
		"last_updated": "` + time.Now().Format(time.RFC3339) + `"
	}`

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, summary)
}

func main() {
	router := mux.NewRouter()

	// Tiered metrics endpoints
	router.HandleFunc("/api/servers/{server_id}/metrics/tiered", handleTieredMetrics).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics/realtime", handleRealTimeMetrics).Methods("GET")
	router.HandleFunc("/api/metrics/summary", handleMetricsSummary).Methods("GET")

	// Health endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	}).Methods("GET")

	// Test endpoint
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"ServerEye API Metrics Test Server","endpoints":3}`)
	}).Methods("GET")

	fmt.Println("🚀 Starting ServerEye API Metrics Test Server on :8083")
	fmt.Println("📊 Available endpoints:")
	fmt.Println("  GET /health - Health check")
	fmt.Println("  GET /test - Test endpoint")
	fmt.Println("  GET /api/servers/{id}/metrics/tiered - Tiered metrics")
	fmt.Println("  GET /api/servers/{id}/metrics/realtime - Real-time metrics")
	fmt.Println("  GET /api/metrics/summary - Metrics summary")

	if err := http.ListenAndServe(":8083", router); err != nil {
		panic(err)
	}
}
