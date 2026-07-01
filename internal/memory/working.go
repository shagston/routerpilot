package memory

import (
	"context"
	"sync"

	"github.com/shagston/routerpilot/sdk/memory"
)

type WorkingMemory struct {
	mu   sync.RWMutex
	data map[string]memory.Record
}

func NewWorkingMemory() *WorkingMemory {
	return &WorkingMemory{
		data: make(map[string]memory.Record),
	}
}

func (w *WorkingMemory) Get(_ context.Context, key string) (memory.Record, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	r, ok := w.data[key]
	return r, ok
}

func (w *WorkingMemory) Set(_ context.Context, record memory.Record) {
	if record.Key == "" {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.data[record.Key] = record
}

func (w *WorkingMemory) Delete(_ context.Context, key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.data, key)
}

func (w *WorkingMemory) Search(_ context.Context, prefix string, limit int) []memory.Record {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if limit < 1 {
		limit = 100
	}

	var results []memory.Record
	for k, r := range w.data {
		if len(prefix) == 0 || (len(k) >= len(prefix) && k[:len(prefix)] == prefix) {
			results = append(results, r)
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}

func (w *WorkingMemory) Clear(_ context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.data = make(map[string]memory.Record)
}

func (w *WorkingMemory) Len() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.data)
}
