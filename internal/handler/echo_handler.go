package handler

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"echo-server/internal/config"
	"echo-server/internal/model"
	"echo-server/pkg/logger"
)

type EchoHandler struct {
	config *config.ServerConfig
	mu     sync.RWMutex
}

func New(cfg *config.ServerConfig) *EchoHandler {
	return &EchoHandler{
		config: cfg,
	}
}

func (h *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract request data
	data, err := model.ExtractRequestData(r)
	if err != nil {
		logger.Error("Failed to extract request data: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Look up path configuration
	pathConfig, matched := h.config.PathMatcher.Match(r.URL.Path, r.Method)

	// Determine which response config to use
	var responseConfig config.ResponseConfig
	if matched {
		responseConfig = pathConfig.Response
	} else {
		responseConfig = h.config.DefaultResponse
	}

	// Apply configured delay if any
	if responseConfig.Delay.Duration > 0 {
		time.Sleep(responseConfig.Delay.Duration)
	}

	// Set configured headers
	for key, value := range responseConfig.Headers {
		w.Header().Set(key, value)
	}

	// Create response
	response := struct {
		Request  *model.RequestData `json:"request"`
		Response interface{}        `json:"response,omitempty"`
	}{
		Request: data,
	}

	// Add configured response body if specified
	if responseConfig.Body != "" {
		var responseBody interface{}
		if err := json.Unmarshal([]byte(responseConfig.Body), &responseBody); err != nil {
			response.Response = responseConfig.Body
		} else {
			response.Response = responseBody
		}
	}

	// Set status code and write response
	w.WriteHeader(responseConfig.StatusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
