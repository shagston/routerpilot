package runtime

import (
	"context"
	"testing"

	eventbus "github.com/shagston/routerpilot/internal/events"
	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/sdk/types"
)

func TestEngineMergesContextTaskResultsIntoSnapshot(t *testing.T) {
	reg := registry.NewToolRegistry()
	readTool := &testTool{id: "network.route_get"}
	if err := reg.Register(readTool); err != nil {
		t.Fatal(err)
	}

	bus := eventbus.NewBus()
	engine := NewEngine(reg, bus)

	execution, err := engine.Execute(context.Background(), types.Plan{
		ID:     "plan-context",
		Intent: "inspect routes",
		Steps: []types.Task{
			{
				ID:      "gather-routes",
				Tool:    "network.route_get",
				Purpose: types.TaskPurposeContext,
			},
		},
		Risk: types.RiskLow,
	}, types.ContextSnapshot{"source": "test"})
	if err != nil {
		t.Fatalf("execute plan: %v", err)
	}

	if execution.Context["network.route_get"] == nil {
		t.Fatalf("expected merged context entry for tool, got %#v", execution.Context)
	}
	if execution.Context["task.gather-routes"] == nil {
		t.Fatalf("expected merged context entry for task, got %#v", execution.Context)
	}
}

func TestEngineDoesNotMergeActionTaskResultsIntoSnapshot(t *testing.T) {
	reg := registry.NewToolRegistry()
	readTool := &testTool{id: "network.ping"}
	if err := reg.Register(readTool); err != nil {
		t.Fatal(err)
	}

	engine := NewEngine(reg, eventbus.NewBus())
	execution, err := engine.Execute(context.Background(), types.Plan{
		ID:     "plan-action",
		Intent: "ping host",
		Steps: []types.Task{
			{ID: "ping", Tool: "network.ping"},
		},
		Risk: types.RiskLow,
	}, types.ContextSnapshot{"source": "test"})
	if err != nil {
		t.Fatalf("execute plan: %v", err)
	}

	if execution.Context["network.ping"] != nil {
		t.Fatalf("did not expect action task to merge into context: %#v", execution.Context)
	}
}
