package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"echo-server/internal/config"
	"echo-server/internal/model"
	"github.com/stretchr/testify/assert"
)

// HistoryStatusResponse is defined in history_handler.go,
// but we might need it here for decoding responses if it's not exported,
// or if we want to avoid direct dependency on the unexported type for testing.
// However, history_handler.HistoryStatusResponse is exported, so we can use it.

func setupHistoryHandlerTest() (*HistoryHandler, *model.HistoryStorage) {
	hs := model.NewHistoryStorage(5) // Default size for tests
	
	// Create a minimal valid ServerConfig
	defaultConfig := config.ServerConfig{
		Host: "localhost",
		Port: 8080,
		History: config.HistoryConfig{
			Enabled:        true,
			DefaultMaxSize: 5,
		},
	}
	// Initialize ConfigManager with this default config
	// Assuming ConfigManager has a way to be initialized with a config directly,
	// or we use its loader and a temporary config file/data.
	// For simplicity, if NewConfigManager just creates an instance, and
	// if the handler doesn't immediately read from it in NewHistoryHandler, this is fine.
	// Let's assume NewConfigManager is simple or we load a default.
	// The current NewHistoryHandler doesn't use cm for much, so a basic one is fine.
	cm := config.NewConfigManager(nil) // Passing nil for loader, as we won't load from file here.
	                                   // This might need adjustment if ConfigManager cannot handle a nil loader.
                                       // Or, provide a loader that can provide our defaultConfig.
                                       // For now, let's assume the handler doesn't immediately need config from cm.


	handler := NewHistoryHandler(hs, cm)
	return handler, hs
}

func TestHandleStartRecording(t *testing.T) {
	handler, hs := setupHistoryHandlerTest()

	t.Run("Start with new maxSize", func(t *testing.T) {
		reqBody := bytes.NewBufferString(`{"maxSize": 10}`)
		req := httptest.NewRequest(http.MethodPost, "/history/start", reqBody)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Status code should be OK")
		assert.True(t, hs.IsRecordingActive(), "History storage should be active")
		assert.Equal(t, 10, hs.GetMaxSize(), "History storage maxSize should be updated")

		var resp HistoryStatusResponse // Using the exported type from handler pkg
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err, "Should decode response body")
		assert.True(t, resp.RecordingActive, "Response should indicate recording active")
		assert.Equal(t, 10, resp.MaxSize, "Response should reflect new maxSize")
	})

	t.Run("Start without maxSize (empty body)", func(t *testing.T) {
		// Reset for this sub-test if necessary, or use a new handler instance
		handler, hs = setupHistoryHandlerTest() // Reset to initial state (maxSize 5, not recording)
		
		req := httptest.NewRequest(http.MethodPost, "/history/start", nil) // No body
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, hs.IsRecordingActive())
		assert.Equal(t, 5, hs.GetMaxSize(), "MaxSize should remain the default from setup") // Default from setupHistoryHandlerTest

		var resp HistoryStatusResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.True(t, resp.RecordingActive)
		assert.Equal(t, 5, resp.MaxSize)
	})
    
    t.Run("Start with invalid JSON body", func(t *testing.T) {
		handler, hs = setupHistoryHandlerTest()
		reqBody := bytes.NewBufferString(`{"maxSize": "not-a-number"}`) // Invalid type for maxSize
		req := httptest.NewRequest(http.MethodPost, "/history/start", reqBody)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Starting recording should still succeed")
		assert.True(t, hs.IsRecordingActive(), "Recording should be active")
        // MaxSize should not change due to invalid body for maxSize, should use existing.
		assert.Equal(t, 5, hs.GetMaxSize(), "MaxSize should remain initial value") 
		
		var resp HistoryStatusResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.True(t, resp.RecordingActive)
		assert.Equal(t, 5, resp.MaxSize, "Response maxSize should be initial value")
	})

	t.Run("Start with negative maxSize", func(t *testing.T) {
		handler, hs = setupHistoryHandlerTest()
		initialMaxSize := hs.GetMaxSize()

		reqBody := bytes.NewBufferString(`{"maxSize": -5}`)
		req := httptest.NewRequest(http.MethodPost, "/history/start", reqBody)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, hs.IsRecordingActive())
		// HistoryStorage treats negative maxSize as 0.
		assert.Equal(t, 0, hs.GetMaxSize(), "History storage maxSize should be 0 after negative input")

		var resp HistoryStatusResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.True(t, resp.RecordingActive)
		assert.Equal(t, 0, resp.MaxSize)
	})
}

func TestHandleStopRecording(t *testing.T) {
	handler, hs := setupHistoryHandlerTest()
	hs.StartRecording() // Start it first
	assert.True(t, hs.IsRecordingActive(), "Pre-condition: recording should be active")

	req := httptest.NewRequest(http.MethodPost, "/history/stop", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, hs.IsRecordingActive(), "Recording should be stopped")

	var resp HistoryStatusResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.False(t, resp.RecordingActive)
	assert.Equal(t, hs.GetMaxSize(), resp.MaxSize) // MaxSize should not change
}

func TestHandleConfigureHistory(t *testing.T) {
	handler, hs := setupHistoryHandlerTest()

	t.Run("Valid maxSize", func(t *testing.T) {
		reqBody := bytes.NewBufferString(`{"maxSize": 20}`)
		req := httptest.NewRequest(http.MethodPut, "/history/config", reqBody)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, 20, hs.GetMaxSize())

		var resp HistoryStatusResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, 20, resp.MaxSize)
		assert.Equal(t, hs.IsRecordingActive(), resp.RecordingActive) // Recording status shouldn't change
	})

	t.Run("maxSize zero", func(t *testing.T) {
		handler, hs = setupHistoryHandlerTest() // Reset
		hs.AddEntry(model.HistoryEntry{Timestamp: time.Now(), Request: model.RequestData{Path: "/test1"}})
		hs.StartRecording() // ensure some data and recording is active

		reqBody := bytes.NewBufferString(`{"maxSize": 0}`)
		req := httptest.NewRequest(http.MethodPut, "/history/config", reqBody)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, 0, hs.GetMaxSize())
		assert.Empty(t, hs.GetEntries(), "Entries should be cleared when maxSize is set to 0")

		var resp HistoryStatusResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.MaxSize)
	})
	
	t.Run("Negative maxSize", func(t *testing.T) {
		handler, hs = setupHistoryHandlerTest() // Reset
		reqBody := bytes.NewBufferString(`{"maxSize": -5}`)
		req := httptest.NewRequest(http.MethodPut, "/history/config", reqBody)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		// HistoryStorage treats negative maxSize as 0.
		assert.Equal(t, 0, hs.GetMaxSize())

		var resp HistoryStatusResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.MaxSize)
	})

	t.Run("Invalid JSON body", func(t *testing.T) {
		handler, _ = setupHistoryHandlerTest() // Reset
		reqBody := bytes.NewBufferString(`{"maxSize": "not-a-number"}`)
		req := httptest.NewRequest(http.MethodPut, "/history/config", reqBody)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code) // Expect bad request for malformed JSON
	})
    
    t.Run("Empty JSON body", func(t *testing.T) {
		handler, _ = setupHistoryHandlerTest()
		reqBody := bytes.NewBufferString(`{}`) // Missing maxSize
		req := httptest.NewRequest(http.MethodPut, "/history/config", reqBody)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

        // The handler expects maxSize. If it's missing, it defaults to 0 for the int.
        // HistoryStorage sets maxSize to 0 if it receives 0.
		assert.Equal(t, http.StatusOK, rr.Code) 
        var resp HistoryStatusResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
        assert.Equal(t, 0, resp.MaxSize) 
	})
}

func TestHandleGetHistory(t *testing.T) {
	handler, hs := setupHistoryHandlerTest()
	hs.StartRecording()
	
	entryTime1 := time.Now()
	entryTime2 := time.Now().Add(1 * time.Second)
	hs.AddEntry(model.HistoryEntry{Timestamp: entryTime1, Request: model.RequestData{Path: "/test1"}})
	hs.AddEntry(model.HistoryEntry{Timestamp: entryTime2, Request: model.RequestData{Path: "/test2"}})

	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var entries []model.HistoryEntry
	err := json.NewDecoder(rr.Body).Decode(&entries)
	assert.NoError(t, err)
	assert.Len(t, entries, 2, "Should retrieve all entries")
	if len(entries) == 2 {
		assert.Equal(t, "/test1", entries[0].Request.Path)
		assert.Equal(t, "/test2", entries[1].Request.Path)
	}

	// Test with no entries
	hs.ClearEntries()
	req = httptest.NewRequest(http.MethodGet, "/history", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	bodyBytes, _ := io.ReadAll(rr.Body)
	assert.Equal(t, "[]\n", string(bodyBytes), "Response should be an empty JSON array") // Or "null\n" if that's the behavior
}

func TestHandleClearHistory(t *testing.T) {
	handler, hs := setupHistoryHandlerTest()
	hs.StartRecording()
	hs.AddEntry(model.HistoryEntry{Request: model.RequestData{Path: "/test1"}})
	assert.Len(t, hs.GetEntries(), 1, "Pre-condition: history should have one entry")

	req := httptest.NewRequest(http.MethodDelete, "/history", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code, "Status code should be No Content")
	assert.Empty(t, hs.GetEntries(), "History storage should be empty after clear")
}

func TestHistoryHandler_Routing(t *testing.T) {
	handler, _ := setupHistoryHandlerTest()

	tests := []struct {
		method     string
		path       string
		statusCode int
	}{
		{http.MethodGet, "/history/nonexistent", http.StatusNotFound},
		{http.MethodPost, "/history/nonexistent", http.StatusNotFound},
		{http.MethodPut, "/history/start", http.StatusNotFound},      // Method not allowed for this subpath
		{http.MethodDelete, "/history/start", http.StatusNotFound}, // Method not allowed
		{http.MethodGet, "/history/start", http.StatusNotFound},    // Method not allowed
		{http.MethodPut, "/history/stop", http.StatusNotFound},
		{http.MethodGet, "/history/config", http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, tc.statusCode, rr.Code, "Expected status code %d for %s %s, got %d", tc.statusCode, tc.method, tc.path, rr.Code)
		})
	}
}
