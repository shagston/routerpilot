package system

import (
	"context"
	"os/exec"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type RebootTool struct{}

func (RebootTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "system.reboot",
		Version:        "0.1.0",
		Category:       "system",
		Description:    "Reboot the system. Requires admin/write permission.",
		Permissions:    []types.Permission{types.PermissionWrite, types.PermissionAdmin},
		Timeout:        10 * time.Second,
		Risk:           types.RiskHigh,
		SupportsDryRun: true,
	}
}

func (RebootTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"force": {Type: types.FieldBoolean, Description: "Force immediate reboot (default false)."},
		},
	}
}

func (RebootTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"reboot_initiated": {Type: types.FieldBoolean},
		},
	}
}

func (t RebootTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t RebootTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	args := []string{"reboot"}
	if force, ok := input["force"].(bool); ok && force {
		args = append(args, "-f")
	}

	err := exec.CommandContext(ctx, args[0], args[1:]...).Run()
	if err != nil {
		return types.ToolResult{
			Success: false,
			Error:   err.Error(),
			Output:  types.ToolOutput{"reboot_initiated": false},
		}, err
	}

	return types.ToolResult{
		Success: true,
		Output:  types.ToolOutput{"reboot_initiated": true},
	}, nil
}
