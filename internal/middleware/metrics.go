// middleware/metrics.go
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"echo-server/metrics"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rww, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		handlerName := r.URL.Path
		status := strconv.Itoa(rww.statusCode)

		metrics.RequestCounter.WithLabelValues(handlerName, r.Method, status).Inc()
		metrics.RequestDuration.WithLabelValues(handlerName, r.Method).Observe(duration)
	})
}

// responseWriterWrapper captures the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rww *responseWriterWrapper) WriteHeader(code int) {
	rww.statusCode = code
	rww.ResponseWriter.WriteHeader(code)
}
