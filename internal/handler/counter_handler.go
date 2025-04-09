package handler

import (
	"encoding/json"
	"net/http"

	"echo-server/internal/counter"
	"echo-server/pkg/logger"
)

type CounterResponse struct {
	GlobalCount uint64            `json:"globalCount"`
	PathCounts  map[string]uint64 `json:"pathCounts,omitempty"`
}

func CounterHandler(w http.ResponseWriter, r *http.Request) {
	c := counter.GetGlobalCounter()

	response := CounterResponse{
		GlobalCount: c.GetCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode counter response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
