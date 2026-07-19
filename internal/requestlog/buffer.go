package requestlog

import (
	"sync"
	"time"
)

type Entry struct {
	ID              int64               `json:"id"`
	Timestamp       time.Time           `json:"timestamp"`
	Method          string              `json:"method"`
	Path            string              `json:"path"`
	Query           string              `json:"query,omitempty"`
	RequestHeaders  map[string][]string `json:"requestHeaders,omitempty"`
	RequestBody     string              `json:"requestBody,omitempty"`
	StatusCode      int                 `json:"statusCode"`
	ResponseHeaders map[string][]string `json:"responseHeaders,omitempty"`
	ResponseBody    string              `json:"responseBody,omitempty"`
	DurationMs      float64             `json:"durationMs"`
}

type Buffer struct {
	mu      sync.RWMutex
	entries []*Entry
	size    int
	pos     int // next write index
	count   int // number of valid entries
	nextID  int64
	subs    []chan *Entry
}

var global = New(200)

func Global() *Buffer { return global }

func InitGlobal(size int) { global = New(size) }

func New(size int) *Buffer {
	if size < 1 {
		size = 1
	}
	return &Buffer{size: size, entries: make([]*Entry, size)}
}

func (b *Buffer) Add(e *Entry) {
	b.mu.Lock()
	b.nextID++
	e.ID = b.nextID
	b.entries[b.pos] = e
	b.pos = (b.pos + 1) % b.size
	if b.count < b.size {
		b.count++
	}
	subs := make([]chan *Entry, len(b.subs))
	copy(subs, b.subs)
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- e:
		default:
		}
	}
}

func (b *Buffer) All() []*Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.snapshot()
}

// snapshot returns entries oldest-first; must be called with lock held.
func (b *Buffer) snapshot() []*Entry {
	out := make([]*Entry, b.count)
	start := (b.pos - b.count + b.size) % b.size
	for i := 0; i < b.count; i++ {
		out[i] = b.entries[(start+i)%b.size]
	}
	return out
}

func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = make([]*Entry, b.size)
	b.pos = 0
	b.count = 0
}

func (b *Buffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

func (b *Buffer) SetSize(newSize int) {
	if newSize < 1 {
		newSize = 1
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	current := b.snapshot()
	if len(current) > newSize {
		current = current[len(current)-newSize:]
	}

	b.entries = make([]*Entry, newSize)
	b.size = newSize
	b.count = len(current)
	copy(b.entries, current)
	b.pos = b.count % newSize
}

func (b *Buffer) Subscribe() chan *Entry {
	ch := make(chan *Entry, 64)
	b.mu.Lock()
	b.subs = append(b.subs, ch)
	b.mu.Unlock()
	return ch
}

func (b *Buffer) Unsubscribe(ch chan *Entry) {
	b.mu.Lock()
	for i, s := range b.subs {
		if s == ch {
			b.subs = append(b.subs[:i], b.subs[i+1:]...)
			break
		}
	}
	b.mu.Unlock()
	close(ch)
}
