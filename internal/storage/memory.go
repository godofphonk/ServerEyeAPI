package storage

import (
	"context"
	"time"

	"github.com/godofphonk/ServerEyeAPI/pkg/models"
	"github.com/sirupsen/logrus"
)

// MemoryStorage is a simple in-memory storage implementation
type MemoryStorage struct {
	metrics map[string][]*models.Metric
	logger  *logrus.Logger
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage(logger *logrus.Logger) *MemoryStorage {
	return &MemoryStorage{
		metrics: make(map[string][]*models.Metric),
		logger:  logger,
	}
}

func (s *MemoryStorage) StoreMetric(ctx context.Context, metric *models.Metric) error {
	s.metrics[metric.ServerID] = append(s.metrics[metric.ServerID], metric)
	s.logger.WithField("server_id", metric.ServerID).Info("Metric stored in memory")
	return nil
}

func (s *MemoryStorage) GetLatestMetrics(ctx context.Context, serverID string) ([]*models.Metric, error) {
	if metrics, exists := s.metrics[serverID]; exists {
		if len(metrics) > 0 {
			return []*models.Metric{metrics[len(metrics)-1]}, nil
		}
	}
	return []*models.Metric{}, nil
}

func (s *MemoryStorage) GetMetricsHistory(ctx context.Context, serverID string, metricType string, from, to time.Time) ([]*models.Metric, error) {
	if metrics, exists := s.metrics[serverID]; exists {
		var result []*models.Metric
		for _, metric := range metrics {
			if metric.Type == metricType && metric.Timestamp.After(from) && metric.Timestamp.Before(to) {
				result = append(result, metric)
			}
		}
		return result, nil
	}
	return []*models.Metric{}, nil
}

func (s *MemoryStorage) GetServers(ctx context.Context) ([]string, error) {
	var servers []string
	for serverID := range s.metrics {
		servers = append(servers, serverID)
	}
	return servers, nil
}

func (s *MemoryStorage) GetPendingCommands(ctx context.Context, serverID string) ([]string, error) {
	// Return empty commands for demo
	return []string{}, nil
}

func (s *MemoryStorage) StoreDLQMessage(ctx context.Context, topic string, partition int, offset int64, message []byte, errorMsg string) error {
	s.logger.Info("DLQ message stored in memory (not implemented)")
	return nil
}

func (s *MemoryStorage) InsertGeneratedKey(ctx context.Context, secretKey, agentVersion, osInfo, hostname string) error {
	s.logger.Info("Generated key inserted in memory (not implemented)")
	return nil
}

func (s *MemoryStorage) Ping() error {
	return nil
}

func (s *MemoryStorage) Close() error {
	return nil
}
