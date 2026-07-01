package memory

import (
	"context"
	"sync"
	"time"

	"github.com/shagston/routerpilot/sdk/memory"
)

type sessionEntry struct {
	record    memory.Record
	expiresAt time.Time
}

type SessionMemory struct {
	mu      sync.RWMutex
	data    map[string]*sessionEntry
	order   []string
	ttl     time.Duration
	maxSize int
}

type SessionOption func(*SessionMemory)

func WithSessionTTL(ttl time.Duration) SessionOption {
	return func(s *SessionMemory) { s.ttl = ttl }
}

func WithSessionMaxSize(n int) SessionOption {
	return func(s *SessionMemory) { s.maxSize = n }
}

func NewSessionMemory(opts ...SessionOption) *SessionMemory {
	s := &SessionMemory{
		data:    make(map[string]*sessionEntry),
		ttl:     5 * time.Minute,
		maxSize: 1000,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *SessionMemory) Get(_ context.Context, key string) (memory.Record, bool) {
	s.mu.RLock()
	entry, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return memory.Record{}, false
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return memory.Record{}, false
	}

	return entry.record, true
}

func (s *SessionMemory) Set(ctx context.Context, record memory.Record) {
	if record.Key == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[record.Key]; !exists && len(s.data) >= s.maxSize {
		oldest := s.order[0]
		delete(s.data, oldest)
		s.order = s.order[1:]
	}

	var expiresAt time.Time
	if s.ttl > 0 {
		expiresAt = time.Now().Add(s.ttl)
	}

	if _, exists := s.data[record.Key]; !exists {
		s.order = append(s.order, record.Key)
	}
	s.data[record.Key] = &sessionEntry{
		record:    record,
		expiresAt: expiresAt,
	}
}

func (s *SessionMemory) SetWithTTL(ctx context.Context, record memory.Record, ttl time.Duration) {
	if record.Key == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[record.Key]; !exists {
		s.order = append(s.order, record.Key)
	}
	s.data[record.Key] = &sessionEntry{
		record:    record,
		expiresAt: time.Now().Add(ttl),
	}
}

func (s *SessionMemory) Delete(_ context.Context, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	s.removeOrder(key)
}

func (s *SessionMemory) removeOrder(key string) {
	for i, k := range s.order {
		if k == key {
			s.order = append(s.order[:i], s.order[i+1:]...)
			return
		}
	}
}

func (s *SessionMemory) Search(_ context.Context, prefix string, limit int) []memory.Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit < 1 {
		limit = 100
	}

	now := time.Now()
	var results []memory.Record
	for k, entry := range s.data {
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			continue
		}
		if len(prefix) == 0 || (len(k) >= len(prefix) && k[:len(prefix)] == prefix) {
			results = append(results, entry.record)
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}

func (s *SessionMemory) Clear(_ context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]*sessionEntry)
	s.order = nil
}

func (s *SessionMemory) Cleanup(_ context.Context) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, entry := range s.data {
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			delete(s.data, k)
			s.removeOrder(k)
		}
	}
}
