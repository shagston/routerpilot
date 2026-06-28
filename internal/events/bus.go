package events

import (
	"sync"
	"sync/atomic"

	"github.com/shagston/routerpilot/sdk/types"
)

type Bus struct {
	mu          sync.RWMutex
	subscribers []Subscriber
	events      []types.Event
}

type Subscriber interface {
	Handle(types.Event) error
}

type chanSubscriber struct {
	ch    chan types.Event
	drops atomic.Uint64
}

func (s *chanSubscriber) Handle(event types.Event) error {
	select {
	case s.ch <- event:
	default:
		s.drops.Add(1)
	}
	return nil
}

func (s *chanSubscriber) Drops() uint64 {
	return s.drops.Load()
}

func NewBus() *Bus {
	return &Bus{}
}

func (b *Bus) Publish(event types.Event) error {
	b.mu.Lock()
	b.events = append(b.events, event)
	subscribers := append([]Subscriber(nil), b.subscribers...)
	b.mu.Unlock()

	for _, subscriber := range subscribers {
		_ = subscriber.Handle(event)
	}
	return nil
}

func (b *Bus) Subscribe(subscriber Subscriber) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers = append(b.subscribers, subscriber)
	return nil
}

func (b *Bus) SubscribeChan(buffer int) (<-chan types.Event, func()) {
	if buffer < 1 {
		buffer = 256
	}

	ch := make(chan types.Event, buffer)
	sub := &chanSubscriber{ch: ch}

	b.mu.Lock()
	b.subscribers = append(b.subscribers, sub)
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		for i, subscriber := range b.subscribers {
			if subscriber == sub {
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
