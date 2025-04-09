package counter

import (
	"sync"
	"sync/atomic"

	"echo-server/pkg/logger"
)

type Counter struct {
	count      uint64
	pathCounts map[string]*uint64
	mu         sync.RWMutex
}

var (
	globalCounter *Counter
	once          sync.Once
)

// GetGlobalCounter returns the singleton instance of the global counter
func GetGlobalCounter() *Counter {
	once.Do(func() {
		globalCounter = &Counter{
			pathCounts: make(map[string]*uint64),
		}
	})
	return globalCounter
}

// Increment atomically increments the global counter
func (c *Counter) Increment() uint64 {
	return atomic.AddUint64(&c.count, 1)
}

// IncrementPath atomically increments the counter for a specific path
func (c *Counter) IncrementPath(path string) uint64 {
	c.mu.Lock()
	if _, exists := c.pathCounts[path]; !exists {
		var count uint64
		c.pathCounts[path] = &count
	}
	c.mu.Unlock()

	return atomic.AddUint64(c.pathCounts[path], 1)
}

// GetCount returns the current global count
func (c *Counter) GetCount() uint64 {
	return atomic.LoadUint64(&c.count)
}

// GetPathCount returns the current count for a specific path
func (c *Counter) GetPathCount(path string) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if count, exists := c.pathCounts[path]; exists {
		return atomic.LoadUint64(count)
	}
	return 0
}

// GetAllPathCounts returns a map of all path counts
func (c *Counter) GetAllPathCounts() map[string]uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	counts := make(map[string]uint64, len(c.pathCounts))
	for path, count := range c.pathCounts {
		counts[path] = atomic.LoadUint64(count)
	}
	return counts
}

// Reset atomically resets the global counter and all path counters to zero
func (c *Counter) Reset() {
	atomic.StoreUint64(&c.count, 0)

	c.mu.Lock()
	c.pathCounts = make(map[string]*uint64)
	c.mu.Unlock()

	logger.Info("All counters reset to 0")
}

// ResetPath resets the counter for a specific path
func (c *Counter) ResetPath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if count, exists := c.pathCounts[path]; exists {
		atomic.StoreUint64(count, 0)
		logger.Info("Counter reset for path: %s", path)
	}
}

// ResetAll resets both global and path-specific counters
func (c *Counter) ResetAll() {
	atomic.StoreUint64(&c.count, 0)

	c.mu.Lock()
	c.pathCounts = make(map[string]*uint64)
	c.mu.Unlock()

	logger.Info("All counters reset")
}
