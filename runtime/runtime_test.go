package runtime

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	eventbus "github.com/shagston/routerpilot/internal/events"
	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/internal/safety"
	"github.com/shagston/routerpilot/runtime/engine"
	"github.com/shagston/routerpilot/runtime/scheduler"
	"github.com/shagston/routerpilot/sdk/types"
)

type scheduleTestTool struct {
	id    types.ToolID
	exec  func(context.Context, types.ToolInput) (types.ToolResult, error)
}

func (t *scheduleTestTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:          t.id,
		Version:     "0.1.0",
		Category:    "test",
		Permissions: []types.Permission{types.PermissionRead},
		Timeout:     time.Second,
		Risk:        types.RiskLow,
	}
}
func (t *scheduleTestTool) InputSchema() types.Schema  { return types.Schema{} }
func (t *scheduleTestTool) OutputSchema() types.Schema { return types.Schema{} }
func (t *scheduleTestTool) Validate(context.Context, types.ToolInput) error { return nil }
func (t *scheduleTestTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	return t.exec(ctx, input)
}

func TestRuntimeSchedulerCreated(t *testing.T) {
	reg := registry.NewToolRegistry()
	bus := eventbus.NewBus()
	rt := New(reg, bus)

	if rt.Scheduler() == nil {
		t.Fatal("expected scheduler to be created")
	}
}

func TestRuntimeSchedulerLifecycle(t *testing.T) {
	reg := registry.NewToolRegistry()
	bus := eventbus.NewBus()
	rt := New(reg, bus)

	ctx := context.Background()
	if err := rt.Start(ctx); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	sched := rt.Scheduler()
	list := sched.List()
	if list == nil {
		t.Fatal("expected Scheduler.List() to return non-nil")
	}

	if err := rt.Stop(ctx); err != nil {
		t.Fatalf("stop runtime: %v", err)
	}
}

func TestRuntimeSchedulerExecutesCapability(t *testing.T) {
	var callCount atomic.Int32

	tool := &scheduleTestTool{
		id: "test.ping",
		exec: func(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
			callCount.Add(1)
			return types.ToolResult{Success: true, Output: types.ToolOutput{"pong": true}}, nil
		},
	}

	reg := registry.NewToolRegistry()
	if err := reg.Register(tool); err != nil {
		t.Fatal(err)
	}

	bus := eventbus.NewBus()
	rt := New(reg, bus, engine.WithValidator(safety.NewValidator(reg, safety.Config{
		Permissions: []types.Permission{types.PermissionRead},
	})))

	ctx := context.Background()
	if err := rt.Start(ctx); err != nil {
		t.Fatalf("start runtime: %v", err)
	}
	defer rt.Stop(ctx)

	sched := rt.Scheduler()
	oneshot := scheduler.OneshotSchedule("test-exec", time.Now())
	oneshot.Capability = "test.ping"
	oneshot.Input = map[string]any{}

	if err := sched.Register(oneshot); err != nil {
		t.Fatalf("register schedule: %v", err)
	}

	if err := sched.Trigger("test-exec"); err != nil {
		t.Fatalf("trigger schedule: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if n := callCount.Load(); n != 1 {
		t.Fatalf("expected capability to be called once, got %d", n)
	}
}

func TestRuntimeSchedulerEventTriggered(t *testing.T) {
	var callCount atomic.Int32

	tool := &scheduleTestTool{
		id: "test.event_listener",
		exec: func(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
			callCount.Add(1)
			return types.ToolResult{Success: true, Output: types.ToolOutput{"received": true}}, nil
		},
	}

	reg := registry.NewToolRegistry()
	if err := reg.Register(tool); err != nil {
		t.Fatal(err)
	}

	bus := eventbus.NewBus()
	rt := New(reg, bus, engine.WithValidator(safety.NewValidator(reg, safety.Config{
		Permissions: []types.Permission{types.PermissionRead},
	})))

	ctx := context.Background()
	if err := rt.Start(ctx); err != nil {
		t.Fatalf("start runtime: %v", err)
	}
	defer rt.Stop(ctx)

	sched := rt.Scheduler()
	evSched := scheduler.EventSchedule("test-event-listener", "custom.test.event")
	evSched.Capability = "test.event_listener"
	evSched.Input = map[string]any{}

	if err := sched.Register(evSched); err != nil {
		t.Fatalf("register event schedule: %v", err)
	}

	bus.Publish(types.Event{
		Type:      "custom.test.event",
		Source:    "test",
		Severity:  types.SeverityInfo,
		Priority:  types.PriorityNormal,
		Timestamp: time.Now(),
	})

	time.Sleep(500 * time.Millisecond)

	if n := callCount.Load(); n != 1 {
		t.Fatalf("expected event-triggered capability to be called once, got %d", n)
	}
}
