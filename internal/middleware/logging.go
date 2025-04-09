package middleware

import (
	"net/http"
	"time"

	"echo-server/internal/counter"
	"echo-server/pkg/logger"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	size        int64
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += int64(n)
	return n, err
}

func RequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		// Get counter instance
		c := counter.GetGlobalCounter()

		// Increment both global and path-specific counters
		globalCount := c.Increment()
		pathCount := c.IncrementPath(r.URL.Path)

		// Log request details with counter information
		logger.Info("Request #%d (path #%d) started: %s %s %s",
			globalCount,
			pathCount,
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
		)

		// Process request
		next.ServeHTTP(rw, r)

		// Log completion
		duration := time.Since(start)
		logger.Info("Request #%d completed: %s %s %s status=%d size=%d duration=%v",
			globalCount,
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			rw.status,
			rw.size,
			duration,
		)
	})
}
