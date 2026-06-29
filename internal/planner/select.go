package planner

import (
	"strings"

	"github.com/shagston/routerpilot/internal/config"
	"github.com/shagston/routerpilot/internal/registry"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
)

func SelectPlanner(reg *registry.ToolRegistry, cfg *config.Config) sdkPlanner.Planner {
	if cfg == nil {
		return NewSimplePlanner()
	}

	switch strings.ToLower(cfg.Planner.Type) {
	case "simple":
		return NewSimplePlanner()
	case "llm":
		return NewLLMPlanner(reg, cfg)
	default:
		if cfg.Planner.APIKey != "" {
			return NewLLMPlanner(reg, cfg)
		}
		return NewSimplePlanner()
	}
}
