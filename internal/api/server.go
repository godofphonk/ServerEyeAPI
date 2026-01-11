package api

import (
	"context"
	"net/http"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/api/middleware"
	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/handlers"
	"github.com/godofphonk/ServerEyeAPI/internal/services"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/godofphonk/ServerEyeAPI/internal/storage/memory"
	postgresStorage "github.com/godofphonk/ServerEyeAPI/internal/storage/postgres"
	redisStorage "github.com/godofphonk/ServerEyeAPI/internal/storage/redis"
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
	// Initialize storage
	var storageImpl storage.Storage
	var redisClient *redisStorage.Client

	if cfg.DatabaseURL == "" {
		logger.Info("Using in-memory storage (DATABASE_URL not set)")
		storageImpl = memory.NewStorage(logger)
	} else {
		// Initialize PostgreSQL
		pgClient, err := postgresStorage.NewClient(cfg.DatabaseURL, logger)
		if err != nil {
			return nil, err
		}

		// Initialize Redis if URL is provided
		if cfg.RedisURL != "" {
			redisClient, err = redisStorage.NewClient(cfg.RedisURL, "", 0, logger)
			if err != nil {
				return nil, err
			}
		}

		// Use combined storage
		storageImpl = storage.NewCombinedStorage(pgClient, redisClient)
	}

	// Initialize WebSocket server
	wsServer := websocket.NewServer(storageImpl, logger)

	// Initialize services
	authService := services.NewAuthService(storageImpl, logger)
	metricsService := services.NewMetricsService(storageImpl, logger)
	commandsService := services.NewCommandsService(storageImpl, logger)

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
	rateLimiter := middleware.NewRateLimiter(100, time.Minute, logger)
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

	// Close storage
	if err := s.storage.Close(); err != nil {
		s.logger.WithError(err).Error("Failed to close storage")
	}

	return s.server.Shutdown(ctx)
}

// GetWebSocketServer returns the WebSocket server instance
func (s *Server) GetWebSocketServer() *websocket.Server {
	return s.wsServer
}
