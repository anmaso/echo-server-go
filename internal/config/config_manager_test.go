package config

import (
	"fmt"
	"sync"
	"testing"
)

func TestConfigManagerConcurrency(t *testing.T) {
	cm := NewConfigManager()

	const (
		numGoroutines = 100
		numUpdates    = 50
	)

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // For both readers and writers

	// Test concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numUpdates; j++ {
				cfg := cm.GetConfig()
				if cfg == nil {
					t.Error("Configuration should never be nil")
				}
			}
		}()
	}

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numUpdates; j++ {
				pathCfg := PathConfig{
					Pattern: fmt.Sprintf("/test/%d/%d", id, j),
					Methods: []string{"GET"},
				}
				if err := cm.UpdatePathConfig(pathCfg); err != nil {
					t.Errorf("Failed to update path config: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()
}
