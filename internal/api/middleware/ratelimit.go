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

package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/config"
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
func NewRateLimiter(cfg *config.Config, logger *logrus.Logger) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*ClientLimiter),
		rate:    rate{limit: cfg.RateLimit.Limit, window: cfg.RateLimit.Window},
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
