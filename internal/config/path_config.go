package config

import (
	"regexp"
	"sync"
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
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Delay      Duration          `json:"delay"`
}

// PathMatcher handles path configuration matching and storage
type PathMatcher struct {
	configs []PathConfig
	mu      sync.RWMutex
}

// NewPathMatcher creates a new PathMatcher instance
func NewPathMatcher() *PathMatcher {
	return &PathMatcher{
		configs: make([]PathConfig, 0),
	}
}

// CompilePatterns compiles all regex patterns
func (pm *PathMatcher) CompilePatterns() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i := range pm.configs {
		regex, err := regexp.Compile(pm.configs[i].Pattern)
		if err != nil {
			return err
		}
		pm.configs[i].regex = regex
	}
	return nil
}

// AddConfig adds a new path configuration
func (pm *PathMatcher) AddConfig(cfg PathConfig) error {
	regex, err := regexp.Compile(cfg.Pattern)
	if err != nil {
		return err
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	cfg.regex = regex
	pm.configs = append(pm.configs, cfg)
	return nil
}

// FindMatch finds the first matching configuration for a path
func (pm *PathMatcher) FindMatch(path string, method string) *PathConfig {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for i := range pm.configs {
		if pm.configs[i].regex.MatchString(path) {
			if len(pm.configs[i].Methods) == 0 || contains(pm.configs[i].Methods, method) {
				return &pm.configs[i]
			}
		}
	}
	return nil
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
