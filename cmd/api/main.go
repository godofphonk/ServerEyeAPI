package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/config"
	"github.com/godofphonk/ServerEyeAPI/internal/wire"
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

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.WithError(err).Fatal("Configuration validation failed")
	}

	// Initialize API server using Wire DI container
	server, err := wire.InitializeApp(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize server")
	}

	// CI/CD TEST: Log at startup for verification
	logger.Info("CI/CD TEST: Application started successfully at " + time.Now().Format("2006-01-02 15:04:05"))

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
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
