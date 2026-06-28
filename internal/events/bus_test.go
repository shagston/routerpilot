package events

import (
	"testing"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

func TestBusSubscribeChanReceivesPublishedEvents(t *testing.T) {
	bus := NewBus()
	ch, unsubscribe := bus.SubscribeChan(4)
	defer unsubscribe()

	event := types.Event{
		ID:        "evt-1",
		Timestamp: time.Now(),
		Type:      "test.event",
		Source:    "test",
		Severity:  types.SeverityInfo,
	}

	if err := bus.Publish(event); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case got := <-ch:
		if got.ID != event.ID {
			t.Fatalf("expected event %q, got %q", event.ID, got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestBusEventsFilteredByExecutionID(t *testing.T) {
	bus := NewBus()

	_ = bus.Publish(types.Event{ID: "1", ExecutionID: "exec-a", Type: "a"})
	_ = bus.Publish(types.Event{ID: "2", ExecutionID: "exec-b", Type: "b"})
	_ = bus.Publish(types.Event{ID: "3", ExecutionID: "exec-a", Type: "c"})

	filtered := bus.EventsFiltered("exec-a")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 events, got %d", len(filtered))
	}
}
