package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/godofphonk/ServerEyeAPI/internal/storage"
)

type contextKey string

const (
	ServiceIDKey   contextKey = "service_id"
	PermissionsKey contextKey = "permissions"
	KeyIDKey       contextKey = "key_id"
)

type APIKeyAuthMiddleware struct {
	storage *storage.APIKeyStorage
	logger  *logrus.Logger
}

func NewAPIKeyAuthMiddleware(storage *storage.APIKeyStorage, logger *logrus.Logger) *APIKeyAuthMiddleware {
	return &APIKeyAuthMiddleware{
		storage: storage,
		logger:  logger,
	}
}

// Authenticate validates the API key from the request header
func (m *APIKeyAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			m.logger.Warn("Missing API key in request")
			http.Error(w, `{"error":"API key required"}`, http.StatusUnauthorized)
			return
		}

		// Validate the API key
		key, err := m.storage.ValidateAPIKey(r.Context(), apiKey)
		if err != nil {
			m.logger.WithError(err).Warn("Invalid API key")

			// Log failed attempt
			m.storage.LogAPIKeyUsage(
				r.Context(),
				"unknown",
				r.URL.Path,
				getClientIP(r),
				r.UserAgent(),
				false,
				"Invalid API key",
			)

			http.Error(w, `{"error":"Invalid API key"}`, http.StatusUnauthorized)
			return
		}

		// Update last used timestamp
		go m.storage.UpdateLastUsed(context.Background(), key.KeyID)

		// Log successful usage
		go m.storage.LogAPIKeyUsage(
			context.Background(),
			key.KeyID,
			r.URL.Path,
			getClientIP(r),
			r.UserAgent(),
			true,
			"",
		)

		// Add key information to request context
		ctx := context.WithValue(r.Context(), ServiceIDKey, key.ServiceID)
		ctx = context.WithValue(ctx, PermissionsKey, key.Permissions)
		ctx = context.WithValue(ctx, KeyIDKey, key.KeyID)

		m.logger.WithFields(logrus.Fields{
			"service_id": key.ServiceID,
			"key_id":     key.KeyID,
			"endpoint":   r.URL.Path,
		}).Debug("API key authenticated")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission checks if the authenticated service has the required permission
func (m *APIKeyAuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, ok := r.Context().Value(PermissionsKey).([]string)
			if !ok {
				m.logger.Error("Permissions not found in context")
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Check if the required permission exists
			hasPermission := false
			for _, p := range permissions {
				if p == permission || p == "*" {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				serviceID := r.Context().Value(ServiceIDKey).(string)
				m.logger.WithFields(logrus.Fields{
					"service_id": serviceID,
					"permission": permission,
					"endpoint":   r.URL.Path,
				}).Warn("Permission denied")

				http.Error(w, `{"error":"Permission denied"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if multiple are present
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// GetServiceID retrieves the service ID from the request context
func GetServiceID(ctx context.Context) string {
	if serviceID, ok := ctx.Value(ServiceIDKey).(string); ok {
		return serviceID
	}
	return ""
}

// GetPermissions retrieves the permissions from the request context
func GetPermissions(ctx context.Context) []string {
	if permissions, ok := ctx.Value(PermissionsKey).([]string); ok {
		return permissions
	}
	return []string{}
}

// GetKeyID retrieves the key ID from the request context
func GetKeyID(ctx context.Context) string {
	if keyID, ok := ctx.Value(KeyIDKey).(string); ok {
		return keyID
	}
	return ""
}
