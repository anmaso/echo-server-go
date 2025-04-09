package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"echo-server/internal/config"
)

func TestConfigHandler(t *testing.T) {
	cm := config.NewConfigManager()
	handler := NewConfigurationHandler(cm)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "get configurations",
			method:     "GET",
			path:       "/config/paths",
			wantStatus: http.StatusOK,
		},
		{
			name:   "add new configuration",
			method: "POST",
			path:   "/config/paths",
			body: `{
                "pattern": "/test/.*",
                "methods": ["GET"],
                "response": {
                    "statusCode": 200,
                    "body": "{\"status\":\"ok\"}"
                }
            }`,
			wantStatus: http.StatusCreated,
		},
		{
			name:   "update configuration",
			method: "PUT",
			path:   "/config/paths/test",
			body: `{
                "methods": ["GET", "POST"],
                "response": {
                    "statusCode": 200,
                    "body": "{\"status\":\"updated\"}"
                }
            }`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
