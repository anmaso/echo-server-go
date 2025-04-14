package config

import (
	"sync"
)

// ConfigManager handles thread-safe access to configurations
type ConfigManager struct {
	mu     sync.RWMutex
	config *ServerConfig
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		config: &ServerConfig{
			PathMatcher: NewPathMatcher(),
		},
	}
}

func (cm *ConfigManager) GetConfig() *ServerConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

func (cm *ConfigManager) UpdateConfig(cfg *ServerConfig) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config = cfg
}

func (cm *ConfigManager) UpdatePathConfig(cfg PathConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.PathMatcher.DeleteByName(cfg.Name)
	return cm.config.PathMatcher.Add(&cfg)
}
