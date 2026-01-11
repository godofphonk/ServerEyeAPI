package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	// Server
	Host string
	Port int

	// Database
	DatabaseURL     string
	KeysDatabaseURL string
	RedisURL        string

	// Metrics
	MetricsTopic string

	// Security
	JWTSecret     string
	WebhookSecret string

	// Web
	WebURL string

	// Kafka
	KafkaBrokers []string
	KafkaGroupID string
}

func Load() (*Config, error) {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		logrus.Info("No .env file found, using environment variables")
	}

	cfg := &Config{
		Host: getEnv("HOST", "0.0.0.0"),
		Port: getEnvInt("PORT", 8080),

		DatabaseURL:     getEnv("DATABASE_URL", ""),
		KeysDatabaseURL: getEnv("KEYS_DATABASE_URL", ""),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),

		MetricsTopic: getEnv("METRICS_TOPIC", "metrics"),

		JWTSecret:     getEnv("JWT_SECRET", ""),
		WebhookSecret: getEnv("WEBHOOK_SECRET", ""),

		WebURL: getEnv("WEB_URL", ""),

		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "")},
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", ""),
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
		BatchSize:         100,
		BatchTimeout:      1 * time.Second,
		CommitInterval:    1 * time.Second,
		RebalanceTimeout:  30 * time.Second,
		StartOffset:       -2, // Start from earliest
		MaxProcessingTime: 30 * time.Second,
		EnableAutoCommit:  false,
		SessionTimeout:    30 * time.Second,
		HeartbeatInterval: 3 * time.Second,
		MaxPollRecords:    500,
	}
}
