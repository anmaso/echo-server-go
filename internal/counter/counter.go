package counter

import (
	"sync"
	"sync/atomic"

	"echo-server/pkg/logger"
)

type Counter struct {
	count      uint64
	pathCounts map[string]uint64
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
			pathCounts: make(map[string]uint64),
		}
	})
	return globalCounter
}

// Increment atomically increments the global counter
func (c *Counter) Increment() uint64 {
	return atomic.AddUint64(&c.count, 1)
}

// GetCount returns the current global count
func (c *Counter) GetCount() uint64 {
	return atomic.LoadUint64(&c.count)
}

// Reset atomically resets the global counter to zero
func (c *Counter) Reset() {
	atomic.StoreUint64(&c.count, 0)
	logger.Info("Global counter reset to 0")
}
