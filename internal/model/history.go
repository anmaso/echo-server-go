package model

import (
	"sync"
	"time"
)

// HistoryEntry represents a single recorded request and response.
type HistoryEntry struct {
	Timestamp    time.Time   `json:"timestamp"`
	Request      RequestData `json:"request"`
	Response     ResponseSummary `json:"response"` // Using a summary to avoid storing large response bodies if not needed, or we can make it configurable
	ResponseSize int64       `json:"responseSize"`
}

// ResponseSummary holds key details of an HTTP response.
type ResponseSummary struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"` // Consider if full body is always needed or if a truncated/summary version is better for history
}

// HistoryStorage manages the collection of HistoryEntry items.
type HistoryStorage struct {
	entries       []HistoryEntry
	maxSize       int
	mu            sync.RWMutex
	recordingActive bool
}

// NewHistoryStorage creates a new HistoryStorage instance.
func NewHistoryStorage(maxSize int) *HistoryStorage {
	return &HistoryStorage{
		entries:       make([]HistoryEntry, 0, maxSize),
		maxSize:       maxSize,
		recordingActive: false, // Recording is off by default
	}
}

// AddEntry adds a new entry to the history.
// If the history is full, it removes the oldest entry (FIFO).
func (hs *HistoryStorage) AddEntry(entry HistoryEntry) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if !hs.recordingActive {
		return // Do not add if recording is not active
	}

	if hs.maxSize <= 0 { // If maxSize is 0 or negative, recording is effectively disabled or unlimited (handle appropriately)
		// Decide on behavior: either don't add, or add without limit (which can be dangerous)
		// For now, let's assume maxSize > 0 means limited storage.
		// If maxSize is 0, we can treat it as "don't store".
		return
	}

	if len(hs.entries) >= hs.maxSize {
		// Remove the oldest entry
		hs.entries = hs.entries[1:]
	}
	hs.entries = append(hs.entries, entry)
}

// GetEntries returns a copy of all entries in the history.
func (hs *HistoryStorage) GetEntries() []HistoryEntry {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	// Return a copy to prevent external modification
	entriesCopy := make([]HistoryEntry, len(hs.entries))
	copy(entriesCopy, hs.entries)
	return entriesCopy
}

// ClearEntries removes all entries from the history.
func (hs *HistoryStorage) ClearEntries() {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.entries = make([]HistoryEntry, 0, hs.maxSize)
}

// SetMaxSize updates the maximum size of the history.
// If the new size is smaller than the current number of entries,
// older entries are truncated.
func (hs *HistoryStorage) SetMaxSize(newSize int) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if newSize < 0 {
		newSize = 0 // Or handle as an error
	}
	
	hs.maxSize = newSize
	if newSize == 0 {
		hs.entries = make([]HistoryEntry, 0) // Clear entries if new size is 0
		return
	}

	if len(hs.entries) > newSize {
		hs.entries = hs.entries[len(hs.entries)-newSize:]
	}
	// Ensure capacity is also adjusted if necessary, though append handles this.
    // Create a new slice with the new capacity and copy elements.
    newEntries := make([]HistoryEntry, len(hs.entries), newSize)
    copy(newEntries, hs.entries)
    hs.entries = newEntries
}

// StartRecording enables history recording.
func (hs *HistoryStorage) StartRecording() {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.recordingActive = true
}

// StopRecording disables history recording.
func (hs *HistoryStorage) StopRecording() {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.recordingActive = false
}

// IsRecordingActive returns the current recording status.
func (hs *HistoryStorage) IsRecordingActive() bool {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.recordingActive
}

// GetMaxSize returns the current max size.
func (hs *HistoryStorage) GetMaxSize() int {
    hs.mu.RLock()
    defer hs.mu.RUnlock()
    return hs.maxSize
}
