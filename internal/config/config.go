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

package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	// Server
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	Port int    `env:"PORT" envDefault:"8080"`

	// Database
	DatabaseURL     string `env:"DATABASE_URL"`
	KeysDatabaseURL string `env:"KEYS_DATABASE_URL"`
	TimescaleDBURL  string `env:"TIMESCALEDB_URL"` // TimescaleDB URL for time-series data

	// Metrics
	MetricsTopic string `env:"METRICS_TOPIC" envDefault:"metrics"`

	// Security
	JWTSecret     string `env:"JWT_SECRET"`
	WebhookSecret string `env:"WEBHOOK_SECRET"`

	// Web
	WebURL string `env:"WEB_URL"`

	// Kafka
	KafkaBrokers []string `env:"KAFKA_BROKERS" envSeparator:","`
	KafkaGroupID string   `env:"KAFKA_GROUP_ID"`

	// Redis Configuration (deprecated - use TimescaleDB for new deployments)
	Redis struct {
		TTL         time.Duration `env:"REDIS_TTL" envDefault:"5m"`
		ConnTimeout time.Duration `env:"REDIS_CONN_TIMEOUT" envDefault:"5s"`
		MaxRetries  int           `env:"REDIS_MAX_RETRIES" envDefault:"3"`
	}

	// TimescaleDB Configuration
	TimescaleDB struct {
		MaxConnections      int           `env:"TIMESCALEDB_MAX_CONNECTIONS" envDefault:"20"`
		ConnTimeout         time.Duration `env:"TIMESCALEDB_CONN_TIMEOUT" envDefault:"30s"`
		QueryTimeout        time.Duration `env:"TIMESCALEDB_QUERY_TIMEOUT" envDefault:"10s"`
		HealthCheckInterval time.Duration `env:"TIMESCALEDB_HEALTH_CHECK_INTERVAL" envDefault:"30s"`
	}

	// WebSocket Configuration
	WebSocket struct {
		BufferSize   int           `env:"WS_BUFFER_SIZE" envDefault:"256"`
		WriteTimeout time.Duration `env:"WS_WRITE_TIMEOUT" envDefault:"30s"`
		ReadTimeout  time.Duration `env:"WS_READ_TIMEOUT" envDefault:"300s"`
		PingInterval time.Duration `env:"WS_PING_INTERVAL" envDefault:"60s"`
		PongWait     time.Duration `env:"WS_PONG_WAIT" envDefault:"600s"`
	}

	// Rate Limiting Configuration
	RateLimit struct {
		Limit  int           `env:"RATE_LIMIT" envDefault:"100"`
		Window time.Duration `env:"RATE_WINDOW" envDefault:"1m"`
	}

	// Data Retention Configuration
	Retention struct {
		MetricsDays  int `env:"METRICS_RETENTION_DAYS" envDefault:"30"`
		StatusDays   int `env:"STATUS_RETENTION_DAYS" envDefault:"7"`
		EventsDays   int `env:"EVENTS_RETENTION_DAYS" envDefault:"14"`
		CommandsDays int `env:"COMMANDS_RETENTION_DAYS" envDefault:"14"`
	}

	// Consumer Configuration
	Consumer struct {
		BatchSize         int           `env:"CONSUMER_BATCH_SIZE" envDefault:"100"`
		BatchTimeout      time.Duration `env:"CONSUMER_BATCH_TIMEOUT" envDefault:"1s"`
		CommitInterval    time.Duration `env:"CONSUMER_COMMIT_INTERVAL" envDefault:"1s"`
		RebalanceTimeout  time.Duration `env:"CONSUMER_REBALANCE_TIMEOUT" envDefault:"30s"`
		MaxProcessingTime time.Duration `env:"CONSUMER_MAX_PROCESSING_TIME" envDefault:"30s"`
		SessionTimeout    time.Duration `env:"CONSUMER_SESSION_TIMEOUT" envDefault:"30s"`
		HeartbeatInterval time.Duration `env:"CONSUMER_HEARTBEAT_INTERVAL" envDefault:"3s"`
		MaxPollRecords    int           `env:"CONSUMER_MAX_POLL_RECORDS" envDefault:"500"`
	}
}

func Load() (*Config, error) {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		logrus.Info("No .env file found, using environment variables")
	}

	cfg := &Config{}

	// Load environment variables with env tags
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Use DATABASE_URL for KEYS_DATABASE_URL if not provided (backward compatibility)
	if cfg.KeysDatabaseURL == "" {
		cfg.KeysDatabaseURL = cfg.DatabaseURL
		logrus.Info("Using DATABASE_URL for KEYS_DATABASE_URL (backward compatibility)")
	}

	// Validate required fields
	if cfg.JWTSecret == "" {
		logrus.Fatal("JWT_SECRET is required")
	}

	return cfg, nil
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		return []string{value}
	}
	return defaultValue
}

// Consumer configuration
type ConsumerConfig struct {
	Brokers           []string
	GroupID           string
	Topic             string
	BatchSize         int
	BatchTimeout      time.Duration
	CommitInterval    time.Duration
	RebalanceTimeout  time.Duration
	StartOffset       int64 // -2 = earliest, -1 = latest, 0+ = specific
	MaxProcessingTime time.Duration
	EnableAutoCommit  bool
	SessionTimeout    time.Duration
	HeartbeatInterval time.Duration
	MaxPollRecords    int
}

func NewConsumerConfig(cfg *Config) ConsumerConfig {
	return ConsumerConfig{
		Brokers:           cfg.KafkaBrokers,
		GroupID:           cfg.KafkaGroupID,
		Topic:             cfg.MetricsTopic,
		BatchSize:         cfg.Consumer.BatchSize,
		BatchTimeout:      cfg.Consumer.BatchTimeout,
		CommitInterval:    cfg.Consumer.CommitInterval,
		RebalanceTimeout:  cfg.Consumer.RebalanceTimeout,
		StartOffset:       -2, // Start from earliest
		MaxProcessingTime: cfg.Consumer.MaxProcessingTime,
		EnableAutoCommit:  false,
		SessionTimeout:    cfg.Consumer.SessionTimeout,
		HeartbeatInterval: cfg.Consumer.HeartbeatInterval,
		MaxPollRecords:    cfg.Consumer.MaxPollRecords,
	}
}

// Validate validates critical configuration values
func (c *Config) Validate() error {
	var errors []string

	// Validate required security fields
	if c.JWTSecret == "" {
		errors = append(errors, "JWT_SECRET is required")
	} else if len(c.JWTSecret) < 32 {
		errors = append(errors, "JWT_SECRET must be at least 32 characters")
	}

	if c.WebhookSecret == "" {
		errors = append(errors, "WEBHOOK_SECRET is required")
	} else if len(c.WebhookSecret) < 16 {
		errors = append(errors, "WEBHOOK_SECRET must be at least 16 characters")
	}

	// Validate database URLs
	if c.DatabaseURL == "" {
		errors = append(errors, "DATABASE_URL is required")
	}

	// KEYS_DATABASE_URL is optional for backward compatibility
	// If not provided, will use the same database as DATABASE_URL

	// Validate port range
	if c.Port < 1 || c.Port > 65535 {
		errors = append(errors, "PORT must be between 1 and 65535")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %v", errors)
	}

	return nil
}
