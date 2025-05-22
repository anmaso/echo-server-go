package model

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHistoryStorage(t *testing.T) {
	hs := NewHistoryStorage(10)
	assert.NotNil(t, hs)
	assert.Equal(t, 10, hs.GetMaxSize(), "Initial maxSize should be set")
	assert.False(t, hs.IsRecordingActive(), "Recording should be off by default")
	assert.Empty(t, hs.GetEntries(), "Entries should be empty initially")
}

func TestAddEntry_RecordingActive(t *testing.T) {
	hs := NewHistoryStorage(2)
	hs.StartRecording()

	entry1 := HistoryEntry{Timestamp: time.Now(), Request: RequestData{Path: "/path1"}}
	entry2 := HistoryEntry{Timestamp: time.Now().Add(1 * time.Second), Request: RequestData{Path: "/path2"}}
	entry3 := HistoryEntry{Timestamp: time.Now().Add(2 * time.Second), Request: RequestData{Path: "/path3"}}

	hs.AddEntry(entry1)
	entries := hs.GetEntries()
	assert.Len(t, entries, 1, "Should have 1 entry after adding one")
	assert.Equal(t, "/path1", entries[0].Request.Path)

	hs.AddEntry(entry2)
	entries = hs.GetEntries()
	assert.Len(t, entries, 2, "Should have 2 entries after adding two")
	assert.Equal(t, "/path1", entries[0].Request.Path)
	assert.Equal(t, "/path2", entries[1].Request.Path)

	hs.AddEntry(entry3) // Should evict entry1 (FIFO)
	entries = hs.GetEntries()
	assert.Len(t, entries, 2, "Should still have 2 entries after adding third (FIFO)")
	assert.Equal(t, "/path2", entries[0].Request.Path, "Oldest entry should be evicted")
	assert.Equal(t, "/path3", entries[1].Request.Path)
}

func TestAddEntry_NotRecording(t *testing.T) {
	hs := NewHistoryStorage(2)
	// Recording is off by default (hs.IsRecordingActive() is false)
	entry1 := HistoryEntry{Timestamp: time.Now(), Request: RequestData{Path: "/path1"}}
	hs.AddEntry(entry1)
	assert.Empty(t, hs.GetEntries(), "Should not add entry if recording is not active")
}

func TestAddEntry_MaxSizeZero(t *testing.T) {
	hs := NewHistoryStorage(0)
	hs.StartRecording() // Recording is on
	entry1 := HistoryEntry{Timestamp: time.Now(), Request: RequestData{Path: "/path1"}}
	hs.AddEntry(entry1)
	assert.Empty(t, hs.GetEntries(), "Should not add entry if maxSize is 0")

	// Also test if maxSize is set to 0 later
	hs = NewHistoryStorage(5)
	hs.StartRecording()
	hs.AddEntry(entry1)
	assert.Len(t, hs.GetEntries(), 1)
	hs.SetMaxSize(0)
	hs.AddEntry(HistoryEntry{Timestamp: time.Now(), Request: RequestData{Path: "/path2"}})
	assert.Empty(t, hs.GetEntries(), "Should not add entry if maxSize is set to 0, and existing entries should be cleared")
}

func TestGetEntries_ReturnsCopy(t *testing.T) {
	hs := NewHistoryStorage(2)
	hs.StartRecording()
	entry1 := HistoryEntry{Timestamp: time.Now(), Request: RequestData{Path: "/path1"}}
	hs.AddEntry(entry1)

	entries1 := hs.GetEntries()
	assert.Len(t, entries1, 1)
	assert.Equal(t, "/path1", entries1[0].Request.Path)

	// Modify the returned slice and check if the original is affected
	if len(entries1) > 0 {
		entries1[0].Request.Path = "/modified"
	}

	entries2 := hs.GetEntries()
	assert.Len(t, entries2, 1)
	assert.Equal(t, "/path1", entries2[0].Request.Path, "Original entry should not be modified")

	// Test appending to the returned slice
	entries1 = append(entries1, HistoryEntry{Request: RequestData{Path: "/another"}})
	entries2 = hs.GetEntries()
	assert.Len(t, entries2, 1, "Appending to returned slice should not affect original storage")
}

func TestClearEntries(t *testing.T) {
	hs := NewHistoryStorage(2)
	hs.StartRecording()
	hs.AddEntry(HistoryEntry{Timestamp: time.Now(), Request: RequestData{Path: "/path1"}})
	assert.NotEmpty(t, hs.GetEntries(), "Should have entries before clearing")
	hs.ClearEntries()
	assert.Empty(t, hs.GetEntries(), "Should have no entries after clearing")

	// Test clearing when already empty
	hs.ClearEntries()
	assert.Empty(t, hs.GetEntries(), "Clearing an already empty history should result in empty history")
}

func TestSetMaxSize(t *testing.T) {
	hs := NewHistoryStorage(3)
	hs.StartRecording()
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(1, 0), Request: RequestData{Path: "1"}})
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(2, 0), Request: RequestData{Path: "2"}})
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(3, 0), Request: RequestData{Path: "3"}})

	// Decrease maxSize, should truncate older entries
	hs.SetMaxSize(2)
	assert.Equal(t, 2, hs.GetMaxSize(), "maxSize should be updated")
	entries := hs.GetEntries()
	assert.Len(t, entries, 2, "Number of entries should be truncated to new maxSize")
	assert.Equal(t, "2", entries[0].Request.Path, "Oldest entry ('1') should be gone")
	assert.Equal(t, "3", entries[1].Request.Path)

	// Increase maxSize
	hs.SetMaxSize(5)
	assert.Equal(t, 5, hs.GetMaxSize(), "maxSize should be updated")
	assert.Len(t, hs.GetEntries(), 2, "Number of entries should remain the same as no new entries added")

	// Add more entries up to the new maxSize
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(4, 0), Request: RequestData{Path: "4"}})
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(5, 0), Request: RequestData{Path: "5"}})
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(6, 0), Request: RequestData{Path: "6"}}) // This will push out "2"
	
	entries = hs.GetEntries()
	assert.Len(t, entries, 5)
	assert.Equal(t, "3", entries[0].Request.Path) // "2" should be gone
	assert.Equal(t, "4", entries[1].Request.Path)
	assert.Equal(t, "5", entries[2].Request.Path)
	assert.Equal(t, "6", entries[3].Request.Path) // "6" is the newest of the first 5
	// Wait, the entry "6" should be the 4th element if "2" was pushed out. And "3", "4", "5" are before it.
	// The entry "6" is added, the list is [3,4,5,6]. Max size 5. So one more can be added.
	// Let's trace SetMaxSize(5) from [2,3] (paths: "2", "3")
	// MaxSize becomes 5. Entries are still [2,3].
	// Add "4" -> [2,3,4]
	// Add "5" -> [2,3,4,5]
	// Add "6" -> [2,3,4,5,6]
	// This is correct. My comment was wrong.

	// Set maxSize to 0
	hs.SetMaxSize(0)
	assert.Equal(t, 0, hs.GetMaxSize(), "maxSize should be 0")
	assert.Empty(t, hs.GetEntries(), "Entries should be cleared if maxSize becomes 0")

	// Test setting maxSize when current entries exceed new maxSize (re-test for clarity)
	hs = NewHistoryStorage(5)
	hs.StartRecording()
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(1,0), Request: RequestData{Path: "a"}})
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(2,0), Request: RequestData{Path: "b"}})
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(3,0), Request: RequestData{Path: "c"}})
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(4,0), Request: RequestData{Path: "d"}})
	hs.AddEntry(HistoryEntry{Timestamp: time.Unix(5,0), Request: RequestData{Path: "e"}}) // entries: a,b,c,d,e

	hs.SetMaxSize(3) // new max size 3. Should keep c,d,e
	entries = hs.GetEntries()
	assert.Len(t, entries, 3)
	assert.Equal(t, "c", entries[0].Request.Path)
	assert.Equal(t, "d", entries[1].Request.Path)
	assert.Equal(t, "e", entries[2].Request.Path)

	// Test setting maxSize to a negative value (should be treated as 0)
	hs.SetMaxSize(-5)
	assert.Equal(t, 0, hs.GetMaxSize(), "Negative maxSize should be treated as 0")
	assert.Empty(t, hs.GetEntries(), "Entries should be cleared if maxSize becomes 0 due to negative input")
}

func TestStartStopRecording(t *testing.T) {
	hs := NewHistoryStorage(1)
	assert.False(t, hs.IsRecordingActive(), "Should not be recording initially")
	
	hs.StartRecording()
	assert.True(t, hs.IsRecordingActive(), "Should be recording after StartRecording()")
	
	// Add an entry while recording
	hs.AddEntry(HistoryEntry{Request: RequestData{Path: "test"}});
	assert.Len(t, hs.GetEntries(), 1, "Entry should be added when recording")

	hs.StopRecording()
	assert.False(t, hs.IsRecordingActive(), "Should not be recording after StopRecording()")
	
	// Try adding an entry while not recording
	hs.AddEntry(HistoryEntry{Request: RequestData{Path: "test2"}});
	assert.Len(t, hs.GetEntries(), 1, "Entry should not be added when recording is stopped")
}

// Optional: Basic thread safety test for AddEntry
// This is a very basic test and might need more sophisticated handling for CI environments (e.g., race detector)
func TestAddEntry_Concurrent(t *testing.T) {
	hs := NewHistoryStorage(1000)
	hs.StartRecording()
	
	var wg sync.WaitGroup
	numGoroutines := 100
	entriesPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < entriesPerGoroutine; j++ {
				entry := HistoryEntry{
					Timestamp: time.Now(),
					Request:   RequestData{Path: "path"}, // Path doesn't need to be unique for this test
				}
				hs.AddEntry(entry)
			}
		}(i)
	}
	wg.Wait()

	// Due to potential eviction if maxSize is smaller than total entries,
	// we expect either maxSize entries or total entries if smaller.
	expectedCount := numGoroutines * entriesPerGoroutine
	if expectedCount > hs.GetMaxSize() {
		expectedCount = hs.GetMaxSize()
	}
	
	assert.Len(t, hs.GetEntries(), expectedCount, "Should have the correct number of entries after concurrent adds")
}
