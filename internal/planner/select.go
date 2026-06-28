package planner

import (
	"os"
	"strings"

	"github.com/shagston/routerpilot/internal/registry"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
)

func SelectPlanner(reg *registry.ToolRegistry) sdkPlanner.Planner {
	switch strings.ToLower(os.Getenv("ROUTERPILOT_PLANNER")) {
	case "simple":
		return NewSimplePlanner()
	case "llm":
		return NewLLMPlanner(reg)
	default:
		if os.Getenv("ROUTERPILOT_API_KEY") != "" {
			return NewLLMPlanner(reg)
		}
		return NewSimplePlanner()
	}
}
