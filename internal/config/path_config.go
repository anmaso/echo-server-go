package config

import (
	"regexp"
	"sync"

	"echo-server/pkg/logger"
)

// ProxyConfig defines upstream proxy configuration for a path
type ProxyConfig struct {
	URL            string   `json:"url"`
	Timeout        Duration `json:"timeout,omitempty"`
	StripPrefix    bool     `json:"stripPrefix,omitempty"`
	PreserveHost   bool     `json:"preserveHost,omitempty"`
	AllowedHeaders []string `json:"allowedHeaders,omitempty"`
	ForwardHeaders []string `json:"forwardHeaders,omitempty"`
}

// PathConfig represents configuration for a specific path pattern
type PathConfig struct {
	Name           string          `json:"name"`
	Pattern        string          `json:"pattern"`
	Methods        []string        `json:"methods"`
	Response       ResponseConfig  `json:"response"`
	ErrorResponse  *ResponseConfig `json:"errorResponse,omitempty"`
	ErrorEvery     int             `json:"errorEvery"`
	CounterEnabled bool            `json:"counterEnabled"`
	regex          *regexp.Regexp
	Proxy          *ProxyConfig `json:"proxy,omitempty"` // Add this field
}

// ResponseConfig defines the response behavior
type ResponseConfig struct {
	StatusCode     int               `json:"statusCode"`
	Headers        map[string]string `json:"headers"`
	Body           string            `json:"body"`
	Delay          Duration          `json:"delay"`
	IncludeRequest bool              `json:"includeRequest"`
}

func NewResponseConfig() ResponseConfig {
	return ResponseConfig{
		StatusCode:     200,
		Headers:        make(map[string]string),
		Body:           "",
		Delay:          Duration{},
		IncludeRequest: false,
	}
}

// PathMatcher interface for path configuration matching and storage
type PathMatcher interface {
	Add(cfg *PathConfig) error
	Match(path, method string) (*PathConfig, bool)
	Clear()
	GetAllConfigs() []PathConfig // New method
	DeleteByName(name string) bool
}

// pathMatcherImpl implements the PathMatcher interface
type pathMatcherImpl struct {
	configs []PathConfig
	mu      sync.RWMutex
}

// NewPathMatcher creates a new PathMatcher instance
func NewPathMatcher() PathMatcher {
	return &pathMatcherImpl{
		configs: make([]PathConfig, 0),
	}
}

// Add adds a new path configuration
func (pm *pathMatcherImpl) Add(cfg *PathConfig) error {
	regex, err := regexp.Compile(cfg.Pattern)
	if err != nil {
		return err
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	cfg.regex = regex
	pm.configs = append(pm.configs, *cfg)
	logger.Info("Added path pattern: %s", cfg.Pattern)
	return nil
}

// Match finds the first matching configuration for a path
func (pm *pathMatcherImpl) Match(path, method string) (*PathConfig, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for i := range pm.configs {
		if pm.configs[i].regex.MatchString(path) {
			if len(pm.configs[i].Methods) == 0 || contains(pm.configs[i].Methods, method) {
				return &pm.configs[i], true
			}
		}
	}
	return nil, false
}

func (pm *pathMatcherImpl) DeleteByName(name string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i, cfg := range pm.configs {
		if cfg.Name == name {
			pm.configs = append(pm.configs[:i], pm.configs[i+1:]...)
			logger.Info("Deleted path pattern: %s", cfg.Pattern)
			return true
		}
	}
	logger.Warn("No path pattern found with name: %s", name)
	return false
}

// Clear removes all path configurations
func (pm *pathMatcherImpl) Clear() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.configs = make([]PathConfig, 0)
	logger.Info("Cleared all path patterns")
}

// GetAllConfigs retrieves all path configurations
func (pm *pathMatcherImpl) GetAllConfigs() []PathConfig {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	configs := make([]PathConfig, len(pm.configs))
	copy(configs, pm.configs)
	return configs
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
