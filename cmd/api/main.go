package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/api"
	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Initialize storage
	var storageImpl storage.Storage
	if cfg.DatabaseURL == "" {
		logger.Info("Using in-memory storage (DATABASE_URL not set)")
		storageImpl = storage.NewMemoryStorage(logger)
	} else {
		postgresStorage, err := storage.NewPostgresStorage(cfg.DatabaseURL, logger)
		if err != nil {
			logger.WithError(err).Fatal("Failed to initialize storage")
		}
		storageImpl = postgresStorage
	}
	defer storageImpl.Close()

	// Initialize API server
	server := api.New(cfg, storageImpl, logger)

	// Start server in goroutine
	go func() {
		logger.WithField("addr", ":8080").Info("Starting API server")
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exited")
}
