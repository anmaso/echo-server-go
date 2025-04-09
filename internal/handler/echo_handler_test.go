package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"echo-server/internal/config"
)

func TestEchoHandler_ConfigLookup(t *testing.T) {
	// Create test configuration
	cfg := &config.ServerConfig{
		DefaultResponse: config.ResponseConfig{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		},
		PathMatcher: config.NewPathMatcher(),
	}

	// Add test path configuration
	testPathConfig := &config.PathConfig{
		Pattern: "^/test/.*",
		Methods: []string{"GET"},
		Response: config.ResponseConfig{
			StatusCode: http.StatusCreated,
			Headers: map[string]string{
				"X-Custom": "test",
			},
			Body:  `{"test": true}`,
			Delay: config.Duration{Duration: 100 * time.Millisecond},
		},
	}
	cfg.PathMatcher.Add(testPathConfig)

	handler := NewEchoHandler(cfg)

	tests := []struct {
		name          string
		path          string
		method        string
		wantStatus    int
		wantHeader    string
		wantHeaderVal string
	}{
		{
			name:          "matching path config",
			path:          "/test/123",
			method:        "GET",
			wantStatus:    http.StatusCreated,
			wantHeader:    "X-Custom",
			wantHeaderVal: "test",
		},
		{
			name:          "default config",
			path:          "/other",
			method:        "GET",
			wantStatus:    http.StatusOK,
			wantHeader:    "Content-Type",
			wantHeaderVal: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantStatus)
			}

			if got := w.Header().Get(tt.wantHeader); got != tt.wantHeaderVal {
				t.Errorf("header %s = %s, want %s", tt.wantHeader, got, tt.wantHeaderVal)
			}
		})
	}
}

func TestCustomStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		pathConfig config.PathConfig
		path       string
		method     string
		wantStatus int
	}{
		{
			name: "custom success status",
			pathConfig: config.PathConfig{
				Pattern: "^/created$",
				Methods: []string{"POST"},
				Response: config.ResponseConfig{
					StatusCode: http.StatusCreated,
				},
			},
			path:       "/created",
			method:     "POST",
			wantStatus: http.StatusCreated,
		},
		{
			name: "custom error status",
			pathConfig: config.PathConfig{
				Pattern: "^/forbidden$",
				Methods: []string{"GET"},
				Response: config.ResponseConfig{
					StatusCode: http.StatusForbidden,
				},
			},
			path:       "/forbidden",
			method:     "GET",
			wantStatus: http.StatusForbidden,
		},
		{
			name: "default status",
			pathConfig: config.PathConfig{
				Pattern:  "^/default$",
				Methods:  []string{"GET"},
				Response: config.ResponseConfig{},
			},
			path:       "/default",
			method:     "GET",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with test configuration
			cfg := &config.ServerConfig{
				DefaultResponse: config.ResponseConfig{
					StatusCode: http.StatusOK,
				},
				PathMatcher: config.NewPathMatcher(),
			}

			if err := cfg.PathMatcher.Add(&tt.pathConfig); err != nil {
				t.Fatalf("Failed to add path config: %v", err)
			}

			handler := NewEchoHandler(cfg)

			// Create test request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Handle request
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.wantStatus {
				t.Errorf("Status code = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
