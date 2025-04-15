package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"echo-server/internal/config"
	"echo-server/internal/server"
)

type testServer struct {
	srv    *server.Server
	client *http.Client
	port   int
}

func setupTestServer(t *testing.T) *testServer {
	port := 8081 // Use a different port than default for testing
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: port,
		DefaultResponse: config.ResponseConfig{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		},
		PathMatcher: config.NewPathMatcher(),
	}

	cm := config.NewConfigManager()
	cm.UpdateConfig(cfg)

	srv := server.New(cm)
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	return &testServer{
		srv:    srv,
		client: &http.Client{},
		port:   port,
	}
}

func (ts *testServer) cleanup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ts.srv.Stop(ctx); err != nil {
		t.Errorf("Error stopping server: %v", err)
	}
}

func (ts *testServer) configureEndpoint(t *testing.T, config map[string]interface{}) {
	body, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	resp, err := ts.client.Post(
		fmt.Sprintf("http://localhost:%d/config/paths", ts.port),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatalf("Failed to configure endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to configure endpoint, status: %d", resp.StatusCode)
	}
}

func TestEndpoints(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup(t)

	// Configure all endpoints
	endpoints := []map[string]interface{}{
		{
			"name":    "api",
			"pattern": "^/api$",
			"methods": []string{"GET"},
			"response": map[string]interface{}{
				"statusCode": 200,
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
				"includeRequest": false,
				"body":           `{"message":"API endpoint"}`,
			},
		},
		{
			"name":    "custom-json",
			"pattern": "^/custom-json$",
			"methods": []string{"POST"},
			"response": map[string]interface{}{
				"statusCode": 201,
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
				"includeRequest": false,
				"body":           `{"status":"created"}`,
			},
		},
		{
			"name":    "custom-text",
			"pattern": "^/custom-text$",
			"methods": []string{"GET"},
			"response": map[string]interface{}{
				"statusCode": 200,
				"headers": map[string]string{
					"Content-Type": "text/plain",
				},
				"body": "Hello, World!",
			},
		},
		{
			"name":    "delay",
			"pattern": "^/delay$",
			"methods": []string{"GET"},
			"response": map[string]interface{}{
				"statusCode": 200,
				"body":       `{"message":"delayed response"}`,
				"delay":      "500ms",
			},
		},
		{
			"name":    "error-every",
			"pattern": "^/error-every$",
			"methods": []string{"GET"},
			"response": map[string]interface{}{
				"statusCode": 200,
				"body":       `{"status":"ok"}`,
			},
			"errorResponse": map[string]interface{}{
				"statusCode": 500,
				"body":       `{"error":"periodic error"}`,
			},
			"errorEvery": 2,
		},
		{
			"name":    "error-test",
			"pattern": "^/error-test$",
			"methods": []string{"GET"},
			"response": map[string]interface{}{
				"statusCode": 200,
				"body":       `{"status":"ok"}`,
			},
			"errorResponse": map[string]interface{}{
				"statusCode": 503,
				"body":       `{"error":"test error"}`,
			},
			"errorEvery": 1,
		},
		{
			"name":    "proxy",
			"pattern": "^/proxy$",
			"methods": []string{"GET"},
			"response": map[string]interface{}{
				"proxy": "https://api.github.com/zen",
			},
		},
		{
			"name":    "limited",
			"pattern": "^/limited$",
			"methods": []string{"GET"},
			"response": map[string]interface{}{
				"statusCode": 200,
				"body":       `{"status":"ok"}`,
			},
			"rateLimit": map[string]interface{}{
				"requests": 2,
				"window":   "1s",
			},
		},
	}

	for _, endpoint := range endpoints {
		ts.configureEndpoint(t, endpoint)
	}

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		expectedBody   string
		preDelay       time.Duration
	}{
		{
			name:           "API endpoint",
			method:         "GET",
			path:           "/api",
			expectedStatus: 200,
			expectedBody:   `{"message":"API endpoint"}`,
		},
		{
			name:           "Custom JSON endpoint",
			method:         "POST",
			path:           "/custom-json",
			expectedStatus: 201,
			expectedBody:   `{"status":"created"}`,
		},
		{
			name:           "Custom text endpoint",
			method:         "GET",
			path:           "/custom-text",
			expectedStatus: 200,
			expectedBody:   "Hello, World!",
		},
		{
			name:           "Delayed endpoint",
			method:         "GET",
			path:           "/delay",
			expectedStatus: 200,
			expectedBody:   `{"message":"delayed response"}`,
		},
		{
			name:           "Error every endpoint (first request)",
			method:         "GET",
			path:           "/error-every",
			expectedStatus: 200,
			expectedBody:   `{"status":"ok"}`,
		},
		{
			name:           "Error every endpoint (second request)",
			method:         "GET",
			path:           "/error-every",
			expectedStatus: 500,
			expectedBody:   `{"error":"periodic error"}`,
		},
		{
			name:           "Error test endpoint",
			method:         "GET",
			path:           "/error-test",
			expectedStatus: 503,
			expectedBody:   `{"error":"test error"}`,
		},
		{
			name:           "Limited endpoint (first request)",
			method:         "GET",
			path:           "/limited",
			expectedStatus: 200,
			expectedBody:   `{"status":"ok"}`,
		},
		{
			name:           "Limited endpoint (second request)",
			method:         "GET",
			path:           "/limited",
			expectedStatus: 200,
			expectedBody:   `{"status":"ok"}`,
		},
		{
			name:           "Limited endpoint (third request - should be rate limited)",
			method:         "GET",
			path:           "/limited",
			expectedStatus: 429,
			expectedBody:   `{"error": "rate limit exceeded"}`,
		},
		{
			name:           "Limited endpoint (after waiting - should work)",
			method:         "GET",
			path:           "/limited",
			preDelay:       time.Second,
			expectedStatus: 200,
			expectedBody:   `{"status":"ok"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preDelay > 0 {
				time.Sleep(tt.preDelay)
			}

			var resp *http.Response
			var err error

			if tt.method == "GET" {
				resp, err = ts.client.Get(fmt.Sprintf("http://localhost:%d%s", ts.port, tt.path))
			} else {
				resp, err = ts.client.Post(
					fmt.Sprintf("http://localhost:%d%s", ts.port, tt.path),
					"application/json",
					bytes.NewBufferString(tt.body),
				)
			}

			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Status = %d, want %d", resp.StatusCode, tt.expectedStatus)
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			body := buf.String()

			if body != tt.expectedBody {
				t.Errorf("Body = %s, want %s", body, tt.expectedBody)
			}
		})
	}
}
