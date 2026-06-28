package runtime

import (
	"context"

	"github.com/shagston/routerpilot/sdk/types"
)

type Validator interface {
	Validate(context.Context, types.Plan) error
}

type Executor interface {
	Execute(context.Context, types.Task) (types.ToolResult, error)
}

type Runtime interface {
	Execute(context.Context, types.Plan, types.ContextSnapshot) (types.Execution, error)
}
