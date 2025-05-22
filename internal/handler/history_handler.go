package handler

import (
	"encoding/json"
	"io" // Added for io.EOF
	"net/http"
	"strings"

	"echo-server/internal/config"
	"echo-server/internal/model"
	"echo-server/pkg/logger"
)

type HistoryHandler struct {
	historyStorage *model.HistoryStorage
	configManager  *config.ConfigManager // To potentially interact with global server config if needed
}

func NewHistoryHandler(hs *model.HistoryStorage, cm *config.ConfigManager) *HistoryHandler {
	return &HistoryHandler{
		historyStorage: hs,
		configManager:  cm,
	}
}

// ServeHTTP routes requests to the appropriate internal handler method.
func (h *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/history")
	path = strings.TrimSuffix(path, "/")

	logger.Debug("History handler received path: %s, method: %s", path, r.Method)

	switch {
	case path == "/start" && r.Method == http.MethodPost:
		h.handleStartRecording(w, r)
	case path == "/stop" && r.Method == http.MethodPost:
		h.handleStopRecording(w, r)
	case path == "/config" && r.Method == http.MethodPut:
		h.handleConfigureHistory(w, r)
	case path == "" && r.Method == http.MethodGet:
		h.handleGetHistory(w, r)
	case path == "" && r.Method == http.MethodDelete:
		h.handleClearHistory(w, r)
	default:
		logger.Warn("History handler: Path not found or method not allowed: /history%s", path)
		http.NotFound(w, r)
	}
}

type HistoryStatusResponse struct {
	RecordingActive bool `json:"recordingActive"`
	MaxSize         int  `json:"maxSize"`
}

type ConfigureHistoryRequest struct {
	MaxSize int `json:"maxSize"`
}

func (h *HistoryHandler) handleStartRecording(w http.ResponseWriter, r *http.Request) {
	h.historyStorage.StartRecording()
	logger.Info("History recording started.")

	// Optionally allow setting maxSize on start
	var req ConfigureHistoryRequest
	if r.Body != http.NoBody {
		// Attempt to decode, but don't error out if body is empty or not JSON
		// This makes setting maxSize on start purely optional
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&req); err == nil { // Check if decoding was successful
			if req.MaxSize > 0 { // Basic validation if MaxSize was provided
				h.historyStorage.SetMaxSize(req.MaxSize)
				logger.Info("History max size set to %d on start.", req.MaxSize)
			} else if req.MaxSize == 0 { // Explicitly setting to 0
				h.historyStorage.SetMaxSize(0)
				logger.Info("History max size set to 0 on start. No new entries will be stored.")
            } else { // MaxSize < 0
				h.historyStorage.SetMaxSize(req.MaxSize) // relies on SetMaxSize to handle negative as 0
				logger.Info("History max size set to %d on start (will be treated as 0 by storage).", req.MaxSize)
			}
		} else if err != io.EOF && !strings.Contains(err.Error(), "unexpected EOF") && !strings.Contains(err.Error(), "EOF") { // Log unexpected error, ignore expected EOF variations
            logger.Warnf("Error decoding request body for /history/start: %v", err)
        }
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HistoryStatusResponse{
		RecordingActive: h.historyStorage.IsRecordingActive(),
		MaxSize:         h.historyStorage.GetMaxSize(),
	})
}

func (h *HistoryHandler) handleStopRecording(w http.ResponseWriter, r *http.Request) {
	h.historyStorage.StopRecording()
	logger.Info("History recording stopped.")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HistoryStatusResponse{
		RecordingActive: h.historyStorage.IsRecordingActive(),
		MaxSize:         h.historyStorage.GetMaxSize(),
	})
}

func (h *HistoryHandler) handleConfigureHistory(w http.ResponseWriter, r *http.Request) {
	var req ConfigureHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// HistoryStorage.SetMaxSize handles newSize < 0 by setting it to 0.
	// So, no specific check for req.MaxSize < 0 is strictly needed here unless 
	// we want to return a different error message for negative values.
	// Current HistoryStorage logic: newSize < 0 becomes 0. newSize == 0 clears entries.
	h.historyStorage.SetMaxSize(req.MaxSize)
	if req.MaxSize < 0 {
		logger.Info("History max size configured to %d (will be treated as 0 by storage).", req.MaxSize)
	} else {
		logger.Info("History max size configured to %d.", req.MaxSize)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HistoryStatusResponse{
		RecordingActive: h.historyStorage.IsRecordingActive(),
		MaxSize:         h.historyStorage.GetMaxSize(),
	})
}

func (h *HistoryHandler) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	entries := h.historyStorage.GetEntries()
	logger.Info("Retrieved %d history entries.", len(entries))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		logger.Error("Failed to encode history entries: %v", err)
		http.Error(w, "Failed to encode history", http.StatusInternalServerError)
	}
}

func (h *HistoryHandler) handleClearHistory(w http.ResponseWriter, r *http.Request) {
	h.historyStorage.ClearEntries()
	logger.Info("History cleared.")
	w.WriteHeader(http.StatusNoContent)
}
