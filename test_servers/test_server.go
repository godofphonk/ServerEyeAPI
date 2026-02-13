package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/handlers"
	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create test handlers without database
	tieredMetricsService := &services.TieredMetricsService{}
	tieredMetricsHandler := handlers.NewTieredMetricsHandler(tieredMetricsService, logger)

	// Create a simple router for testing
	router := mux.NewRouter()

	// Add tiered metrics endpoints
	router.HandleFunc("/api/servers/{server_id}/metrics/tiered", tieredMetricsHandler.GetMetrics).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics/realtime", tieredMetricsHandler.GetRealTimeMetrics).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics/historical", tieredMetricsHandler.GetHistoricalMetrics).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics/dashboard", tieredMetricsHandler.GetDashboardMetrics).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics/comparison", tieredMetricsHandler.GetMetricsComparison).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics/heatmap", tieredMetricsHandler.GetMetricsHeatmap).Methods("GET")
	router.HandleFunc("/api/metrics/summary", tieredMetricsHandler.GetMetricsSummary).Methods("GET")

	// Add health endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	}).Methods("GET")

	// Add test endpoints
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"ServerEye API Test Server","endpoints":7}`)
	}).Methods("GET")

	fmt.Println("🚀 Starting ServerEye API Test Server on :8082")
	fmt.Println("📊 Available endpoints:")
	fmt.Println("  GET /health - Health check")
	fmt.Println("  GET /test - Test endpoint")
	fmt.Println("  GET /api/servers/{id}/metrics/tiered - Tiered metrics")
	fmt.Println("  GET /api/servers/{id}/metrics/realtime - Real-time metrics")
	fmt.Println("  GET /api/servers/{id}/metrics/historical - Historical metrics")
	fmt.Println("  GET /api/servers/{id}/metrics/dashboard - Dashboard metrics")
	fmt.Println("  GET /api/servers/{id}/metrics/comparison - Metrics comparison")
	fmt.Println("  GET /api/servers/{id}/metrics/heatmap - Heatmap data")
	fmt.Println("  GET /api/metrics/summary - Metrics summary")

	server := &http.Server{
		Addr:         ":8082",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
