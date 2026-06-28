package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/shagston/routerpilot/sdk/memory"
	"github.com/shagston/routerpilot/sdk/types"
)

type InMemoryProvider struct {
	mu   sync.RWMutex
	data map[string]memory.Record
}

func NewInMemoryProvider() *InMemoryProvider {
	return &InMemoryProvider{
		data: make(map[string]memory.Record),
	}
}

func (p *InMemoryProvider) Read(_ context.Context, key string) (memory.Record, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	record, exists := p.data[key]
	if !exists {
		return memory.Record{}, fmt.Errorf("%w: key %q not found", types.ErrNotFound, key)
	}
	return record, nil
}

func (p *InMemoryProvider) Write(_ context.Context, record memory.Record) error {
	if record.Key == "" {
		return fmt.Errorf("%w: record key is required", types.ErrInvalidInput)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.data[record.Key] = record
	return nil
}

func (p *InMemoryProvider) Search(_ context.Context, prefix string, limit int) ([]memory.Record, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if limit < 1 {
		limit = 100
	}

	var results []memory.Record
	for key, record := range p.data {
		if strings.HasPrefix(key, prefix) {
			results = append(results, record)
			if len(results) >= limit {
				break
			}
		}
	}

	if results == nil {
		return []memory.Record{}, nil
	}
	return results, nil
}
