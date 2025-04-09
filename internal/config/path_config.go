package config

import (
	"regexp"
	"sync"

	"echo-server/pkg/logger"
)

// PathConfig represents configuration for a specific path pattern
type PathConfig struct {
	Pattern        string          `json:"pattern"`
	Methods        []string        `json:"methods"`
	Response       ResponseConfig  `json:"response"`
	ErrorResponse  *ResponseConfig `json:"errorResponse,omitempty"`
	ErrorFrequency float64         `json:"errorFrequency"`
	CounterEnabled bool            `json:"counterEnabled"`
	regex          *regexp.Regexp
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

// Clear removes all path configurations
func (pm *pathMatcherImpl) Clear() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.configs = make([]PathConfig, 0)
	logger.Info("Cleared all path patterns")
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
