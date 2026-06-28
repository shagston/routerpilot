package planner

import (
	"testing"

	"github.com/shagston/routerpilot/internal/registry"
	networktools "github.com/shagston/routerpilot/tools/network"
	"github.com/shagston/routerpilot/sdk/types"
)

func testRegistry(t *testing.T) *registry.ToolRegistry {
	t.Helper()

	reg := registry.NewToolRegistry()
	if err := reg.Register(networktools.PingTool{}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(networktools.RouteGetTool{}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(networktools.RouteAddTool{}); err != nil {
		t.Fatal(err)
	}
	return reg
}

func TestValidatePlanAcceptsValidPlan(t *testing.T) {
	reg := testRegistry(t)

	plan, err := ValidatePlan(reg, types.Plan{
		ID:     "plan-1",
		Intent: "ping host",
		Risk:   types.RiskLow,
		Steps: []types.Task{
			{
				ID:        "task-1",
				Tool:      "network.ping",
				Arguments: types.ToolInput{"host": "127.0.0.1", "count": 1},
			},
		},
	})
	if err != nil {
		t.Fatalf("ValidatePlan() error: %v", err)
	}
	if plan.Risk != types.RiskLow {
		t.Fatalf("expected low risk, got %q", plan.Risk)
	}
}

func TestValidatePlanRejectsUnknownTool(t *testing.T) {
	reg := testRegistry(t)

	_, err := ValidatePlan(reg, types.Plan{
		ID:     "plan-1",
		Intent: "invalid",
		Steps: []types.Task{
			{ID: "task-1", Tool: "network.nonexistent"},
		},
	})
	if err == nil {
		t.Fatal("expected validation error for unknown tool")
	}
}

func TestValidatePlanRejectsInvalidArguments(t *testing.T) {
	reg := testRegistry(t)

	_, err := ValidatePlan(reg, types.Plan{
		ID:     "plan-1",
		Intent: "ping host",
		Steps: []types.Task{
			{ID: "task-1", Tool: "network.ping", Arguments: types.ToolInput{"count": 1}},
		},
	})
	if err == nil {
		t.Fatal("expected validation error for missing host")
	}
}

func TestValidatePlanRejectsWriteToolMarkedAsContext(t *testing.T) {
	reg := testRegistry(t)

	_, err := ValidatePlan(reg, types.Plan{
		ID:     "plan-1",
		Intent: "add route",
		Steps: []types.Task{
			{
				ID:      "task-1",
				Tool:    "network.route_add",
				Purpose: types.TaskPurposeContext,
				Arguments: types.ToolInput{
					"destination": "10.0.0.0/24",
					"gateway":     "192.168.1.1",
					"interface":   "eth0",
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected validation error for write tool used as context")
	}
}

func TestValidatePlanCoercesRiskUpward(t *testing.T) {
	reg := testRegistry(t)

	plan, err := ValidatePlan(reg, types.Plan{
		ID:     "plan-1",
		Intent: "add route",
		Risk:   types.RiskLow,
		Steps: []types.Task{
			{
				ID:   "task-1",
				Tool: "network.route_add",
				Arguments: types.ToolInput{
					"destination": "10.0.0.0/24",
					"gateway":     "192.168.1.1",
					"interface":   "eth0",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("ValidatePlan() error: %v", err)
	}
	if plan.Risk != types.RiskMedium {
		t.Fatalf("expected coerced medium risk, got %q", plan.Risk)
	}
}

func TestValidatePlanRejectsDependencyCycle(t *testing.T) {
	reg := testRegistry(t)

	_, err := ValidatePlan(reg, types.Plan{
		ID:     "plan-1",
		Intent: "cycle",
		Steps: []types.Task{
			{ID: "a", Tool: "network.ping", Arguments: types.ToolInput{"host": "127.0.0.1"}, Dependencies: []types.TaskID{"b"}},
			{ID: "b", Tool: "network.ping", Arguments: types.ToolInput{"host": "127.0.0.1"}, Dependencies: []types.TaskID{"a"}},
		},
	})
	if err == nil {
		t.Fatal("expected validation error for dependency cycle")
	}
}
