package counter

import (
	"sync"
	"sync/atomic"

	"echo-server/pkg/logger"
)

type Counter struct {
	globalCount uint64
	pathCounts  sync.Map
	mu          sync.RWMutex
}

var (
	globalCounter *Counter
	once          sync.Once
)

func GetGlobalCounter() *Counter {
	once.Do(func() {
		globalCounter = &Counter{}
	})
	return globalCounter
}

func (c *Counter) Increment() uint64 {
	return atomic.AddUint64(&c.globalCount, 1)
}

func (c *Counter) GetCount() uint64 {
	return atomic.LoadUint64(&c.globalCount)
}

func (c *Counter) IncrementPath(path string) uint64 {
	var count uint64
	actual, _ := c.pathCounts.LoadOrStore(path, &count)
	return atomic.AddUint64(actual.(*uint64), 1)
}

func (c *Counter) GetPathCount(path string) uint64 {
	if count, ok := c.pathCounts.Load(path); ok {
		return atomic.LoadUint64(count.(*uint64))
	}
	return 0
}

func (c *Counter) GetAllPathCounts() map[string]uint64 {
	counts := make(map[string]uint64)
	c.pathCounts.Range(func(key, value interface{}) bool {
		counts[key.(string)] = atomic.LoadUint64(value.(*uint64))
		return true
	})
	return counts
}

func (c *Counter) ResetPath(path string) {
	if count, ok := c.pathCounts.Load(path); ok {
		atomic.StoreUint64(count.(*uint64), 0)
		logger.Info("Reset counter for path: %s", path)
	}
}

func (c *Counter) Reset() {
	atomic.StoreUint64(&c.globalCount, 0)
	c.pathCounts.Range(func(key, value interface{}) bool {
		c.pathCounts.Delete(key)
		return true
	})
	logger.Info("Reset all counters")
}
