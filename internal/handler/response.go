package handler

import (
	"encoding/json"
	"net/http"

	"echo-server/internal/config"
	"echo-server/internal/model"
	"echo-server/pkg/logger"
)

type Response struct {
	RequestData *model.RequestData `json:"request"`
	Message     string             `json:"message,omitempty"`
	Status      int                `json:"status"`
}

func formatResponse(data *model.RequestData, cfg *config.Response) *Response {
	status := http.StatusOK
	if cfg != nil && cfg.StatusCode != 0 {
		status = cfg.StatusCode
	}

	return &Response{
		RequestData: data,
		Status:      status,
		Message:     "Request processed successfully",
	}
}

func writeResponse(w http.ResponseWriter, resp *Response, cfg *config.Response) {
	// Set default headers
	w.Header().Set("Content-Type", "application/json")

	// Set configured headers
	if cfg != nil {
		for key, value := range cfg.Headers {
			w.Header().Set(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.Status)

	// Write response body
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
