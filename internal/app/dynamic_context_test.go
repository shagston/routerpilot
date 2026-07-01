package app

import (
	"context"
	"errors"
	"testing"
	"time"

	eventbus "github.com/shagston/routerpilot/internal/events"
	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/internal/safety"
	"github.com/shagston/routerpilot/runtime"
	runtimeengine "github.com/shagston/routerpilot/runtime/engine"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/types"
)

type adaptiveTestTool struct {
	id    types.ToolID
	calls int
}

func (t *adaptiveTestTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:          t.id,
		Version:     "0.1.0",
		Category:    "test",
		Description: "adaptive test tool",
		Permissions: []types.Permission{types.PermissionRead},
		Timeout:     time.Second,
		Risk:        types.RiskLow,
	}
}

func (t *adaptiveTestTool) InputSchema() types.Schema  { return types.Schema{} }
func (t *adaptiveTestTool) OutputSchema() types.Schema { return types.Schema{} }
func (t *adaptiveTestTool) Validate(context.Context, types.ToolInput) error {
	return nil
}
func (t *adaptiveTestTool) Execute(context.Context, types.ToolInput) (types.ToolResult, error) {
	t.calls++
	return types.ToolResult{
		Success: true,
		Output:  types.ToolOutput{"calls": t.calls},
	}, nil
}

type sequencePlanner struct {
	plans []types.Plan
	call  int
}

func (p *sequencePlanner) Name() string    { return "sequence" }
func (p *sequencePlanner) Version() string { return "0.0.0" }

func (p *sequencePlanner) Plan(context.Context, sdkPlanner.Intent, types.ContextSnapshot) (types.Plan, error) {
	if p.call >= len(p.plans) {
		return types.Plan{}, errors.New("no more plans")
	}
	plan := p.plans[p.call]
	p.call++
	return plan, nil
}

func TestExecuteAdaptivePlanReplansAfterContextSegment(t *testing.T) {
	reg := registry.NewToolRegistry()
	contextTool := &adaptiveTestTool{id: "network.route_get"}
	actionTool := &adaptiveTestTool{id: "network.ping"}
	if err := reg.Register(contextTool); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(actionTool); err != nil {
		t.Fatal(err)
	}

	bus := eventbus.NewBus()
	rt := runtime.New(reg, bus, runtimeengine.WithValidator(safety.NewValidator(reg, safety.Config{
		Permissions: []types.Permission{types.PermissionRead},
	})))
	app := &App{
		Registry: reg,
		Events:   bus,
		Runtime:  rt,
	}

	planGen := &sequencePlanner{
		plans: []types.Plan{
			{
				ID:     "plan-replan",
				Intent: "adaptive test",
				Steps: []types.Task{
					{ID: "fresh", Tool: "network.ping"},
				},
				Risk: types.RiskLow,
			},
		},
	}

	initialPlan := types.Plan{
		ID:     "plan-initial",
		Intent: "adaptive test",
		Steps: []types.Task{
			{ID: "gather", Tool: "network.route_get", Purpose: types.TaskPurposeContext},
			{ID: "stale", Tool: "network.ping"},
		},
		Risk: types.RiskLow,
	}

	execution, err := app.executeAdaptivePlan(
		context.Background(),
		planGen,
		safety.NewSimpleSafetyGuard(types.RiskLow),
		sdkPlanner.Intent{Name: "adaptive test"},
		types.ContextSnapshot{},
		initialPlan,
	)
	if err != nil {
		t.Fatalf("executeAdaptivePlan() error: %v", err)
	}
	if execution.Context["network.route_get"] == nil {
		t.Fatalf("expected context tool output in snapshot: %#v", execution.Context)
	}
	if contextTool.calls != 1 {
		t.Fatalf("expected context tool called once, got %d", contextTool.calls)
	}
	if actionTool.calls != 1 {
		t.Fatalf("expected replanned action tool called once, got %d", actionTool.calls)
	}
	if planGen.call != 1 {
		t.Fatalf("expected planner called once for replan, got %d", planGen.call)
	}
}
