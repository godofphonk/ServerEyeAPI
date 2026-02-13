package services

import (
	"context"
	"fmt"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/storage/timescaledb"
	"github.com/sirupsen/logrus"
)

// TieredMetricsService handles tiered metrics with automatic granularity selection
type TieredMetricsService struct {
	timescaleDB *timescaledb.Client
	logger      *logrus.Logger
}

// NewTieredMetricsService creates a new tiered metrics service
func NewTieredMetricsService(timescaleDB *timescaledb.Client, logger *logrus.Logger) *TieredMetricsService {
	return &TieredMetricsService{
		timescaleDB: timescaleDB,
		logger:      logger,
	}
}

// GetMetricsWithAutoGranularity automatically selects the best granularity based on time range
func (s *TieredMetricsService) GetMetricsWithAutoGranularity(
	ctx context.Context,
	serverID string,
	startTime time.Time,
	endTime time.Time,
) (*timescaledb.TieredMetricsResponse, error) {
	req := &timescaledb.TieredMetricsRequest{
		ServerID:  serverID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	response, err := s.timescaleDB.GetTieredMetrics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tiered metrics: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"server_id":   serverID,
		"granularity": response.Granularity,
		"data_points": response.TotalPoints,
		"duration":    endTime.Sub(startTime),
	}).Info("Retrieved metrics with auto-granularity")

	return response, nil
}

// GetRealTimeMetrics gets the most recent metrics (last hour with 1-minute granularity)
func (s *TieredMetricsService) GetRealTimeMetrics(
	ctx context.Context,
	serverID string,
	duration time.Duration,
) (*timescaledb.TieredMetricsResponse, error) {
	// Limit to maximum 1 hour for real-time data
	if duration > time.Hour {
		duration = time.Hour
	}

	endTime := time.Now()
	startTime := endTime.Add(-duration)

	req := &timescaledb.TieredMetricsRequest{
		ServerID:    serverID,
		StartTime:   startTime,
		EndTime:     endTime,
		Granularity: timescaledb.Granularity1Min,
	}

	response, err := s.timescaleDB.GetTieredMetrics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get real-time metrics: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"server_id":   serverID,
		"data_points": response.TotalPoints,
		"duration":    duration,
	}).Debug("Retrieved real-time metrics")

	return response, nil
}

// GetHistoricalMetrics gets historical metrics with specified granularity
func (s *TieredMetricsService) GetHistoricalMetrics(
	ctx context.Context,
	serverID string,
	startTime time.Time,
	endTime time.Time,
	granularity timescaledb.MetricsGranularity,
) (*timescaledb.TieredMetricsResponse, error) {
	req := &timescaledb.TieredMetricsRequest{
		ServerID:    serverID,
		StartTime:   startTime,
		EndTime:     endTime,
		Granularity: granularity,
	}

	response, err := s.timescaleDB.GetTieredMetrics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical metrics: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"server_id":   serverID,
		"granularity": granularity,
		"data_points": response.TotalPoints,
	}).Info("Retrieved historical metrics")

	return response, nil
}

// GetDashboardMetrics gets optimized metrics for dashboard display
func (s *TieredMetricsService) GetDashboardMetrics(
	ctx context.Context,
	serverID string,
) (*DashboardMetrics, error) {
	// Get current status
	current, err := s.timescaleDB.GetCurrentSystemStatus(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current status: %w", err)
	}

	// Get last 24 hours with appropriate granularity
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	points, actualGranularity, err := s.timescaleDB.GetMetricsByGranularity(ctx, serverID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get 24h metrics: %w", err)
	}

	// Calculate trends
	trends := s.calculateTrends(points)

	// Get heatmap data
	heatmapPoints, err := s.timescaleDB.GetMetricsHeatmap(ctx, serverID, startTime, endTime)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get heatmap data")
		heatmapPoints = []*timescaledb.HeatmapPoint{}
	}

	return &DashboardMetrics{
		ServerID:    serverID,
		Current:     current,
		Granularity: actualGranularity,
		Points24h:   points,
		Trends:      trends,
		HeatmapData: heatmapPoints,
		LastUpdated: time.Now(),
	}, nil
}

// GetMetricsComparison compares metrics between two time periods
func (s *TieredMetricsService) GetMetricsComparison(
	ctx context.Context,
	serverID string,
	period1Start, period1End time.Time,
	period2Start, period2End time.Time,
) (*MetricsComparison, error) {
	// Get metrics for both periods
	points1, gran1, err := s.timescaleDB.GetMetricsByGranularity(ctx, serverID, period1Start, period1End)
	if err != nil {
		return nil, fmt.Errorf("failed to get period 1 metrics: %w", err)
	}

	points2, gran2, err := s.timescaleDB.GetMetricsByGranularity(ctx, serverID, period2Start, period2End)
	if err != nil {
		return nil, fmt.Errorf("failed to get period 2 metrics: %w", err)
	}

	// Calculate averages for comparison
	avg1 := s.calculateAverages(points1)
	avg2 := s.calculateAverages(points2)

	// Calculate percentage changes
	changes := &MetricChanges{
		CPUChange:     s.calculatePercentageChange(avg1.CPUAvg, avg2.CPUAvg),
		MemoryChange:  s.calculatePercentageChange(avg1.MemoryAvg, avg2.MemoryAvg),
		DiskChange:    s.calculatePercentageChange(avg1.DiskAvg, avg2.DiskAvg),
		NetworkChange: s.calculatePercentageChange(avg1.NetworkAvg, avg2.NetworkAvg),
	}

	return &MetricsComparison{
		ServerID:  serverID,
		Period1:   TimePeriod{Start: period1Start, End: period1End, Granularity: gran1},
		Period2:   TimePeriod{Start: period2Start, End: period2End, Granularity: gran2},
		Averages1: avg1,
		Averages2: avg2,
		Changes:   changes,
		Points1:   points1,
		Points2:   points2,
	}, nil
}

// GetMetricsSummary returns a summary of metrics statistics
func (s *TieredMetricsService) GetMetricsSummary(ctx context.Context) (*MetricsSummary, error) {
	// Get stats for each granularity level
	stats1m, _ := s.timescaleDB.GetMetricsStatsByGranularity(ctx, "1m")
	stats5m, _ := s.timescaleDB.GetMetricsStatsByGranularity(ctx, "5m")
	stats10m, _ := s.timescaleDB.GetMetricsStatsByGranularity(ctx, "10m")
	stats1h, _ := s.timescaleDB.GetMetricsStatsByGranularity(ctx, "1h")

	summary := &MetricsSummary{
		GranularityStats: make(map[string]*timescaledb.MetricsStats),
		TotalDataPoints:  0,
		TotalServers:     0,
		StorageSize:      "0",
		LastUpdated:      time.Now(),
	}

	// Add stats for each granularity
	if stats1m != nil {
		summary.GranularityStats["1m"] = stats1m
		summary.TotalDataPoints += stats1m.TotalRecords
		if stats1m.UniqueServers > summary.TotalServers {
			summary.TotalServers = stats1m.UniqueServers
		}
	}

	if stats5m != nil {
		summary.GranularityStats["5m"] = stats5m
		summary.TotalDataPoints += stats5m.TotalRecords
		if stats5m.UniqueServers > summary.TotalServers {
			summary.TotalServers = stats5m.UniqueServers
		}
	}

	if stats10m != nil {
		summary.GranularityStats["10m"] = stats10m
		summary.TotalDataPoints += stats10m.TotalRecords
		if stats10m.UniqueServers > summary.TotalServers {
			summary.TotalServers = stats10m.UniqueServers
		}
	}

	if stats1h != nil {
		summary.GranularityStats["1h"] = stats1h
		summary.TotalDataPoints += stats1h.TotalRecords
		if stats1h.UniqueServers > summary.TotalServers {
			summary.TotalServers = stats1h.UniqueServers
		}
	}

	return summary, nil
}

// GetMetricsHeatmap returns data for heatmap visualization
func (s *TieredMetricsService) GetMetricsHeatmap(ctx context.Context, serverID string, start, end time.Time) ([]*timescaledb.HeatmapPoint, error) {
	return s.timescaleDB.GetMetricsHeatmap(ctx, serverID, start, end)
}

// Helper types and functions

type DashboardMetrics struct {
	ServerID    string                           `json:"server_id"`
	Current     *timescaledb.TieredMetricsPoint  `json:"current"`
	Granularity timescaledb.MetricsGranularity   `json:"granularity"`
	Points24h   []timescaledb.TieredMetricsPoint `json:"points_24h"`
	Trends      *MetricTrends                    `json:"trends"`
	HeatmapData []*timescaledb.HeatmapPoint      `json:"heatmap_data"`
	LastUpdated time.Time                        `json:"last_updated"`
}

type MetricTrends struct {
	CPUTrend         float64 `json:"cpu_trend"`         // Percentage change over last 24h
	MemoryTrend      float64 `json:"memory_trend"`      // Percentage change over last 24h
	DiskTrend        float64 `json:"disk_trend"`        // Percentage change over last 24h
	NetworkTrend     float64 `json:"network_trend"`     // Percentage change over last 24h
	TemperatureTrend float64 `json:"temperature_trend"` // Percentage change over last 24h
}

type MetricsComparison struct {
	ServerID  string                           `json:"server_id"`
	Period1   TimePeriod                       `json:"period1"`
	Period2   TimePeriod                       `json:"period2"`
	Averages1 *MetricAverages                  `json:"averages1"`
	Averages2 *MetricAverages                  `json:"averages2"`
	Changes   *MetricChanges                   `json:"changes"`
	Points1   []timescaledb.TieredMetricsPoint `json:"points1"`
	Points2   []timescaledb.TieredMetricsPoint `json:"points2"`
}

type TimePeriod struct {
	Start       time.Time                      `json:"start"`
	End         time.Time                      `json:"end"`
	Granularity timescaledb.MetricsGranularity `json:"granularity"`
}

type MetricAverages struct {
	CPUAvg     float64 `json:"cpu_avg"`
	MemoryAvg  float64 `json:"memory_avg"`
	DiskAvg    float64 `json:"disk_avg"`
	NetworkAvg float64 `json:"network_avg"`
	TempAvg    float64 `json:"temp_avg"`
	LoadAvg    float64 `json:"load_avg"`
}

type MetricChanges struct {
	CPUChange     float64 `json:"cpu_change"`
	MemoryChange  float64 `json:"memory_change"`
	DiskChange    float64 `json:"disk_change"`
	NetworkChange float64 `json:"network_change"`
}

type MetricsSummary struct {
	GranularityStats map[string]*timescaledb.MetricsStats `json:"granularity_stats"`
	TotalDataPoints  int64                                `json:"total_data_points"`
	TotalServers     int64                                `json:"total_servers"`
	StorageSize      string                               `json:"storage_size"`
	LastUpdated      time.Time                            `json:"last_updated"`
}

func (s *TieredMetricsService) calculateTrends(points []timescaledb.TieredMetricsPoint) *MetricTrends {
	if len(points) < 2 {
		return &MetricTrends{}
	}

	// Compare first 10% with last 10% of points
	firstQuarter := len(points) / 4
	lastQuarter := len(points) - firstQuarter

	firstAvg := s.calculateAverageSlice(points[:firstQuarter])
	lastAvg := s.calculateAverageSlice(points[lastQuarter:])

	return &MetricTrends{
		CPUTrend:         s.calculatePercentageChange(firstAvg.CPUAvg, lastAvg.CPUAvg),
		MemoryTrend:      s.calculatePercentageChange(firstAvg.MemoryAvg, lastAvg.MemoryAvg),
		DiskTrend:        s.calculatePercentageChange(firstAvg.DiskAvg, lastAvg.DiskAvg),
		NetworkTrend:     s.calculatePercentageChange(firstAvg.NetworkAvg, lastAvg.NetworkAvg),
		TemperatureTrend: s.calculatePercentageChange(firstAvg.TempAvg, lastAvg.TempAvg),
	}
}

func (s *TieredMetricsService) calculateAverageSlice(points []timescaledb.TieredMetricsPoint) *MetricAverages {
	if len(points) == 0 {
		return &MetricAverages{}
	}

	var sumCPU, sumMem, sumDisk, sumNet, sumTemp, sumLoad float64
	count := float64(len(points))

	for _, p := range points {
		sumCPU += p.CPUAvg
		sumMem += p.MemoryAvg
		sumDisk += p.DiskAvg
		sumNet += p.NetworkAvg
		sumTemp += p.TempAvg
		sumLoad += p.LoadAvg
	}

	return &MetricAverages{
		CPUAvg:     sumCPU / count,
		MemoryAvg:  sumMem / count,
		DiskAvg:    sumDisk / count,
		NetworkAvg: sumNet / count,
		TempAvg:    sumTemp / count,
		LoadAvg:    sumLoad / count,
	}
}

func (s *TieredMetricsService) calculateAverages(points []timescaledb.TieredMetricsPoint) *MetricAverages {
	return s.calculateAverageSlice(points)
}

func (s *TieredMetricsService) calculatePercentageChange(old, new float64) float64 {
	if old == 0 {
		return 0
	}
	return ((new - old) / old) * 100
}
