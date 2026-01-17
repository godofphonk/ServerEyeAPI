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

//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/godofphonk/ServerEyeAPI/internal/api"
	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/handlers"
	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	postgresStorage "github.com/godofphonk/ServerEyeAPI/internal/storage/postgres"
	redisStorage "github.com/godofphonk/ServerEyeAPI/internal/storage/redis"
	postgresRepo "github.com/godofphonk/ServerEyeAPI/internal/storage/repositories/postgres"
	"github.com/godofphonk/ServerEyeAPI/internal/websocket"
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
)

// ProviderSet contains all the providers for dependency injection
var ProviderSet = wire.NewSet(
	// Core dependencies
	NewLogger,

	// Storage layer
	NewPostgresClient,
	NewTimescaleDBClient,
	NewRedisClient,
	NewTimescaleDBStorageAdapter,

	// Repository layer
	postgresRepo.NewGeneratedKeyRepository,
	postgresRepo.NewServerRepository,

	// Service layer
	services.NewServerService,
	services.NewMetricsService,
	services.NewCommandsService,
	services.NewAuthService,

	// WebSocket
	websocket.NewServer,

	// Handlers
	handlers.NewAuthHandler,
	handlers.NewHealthHandler,
	handlers.NewMetricsHandler,
	handlers.NewServersHandler,
	handlers.NewServerSourcesHandler,
	handlers.NewCommandsHandler,

	// API Server
	api.New,
)

// InitializeApp creates a new application with all dependencies injected
func InitializeApp(cfg *config.Config) (*api.Server, error) {
	wire.Build(ProviderSet)
	return nil, nil
}

// NewLogger creates a new logger instance
func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	return logger
}

// NewPostgresClient creates a new PostgreSQL client
func NewPostgresClient(cfg *config.Config, logger *logrus.Logger) (*postgresStorage.Client, error) {
	return postgresStorage.NewClient(cfg.DatabaseURL, logger)
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.Config, logger *logrus.Logger) (*redisStorage.Client, error) {
	if cfg.RedisURL == "" {
		return nil, nil // Redis is optional
	}

	// Extract host:port from redis://host:port format
	redisAddr := "redis:6379" // Default for Docker Compose
	if len(cfg.RedisURL) > 9 && cfg.RedisURL[:9] == "redis://" {
		redisAddr = cfg.RedisURL[9:] // Remove "redis://" prefix
	}

	return redisStorage.NewClient(redisAddr, "", 0, logger, cfg)
}

// NewStorageAdapter creates a storage adapter for backward compatibility
func NewStorageAdapter(
	keyRepo interfaces.GeneratedKeyRepository,
	serverRepo interfaces.ServerRepository,
) storage.Storage {
	return storage.NewStorageAdapter(keyRepo, serverRepo)
}

// NewTimescaleDBClient creates a new TimescaleDB client
func NewTimescaleDBClient(cfg *config.Config, logger *logrus.Logger) (*timescaledbStorage.Client, error) {
	if cfg.TimescaleDBURL == "" {
		return nil, nil // TimescaleDB is optional for backward compatibility
	}

	config := timescaledbStorage.DefaultClientConfig()
	return timescaledbStorage.NewClient(cfg.TimescaleDBURL, logger, config)
}

// NewTimescaleDBStorageAdapter creates a TimescaleDB storage adapter
func NewTimescaleDBStorageAdapter(
	keyRepo interfaces.GeneratedKeyRepository,
	serverRepo interfaces.ServerRepository,
	timescaleDB *timescaledbStorage.Client,
	logger *logrus.Logger,
	cfg *config.Config,
) storage.Storage {
	if timescaleDB != nil {
		return storage.NewTimescaleDBStorageAdapter(keyRepo, serverRepo, timescaleDB, logger, cfg)
	}
	// Fallback to old adapter if TimescaleDB is not available
	return storage.NewStorageAdapter(keyRepo, serverRepo)
}
