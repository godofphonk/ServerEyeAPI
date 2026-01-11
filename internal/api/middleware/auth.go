package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/godofphonk/ServerEyeAPI/internal/utils"
	"github.com/sirupsen/logrus"
)

// Auth middleware handles authentication
func Auth(storage storage.Storage, logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public routes
			if isPublicRoute(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Extract server_id and server_key from Authorization header
			// Expected format: "Bearer server_id:server_key"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			credentials := strings.SplitN(parts[1], ":", 2)
			if len(credentials) != 2 {
				http.Error(w, "Invalid credentials format", http.StatusUnauthorized)
				return
			}

			serverID := credentials[0]
			serverKey := credentials[1]

			// Validate format
			if err := utils.ValidateServerID(serverID); err != nil {
				http.Error(w, "Invalid server ID", http.StatusUnauthorized)
				return
			}

			if err := utils.ValidateServerKey(serverKey); err != nil {
				http.Error(w, "Invalid server key", http.StatusUnauthorized)
				return
			}

			// Check if server exists and is authenticated
			// This would typically check against a database or Redis
			// For now, we'll just validate the format

			// Add server info to context
			ctx := context.WithValue(r.Context(), "server_id", serverID)
			ctx = context.WithValue(ctx, "server_key", serverKey)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isPublicRoute checks if the route is public (doesn't require auth)
func isPublicRoute(path string) bool {
	publicRoutes := []string{
		"/health",
		"/RegisterKey",
	}

	for _, route := range publicRoutes {
		if path == route {
			return true
		}
	}

	return false
}
