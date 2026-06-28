package tool

import (
	"context"

	"github.com/shagston/routerpilot/sdk/types"
)

type Tool interface {
	Metadata() types.ToolMetadata
	InputSchema() types.Schema
	OutputSchema() types.Schema
	Validate(context.Context, types.ToolInput) error
	Execute(context.Context, types.ToolInput) (types.ToolResult, error)
}

type Registry interface {
	Register(Tool) error
	Get(types.ToolID) (Tool, error)
	List() []types.ToolMetadata
}
