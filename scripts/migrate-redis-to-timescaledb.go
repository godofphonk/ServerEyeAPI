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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/timescaledb"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type MigrationConfig struct {
	RedisURL       string
	TimescaleDBURL string
	DryRun         bool
	BatchSize      int
	LogLevel       string
}

type RedisMigrator struct {
	redisClient *redis.Client
	timescaleDB *timescaledb.Client
	logger      *logrus.Logger
	config      *MigrationConfig
}

func main() {
	config := &MigrationConfig{}

	flag.StringVar(&config.RedisURL, "redis-url", "redis://localhost:6379", "Redis connection URL")
	flag.StringVar(&config.TimescaleDBURL, "timescaledb-url", "postgres://postgres:password@localhost:5432/servereye?sslmode=disable", "TimescaleDB connection URL")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Run migration without actually migrating data")
	flag.IntVar(&config.BatchSize, "batch-size", 100, "Batch size for migration")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Setup logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	logger.SetLevel(level)

	// Create migrator
	migrator, err := NewRedisMigrator(config, logger)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer migrator.Close()

	// Run migration
	if err := migrator.MigrateAll(context.Background()); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	logger.Info("Migration completed successfully!")
}

func NewRedisMigrator(config *MigrationConfig, logger *logrus.Logger) (*RedisMigrator, error) {
	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     strings.TrimPrefix(config.RedisURL, "redis://"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Connect to TimescaleDB
	tsConfig := timescaledb.DefaultClientConfig()
	timescaleDB, err := timescaledb.NewClient(config.TimescaleDBURL, logger, tsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TimescaleDB: %w", err)
	}

	return &RedisMigrator{
		redisClient: redisClient,
		timescaleDB: timescaleDB,
		logger:      logger,
		config:      config,
	}, nil
}

func (m *RedisMigrator) Close() error {
	if m.redisClient != nil {
		return m.redisClient.Close()
	}
	if m.timescaleDB != nil {
		return m.timescaleDB.Close()
	}
	return nil
}

func (m *RedisMigrator) MigrateAll(ctx context.Context) error {
	m.logger.Info("Starting migration from Redis to TimescaleDB")

	// Migrate metrics
	if err := m.MigrateMetrics(ctx); err != nil {
		return fmt.Errorf("failed to migrate metrics: %w", err)
	}

	// Migrate status
	if err := m.MigrateStatus(ctx); err != nil {
		return fmt.Errorf("failed to migrate status: %w", err)
	}

	// Migrate commands
	if err := m.MigrateCommands(ctx); err != nil {
		return fmt.Errorf("failed to migrate commands: %w", err)
	}

	m.logger.Info("All data migrated successfully")
	return nil
}

func (m *RedisMigrator) MigrateMetrics(ctx context.Context) error {
	m.logger.Info("Migrating metrics...")

	// Get all metric keys
	keys, err := m.redisClient.Keys(ctx, "metrics:*").Result()
	if err != nil {
		return fmt.Errorf("failed to get metric keys: %w", err)
	}

	m.logger.WithField("count", len(keys)).Info("Found metric keys")

	migrated := 0
	failed := 0

	for _, key := range keys {
		serverID := strings.TrimPrefix(key, "metrics:")

		// Get metric data from Redis
		data, err := m.redisClient.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				m.logger.WithField("key", key).Warn("Metric key not found")
				continue
			}
			m.logger.WithError(err).WithField("key", key).Error("Failed to get metric from Redis")
			failed++
			continue
		}

		// Parse metrics
		var metrics models.ServerMetrics
		if err := json.Unmarshal([]byte(data), &metrics); err != nil {
			m.logger.WithError(err).WithField("key", key).Error("Failed to unmarshal metrics")
			failed++
			continue
		}

		// Migrate to TimescaleDB
		if !m.config.DryRun {
			if err := m.timescaleDB.StoreMetric(ctx, serverID, &metrics); err != nil {
				m.logger.WithError(err).WithField("server_id", serverID).Error("Failed to store metrics in TimescaleDB")
				failed++
				continue
			}
		}

		migrated++
		if migrated%100 == 0 {
			m.logger.WithFields(logrus.Fields{
				"migrated": migrated,
				"failed":   failed,
				"total":    len(keys),
			}).Info("Migration progress")
		}
	}

	m.logger.WithFields(logrus.Fields{
		"migrated": migrated,
		"failed":   failed,
		"total":    len(keys),
	}).Info("Metrics migration completed")

	return nil
}

func (m *RedisMigrator) MigrateStatus(ctx context.Context) error {
	m.logger.Info("Migrating status...")

	// Get all status keys
	keys, err := m.redisClient.Keys(ctx, "status:*").Result()
	if err != nil {
		return fmt.Errorf("failed to get status keys: %w", err)
	}

	m.logger.WithField("count", len(keys)).Info("Found status keys")

	migrated := 0
	failed := 0

	for _, key := range keys {
		serverID := strings.TrimPrefix(key, "status:")

		// Get status data from Redis
		data, err := m.redisClient.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				m.logger.WithField("key", key).Warn("Status key not found")
				continue
			}
			m.logger.WithError(err).WithField("key", key).Error("Failed to get status from Redis")
			failed++
			continue
		}

		// Parse status
		var status models.ServerStatus
		if err := json.Unmarshal([]byte(data), &status); err != nil {
			m.logger.WithError(err).WithField("key", key).Error("Failed to unmarshal status")
			failed++
			continue
		}

		// Migrate to TimescaleDB
		if !m.config.DryRun {
			if err := m.timescaleDB.SetServerStatus(ctx, serverID, &status); err != nil {
				m.logger.WithError(err).WithField("server_id", serverID).Error("Failed to store status in TimescaleDB")
				failed++
				continue
			}
		}

		migrated++
	}

	m.logger.WithFields(logrus.Fields{
		"migrated": migrated,
		"failed":   failed,
		"total":    len(keys),
	}).Info("Status migration completed")

	return nil
}

func (m *RedisMigrator) MigrateCommands(ctx context.Context) error {
	m.logger.Info("Migrating commands...")

	// Get all command keys
	keys, err := m.redisClient.Keys(ctx, "commands:*").Result()
	if err != nil {
		return fmt.Errorf("failed to get command keys: %w", err)
	}

	m.logger.WithField("count", len(keys)).Info("Found command keys")

	migrated := 0
	failed := 0

	for _, key := range keys {
		serverID := strings.TrimPrefix(key, "commands:")

		// Get command data from Redis
		data, err := m.redisClient.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				m.logger.WithField("key", key).Warn("Command key not found")
				continue
			}
			m.logger.WithError(err).WithField("key", key).Error("Failed to get command from Redis")
			failed++
			continue
		}

		// Parse commands (expecting array)
		var commands []map[string]interface{}
		if err := json.Unmarshal([]byte(data), &commands); err != nil {
			m.logger.WithError(err).WithField("key", key).Error("Failed to unmarshal commands")
			failed++
			continue
		}

		// Migrate each command
		for _, cmd := range commands {
			if !m.config.DryRun {
				if err := m.timescaleDB.StoreCommand(ctx, serverID, cmd); err != nil {
					m.logger.WithError(err).WithField("server_id", serverID).Error("Failed to store command in TimescaleDB")
					failed++
					continue
				}
			}
			migrated++
		}
	}

	m.logger.WithFields(logrus.Fields{
		"migrated": migrated,
		"failed":   failed,
		"total":    len(keys),
	}).Info("Commands migration completed")

	return nil
}
