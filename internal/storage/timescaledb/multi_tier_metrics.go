package timescaledb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// MetricsGranularity defines the granularity level for metrics
type MetricsGranularity string

const (
	Granularity1Min  MetricsGranularity = "1m"
	Granularity5Min  MetricsGranularity = "5m"
	Granularity10Min MetricsGranularity = "10m"
	Granularity1Hour MetricsGranularity = "1h"
)

// TieredMetricsRequest represents a request for tiered metrics
type TieredMetricsRequest struct {
	ServerID    string             `json:"server_id"`
	StartTime   time.Time          `json:"start_time"`
	EndTime     time.Time          `json:"end_time"`
	Granularity MetricsGranularity `json:"granularity,omitempty"`
	Metrics     []string           `json:"metrics,omitempty"` // Specific metrics to retrieve
}

// TieredMetricsResponse contains tiered metrics with appropriate granularity
type TieredMetricsResponse struct {
	ServerID    string               `json:"server_id"`
	StartTime   time.Time            `json:"start_time"`
	EndTime     time.Time            `json:"end_time"`
	Granularity MetricsGranularity   `json:"granularity"`
	DataPoints  []TieredMetricsPoint `json:"data_points"`
	TotalPoints int64                `json:"total_points"`
	Message     string               `json:"message,omitempty"`
}

// TieredMetricsPoint represents a single data point in tiered metrics
type TieredMetricsPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUAvg      float64   `json:"cpu_avg,omitempty"`
	CPUMax      float64   `json:"cpu_max,omitempty"`
	CPUMin      float64   `json:"cpu_min,omitempty"`
	MemoryAvg   float64   `json:"memory_avg,omitempty"`
	MemoryMax   float64   `json:"memory_max,omitempty"`
	MemoryMin   float64   `json:"memory_min,omitempty"`
	DiskAvg     float64   `json:"disk_avg,omitempty"`
	DiskMax     float64   `json:"disk_max,omitempty"`
	NetworkAvg  float64   `json:"network_avg,omitempty"`
	NetworkMax  float64   `json:"network_max,omitempty"`
	TempAvg     float64   `json:"temp_avg,omitempty"`
	TempMax     float64   `json:"temp_max,omitempty"`
	LoadAvg     float64   `json:"load_avg,omitempty"`
	LoadMax     float64   `json:"load_max,omitempty"`
	SampleCount int64     `json:"sample_count"`
}

// GetTieredMetrics retrieves metrics with appropriate granularity based on time range
func (c *Client) GetTieredMetrics(ctx context.Context, req *TieredMetricsRequest) (*TieredMetricsResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Determine granularity based on time range if not specified
	granularity := req.Granularity
	if granularity == "" {
		granularity = c.determineGranularity(req.StartTime, req.EndTime)
	}

	// Get the appropriate view
	viewName := c.getViewName(granularity)
	if viewName == "" {
		return nil, fmt.Errorf("unsupported granularity: %s", granularity)
	}

	query := fmt.Sprintf(`
		SELECT 
			bucket,
			avg_cpu, max_cpu, min_cpu,
			avg_memory, max_memory, min_memory,
			avg_disk, max_disk,
			avg_network, max_network,
			avg_cpu_temp, max_cpu_temp,
			avg_load_1m, max_load_1m,
			sample_count
		FROM %s
		WHERE server_id = $1 AND bucket BETWEEN $2 AND $3
		ORDER BY bucket`, viewName)

	rows, err := c.pool.Query(ctx, query, req.ServerID, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query tiered metrics: %w", err)
	}
	defer rows.Close()

	var dataPoints []TieredMetricsPoint
	for rows.Next() {
		var point TieredMetricsPoint
		var cpuAvg, cpuMax, cpuMin sql.NullFloat64
		var memAvg, memMax, memMin sql.NullFloat64
		var diskAvg, diskMax sql.NullFloat64
		var netAvg, netMax sql.NullFloat64
		var tempAvg, tempMax sql.NullFloat64
		var loadAvg, loadMax sql.NullFloat64

		err := rows.Scan(
			&point.Timestamp,
			&cpuAvg, &cpuMax, &cpuMin,
			&memAvg, &memMax, &memMin,
			&diskAvg, &diskMax,
			&netAvg, &netMax,
			&tempAvg, &tempMax,
			&loadAvg, &loadMax,
			&point.SampleCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tiered metrics row: %w", err)
		}

		// Convert NullFloat64 to float64
		if cpuAvg.Valid {
			point.CPUAvg = cpuAvg.Float64
		}
		if cpuMax.Valid {
			point.CPUMax = cpuMax.Float64
		}
		if cpuMin.Valid {
			point.CPUMin = cpuMin.Float64
		}
		if memAvg.Valid {
			point.MemoryAvg = memAvg.Float64
		}
		if memMax.Valid {
			point.MemoryMax = memMax.Float64
		}
		if memMin.Valid {
			point.MemoryMin = memMin.Float64
		}
		if diskAvg.Valid {
			point.DiskAvg = diskAvg.Float64
		}
		if diskMax.Valid {
			point.DiskMax = diskMax.Float64
		}
		if netAvg.Valid {
			point.NetworkAvg = netAvg.Float64
		}
		if netMax.Valid {
			point.NetworkMax = netMax.Float64
		}
		if tempAvg.Valid {
			point.TempAvg = tempAvg.Float64
		}
		if tempMax.Valid {
			point.TempMax = tempMax.Float64
		}
		if loadAvg.Valid {
			point.LoadAvg = loadAvg.Float64
		}
		if loadMax.Valid {
			point.LoadMax = loadMax.Float64
		}

		dataPoints = append(dataPoints, point)
	}

	response := &TieredMetricsResponse{
		ServerID:    req.ServerID,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Granularity: granularity,
		DataPoints:  dataPoints,
		TotalPoints: int64(len(dataPoints)),
	}

	// If no data in requested period, try to get available data
	if len(dataPoints) == 0 {
		availableData, err := c.getAvailableMetrics(ctx, req.ServerID, viewName)
		if err == nil && len(availableData) > 0 {
			response.DataPoints = availableData
			response.TotalPoints = int64(len(availableData))
			response.Message = "Showing available data (requested period had no data)"
		}
	}

	return response, nil
}

// GetMetricsByGranularity uses the optimized function to get metrics
func (c *Client) GetMetricsByGranularity(ctx context.Context, serverID string, start, end time.Time) ([]TieredMetricsPoint, MetricsGranularity, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		SELECT 
			bucket, avg_cpu, max_cpu, min_cpu,
			avg_memory, max_memory, min_memory,
			avg_disk, max_disk, avg_network, max_network,
			sample_count, granularity
		FROM get_metrics_by_granularity($1, $2, $3)`

	rows, err := c.pool.Query(ctx, query, serverID, start, end)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query metrics by granularity: %w", err)
	}
	defer rows.Close()

	var points []TieredMetricsPoint
	var granularity MetricsGranularity

	for rows.Next() {
		var point TieredMetricsPoint
		var granularityStr string
		var cpuAvg, cpuMax, cpuMin sql.NullFloat64
		var memAvg, memMax, memMin sql.NullFloat64
		var diskAvg, diskMax sql.NullFloat64
		var netAvg, netMax sql.NullFloat64

		err := rows.Scan(
			&point.Timestamp,
			&cpuAvg, &cpuMax, &cpuMin,
			&memAvg, &memMax, &memMin,
			&diskAvg, &diskMax,
			&netAvg, &netMax,
			&point.SampleCount,
			&granularityStr,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan metrics row: %w", err)
		}

		// Convert values
		if cpuAvg.Valid {
			point.CPUAvg = cpuAvg.Float64
		}
		if cpuMax.Valid {
			point.CPUMax = cpuMax.Float64
		}
		if cpuMin.Valid {
			point.CPUMin = cpuMin.Float64
		}
		if memAvg.Valid {
			point.MemoryAvg = memAvg.Float64
		}
		if memMax.Valid {
			point.MemoryMax = memMax.Float64
		}
		if memMin.Valid {
			point.MemoryMin = memMin.Float64
		}
		if diskAvg.Valid {
			point.DiskAvg = diskAvg.Float64
		}
		if diskMax.Valid {
			point.DiskMax = diskMax.Float64
		}
		if netAvg.Valid {
			point.NetworkAvg = netAvg.Float64
		}
		if netMax.Valid {
			point.NetworkMax = netMax.Float64
		}

		points = append(points, point)
		granularity = MetricsGranularity(granularityStr)
	}

	return points, granularity, nil
}

// GetCurrentSystemStatus gets the latest system status using 1-minute data
func (c *Client) GetCurrentSystemStatus(ctx context.Context, serverID string) (*TieredMetricsPoint, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		SELECT 
			bucket,
			avg_cpu, max_cpu,
			avg_memory, max_memory,
			avg_disk,
			avg_cpu_temp, max_cpu_temp,
			avg_load_1m, max_load_1m,
			sample_count
		FROM current_system_status
		WHERE server_id = $1
		ORDER BY bucket DESC
		LIMIT 1`

	var point TieredMetricsPoint
	var cpuAvg, cpuMax sql.NullFloat64
	var memAvg, memMax sql.NullFloat64
	var diskAvg sql.NullFloat64
	var tempAvg, tempMax sql.NullFloat64
	var loadAvg, loadMax sql.NullFloat64

	err := c.pool.QueryRow(ctx, query, serverID).Scan(
		&point.Timestamp,
		&cpuAvg, &cpuMax,
		&memAvg, &memMax,
		&diskAvg,
		&tempAvg, &tempMax,
		&loadAvg, &loadMax,
		&point.SampleCount,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no recent data found for server %s", serverID)
		}
		return nil, fmt.Errorf("failed to get current status: %w", err)
	}

	// Convert NullFloat64 to float64
	if cpuAvg.Valid {
		point.CPUAvg = cpuAvg.Float64
	}
	if cpuMax.Valid {
		point.CPUMax = cpuMax.Float64
	}
	if memAvg.Valid {
		point.MemoryAvg = memAvg.Float64
	}
	if memMax.Valid {
		point.MemoryMax = memMax.Float64
	}
	if diskAvg.Valid {
		point.DiskAvg = diskAvg.Float64
	}
	if tempAvg.Valid {
		point.TempAvg = tempAvg.Float64
	}
	if tempMax.Valid {
		point.TempMax = tempMax.Float64
	}
	if loadAvg.Valid {
		point.LoadAvg = loadAvg.Float64
	}
	if loadMax.Valid {
		point.LoadMax = loadMax.Float64
	}

	return &point, nil
}

// GetMetricsHeatmap returns data for heatmap visualization
func (c *Client) GetMetricsHeatmap(ctx context.Context, serverID string, start, end time.Time) ([]*HeatmapPoint, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	granularity := c.determineGranularity(start, end)
	viewName := c.getViewName(granularity)

	query := fmt.Sprintf(`
		SELECT 
			bucket,
			avg_cpu,
			avg_memory,
			avg_disk,
			max_cpu,
			max_memory,
			max_disk,
			sample_count
		FROM %s
		WHERE server_id = $1 AND bucket BETWEEN $2 AND $3
		ORDER BY bucket`, viewName)

	rows, err := c.pool.Query(ctx, query, serverID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query heatmap data: %w", err)
	}
	defer rows.Close()

	var points []*HeatmapPoint
	for rows.Next() {
		var point HeatmapPoint
		err := rows.Scan(
			&point.Timestamp,
			&point.CPUAvg,
			&point.MemoryAvg,
			&point.DiskAvg,
			&point.CPUMax,
			&point.MemoryMax,
			&point.DiskMax,
			&point.SampleCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan heatmap row: %w", err)
		}
		points = append(points, &point)
	}

	return points, nil
}

// HeatmapPoint represents a data point for heatmap visualization
type HeatmapPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUAvg      float64   `json:"cpu_avg"`
	MemoryAvg   float64   `json:"memory_avg"`
	DiskAvg     float64   `json:"disk_avg"`
	CPUMax      float64   `json:"cpu_max"`
	MemoryMax   float64   `json:"memory_max"`
	DiskMax     float64   `json:"disk_max"`
	SampleCount int64     `json:"sample_count"`
}

// Helper functions

// GetMetricsStatsByGranularity returns statistics for a specific granularity
func (c *Client) GetMetricsStatsByGranularity(ctx context.Context, granularity string) (*MetricsStats, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	viewName := "metrics_" + granularity + "_avg"
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_records,
			COUNT(DISTINCT server_id) as unique_servers,
			MIN(bucket) as earliest_record,
			MAX(bucket) as latest_record,
			pg_size_pretty(pg_total_relation_size('%s')) as table_size
		FROM %s`, viewName, viewName)

	var stat MetricsStats
	var earliestRecord, latestRecord sql.NullTime
	var tableSize string

	err := c.pool.QueryRow(ctx, query).Scan(
		&stat.TotalRecords,
		&stat.UniqueServers,
		&earliestRecord,
		&latestRecord,
		&tableSize,
	)

	if err != nil {
		return nil, err
	}

	stat.TableSize = tableSize
	if earliestRecord.Valid {
		stat.EarliestRecord = earliestRecord.Time
	}
	if latestRecord.Valid {
		stat.LatestRecord = latestRecord.Time
	}

	return &stat, nil
}

func (c *Client) determineGranularity(start, end time.Time) MetricsGranularity {
	duration := end.Sub(start)

	if duration <= time.Hour {
		return Granularity1Min
	} else if duration <= 3*time.Hour {
		return Granularity5Min
	} else if duration <= 24*time.Hour {
		return Granularity10Min
	} else {
		return Granularity1Hour
	}
}

func (c *Client) getViewName(granularity MetricsGranularity) string {
	switch granularity {
	case Granularity1Min:
		return "metrics_1m_avg"
	case Granularity5Min:
		return "metrics_5m_avg"
	case Granularity10Min:
		return "metrics_10m_avg"
	case Granularity1Hour:
		return "metrics_1h_avg"
	default:
		return ""
	}
}

// getAvailableMetrics retrieves the latest available metrics when requested period has no data
func (c *Client) getAvailableMetrics(ctx context.Context, serverID string, viewName string) ([]TieredMetricsPoint, error) {
	query := fmt.Sprintf(`
		SELECT 
			bucket,
			avg_cpu, max_cpu, min_cpu,
			avg_memory, max_memory, min_memory,
			avg_disk, max_disk,
			avg_network, max_network,
			avg_cpu_temp, max_cpu_temp,
			avg_load_1m, max_load_1m,
			sample_count
		FROM %s
		WHERE server_id = $1
		ORDER BY bucket DESC
		LIMIT 100`, viewName)

	rows, err := c.pool.Query(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query available metrics: %w", err)
	}
	defer rows.Close()

	var dataPoints []TieredMetricsPoint
	for rows.Next() {
		var point TieredMetricsPoint
		var cpuAvg, cpuMax, cpuMin sql.NullFloat64
		var memAvg, memMax, memMin sql.NullFloat64
		var diskAvg, diskMax sql.NullFloat64
		var netAvg, netMax sql.NullFloat64
		var tempAvg, tempMax sql.NullFloat64
		var loadAvg, loadMax sql.NullFloat64

		err := rows.Scan(
			&point.Timestamp,
			&cpuAvg, &cpuMax, &cpuMin,
			&memAvg, &memMax, &memMin,
			&diskAvg, &diskMax,
			&netAvg, &netMax,
			&tempAvg, &tempMax,
			&loadAvg, &loadMax,
			&point.SampleCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan available metrics row: %w", err)
		}

		if cpuAvg.Valid {
			point.CPUAvg = cpuAvg.Float64
		}
		if cpuMax.Valid {
			point.CPUMax = cpuMax.Float64
		}
		if cpuMin.Valid {
			point.CPUMin = cpuMin.Float64
		}
		if memAvg.Valid {
			point.MemoryAvg = memAvg.Float64
		}
		if memMax.Valid {
			point.MemoryMax = memMax.Float64
		}
		if memMin.Valid {
			point.MemoryMin = memMin.Float64
		}
		if diskAvg.Valid {
			point.DiskAvg = diskAvg.Float64
		}
		if diskMax.Valid {
			point.DiskMax = diskMax.Float64
		}
		if netAvg.Valid {
			point.NetworkAvg = netAvg.Float64
		}
		if netMax.Valid {
			point.NetworkMax = netMax.Float64
		}
		if tempAvg.Valid {
			point.TempAvg = tempAvg.Float64
		}
		if tempMax.Valid {
			point.TempMax = tempMax.Float64
		}
		if loadAvg.Valid {
			point.LoadAvg = loadAvg.Float64
		}
		if loadMax.Valid {
			point.LoadMax = loadMax.Float64
		}

		dataPoints = append(dataPoints, point)
	}

	return dataPoints, nil
}
