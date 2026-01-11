package api

import (
	"context"
	"net/http"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config       *config.Config
	storage      storage.Storage
	redisStorage *storage.RedisStorage
	logger       *logrus.Logger
	server       *http.Server
	wsServer     *WebSocketServer
}

func New(cfg *config.Config, storage storage.Storage, redisStorage *storage.RedisStorage, logger *logrus.Logger) *Server {
	s := &Server{
		config:       cfg,
		storage:      storage,
		redisStorage: redisStorage,
		logger:       logger,
	}

	// Initialize WebSocket server
	s.wsServer = NewWebSocketServer(s, logger)

	// Initialize rate limiter (100 requests per minute)
	rateLimiter := NewRateLimiter(100, 20, logger)

	router := s.setupRoutes()

	// Apply global middleware (not auth)
	router.Use(s.loggingMiddleware)
	router.Use(rateLimiter.Middleware)
	router.Use(s.corsMiddleware)

	s.server = &http.Server{
		Addr:         cfg.GetAddr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// GetWebSocketServer returns the WebSocket server instance
func (s *Server) GetWebSocketServer() *WebSocketServer {
	return s.wsServer
}

func (s *Server) setupRoutes() *mux.Router {
	router := mux.NewRouter()

	s.logger.Info("Setting up routes...")

	// Public routes (no auth required)
	router.HandleFunc("/RegisterKey", s.handleRegisterKey).Methods("POST")
	router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// API endpoints for Telegram bot and web dashboard
	router.HandleFunc("/api/servers", s.handleListServers).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/metrics", s.handleGetServerMetrics).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/status", s.handleGetServerStatus).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/command", s.handleSendCommand).Methods("POST")

	s.logger.Info("Registered routes:")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		s.logger.Infof("Route: %s %s", methods, path)
		return nil
	})

	// Apply global middleware (except for WebSocket)
	router.Use(s.loggingMiddleware)
	router.Use(s.corsMiddleware)

	// Handle WebSocket separately without middleware
	router.Path("/ws").HandlerFunc(s.wsServer.handleWebSocket)

	return router
}

func (s *Server) Start() error {
	s.logger.WithField("addr", s.server.Addr).Info("Starting API server")
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down API server")
	return s.server.Shutdown(ctx)
}
