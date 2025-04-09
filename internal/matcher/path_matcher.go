package matcher

import (
	"fmt"
	"regexp"
	"sync"

	"echo-server/internal/config"
	"echo-server/pkg/logger"
)

type PathMatcher struct {
	patterns sync.Map
	order    []string
	mu       sync.RWMutex
}

type pattern struct {
	config *config.PathConfig
	regex  *regexp.Regexp
}

func New() *PathMatcher {
	return &PathMatcher{
		order: make([]string, 0),
	}
}

func (pm *PathMatcher) Add(cfg *config.PathConfig) error {
	if cfg.Pattern == "" {
		return fmt.Errorf("empty pattern not allowed")
	}

	regex, err := regexp.Compile(cfg.Pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %s: %w", cfg.Pattern, err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Store pattern in sync.Map
	pm.patterns.Store(cfg.Pattern, &pattern{
		config: cfg,
		regex:  regex,
	})

	// Update order slice
	pm.order = append(pm.order, cfg.Pattern)

	logger.Info("Added path pattern: %s", cfg.Pattern)
	return nil
}

func (pm *PathMatcher) Match(path, method string) (*config.PathConfig, bool) {
	// Get ordered patterns
	pm.mu.RLock()
	patterns := make([]string, len(pm.order))
	copy(patterns, pm.order)
	pm.mu.RUnlock()

	// Check patterns in order
	for _, patternKey := range patterns {
		if p, ok := pm.patterns.Load(patternKey); ok {
			pat := p.(*pattern)
			if pat.regex.MatchString(path) {
				if len(pat.config.Methods) == 0 || containsMethod(pat.config.Methods, method) {
					return pat.config, true
				}
			}
		}
	}
	return nil, false
}

func (pm *PathMatcher) Clear() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.patterns.Range(func(key, _ interface{}) bool {
		pm.patterns.Delete(key)
		return true
	})
	pm.order = make([]string, 0)

	logger.Info("Cleared all path patterns")
}

func containsMethod(methods []string, method string) bool {
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}
