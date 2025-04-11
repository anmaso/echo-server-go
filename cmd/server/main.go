package main

import (
	"context"
	"flag"
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
	cfg := getConfig()

	if cfg == nil {
		fmt.Println(helpText)
		os.Exit(0)
	}

	cm := config.NewConfigManager()
	cm.UpdateConfig(cfg)

	// Create and start server
	srv := server.New(cm)

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

func getConfig() *config.ServerConfig {
	// Define command-line flags
	host := flag.String("host", "0.0.0.0", "Server host (overrides config file)")
	port := flag.Int("port", 8080, "Server port (overrides config file)")
	readTimeout := flag.Duration("read-timeout", 0, "Read timeout duration (overrides config file)")
	writeTimeout := flag.Duration("write-timeout", 0, "Write timeout duration (overrides config file)")
	configPath := flag.String("config", "config/server.json", "Path to server configuration file")
	pathsDir := flag.String("paths-dir", "config/paths", "Path to directory containing path configurations")
	logLevel := flag.String("log-level", "info", "Logging level (debug, info, warn, error)")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		return nil
	}

	// Set log level
	switch strings.ToLower(*logLevel) {
	case "debug":
		logger.SetLevel(logger.DEBUG)
	case "info":
		logger.SetLevel(logger.INFO)
	case "warn":
		logger.SetLevel(logger.WARN)
	case "error":
		logger.SetLevel(logger.ERROR)
	}

	// Initialize configuration loader
	loader := config.NewLoader()
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current working directory: %v", err)
	}

	// Resolve config paths
	configPathServer := *configPath
	if strings.Contains(cwd, "cmd/server") {
		configPathServer, _ = url.JoinPath(cwd, "../../", configPathServer)
	}

	configPathRoutes := *pathsDir
	if strings.Contains(cwd, "cmd/server") {
		configPathRoutes, _ = url.JoinPath(cwd, "../../", configPathRoutes)
	}

	// Load server configuration
	if err := loader.LoadServerConfig(configPathServer); err != nil {
		logger.Error("Failed to load server config: %v", err)
		os.Exit(1)
	}

	// Override configuration with command-line flags
	cfg := loader.GetConfig()
	if *host != "" {
		cfg.Host = *host
	}
	if *port != 0 {
		cfg.Port = *port
	}
	if *readTimeout != 0 {
		cfg.ReadTimeout.Duration = *readTimeout
	}
	if *writeTimeout != 0 {
		cfg.WriteTimeout.Duration = *writeTimeout
	}

	// Load path configurations
	if err := loader.LoadPathConfigs(configPathRoutes); err != nil {
		logger.Error("Failed to load path configs: %v", err)
		os.Exit(1)
	}
	return cfg
}

const helpText = `Echo Server - A configurable HTTP mock server

Usage:
  echo-server [options]

Options:
  -port int
        Port to run the server on (default 8080)
  -config string
        Path to configuration directory (default "./config")
  -help
        Show this help message

Examples:
  # Start server on default port 8080
  echo-server

  # Start server on custom port
  echo-server -port 3000

  # Use custom config directory
  echo-server -config /path/to/configs

  # Show help
  echo-server -help

Features:
  - Configure custom endpoint behaviors
  - Response templating
  - Request delay simulation
  - Error injection
  - Request counters
  - Path pattern matching
  - Multiple HTTP methods support

For more information, visit: https://github.com/anmaso/echo-server-go`
