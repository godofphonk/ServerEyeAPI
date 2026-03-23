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
	tieredMetricsHandler *handlers.TieredMetricsHandler,
	unifiedServerHandler *handlers.UnifiedServerHandler,
	serversHandler *handlers.ServersHandler,
	serverSourcesHandler *handlers.ServerSourcesHandler,
	commandsHandler *handlers.CommandsHandler,
	apiKeyHandler *handlers.APIKeyHandler,
	staticInfoHandler *handlers.StaticInfoHandler,
	metricsPushHandler *handlers.MetricsPushHandler,
	serverMetricsHandler *handlers.ServerMetricsHandler,
	alertHandler *handlers.AlertHandler,
	wsServer *websocket.Server,
	apiKeyMiddleware interface{},
	storageImpl storage.Storage,
	logger *logrus.Logger,
) *mux.Router {
	router := mux.NewRouter()

	// Public routes (no auth required)
	router.HandleFunc("/RegisterKey", authHandler.RegisterKey).Methods("POST")
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")

	// HTTP endpoints for agent metrics push (replacing WebSocket)
	router.HandleFunc("/api/servers/by-key/{server_key}/metrics", metricsPushHandler.PushMetrics).Methods("POST")
	router.HandleFunc("/api/servers/by-key/{server_key}/heartbeat", metricsPushHandler.PushHeartbeat).Methods("POST")
	router.HandleFunc("/api/servers/{server_id}/metrics", metricsPushHandler.PushMetricsByID).Methods("POST")
	router.HandleFunc("/api/servers/{server_id}/heartbeat", metricsPushHandler.PushHeartbeatByID).Methods("POST")

	// Metrics endpoint by key (public for TG bot)
	router.HandleFunc("/api/servers/by-key/{server_key}/metrics", metricsHandler.GetServerMetricsByKey).Methods("GET")

	// Server status endpoints (public)
	router.HandleFunc("/api/servers/{server_id}/status", metricsHandler.GetServerStatus).Methods("GET")
	router.HandleFunc("/api/servers/by-key/{server_key}/status", metricsHandler.GetServerStatusByKey).Methods("GET")

	// Server sources endpoints (public for TG bot and web)
	router.HandleFunc("/api/servers/{server_id}/sources", serverSourcesHandler.AddServerSource).Methods("POST")
	router.HandleFunc("/api/servers/{server_id}/sources", serverSourcesHandler.GetServerSources).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/sources/{source}", serverSourcesHandler.RemoveServerSource).Methods("DELETE")

	// Server sources by key endpoints (public for TG bot)
	router.HandleFunc("/api/servers/by-key/{server_key}/sources", serverSourcesHandler.AddServerSourceByKey).Methods("POST")
	router.HandleFunc("/api/servers/by-key/{server_key}/sources", serverSourcesHandler.GetServerSourcesByKey).Methods("GET")
	router.HandleFunc("/api/servers/by-key/{server_key}/sources/{source}", serverSourcesHandler.RemoveServerSourceByKey).Methods("DELETE")
	router.HandleFunc("/api/servers/by-key/{server_key}/sources/{source_type}/identifiers", serverSourcesHandler.RemoveServerSourceIdentifiersByKey).Methods("DELETE")

	// Server source identifiers endpoints (public for TG bot and web)
	router.HandleFunc("/api/servers/{server_id}/sources/identifiers", serverSourcesHandler.AddServerSourceIdentifiers).Methods("POST")
	router.HandleFunc("/api/servers/{server_id}/sources/identifiers", serverSourcesHandler.GetServerSourceIdentifiers).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/sources/{source_type}/identifiers", serverSourcesHandler.RemoveServerSourceIdentifiers).Methods("DELETE")
	router.HandleFunc("/api/servers/{server_id}/sources/{source_type}/identifiers/{identifier}/telegram-id", serverSourcesHandler.UpdateTelegramID).Methods("PUT")

	// Server source identifiers by key endpoints (public for TG bot)
	router.HandleFunc("/api/servers/by-key/{server_key}/sources/identifiers", serverSourcesHandler.AddServerSourceIdentifiersByKey).Methods("POST")
	router.HandleFunc("/api/servers/by-key/{server_key}/sources/identifiers", serverSourcesHandler.GetServerSourceIdentifiersByKey).Methods("GET")

	// Get servers by Telegram ID (public for TG bot)
	router.HandleFunc("/api/servers/by-telegram/{telegramId}", serverSourcesHandler.GetServersByTelegramID).Methods("GET")

	// API Key management routes (admin only) - TODO: Add middleware protection
	router.HandleFunc("/api/admin/keys", apiKeyHandler.CreateAPIKey).Methods("POST")
	router.HandleFunc("/api/admin/keys", apiKeyHandler.ListAPIKeys).Methods("GET")
	router.HandleFunc("/api/admin/keys/{keyId}", apiKeyHandler.GetAPIKey).Methods("GET")
	router.HandleFunc("/api/admin/keys/{keyId}", apiKeyHandler.RevokeAPIKey).Methods("DELETE")

	// Unified metrics endpoint (public)
	router.HandleFunc("/api/servers/{server_id}/metrics/tiered", tieredMetricsHandler.GetMetrics).Methods("GET")
	router.HandleFunc("/api/servers/by-key/{server_key}/metrics/tiered", tieredMetricsHandler.GetMetricsByKey).Methods("GET")

	// Unified server data endpoint (public) - combines metrics, status, and static info
	router.HandleFunc("/api/servers/by-key/{server_key}/unified", unifiedServerHandler.GetUnifiedServerData).Methods("GET")

	// Static server information endpoints (public)
	router.HandleFunc("/api/servers/{server_id}/static-info", staticInfoHandler.UpsertStaticInfo).Methods("POST", "PUT")
	router.HandleFunc("/api/servers/{server_id}/static-info", staticInfoHandler.GetStaticInfo).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/static-info/server", staticInfoHandler.GetServerInfo).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/static-info/hardware", staticInfoHandler.GetHardwareInfo).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/static-info/network", staticInfoHandler.GetNetworkInterfaces).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/static-info/disks", staticInfoHandler.GetDiskInfo).Methods("GET")

	// Static server information endpoints by server key (for agents)
	router.HandleFunc("/api/servers/by-key/{server_key}/static-info", staticInfoHandler.UpsertStaticInfoByKey).Methods("POST", "PUT")
	router.HandleFunc("/api/servers/by-key/{server_key}/static-info", staticInfoHandler.GetStaticInfoByKey).Methods("GET")
	router.HandleFunc("/api/servers/by-key/{server_key}/static-info/server", staticInfoHandler.GetServerInfoByKey).Methods("GET")
	router.HandleFunc("/api/servers/by-key/{server_key}/static-info/hardware", staticInfoHandler.GetHardwareInfoByKey).Methods("GET")
	router.HandleFunc("/api/servers/by-key/{server_key}/static-info/network", staticInfoHandler.GetNetworkInterfacesByKey).Methods("GET")
	router.HandleFunc("/api/servers/by-key/{server_key}/static-info/disks", staticInfoHandler.GetDiskInfoByKey).Methods("GET")

	// Server metrics with storage temperatures (public)
	router.HandleFunc("/api/servers/{server_id}/metrics/temperatures", serverMetricsHandler.GetServerMetricsWithTemperatures).Methods("GET")
	router.HandleFunc("/api/servers/by-key/{server_key}/metrics/temperatures", serverMetricsHandler.GetServerMetricsWithTemperaturesByKey).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/alerts/storage-temperature", serverMetricsHandler.GetStorageTemperatureAlerts).Methods("GET")

	// Alert endpoints (public)
	router.HandleFunc("/api/servers/{server_id}/alerts", alertHandler.GetActiveAlerts).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/alerts/all", alertHandler.GetAllAlerts).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/alerts/type/{type}", alertHandler.GetAlertsByType).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/alerts/range", alertHandler.GetAlertsByTimeRange).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/alerts/stats", alertHandler.GetAlertStats).Methods("GET")
	router.HandleFunc("/api/servers/{server_id}/alerts/{alert_id}/resolve", alertHandler.ResolveAlert).Methods("POST")
	router.HandleFunc("/api/servers/{server_id}/alerts/type/{type}/resolve", alertHandler.ResolveAlertsByType).Methods("POST")

	// WebSocket endpoint for testing
	router.HandleFunc("/ws", wsServer.HandleConnection).Methods("GET")

	// API endpoints for Telegram bot and web dashboard (with auth)
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.Auth(storageImpl, logger))

	// Server endpoints (protected)
	api.HandleFunc("/servers", serversHandler.ListServers).Methods("GET")
	api.HandleFunc("/servers/{server_id}/status", serversHandler.GetServerStatus).Methods("GET")
	api.HandleFunc("/servers/{server_id}/command", commandsHandler.SendCommand).Methods("POST")

	return router
}
