package matcher

import (
	"fmt"
	"sync"
	"testing"

	"echo-server/internal/config"
)

func TestPathMatcherConcurrency(t *testing.T) {
	pm := New()

	const (
		numGoroutines = 100
		numPatterns   = 50
	)

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // For both adding and matching

	// Test concurrent pattern addition
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numPatterns; j++ {
				cfg := &config.PathConfig{
					Pattern: fmt.Sprintf("/test/%d/%d/.*", id, j),
					Methods: []string{"GET"},
				}
				if err := pm.Add(cfg); err != nil {
					t.Errorf("Failed to add pattern: %v", err)
				}
			}
		}(i)
	}

	// Test concurrent matching
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numPatterns; j++ {
				_, matched := pm.Match("/test/1/1/endpoint", "GET")
				if matched {
					// Just ensure no panic occurs during concurrent access
					continue
				}
			}
		}()
	}

	wg.Wait()
}

func TestPathMatcherOrdering(t *testing.T) {
	pm := New()

	patterns := []struct {
		pattern string
		method  string
	}{
		{"/specific/path", "GET"},
		{"/specific/.*", "GET"},
		{"/.*", "GET"},
	}

	// Add patterns in specific order
	for _, p := range patterns {
		err := pm.Add(&config.PathConfig{
			Pattern: p.pattern,
			Methods: []string{p.method},
		})
		if err != nil {
			t.Fatalf("Failed to add pattern: %v", err)
		}
	}

	// Test that more specific patterns match first
	cfg, matched := pm.Match("/specific/path", "GET")
	if !matched {
		t.Error("Expected to match specific path")
	}
	if cfg.Pattern != "/specific/path" {
		t.Errorf("Got pattern %s, want /specific/path", cfg.Pattern)
	}
}
