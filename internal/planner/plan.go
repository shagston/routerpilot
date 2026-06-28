package planner

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/types"
)

func parsePlanResponse(content string, intent planner.Intent) (types.Plan, error) {
	content = stripMarkdownJSON(content)

	var plan types.Plan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return types.Plan{}, fmt.Errorf("failed to parse LLM JSON response: %w. Content: %s", err, content)
	}

	return normalizePlan(plan, intent), nil
}

func normalizePlan(plan types.Plan, intent planner.Intent) types.Plan {
	if plan.ID == "" {
		plan.ID = types.PlanID(fmt.Sprintf("plan-llm-%d", time.Now().UnixNano()))
	}
	if plan.Intent == "" {
		plan.Intent = intent.Name
	}
	if plan.Risk == "" {
		plan.Risk = types.RiskLow
	}
	return plan
}
