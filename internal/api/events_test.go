package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shagston/routerpilot/internal/app"
	"github.com/shagston/routerpilot/sdk/types"
)

func TestHandleEventsReturnsPublishedEvents(t *testing.T) {
	instance, err := app.New()
	if err != nil {
		t.Fatalf("app.New() error: %v", err)
	}

	_ = instance.Events.Publish(types.Event{
		ID:        "evt-1",
		Timestamp: time.Now(),
		Type:      "execution.started",
		Source:    "test",
		Severity:  types.SeverityInfo,
	})

	server := NewServer(instance)
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rec := httptest.NewRecorder()

	server.handleEvents(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var events []types.Event
	if err := json.Unmarshal(rec.Body.Bytes(), &events); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(events) != 1 || events[0].ID != "evt-1" {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestHandleEventsStreamReplaysAndStreams(t *testing.T) {
	instance, err := app.New()
	if err != nil {
		t.Fatalf("app.New() error: %v", err)
	}

	_ = instance.Events.Publish(types.Event{
		ID:        "evt-replay",
		Timestamp: time.Now(),
		Type:      "execution.created",
		Source:    "test",
		Severity:  types.SeverityInfo,
	})

	server := NewServer(instance)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/events/stream?replay=false", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	go func() {
		time.Sleep(20 * time.Millisecond)
		_ = instance.Events.Publish(types.Event{
			ID:        "evt-live",
			Timestamp: time.Now(),
			Type:      "execution.completed",
			Source:    "test",
			Severity:  types.SeverityInfo,
		})
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	server.handleEventsStream(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "evt-replay") {
		t.Fatalf("did not expect replayed event in body: %s", body)
	}
	if !strings.Contains(body, "evt-live") {
		t.Fatalf("expected live event in body: %s", body)
	}
}
