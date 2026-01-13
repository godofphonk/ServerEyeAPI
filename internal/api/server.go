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
	"github.com/godofphonk/ServerEyeAPI/internal/storage/redis"
	redisStorage "github.com/godofphonk/ServerEyeAPI/internal/storage/redis"
	postgresRepo "github.com/godofphonk/ServerEyeAPI/internal/storage/repositories/postgres"
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

	// Initialize Redis if URL is provided
	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		// Extract host:port from redis://host:port format
		redisAddr := "redis:6379" // Default for Docker Compose
		if cfg.RedisURL != "" {
			// Parse URL to get host:port
			if len(cfg.RedisURL) > 9 && cfg.RedisURL[:9] == "redis://" {
				redisAddr = cfg.RedisURL[9:] // Remove "redis://" prefix
			}
		}
		var err error
		redisClient, err = redisStorage.NewClient(redisAddr, "", 0, logger, cfg)
		if err != nil {
			return nil, err
		}
	}

	// Create storage adapter with Redis support
	storageImpl = storage.NewStorageAdapterWithRedis(keyRepo, serverRepo, redisClient)

	// Initialize WebSocket server
	wsServer := websocket.NewServer(storageImpl, logger, cfg)

	// Initialize services with repositories
	authService := services.NewAuthService(keyRepo, serverRepo, logger)
	metricsService := services.NewMetricsService(keyRepo, logger)
	commandsService := services.NewCommandsService(keyRepo, logger)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, logger)
	healthHandler := handlers.NewHealthHandler(storageImpl, logger)
	metricsHandler := handlers.NewMetricsHandler(metricsService, logger)
	serversHandler := handlers.NewServersHandler(storageImpl, logger)
	commandsHandler := handlers.NewCommandsHandler(commandsService, logger)

	// Setup routes
	router := SetupRoutes(
		authHandler,
		healthHandler,
		metricsHandler,
		serversHandler,
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
