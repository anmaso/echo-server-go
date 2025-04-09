package counter

import (
	"sync"
	"testing"
)

func TestCounterThreadSafety(t *testing.T) {
	counter := GetGlobalCounter()
	counter.Reset()

	const (
		numGoroutines = 100
		numIterations = 1000
	)

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // For both global and path counters

	// Test global counter
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				counter.Increment()
			}
		}()
	}

	// Test path counter
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				counter.IncrementPath("/test")
			}
		}()
	}

	wg.Wait()

	expectedCount := uint64(numGoroutines * numIterations)

	// Check global counter
	if count := counter.GetCount(); count != expectedCount {
		t.Errorf("Global counter = %d, want %d", count, expectedCount)
	}

	// Check path counter
	if count := counter.GetPathCount("/test"); count != expectedCount {
		t.Errorf("Path counter = %d, want %d", count, expectedCount)
	}
}
