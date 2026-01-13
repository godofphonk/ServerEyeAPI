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
	RedisURL        string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`

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

	// Redis Configuration
	Redis struct {
		TTL         time.Duration `env:"REDIS_TTL" envDefault:"60s"`
		ConnTimeout time.Duration `env:"REDIS_CONN_TIMEOUT" envDefault:"5s"`
		MaxRetries  int           `env:"REDIS_MAX_RETRIES" envDefault:"3"`
	}

	// WebSocket Configuration
	WebSocket struct {
		BufferSize   int           `env:"WS_BUFFER_SIZE" envDefault:"256"`
		WriteTimeout time.Duration `env:"WS_WRITE_TIMEOUT" envDefault:"30s"`
		ReadTimeout  time.Duration `env:"WS_READ_TIMEOUT" envDefault:"120s"`
		PingInterval time.Duration `env:"WS_PING_INTERVAL" envDefault:"30s"`
		PongWait     time.Duration `env:"WS_PONG_WAIT" envDefault:"120s"`
	}

	// Rate Limiting Configuration
	RateLimit struct {
		Limit  int           `env:"RATE_LIMIT" envDefault:"100"`
		Window time.Duration `env:"RATE_WINDOW" envDefault:"1m"`
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
