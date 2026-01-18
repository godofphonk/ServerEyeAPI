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

package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/api/middleware"
	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/handlers"
	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/interfaces"
	postgresStorage "github.com/godofphonk/ServerEyeAPI/internal/storage/postgres"
	postgresRepo "github.com/godofphonk/ServerEyeAPI/internal/storage/repositories/postgres"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/timescaledb"
	"github.com/godofphonk/ServerEyeAPI/internal/version"
	"github.com/godofphonk/ServerEyeAPI/internal/websocket"
	"github.com/sirupsen/logrus"
)

// Server represents the HTTP server
type Server struct {
	server   *http.Server
	logger   *logrus.Logger
	wsServer *websocket.Server
	storage  storage.Storage
}

// New creates a new server instance
func New(cfg *config.Config, logger *logrus.Logger) (*Server, error) {
	// Initialize repositories
	var keyRepo interfaces.GeneratedKeyRepository
	var serverRepo interfaces.ServerRepository
	var storageImpl storage.Storage

	if cfg.DatabaseURL == "" {
		logger.Info("DATABASE_URL not set - cannot start without database")
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// Initialize PostgreSQL
	pgClient, err := postgresStorage.NewClient(cfg.DatabaseURL, logger)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	keyRepo = postgresRepo.NewGeneratedKeyRepository(pgClient.DB(), logger)
	serverRepo = postgresRepo.NewServerRepository(pgClient.DB(), logger)

	// Initialize TimescaleDB client
	var timescaleDBClient *timescaledb.Client
	if cfg.TimescaleDBURL != "" {
		tsConfig := timescaledb.DefaultClientConfig()
		var err error
		timescaleDBClient, err = timescaledb.NewClient(cfg.TimescaleDBURL, logger, tsConfig)
		if err != nil {
			return nil, err
		}
	}

	// Create storage adapter with TimescaleDB
	storageImpl = storage.NewTimescaleDBStorageAdapter(keyRepo, serverRepo, timescaleDBClient, logger, cfg)

	// Initialize WebSocket server
	wsServer := websocket.NewServer(storageImpl, logger, cfg)

	// Initialize services with repositories
	authService := services.NewAuthService(keyRepo, serverRepo, logger)
	serverService := services.NewServerService(serverRepo, keyRepo, logger)
	metricsService := services.NewMetricsService(keyRepo, storageImpl, logger)
	commandsService := services.NewCommandsService(keyRepo, logger)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, logger)
	healthHandler := handlers.NewHealthHandler(storageImpl, logger)
	metricsHandler := handlers.NewMetricsHandler(metricsService, logger)
	serversHandler := handlers.NewServersHandler(storageImpl, logger)
	serverSourcesHandler := handlers.NewServerSourcesHandler(serverService, logger)
	commandsHandler := handlers.NewCommandsHandler(commandsService, logger)

	// Setup routes
	router := SetupRoutes(
		authHandler,
		healthHandler,
		metricsHandler,
		serversHandler,
		serverSourcesHandler,
		commandsHandler,
		wsServer,
		storageImpl,
		logger,
	)

	// Apply middleware
	router.Use(middleware.Logging(logger))
	router.Use(middleware.CORS)

	// Apply rate limiting
	rateLimiter := middleware.NewRateLimiter(cfg, logger)
	router.Use(rateLimiter.RateLimit)

	server := &http.Server{
		Addr:         cfg.GetAddr(),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		server:   server,
		logger:   logger,
		wsServer: wsServer,
		storage:  storageImpl,
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.WithFields(logrus.Fields{
		"addr":    s.server.Addr,
		"version": version.Version,
	}).Info("Starting API server")

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server")

	// 1. Shutdown HTTP server first (stops accepting new connections)
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to shutdown HTTP server")
	}

	// 2. Close WebSocket connections
	if err := s.wsServer.Close(); err != nil {
		s.logger.WithError(err).Error("Failed to close WebSocket server")
	}

	// 3. Close storage last
	if err := s.storage.Close(); err != nil {
		s.logger.WithError(err).Error("Failed to close storage")
	}

	s.logger.Info("Server shutdown complete")
	return nil
}

// GetWebSocketServer returns the WebSocket server instance
func (s *Server) GetWebSocketServer() *websocket.Server {
	return s.wsServer
}
