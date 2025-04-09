package config

import "testing"

func TestPathMatcher(t *testing.T) {

	tests := []struct {
		name        string
		config      *PathConfig
		path        string
		method      string
		shouldMatch bool
	}{
		{
			name: "exact match",
			config: &PathConfig{
				Pattern: "^/test$",
				Methods: []string{"GET"},
			},
			path:        "/test",
			method:      "GET",
			shouldMatch: true,
		},
		{
			name: "method not allowed",
			config: &PathConfig{
				Pattern: "^/test$",
				Methods: []string{"GET"},
			},
			path:        "/test",
			method:      "POST",
			shouldMatch: false,
		},
		{
			name: "pattern no match",
			config: &PathConfig{
				Pattern: "^/api/.*",
				Methods: []string{"GET"},
			},
			path:        "/test",
			method:      "GET",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPathMatcher()
			if err := pm.Add(tt.config); err != nil {
				t.Fatalf("Failed to add pattern: %v", err)
			}

			cfg, matched := pm.Match(tt.path, tt.method)
			if matched != tt.shouldMatch {
				t.Errorf("%s Match() = %v, want %v", tt.path, matched, tt.shouldMatch)
			}

			if matched && cfg == nil {
				t.Error("Match() returned true but nil config")
			}
		})
	}
}
