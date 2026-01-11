package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Logging middleware logs HTTP requests
func Logging(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for WebSocket upgrade requests
			if r.Header.Get("Upgrade") == "websocket" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"status":      wrapped.statusCode,
				"duration_ms": duration.Milliseconds(),
				"user_agent":  r.UserAgent(),
			}).Info("HTTP request")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
