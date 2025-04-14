package middleware

import (
	"net/http"
	"sync"
	"time"

	"echo-server/internal/config"
	"echo-server/pkg/logger"
)

type rateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
}

func NewRateLimiter() *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
	}
}

// cleanup removes expired timestamps for a given key
func (rl *rateLimiter) cleanup(key string, window time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if timestamps, exists := rl.requests[key]; exists {
		now := time.Now()
		valid := 0
		for _, ts := range timestamps {
			if now.Sub(ts) <= window {
				timestamps[valid] = ts
				valid++
			}
		}
		if valid > 0 {
			rl.requests[key] = timestamps[:valid]
		} else {
			delete(rl.requests, key)
		}
	}
}

// isAllowed checks if a request is allowed based on rate limiting configuration
func (rl *rateLimiter) isAllowed(key string, cfg *config.RateLimitConfig) bool {
	if cfg == nil {
		return true
	}

	rl.cleanup(key, cfg.Window.Duration)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	timestamps := rl.requests[key]
	if len(timestamps) < cfg.Requests {
		rl.requests[key] = append(timestamps, time.Now())
		return true
	}
	return false
}

// RateLimit middleware function that checks rate limits based on path configuration
func RateLimit(pathMatcher config.PathMatcher) func(http.Handler) http.Handler {
	limiter := NewRateLimiter()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pathConfig, matched := pathMatcher.Match(r.URL.Path, r.Method)
			if !matched || pathConfig.RateLimit == nil {
				next.ServeHTTP(w, r)
				return
			}

			if !limiter.isAllowed(pathConfig.Name, pathConfig.RateLimit) {
				logger.Debug("Rate limit exceeded for %s on path %s", pathConfig.Name, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "rate limit exceeded"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
