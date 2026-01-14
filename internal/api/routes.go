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
	serverSourcesHandler *handlers.ServerSourcesHandler,
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

	// Single metrics endpoint (public for testing)
	router.HandleFunc("/api/servers/{server_id}/metrics", metricsHandler.GetServerMetrics).Methods("GET")

	// Server sources endpoints (public for TG bot and web)
	router.HandleFunc("/api/servers/{server_id}/sources", serverSourcesHandler.AddServerSource).Methods("POST")
	router.HandleFunc("/api/servers/{server_id}/sources", serverSourcesHandler.GetServerSources).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/sources/{source}", serverSourcesHandler.RemoveServerSource).Methods("DELETE")

	// API endpoints for Telegram bot and web dashboard (with auth)
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.Auth(storageImpl, logger))

	// Server endpoints (protected)
	api.HandleFunc("/servers", serversHandler.ListServers).Methods("GET")
	api.HandleFunc("/servers/{server_id}/status", serversHandler.GetServerStatus).Methods("GET")
	api.HandleFunc("/servers/{server_id}/command", commandsHandler.SendCommand).Methods("POST")

	return router
}
