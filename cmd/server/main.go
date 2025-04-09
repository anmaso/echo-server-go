package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"echo-server/internal/config"
	"echo-server/internal/server"
	"echo-server/pkg/logger"
)

func main() {
	// Initialize configuration loader
	loader := config.NewLoader()
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current working directory: %v", err)
	}

	configPathServer := "config/server.json"
	if strings.Contains(cwd, "cmd/server") {
		configPathServer, _ = url.JoinPath(cwd, "../../", configPathServer)
	}
	fmt.Printf("configPath: %s\n", configPathServer)

	configPathRoutes := "config/paths"
	if strings.Contains(cwd, "cmd/server") {
		configPathRoutes, _ = url.JoinPath(cwd, "../../", configPathRoutes)
	}

	// Load server configuration
	if err := loader.LoadServerConfig(configPathServer); err != nil {
		logger.Error("Failed to load server config: %v", err)
		os.Exit(1)
	}

	// Load path configurations
	if err := loader.LoadPathConfigs(configPathRoutes); err != nil {
		logger.Error("Failed to load path configs: %v", err)
		os.Exit(1)
	}

	// Create and start server
	srv := server.New(loader.GetConfig())

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
