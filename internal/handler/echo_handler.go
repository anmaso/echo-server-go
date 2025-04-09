package handler

import (
	"encoding/json"
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

func NewEchoHandler(cfg *config.ServerConfig) *EchoHandler {
	return &EchoHandler{
		config: cfg,
	}
}

func (h *EchoHandler) handleResponse(w http.ResponseWriter, r *http.Request, data *model.RequestData) {
	// Look up path configuration
	pathConfig, matched := h.config.PathMatcher.Match(r.URL.Path, r.Method)

	// Determine which response config to use
	var responseConfig config.ResponseConfig
	if matched {
		responseConfig = pathConfig.Response
	} else {
		responseConfig = h.config.DefaultResponse
	}

	// Set response headers
	for key, value := range responseConfig.Headers {
		w.Header().Set(key, value)
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	// Prepare response body
	response := struct {
		Request  *model.RequestData `json:"request"`
		Response interface{}        `json:"response,omitempty"`
		Status   int                `json:"status"`
	}{
		Request: data,
		Status:  responseConfig.StatusCode,
	}

	// Add configured response body if specified
	if responseConfig.Body != "" {
		var responseBody interface{}
		if err := json.Unmarshal([]byte(responseConfig.Body), &responseBody); err != nil {
			response.Response = responseConfig.Body // Use as string if not valid JSON
		} else {
			response.Response = responseBody
		}
	}

	// Set status code from configuration
	statusCode := responseConfig.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	w.WriteHeader(statusCode)

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	h.handleResponse(w, r, data)
}
