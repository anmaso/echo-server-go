package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"echo-server/internal/requestlog"
)

type LogHandler struct {
	buf *requestlog.Buffer
}

func NewLogHandler() *LogHandler {
	return &LogHandler{buf: requestlog.Global()}
}

func (h *LogHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h.buf.All()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *LogHandler) HandleClear(w http.ResponseWriter, _ *http.Request) {
	h.buf.Clear()
	w.WriteHeader(http.StatusNoContent)
}

func (h *LogHandler) HandleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"bufferSize": h.buf.Size()})
	case http.MethodPut:
		var body struct {
			BufferSize int `json:"bufferSize"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.BufferSize < 1 {
			http.Error(w, `{"error":"bufferSize must be a positive integer"}`, http.StatusBadRequest)
			return
		}
		h.buf.SetSize(body.BufferSize)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *LogHandler) HandleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Disable write deadline so the SSE connection lives indefinitely.
	if err := http.NewResponseController(w).SetWriteDeadline(time.Time{}); err != nil {
		http.Error(w, "failed to disable write deadline", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var lastID int64
	if idStr := r.Header.Get("Last-Event-ID"); idStr != "" {
		lastID, _ = strconv.ParseInt(idStr, 10, 64)
	}

	for _, entry := range h.buf.All() {
		if entry.ID > lastID {
			data, _ := json.Marshal(entry)
			_, _ = fmt.Fprintf(w, "id: %d\ndata: %s\n\n", entry.ID, data)
		}
	}
	flusher.Flush()

	ch := h.buf.Subscribe()
	defer h.buf.Unsubscribe(ch)

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case entry, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(entry)
			_, _ = fmt.Fprintf(w, "id: %d\ndata: %s\n\n", entry.ID, data)
			flusher.Flush()
		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}
