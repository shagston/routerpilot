package main

import (
	"context"
	"fmt"
	"time"

	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/sdk/types"
)

type PingTool struct{}

func (PingTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.ping",
		Version:        "0.1.0",
		Category:       "diagnostics",
		Description:    "Mock ping tool for SDK compilation checks.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (PingTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"host": {Type: types.FieldString, Required: true},
		},
	}
}

func (PingTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"latency_ms":  {Type: types.FieldNumber},
			"packet_loss": {Type: types.FieldNumber},
		},
	}
}

func (PingTool) Validate(_ context.Context, input types.ToolInput) error {
	if _, ok := input["host"].(string); !ok {
		return types.ErrInvalidInput
	}
	return nil
}

func (PingTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	if err := (PingTool{}).Validate(ctx, input); err != nil {
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}
	return types.ToolResult{
		Success: true,
		Output: types.ToolOutput{
			"latency_ms":  1.2,
			"packet_loss": 0,
		},
	}, nil
}

func main() {
	reg := registry.NewToolRegistry()
	if err := reg.Register(PingTool{}); err != nil {
		panic(err)
	}

	for _, metadata := range reg.List() {
		fmt.Printf("%s@%s %s\n", metadata.ID, metadata.Version, metadata.Description)
	}
}
