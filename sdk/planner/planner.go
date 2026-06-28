package planner

import (
	"context"

	"github.com/shagston/routerpilot/sdk/types"
)

type Intent struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type ContextProvider interface {
	Build(context.Context, Intent) (types.ContextSnapshot, error)
}

type Planner interface {
	Plan(context.Context, Intent, types.ContextSnapshot) (types.Plan, error)
}
