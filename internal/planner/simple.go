package planner

import (
	"context"
	"fmt"

	"github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/types"
)

type SimplePlanner struct{}

func NewSimplePlanner() *SimplePlanner {
	return &SimplePlanner{}
}

func (p *SimplePlanner) Plan(ctx context.Context, intent planner.Intent, snapshot types.ContextSnapshot) (types.Plan, error) {
	// Simple rule-based planning
	if intent.Name == "ping" {
		target, ok := intent.Arguments["target"].(string)
		if !ok {
			return types.Plan{}, fmt.Errorf("ping intent requires 'target' argument")
		}

		return types.Plan{
			ID:     types.PlanID("plan-simple-ping"),
			Intent: fmt.Sprintf("Ping host %s", target),
			Steps: []types.Task{
				{
					ID:   types.TaskID("task-ping"),
					Tool: "network.ping",
					Arguments: types.ToolInput{
						"host": target,
					},
				},
			},
			Risk: types.RiskLow,
		}, nil
	}

	return types.Plan{}, fmt.Errorf("unsupported intent: %s", intent.Name)
}

type SimpleContextProvider struct{}

func NewSimpleContextProvider() *SimpleContextProvider {
	return &SimpleContextProvider{}
}

func (c *SimpleContextProvider) Build(ctx context.Context, intent planner.Intent) (types.ContextSnapshot, error) {
	// Return empty snapshot for now
	return types.ContextSnapshot{}, nil
}
