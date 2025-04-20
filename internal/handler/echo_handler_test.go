package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"echo-server/internal/config"
	"echo-server/internal/middleware"
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
				Response: config.NewResponseConfig(),
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

func TestResponseDelay(t *testing.T) {
	tests := []struct {
		name        string
		pathConfig  config.PathConfig
		path        string
		wantMinTime time.Duration
	}{
		{
			name: "with delay",
			pathConfig: config.PathConfig{
				Pattern: "^/delay$",
				Methods: []string{"GET"},
				Response: config.ResponseConfig{
					Delay: config.Duration{Duration: 100 * time.Millisecond},
				},
			},
			path:        "/delay",
			wantMinTime: 100 * time.Millisecond,
		},
		{
			name: "no delay",
			pathConfig: config.PathConfig{
				Pattern: "^/nodelay$",
				Methods: []string{"GET"},
				Response: config.ResponseConfig{
					Delay: config.Duration{Duration: 0},
				},
			},
			path:        "/nodelay",
			wantMinTime: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with test configuration
			cfg := &config.ServerConfig{
				DefaultResponse: config.ResponseConfig{},
				PathMatcher:     config.NewPathMatcher(),
			}

			if err := cfg.PathMatcher.Add(&tt.pathConfig); err != nil {
				t.Fatalf("Failed to add path config: %v", err)
			}

			handler := NewEchoHandler(cfg)

			// Create test request
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			// Measure response time
			start := time.Now()
			handler.ServeHTTP(w, req)
			duration := time.Since(start)

			// Check if delay was respected
			if duration < tt.wantMinTime {
				t.Errorf("Response time %v was shorter than configured delay %v",
					duration, tt.wantMinTime)
			}
		})
	}
}

func TestCustomResponseBody(t *testing.T) {
	tests := []struct {
		name       string
		pathConfig config.PathConfig
		path       string
		body       string
		want       string
	}{
		{
			name: "static json response",
			pathConfig: config.PathConfig{
				Pattern: "^/json$",
				Response: config.ResponseConfig{
					Body: `{"message":"hello"}`,
				},
			},
			path: "/json",
			want: `{"message":"hello"}`,
		},
		{
			name: "template response",
			pathConfig: config.PathConfig{
				Pattern: "^/template$",
				Response: config.ResponseConfig{
					Body: `template:{"path":"{{.Path}}{{.Method}}"}`,
				},
			},
			path: "/template",
			want: `{"path":"/templateGET"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{
				PathMatcher: config.NewPathMatcher(),
			}

			if err := cfg.PathMatcher.Add(&tt.pathConfig); err != nil {
				t.Fatalf("Failed to add path config: %v", err)
			}

			handler := NewEchoHandler(cfg)
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			body := strings.Trim(w.Body.String(), "\n")
			if body != tt.want {
				t.Errorf("Response body = %q, want %q", body, tt.want)
			}
		})
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name         string
		pathConfig   config.PathConfig
		path         string
		wantStatus   int
		wantError    bool
		wantErrorStr string
	}{
		{
			name: "error response",
			pathConfig: config.PathConfig{
				Pattern: "^/error$",
				Methods: []string{"GET"},
				ErrorResponse: &config.ResponseConfig{
					StatusCode: http.StatusInternalServerError,
					Body:       `{"message":"error occurred"}`,
				},
				ErrorEvery: 1,
			},
			path:         "/error",
			wantStatus:   http.StatusInternalServerError,
			wantError:    true,
			wantErrorStr: `{"message":"error occurred"}`,
		},
		{
			name: "normal response",
			pathConfig: config.PathConfig{
				Pattern: "^/normal$",
				Methods: []string{"GET"},
				Response: config.ResponseConfig{
					StatusCode: http.StatusOK,
					Body:       `{"message":"ok"}`,
				},
			},
			path:       "/normal",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{
				PathMatcher: config.NewPathMatcher(),
			}

			if err := cfg.PathMatcher.Add(&tt.pathConfig); err != nil {
				t.Fatalf("Failed to add path config: %v", err)
			}

			handler := NewEchoHandler(cfg)
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			middleware.RequestLoggingHandler()(handler).ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("Status code = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantError {
				body := strings.Trim(w.Body.String(), "\n")
				if body != tt.wantErrorStr {
					t.Errorf("Response body = %q, want %q", body, tt.wantErrorStr)
				}
			}
		})
	}
}

func TestErrorEvery(t *testing.T) {
	pathConfig := config.PathConfig{
		Pattern: "^/error-every$",
		Methods: []string{"GET"},
		Response: config.ResponseConfig{
			StatusCode: http.StatusOK,
			Body:       `{"status":"ok"}`,
		},
		ErrorResponse: &config.ResponseConfig{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error":"periodic error"}`,
		},
		ErrorEvery: 3, // Error every 3rd request
	}

	cfg := &config.ServerConfig{
		PathMatcher: config.NewPathMatcher(),
	}

	if err := cfg.PathMatcher.Add(&pathConfig); err != nil {
		t.Fatalf("Failed to add path config: %v", err)
	}

	handler := NewEchoHandler(cfg)

	tests := []struct {
		name      string
		wantError bool
	}{
		{"first request", false},
		{"second request", false},
		{"third request", true}, // Should error
		{"fourth request", false},
		{"fifth request", false},
		{"sixth request", true}, // Should error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/error-every", nil)
			w := httptest.NewRecorder()

			middleware.RequestLoggingHandler()(handler).ServeHTTP(w, req)

			status := w.Result().StatusCode

			if tt.wantError && status != http.StatusInternalServerError {
				t.Errorf("Error = %v, want %v", status, tt.wantError)
			}
		})
	}
}

func TestTemplateProcessing(t *testing.T) {
	tests := []struct {
		name       string
		pathConfig config.PathConfig
		path       string
		method     string
		body       string
		headers    map[string]string
		want       string
	}{
		{
			name: "template with request data",
			pathConfig: config.PathConfig{
				Pattern: "^/template-test$",
				Methods: []string{"POST"},
				Response: config.ResponseConfig{
					Body: `template:{"method":"{{.Method}}","path":"{{.Path}}","headers":{"X-Test":"{{index .Headers "X-Test"}}"}}`,
				},
			},
			path:    "/template-test",
			method:  "POST",
			body:    "test body",
			headers: map[string]string{"X-Test": "test-value"},
			want:    `{"method":"POST","path":"/template-test","headers":{"X-Test":"test-value"}}`,
		},
		{
			name: "invalid template",
			pathConfig: config.PathConfig{
				Pattern: "^/invalid-template$",
				Methods: []string{"GET"},
				Response: config.ResponseConfig{
					Body: `template:{{.InvalidField}}`,
				},
			},
			path:   "/invalid-template",
			method: "GET",
			want:   "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{
				PathMatcher: config.NewPathMatcher(),
			}

			if err := cfg.PathMatcher.Add(&tt.pathConfig); err != nil {
				t.Fatalf("Failed to add path config: %v", err)
			}

			handler := NewEchoHandler(cfg)
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.name == "invalid template" {
				if w.Code != http.StatusInternalServerError {
					t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
				}
			} else {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.want {
					t.Errorf("Response body = %q, want %q", body, tt.want)
				}
			}
		})
	}
}

func TestIncludeRequest(t *testing.T) {
	tests := []struct {
		name                  string
		pathConfig            config.PathConfig
		path                  string
		method                string
		requestBody           string
		includeRequest        bool
		wantRequestInResponse bool
	}{
		{
			name: "include request enabled",
			pathConfig: config.PathConfig{
				Pattern: "^/include-request$",
				Methods: []string{"POST"},
				Response: config.ResponseConfig{
					Body:           `{"message":"test"}`,
					IncludeRequest: true,
				},
			},
			path:                  "/include-request",
			method:                "POST",
			requestBody:           `{"test":"data"}`,
			includeRequest:        true,
			wantRequestInResponse: true,
		},
		{
			name: "include request disabled",
			pathConfig: config.PathConfig{
				Pattern: "^/no-request$",
				Methods: []string{"POST"},
				Response: config.ResponseConfig{
					Body:           `{"message":"test"}`,
					IncludeRequest: false,
				},
			},
			path:                  "/no-request",
			method:                "POST",
			requestBody:           `{"test":"data"}`,
			includeRequest:        false,
			wantRequestInResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{
				PathMatcher: config.NewPathMatcher(),
			}

			if err := cfg.PathMatcher.Add(&tt.pathConfig); err != nil {
				t.Fatalf("Failed to add path config: %v", err)
			}

			handler := NewEchoHandler(cfg)
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			_, hasRequest := response["request"]
			if hasRequest != tt.wantRequestInResponse {
				t.Errorf("Response has request = %v, want %v", hasRequest, tt.wantRequestInResponse)
			}
		})
	}
}

func TestProxyFunctionality(t *testing.T) {
	oldEnv := os.Getenv("ENABLE_PROXY")
	defer os.Setenv("ENABLE_PROXY", oldEnv)
	os.Setenv("ENABLE_PROXY", "true")

	tests := []struct {
		name       string
		pathConfig config.PathConfig
		path       string
		method     string
		wantStatus int
	}{
		{
			name: "proxy enabled with valid URL",
			pathConfig: config.PathConfig{
				Pattern: "^/proxy$",
				Methods: []string{"GET"},
				Proxy: &config.ProxyConfig{
					URL: "http://localhost:8081/test",
				},
			},
			path:       "/proxy",
			method:     "GET",
			wantStatus: http.StatusOK,
		},
		{
			name: "proxy with invalid URL",
			pathConfig: config.PathConfig{
				Pattern: "^/invalid-proxy$",
				Methods: []string{"GET"},
				Proxy: &config.ProxyConfig{
					URL: "http://invalid-host:9999/test",
				},
			},
			path:       "/invalid-proxy",
			method:     "GET",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{
				PathMatcher: config.NewPathMatcher(),
			}

			if err := cfg.PathMatcher.Add(&tt.pathConfig); err != nil {
				t.Fatalf("Failed to add path config: %v", err)
			}

			handler := NewEchoHandler(cfg)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status code = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestContentTypeHeaders(t *testing.T) {
	tests := []struct {
		name            string
		pathConfig      config.PathConfig
		path            string
		method          string
		wantContentType string
	}{
		{
			name: "explicit JSON content type",
			pathConfig: config.PathConfig{
				Pattern: "^/json$",
				Methods: []string{"GET"},
				Response: config.ResponseConfig{
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: `{"message":"test"}`,
				},
			},
			path:            "/json",
			method:          "GET",
			wantContentType: "application/json",
		},
		{
			name: "explicit text content type",
			pathConfig: config.PathConfig{
				Pattern: "^/text$",
				Methods: []string{"GET"},
				Response: config.ResponseConfig{
					Headers: map[string]string{
						"Content-Type": "text/plain",
					},
					Body: "Hello, world!",
				},
			},
			path:            "/text",
			method:          "GET",
			wantContentType: "text/plain",
		},
		{
			name: "default content type",
			pathConfig: config.PathConfig{
				Pattern: "^/default$",
				Methods: []string{"GET"},
				Response: config.ResponseConfig{
					Body: `{"message":"test"}`,
				},
			},
			path:            "/default",
			method:          "GET",
			wantContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{
				PathMatcher: config.NewPathMatcher(),
			}

			if err := cfg.PathMatcher.Add(&tt.pathConfig); err != nil {
				t.Fatalf("Failed to add path config: %v", err)
			}

			handler := NewEchoHandler(cfg)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if contentType := w.Header().Get("Content-Type"); contentType != tt.wantContentType {
				t.Errorf("Content-Type = %q, want %q", contentType, tt.wantContentType)
			}
		})
	}
}
