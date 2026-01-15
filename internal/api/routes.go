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

	// Metrics endpoint by key (public for TG bot)
	router.HandleFunc("/api/servers/by-key/{server_key}/metrics", metricsHandler.GetServerMetricsByKey).Methods("GET")

	// Server sources endpoints (public for TG bot and web)
	router.HandleFunc("/api/servers/{server_id}/sources", serverSourcesHandler.AddServerSource).Methods("POST")
	router.HandleFunc("/api/servers/{server_id}/sources", serverSourcesHandler.GetServerSources).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/sources/{source}", serverSourcesHandler.RemoveServerSource).Methods("DELETE")

	// Server sources by key endpoints (public for TG bot)
	router.HandleFunc("/api/servers/by-key/{server_key}/sources", serverSourcesHandler.AddServerSourceByKey).Methods("POST")
	router.HandleFunc("/api/servers/by-key/{server_key}/sources", serverSourcesHandler.GetServerSourcesByKey).Methods("GET")
	router.HandleFunc("/api/servers/by-key/{server_key}/sources/{source}", serverSourcesHandler.RemoveServerSourceByKey).Methods("DELETE")

	// API endpoints for Telegram bot and web dashboard (with auth)
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.Auth(storageImpl, logger))

	// Server endpoints (protected)
	api.HandleFunc("/servers", serversHandler.ListServers).Methods("GET")
	api.HandleFunc("/servers/{server_id}/status", serversHandler.GetServerStatus).Methods("GET")
	api.HandleFunc("/servers/{server_id}/command", commandsHandler.SendCommand).Methods("POST")

	return router
}
