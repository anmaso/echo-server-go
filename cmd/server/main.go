package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"echo-server/internal/config"
	"echo-server/internal/server"
	"echo-server/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.Default()
	if err := config.Load("config/config.json"); err == nil {
		cfg = config.Get()
	}

	// Create and start server
	srv := server.New(cfg)

	// Handle shutdown gracefully
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}
