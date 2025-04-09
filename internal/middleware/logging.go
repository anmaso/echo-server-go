package middleware

import (
	"net/http"
	"time"

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

		// Log request details
		logger.Info("Request started: %s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		logger.Debug("Request headers: %v", r.Header)

		// Process request
		next.ServeHTTP(rw, r)

		// Log response details
		duration := time.Since(start)
		logger.Info("Request completed: %s %s %s status=%d size=%d duration=%v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			rw.status,
			rw.size,
			duration,
		)
	})
}
