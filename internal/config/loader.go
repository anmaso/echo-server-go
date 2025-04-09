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
		return fmt.Errorf("reading server config: %w", err)
	}

	var cfg ServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parsing server config: %w", err)
	}

	cfg.PathMatcher = NewPathMatcher()
	l.config = &cfg
	return nil
}

func (l *Loader) LoadPathConfigs(dirPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
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

		if err := l.config.PathMatcher.AddConfig(cfg); err != nil {
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
