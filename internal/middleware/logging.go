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

		// Increment global counter
		count := counter.GetGlobalCounter().Increment()

		// Log request details with counter
		logger.Info("Request #%d started: %s %s %s", count, r.RemoteAddr, r.Method, r.URL.Path)
		logger.Debug("Request headers: %v", r.Header)

		// Process request
		next.ServeHTTP(rw, r)

		// Log response details
		duration := time.Since(start)
		logger.Info("Request #%d completed: %s %s %s status=%d size=%d duration=%v",
			count,
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			rw.status,
			rw.size,
			duration,
		)
	})
}
