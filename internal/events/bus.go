package events

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type DeliveryState string

const (
	DeliveryPending DeliveryState = "pending"
	DeliverySuccess DeliveryState = "success"
	DeliveryFailed  DeliveryState = "failed"
	DeliveryDropped DeliveryState = "dropped"
)

type DeliveryReport struct {
	EventID   types.EventID
	SubID     uint64
	State     DeliveryState
	Attempts  int
	LastError string
}

type Subscriber interface {
	Handle(types.Event) error
}

type Bus struct {
	mu          sync.RWMutex
	subscribers []*subEntry
	events      []types.Event
	dlq         []types.Event
	reports     []DeliveryReport
	seq         uint64
	done        chan struct{}
	wg          sync.WaitGroup
}

type subEntry struct {
	id    uint64
	sub   Subscriber
	ch    chan types.Event
	drops atomic.Uint64
	name  string
}

func NewBus() *Bus {
	b := &Bus{
		done: make(chan struct{}),
	}
	b.wg.Add(1)
	go b.ttlLoop()
	return b
}

func (b *Bus) ttlLoop() {
	defer b.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-b.done:
			return
		case <-ticker.C:
			b.evictExpired()
		}
	}
}

func (b *Bus) evictExpired() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	kept := make([]types.Event, 0, len(b.events))
	for _, e := range b.events {
		if e.TTL > 0 && now.After(e.Timestamp.Add(e.TTL)) {
			continue
		}
		kept = append(kept, e)
	}
	b.events = kept

	dlqKept := make([]types.Event, 0, len(b.dlq))
	for _, e := range b.dlq {
		if e.TTL > 0 && now.After(e.Timestamp.Add(e.TTL)) {
			continue
		}
		dlqKept = append(dlqKept, e)
	}
	b.dlq = dlqKept
}

func (b *Bus) Close() {
	close(b.done)
	b.wg.Wait()
}

func (b *Bus) nextSeq() uint64 {
	return atomic.AddUint64(&b.seq, 1)
}

func (b *Bus) Publish(event types.Event) error {
	b.mu.Lock()
	b.events = append(b.events, event)
	entries := append([]*subEntry(nil), b.subscribers...)
	b.mu.Unlock()

	for _, entry := range entries {
		report := b.deliver(event, entry)
		b.mu.Lock()
		b.reports = append(b.reports, report)
		b.mu.Unlock()
	}
	return nil
}

func (b *Bus) deliver(event types.Event, entry *subEntry) DeliveryReport {
	report := DeliveryReport{
		EventID:  event.ID,
		SubID:    entry.id,
		State:    DeliveryPending,
		Attempts: 0,
	}

	if entry.ch != nil {
		select {
		case entry.ch <- event:
			report.State = DeliverySuccess
			report.Attempts = 1
		default:
			entry.drops.Add(1)
			report.State = DeliveryDropped
			report.Attempts = 1
		}
		return report
	}

	const maxRetries = 3
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					lastErr = fmt.Errorf("subscriber panic: %v", r)
				}
			}()
			err := entry.sub.Handle(event)
			if err == nil {
				lastErr = nil
				return
			}
			lastErr = err
		}()

		if lastErr == nil {
			report.State = DeliverySuccess
			report.Attempts = attempt
			return report
		}

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt*50) * time.Millisecond)
		}
	}

	b.mu.Lock()
	b.dlq = append(b.dlq, event)
	b.mu.Unlock()

	report.State = DeliveryFailed
	report.Attempts = maxRetries
	report.LastError = lastErr.Error()

	slog.Warn("event delivery failed, sent to DLQ",
		"event_id", event.ID,
		"type", event.Type,
		"subscriber", entry.name,
		"error", lastErr,
	)
	return report
}

func (b *Bus) Subscribe(subscriber Subscriber) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers = append(b.subscribers, &subEntry{
		id:   b.nextSeq(),
		sub:  subscriber,
		name: fmt.Sprintf("subscriber-%d", len(b.subscribers)+1),
	})
	return nil
}

func (b *Bus) SubscribeChan(buffer int) (<-chan types.Event, func()) {
	if buffer < 1 {
		buffer = 256
	}

	ch := make(chan types.Event, buffer)
	entry := &subEntry{
		id:   b.nextSeq(),
		ch:   ch,
		name: fmt.Sprintf("chan-%d", b.nextSeq()),
	}

	b.mu.Lock()
	b.subscribers = append(b.subscribers, entry)
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		for i, e := range b.subscribers {
			if e == entry {
				b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
				break
			}
		}
		close(ch)
	}

	return ch, unsubscribe
}

func (b *Bus) Events() []types.Event {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return append([]types.Event(nil), b.events...)
}

func (b *Bus) EventsFiltered(executionID types.ExecutionID) []types.Event {
	all := b.Events()
	if executionID == "" {
		return all
	}

	filtered := make([]types.Event, 0, len(all))
	for _, event := range all {
		if event.ExecutionID == executionID {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func (b *Bus) DLQ() []types.Event {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return append([]types.Event(nil), b.dlq...)
}

func (b *Bus) Reports() []DeliveryReport {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return append([]DeliveryReport(nil), b.reports...)
}

func (b *Bus) Drops(subID uint64) uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, entry := range b.subscribers {
		if entry.id == subID {
			return entry.drops.Load()
		}
	}
	return 0
}
