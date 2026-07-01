package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	eventbus "github.com/shagston/routerpilot/internal/events"
	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/internal/safety"
	"github.com/shagston/routerpilot/sdk/types"
)

type testTool struct {
	id          types.ToolID
	permissions []types.Permission
	failures    int
	calls       int
	retryable   bool
}

func (t *testTool) Metadata() types.ToolMetadata {
	permissions := t.permissions
	if permissions == nil {
		permissions = []types.Permission{types.PermissionRead}
	}
	return types.ToolMetadata{
		ID:          t.id,
		Version:     "0.1.0",
		Category:    "test",
		Description: "test tool",
		Permissions: permissions,
		Timeout:     time.Second,
		Risk:        types.RiskLow,
	}
}

func (t *testTool) InputSchema() types.Schema  { return types.Schema{} }
func (t *testTool) OutputSchema() types.Schema { return types.Schema{} }
func (t *testTool) Validate(context.Context, types.ToolInput) error {
	return nil
}
func (t *testTool) Execute(context.Context, types.ToolInput) (types.ToolResult, error) {
	t.calls++
	if t.calls <= t.failures {
		return types.ToolResult{Success: false, Error: "temporary failure", Retryable: t.retryable}, errors.New("temporary failure")
	}
	return types.ToolResult{Success: true, Output: types.ToolOutput{"calls": t.calls}}, nil
}

func TestEngineExecutesPlanInDependencyOrder(t *testing.T) {
	reg := registry.NewToolRegistry()
	first := &testTool{id: "first"}
	second := &testTool{id: "second"}
	if err := reg.Register(first); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(second); err != nil {
		t.Fatal(err)
	}
	bus := eventbus.NewBus()
	engine := NewEngine(reg, bus)

	execution, err := engine.Execute(context.Background(), types.Plan{
		ID:     "plan-1",
		Intent: "test.intent",
		Steps: []types.Task{
			{ID: "b", Tool: "second", Dependencies: []types.TaskID{"a"}},
			{ID: "a", Tool: "first"},
		},
		Risk: types.RiskLow,
	}, nil)
	if err != nil {
		t.Fatalf("execute plan: %v", err)
	}
	if execution.State != types.ExecutionCompleted {
		t.Fatalf("expected completed, got %s", execution.State)
	}
	if first.calls != 1 || second.calls != 1 {
		t.Fatalf("expected both tools called once, got first=%d second=%d", first.calls, second.calls)
	}
	if len(bus.Events()) == 0 {
		t.Fatal("expected runtime events")
	}
}

func TestEngineRetriesRetryableFailures(t *testing.T) {
	reg := registry.NewToolRegistry()
	flaky := &testTool{id: "flaky", failures: 1, retryable: true}
	if err := reg.Register(flaky); err != nil {
		t.Fatal(err)
	}
	engine := NewEngine(reg, eventbus.NewBus())

	_, err := engine.Execute(context.Background(), types.Plan{
		ID:     "plan-1",
		Intent: "test.intent",
		Steps: []types.Task{
			{ID: "a", Tool: "flaky", Retry: types.RetryPolicy{Attempts: 2}},
		},
		Risk: types.RiskLow,
	}, nil)
	if err != nil {
		t.Fatalf("execute plan: %v", err)
	}
	if flaky.calls != 2 {
		t.Fatalf("expected retry, got %d calls", flaky.calls)
	}
}

func TestEngineRejectsDependencyCycles(t *testing.T) {
	reg := registry.NewToolRegistry()
	if err := reg.Register(&testTool{id: "tool"}); err != nil {
		t.Fatal(err)
	}
	engine := NewEngine(reg, eventbus.NewBus())

	_, err := engine.Execute(context.Background(), types.Plan{
		ID:     "plan-1",
		Intent: "test.intent",
		Steps: []types.Task{
			{ID: "a", Tool: "tool", Dependencies: []types.TaskID{"b"}},
			{ID: "b", Tool: "tool", Dependencies: []types.TaskID{"a"}},
		},
		Risk: types.RiskLow,
	}, nil)
	if !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestEngineStopsBeforeExecutionWhenSafetyFails(t *testing.T) {
	reg := registry.NewToolRegistry()
	adminTool := &testTool{id: "system.reboot", permissions: []types.Permission{types.PermissionAdmin}}
	if err := reg.Register(adminTool); err != nil {
		t.Fatal(err)
	}
	engine := NewEngine(reg, eventbus.NewBus(), WithValidator(safety.NewValidator(reg, safety.Config{
		Permissions: []types.Permission{types.PermissionRead},
	})))

	execution, err := engine.Execute(context.Background(), types.Plan{
		ID:     "plan-1",
		Intent: "system.reboot",
		Steps:  []types.Task{{ID: "reboot", Tool: "system.reboot"}},
		Risk:   types.RiskHigh,
	}, nil)
	if !errors.Is(err, types.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}
	if execution.State != types.ExecutionFailed {
		t.Fatalf("expected failed execution, got %s", execution.State)
	}
	if adminTool.calls != 0 {
		t.Fatalf("expected safety to stop execution, got %d calls", adminTool.calls)
	}
}
