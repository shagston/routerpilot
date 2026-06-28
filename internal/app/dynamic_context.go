package app

import (
	"context"
	"fmt"

	ctxengine "github.com/shagston/routerpilot/internal/context"
	"github.com/shagston/routerpilot/internal/planner"
	"github.com/shagston/routerpilot/internal/safety"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/types"
)

func (a *App) executeAdaptivePlan(
	ctx context.Context,
	planGen sdkPlanner.Planner,
	guard *safety.SimpleSafetyGuard,
	intent sdkPlanner.Intent,
	snapshot types.ContextSnapshot,
	plan types.Plan,
) (types.Execution, error) {
	pending := append([]types.Task(nil), plan.Steps...)
	currentPlan := plan
	var lastExecution types.Execution

	for len(pending) > 0 {
		segment, tail, segmentIsContext := planner.NextSegment(pending)

		for _, task := range segment {
			if task.Purpose != types.TaskPurposeContext {
				continue
			}
			if err := ctxengine.ValidateContextTask(a.Registry, task); err != nil {
				return types.Execution{}, fmt.Errorf("invalid context task: %w", err)
			}
		}

		subPlan := currentPlan
		subPlan.Steps = segment

		execution, err := a.Runtime.Execute(ctx, subPlan, snapshot)
		snapshot = execution.Context
		lastExecution = execution
		if err != nil {
			return execution, err
		}

		if segmentIsContext && len(tail) > 0 {
			_, contextTail := planner.SplitLeadingActions(tail)

			newPlan, err := planGen.Plan(ctx, intent, snapshot)
			if err != nil {
				return lastExecution, fmt.Errorf("replanning after context gather failed: %w", err)
			}
			if newPlan.ID == "" {
				newPlan.ID = currentPlan.ID
			}
			if newPlan.Intent == "" {
				newPlan.Intent = currentPlan.Intent
			}

			safe, err := guard.Validate(newPlan)
			if err != nil {
				return lastExecution, fmt.Errorf("safety validation error after replan: %w", err)
			}
			if !safe {
				return lastExecution, &SafetyError{Plan: newPlan, Snapshot: snapshot}
			}

			pending = append(append([]types.Task(nil), newPlan.Steps...), contextTail...)
			currentPlan = newPlan
			a.publishEvent("context.replanned", types.SeverityInfo, map[string]any{
				"plan_id":               newPlan.ID,
				"risk":                  newPlan.Risk,
				"steps":                 len(newPlan.Steps),
				"pending_context_steps": len(contextTail),
			})
			continue
		}

		pending = tail
	}

	return lastExecution, nil
}
