package config

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"echo-server/pkg/logger"
)

var (
	config *ServerConfig
	mu     sync.RWMutex
)

// Load reads configuration from a JSON file
func Load(filepath string) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(filepath)
	if err != nil {
		logger.Error("Failed to read config file: %v", err)
		return err
	}

	var cfg ServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		logger.Error("Failed to parse config file: %v", err)
		return err
	}

	config = &cfg
	logger.Info("Configuration loaded successfully")
	return nil
}

// Get returns the current server configuration
func Get() *ServerConfig {
	mu.RLock()
	defer mu.RUnlock()
	return config
}

// Default returns a default configuration
func Default() *ServerConfig {
	return &ServerConfig{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  Duration{time.Second * 30},
		WriteTimeout: Duration{time.Second * 30},
		DefaultResponse: Response{
			StatusCode: 200,
			Headers:    Headers{"Content-Type": "application/json"},
		},
		Paths: []PathConfig{},
	}
}
