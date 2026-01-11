package api

import (
	"github.com/godofphonk/ServerEyeAPI/internal/api/middleware"
	"github.com/godofphonk/ServerEyeAPI/internal/handlers"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/godofphonk/ServerEyeAPI/internal/websocket"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	authHandler *handlers.AuthHandler,
	healthHandler *handlers.HealthHandler,
	metricsHandler *handlers.MetricsHandler,
	serversHandler *handlers.ServersHandler,
	commandsHandler *handlers.CommandsHandler,
	wsServer *websocket.Server,
	storageImpl storage.Storage,
	logger *logrus.Logger,
) *mux.Router {
	router := mux.NewRouter()

	// Public routes (no auth required)
	router.HandleFunc("/RegisterKey", authHandler.RegisterKey).Methods("POST")
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")

	// WebSocket route
	router.HandleFunc("/ws", wsServer.HandleConnection).Methods("GET")

	// API endpoints for Telegram bot and web dashboard
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.Auth(storageImpl, logger))
	api.HandleFunc("/servers", serversHandler.ListServers).Methods("GET")
	api.HandleFunc("/servers/{server_id}/metrics", metricsHandler.GetServerMetrics).Methods("GET")
	api.HandleFunc("/servers/{server_id}/status", serversHandler.GetServerStatus).Methods("GET")
	api.HandleFunc("/servers/{server_id}/command", commandsHandler.SendCommand).Methods("POST")

	return router
}
