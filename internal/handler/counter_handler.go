package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"echo-server/internal/counter"
	"echo-server/pkg/logger"
)

type CounterResponse struct {
	GlobalCount uint64            `json:"globalCount"`
	PathCounts  map[string]uint64 `json:"pathCounts,omitempty"`
}

func CounterHandler(w http.ResponseWriter, r *http.Request) {
	c := counter.GetGlobalCounter()

	switch r.Method {
	case http.MethodGet:
		// Return counter values
		response := CounterResponse{
			GlobalCount: c.GetCount(),
			PathCounts:  c.GetAllPathCounts(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Failed to encode counter response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

	case http.MethodDelete:
		// Handle reset operations
		path := r.URL.Query().Get("path")
		if path != "" {
			c.ResetPath(strings.TrimSpace(path))
			logger.Info("Reset counter for path: %s", path)
		} else {
			c.Reset()
			logger.Info("Reset all counters")
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
