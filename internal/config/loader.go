package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"echo-server/pkg/logger"
)

type Loader struct {
	mu     sync.RWMutex
	config *ServerConfig
}

func NewLoader() *Loader {
	return &Loader{
		config: &ServerConfig{
			PathMatcher: NewPathMatcher(),
		},
	}
}

func (l *Loader) LoadServerConfig(filepath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("Server config file does not exist: %s", filepath)
			return nil
		}
		return fmt.Errorf("reading server config: %w", err)
	}

	// Initialize cfg with default values, especially for new nested structs
	// This ensures that if a section like "history" is missing from JSON,
	// the default values are used.
	cfg := ServerConfig{
		PathMatcher: NewPathMatcher(), // Initialize PathMatcher as before
		History: HistoryConfig{ // Default history configuration
			Enabled:        false,
			DefaultMaxSize: 100,
		},
		// Note: Other fields like Host, Port, Timeouts, DefaultResponse 
		// will be zero-valued if not in JSON, which might be acceptable or 
		// might need explicit defaults here too if strict defaults are required for them.
		// For this task, we are focusing on HistoryConfig.
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parsing server config: %w", err)
	}

	// Ensure PathMatcher is re-initialized if it was overwritten by JSON (empty object)
	// or if it's nil (though with the above initialization, it shouldn't be nil).
	// If PathMatcher can be defined in JSON, this logic might need adjustment.
	// Based on current types.go, PathMatcher is not json tagged (`json:"-"`), so it's not directly unmarshalled.
	// However, the `Paths []PathConfig` is used to populate it.
	// The line `cfg.PathMatcher = NewPathMatcher()` after unmarshal might still be needed if JSON
	// could somehow nullify it, but it's safer to initialize it as part of the default struct.
	// Let's ensure it's not nil.
	if cfg.PathMatcher == nil {
		cfg.PathMatcher = NewPathMatcher()
	}
	// The LoadPathConfigs call will populate this PathMatcher.

	l.config = &cfg
	return nil
}

func (l *Loader) LoadPathConfigs(dirPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				logger.Warn("Path config directory does not exist: %s", dirPath)
				return nil
			}
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			logger.Error("Failed to read path config %s: %v", path, err)
			return nil // Continue with other files
		}

		var cfg PathConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			logger.Error("Failed to parse path config %s: %v", path, err)
			return nil // Continue with other files
		}

		if err := l.config.PathMatcher.Add(&cfg); err != nil {
			logger.Error("Failed to add path config %s: %v", path, err)
			return nil // Continue with other files
		}

		logger.Info("Loaded path configuration from %s", path)
		return nil
	})
}

func (l *Loader) GetConfig() *ServerConfig {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}
