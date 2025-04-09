package handler

import (
	"encoding/json"
	"net/http"
	"sync"

	"echo-server/internal/config"
	"echo-server/internal/counter"
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

func (h *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := counter.GetGlobalCounter()

	// Extract request data
	data, err := model.ExtractRequestData(r)
	if err != nil {
		logger.Error("Failed to extract request data: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Add counter information
	data.Counter = model.CounterInfo{
		Global: c.GetCount(),
		Path:   c.GetPathCount(r.URL.Path),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Send response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
