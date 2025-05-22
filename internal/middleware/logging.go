package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"echo-server/internal/counter"
	"echo-server/internal/model" // Added for history
	"echo-server/pkg/logger"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	size        int64
	body        *bytes.Buffer // Added to capture response body
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
		body:           new(bytes.Buffer), // Initialize buffer
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
	// Also write to our internal buffer
	if rw.body != nil {
		rw.body.Write(b)
	}
	return n, err
}

// LoggingAndHistoryHandler combines logging and history recording.
// It takes HistoryStorage as a dependency.
func LoggingAndHistoryHandler(hs *model.HistoryStorage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var requestBodyBytes []byte
			if r.Body != nil && r.Body != http.NoBody { // Check for http.NoBody as well
				var err error
				requestBodyBytes, err = io.ReadAll(r.Body)
				if err != nil {
					logger.Error("Failed to read request body for history: %v", err)
				}
				err = r.Body.Close() // Close the original body
				if err != nil {
					logger.Error("Failed to close request body: %v", err)
				}
				r.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes)) // Replace body with a new reader
			}

			rw := newResponseWriter(w)

			c := counter.GetGlobalCounter()
			globalCount := c.Increment()
			pathCount := c.IncrementPath(r.URL.Path)

			logger.Info("Request #%d (path #%d) started: %s %s %s",
				globalCount, pathCount, r.RemoteAddr, r.Method, r.URL.Path)

			next.ServeHTTP(rw, r) // Process request

			duration := time.Since(start)
			logger.Info("Request #%d completed: %s %s %s status=%d size=%d duration=%v",
				globalCount, r.RemoteAddr, r.Method, r.URL.Path, rw.status, rw.size, duration)

			// Record history if active
			if hs != nil && hs.IsRecordingActive() {
				// Re-assign r.Body with the original content for ExtractRequestData.
				// This ensures that ExtractRequestData gets the full body if it was read before.
				if requestBodyBytes != nil {
					r.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))
				}
				// If requestBodyBytes is nil (e.g. GET request or error during read), 
				// r.Body might be http.NoBody or the original empty body which is fine.
				
				reqData, err := model.ExtractRequestData(r)
				if err != nil {
					logger.Error("Failed to extract request data for history: %v", err)
				} else {
					respHeaders := make(map[string]string)
					for k, v := range rw.Header() {
						if len(v) > 0 {
							respHeaders[k] = v[0] 
						}
					}

					entry := model.HistoryEntry{
						Timestamp:    time.Now(),
						Request:      *reqData,
						Response: model.ResponseSummary{
							StatusCode: rw.status,
							Headers:    respHeaders,
							Body:       rw.body.String(),
						},
						ResponseSize: rw.size,
					}
					hs.AddEntry(entry)
					logger.Debug("Added entry to history for request path: %s", reqData.Path)
				}
			}
		})
	}
}

// RequestLoggingHandler remains for places where only logging is needed.
func RequestLoggingHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			// Using the new responseWriter which has body buffer, 
			// but it won't be used by this specific handler for history.
			rw := newResponseWriter(w) 

			c := counter.GetGlobalCounter()
			globalCount := c.Increment()
			pathCount := c.IncrementPath(r.URL.Path)

			logger.Info("Request #%d (path #%d) started: %s %s %s",
				globalCount,
				pathCount,
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)

			next.ServeHTTP(rw, r)

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
}
