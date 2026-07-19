package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"echo-server/internal/requestlog"
)

const maxCaptureBytes = 64 * 1024

type capturingRW struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
	wrote  bool
}

func newCapturingRW(w http.ResponseWriter) *capturingRW {
	return &capturingRW{ResponseWriter: w, status: http.StatusOK}
}

func (c *capturingRW) WriteHeader(code int) {
	if !c.wrote {
		c.status = code
		c.wrote = true
		c.ResponseWriter.WriteHeader(code)
	}
}

func (c *capturingRW) Write(b []byte) (int, error) {
	if c.body.Len() < maxCaptureBytes {
		remaining := maxCaptureBytes - c.body.Len()
		if len(b) > remaining {
			c.body.Write(b[:remaining])
		} else {
			c.body.Write(b)
		}
	}
	return c.ResponseWriter.Write(b)
}

func RequestCaptureMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var reqBody string
			if r.Body != nil && r.Body != http.NoBody {
				bodyBytes, err := io.ReadAll(r.Body)
				_ = r.Body.Close()
				if err == nil {
					if len(bodyBytes) > maxCaptureBytes {
						reqBody = string(bodyBytes[:maxCaptureBytes]) + " [truncated]"
					} else {
						reqBody = string(bodyBytes)
					}
					r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				}
			}

			crw := newCapturingRW(w)
			next.ServeHTTP(crw, r)

			entry := &requestlog.Entry{
				Timestamp:       start,
				Method:          r.Method,
				Path:            r.URL.Path,
				Query:           r.URL.RawQuery,
				RequestHeaders:  map[string][]string(r.Header),
				RequestBody:     reqBody,
				StatusCode:      crw.status,
				ResponseHeaders: map[string][]string(crw.Header()),
				ResponseBody:    crw.body.String(),
				DurationMs:      float64(time.Since(start).Nanoseconds()) / 1e6,
			}
			requestlog.Global().Add(entry)
		})
	}
}
