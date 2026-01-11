package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	clients map[string]*ClientLimiter
	mutex   sync.RWMutex
	rate    rate
	logger  *logrus.Logger
}

// ClientLimiter represents a rate limiter for a specific client
type ClientLimiter struct {
	tokens   int
	lastSeen time.Time
	mutex    sync.Mutex
}

// rate defines the rate limit configuration
type rate struct {
	limit  int
	window time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration, logger *logrus.Logger) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*ClientLimiter),
		rate:    rate{limit: limit, window: window},
		logger:  logger,
	}
}

// RateLimit middleware applies rate limiting
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr

		rl.mutex.Lock()
		if _, exists := rl.clients[clientIP]; !exists {
			rl.clients[clientIP] = &ClientLimiter{
				tokens:   rl.rate.limit,
				lastSeen: time.Now(),
			}
		}

		client := rl.clients[clientIP]
		rl.mutex.Unlock()

		client.mutex.Lock()
		defer client.mutex.Unlock()

		// Refill tokens based on time passed
		now := time.Now()
		elapsed := now.Sub(client.lastSeen)
		tokensToAdd := int(elapsed.Seconds()) * rl.rate.limit / int(rl.rate.window.Seconds())

		if tokensToAdd > 0 {
			client.tokens = min(client.tokens+tokensToAdd, rl.rate.limit)
			client.lastSeen = now
		}

		// Check if client has tokens
		if client.tokens <= 0 {
			rl.logger.WithField("client_ip", clientIP).Warn("Rate limit exceeded")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Consume a token
		client.tokens--

		next.ServeHTTP(w, r)
	})
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
