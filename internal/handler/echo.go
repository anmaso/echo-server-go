package handler

import (
	"net/http"
	"sync"

	"echo-server/internal/config"
	"echo-server/internal/model"
	"echo-server/pkg/logger"
)

type EchoHandler struct {
	config *config.ServerConfig
	mu     sync.RWMutex
}

func New(cfg *config.ServerConfig) *EchoHandler {
	if cfg == nil {
		cfg = config.Default()
	}
	return &EchoHandler{
		config: cfg,
	}
}

/*
func formatResponse(data *model.RequestData, defaultResponse *config.DefaultResponse) *model.Response {
	// Implement response formatting logic here
	return &model.Response{
		Method: data.Method,
		Path:   data.Path,
		Body:   data.Body,
	}
}

func writeResponse(w http.ResponseWriter, resp *model.Response, defaultResponse *config.DefaultResponse) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
*/

func (h *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Extract request data
	data, err := model.ExtractRequestData(r)
	if err != nil {
		logger.Error("Failed to extract request data: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Format response using default configuration
	resp := formatResponse(data, &h.config.DefaultResponse)

	// Write response
	writeResponse(w, resp, &h.config.DefaultResponse)
}
