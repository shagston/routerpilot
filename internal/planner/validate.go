package planner

import (
	"fmt"

	ctxengine "github.com/shagston/routerpilot/internal/context"
	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/internal/safety"
	"github.com/shagston/routerpilot/sdk/types"
)

func ValidatePlan(reg *registry.ToolRegistry, plan types.Plan) (types.Plan, error) {
	if len(plan.Steps) == 0 {
		return plan, fmt.Errorf("%w: plan has no steps", types.ErrInvalidInput)
	}

	seen := make(map[types.TaskID]struct{}, len(plan.Steps))
	maxToolRisk := types.RiskLow

	for i := range plan.Steps {
		task := &plan.Steps[i]
		if err := normalizeTaskPurpose(task); err != nil {
			return plan, err
		}
		if err := validateTaskIdentity(*task, seen); err != nil {
			return plan, err
		}

		t, err := reg.Get(task.Tool)
		if err != nil {
			return plan, fmt.Errorf("%w: unknown tool %s in task %s", types.ErrNotFound, task.Tool, task.ID)
		}

		metadata := t.Metadata()
		maxToolRisk = maxRiskLevel(maxToolRisk, metadata.Risk)

		if task.Purpose == types.TaskPurposeContext {
			if err := ctxengine.ValidateContextTask(reg, *task); err != nil {
				return plan, fmt.Errorf("context task %s: %w", task.ID, err)
			}
		}

		if err := safety.ValidateInput(t.InputSchema(), task.Arguments); err != nil {
			return plan, fmt.Errorf("%w: task %s", err, task.ID)
		}
	}

	if err := validateDependencies(plan.Steps, seen); err != nil {
		return plan, err
	}
	if types.HasDependencyCycle(plan.Steps) {
		return plan, fmt.Errorf("%w: dependency cycle", types.ErrInvalidInput)
	}

	plan.Risk = coercePlanRisk(plan.Risk, maxToolRisk)
	if !isValidRisk(plan.Risk) {
		return plan, fmt.Errorf("%w: invalid plan risk %q", types.ErrInvalidInput, plan.Risk)
	}

	return plan, nil
}

func normalizeTaskPurpose(task *types.Task) error {
	switch task.Purpose {
	case "":
		task.Purpose = types.TaskPurposeAction
	case types.TaskPurposeAction, types.TaskPurposeContext:
	default:
		return fmt.Errorf("%w: invalid purpose %q on task %s", types.ErrInvalidInput, task.Purpose, task.ID)
	}
	return nil
}

func validateTaskIdentity(task types.Task, seen map[types.TaskID]struct{}) error {
	if task.ID == "" {
		return fmt.Errorf("%w: task missing id", types.ErrInvalidInput)
	}
	if task.Tool == "" {
		return fmt.Errorf("%w: task %s missing tool", types.ErrInvalidInput, task.ID)
	}
	if _, exists := seen[task.ID]; exists {
		return fmt.Errorf("%w: duplicate task %s", types.ErrInvalidInput, task.ID)
	}
	seen[task.ID] = struct{}{}
	return nil
}

func validateDependencies(steps []types.Task, seen map[types.TaskID]struct{}) error {
	for _, task := range steps {
		for _, dependency := range task.Dependencies {
			if _, exists := seen[dependency]; !exists {
				return fmt.Errorf("%w: missing dependency %s", types.ErrInvalidInput, dependency)
			}
		}
	}
	return nil
}

func coercePlanRisk(planRisk, maxToolRisk types.RiskLevel) types.RiskLevel {
	if planRisk == "" {
		return maxToolRisk
	}
	if !isValidRisk(planRisk) {
		return planRisk
	}
	if riskRank(planRisk) < riskRank(maxToolRisk) {
		return maxToolRisk
	}
	return planRisk
}

func isValidRisk(risk types.RiskLevel) bool {
	switch risk {
	case types.RiskLow, types.RiskMedium, types.RiskHigh, types.RiskCritical:
		return true
	default:
		return false
	}
}

func riskRank(risk types.RiskLevel) int {
	switch risk {
	case types.RiskLow:
		return 1
	case types.RiskMedium:
		return 2
	case types.RiskHigh:
		return 3
	case types.RiskCritical:
		return 4
	default:
		return 0
	}
}

func maxRiskLevel(a, b types.RiskLevel) types.RiskLevel {
	if riskRank(a) >= riskRank(b) {
		return a
	}
	return b
}


