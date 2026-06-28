package service

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type RestartTool struct{}

func (RestartTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "service.restart",
		Version:        "0.1.0",
		Category:       "service",
		Description:    "Restart a service (OpenWrt: /etc/init.d/<name> restart, Linux: systemctl restart).",
		Permissions:    []types.Permission{types.PermissionAdmin},
		Timeout:        30 * time.Second,
		Risk:           types.RiskMedium,
		SupportsDryRun: false,
	}
}

func (RestartTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"name": {Type: types.FieldString, Required: true, Description: "Service name."},
		},
	}
}

func (RestartTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"service": {Type: types.FieldString},
			"success": {Type: types.FieldBoolean},
			"output":  {Type: types.FieldString},
		},
	}
}

func (t RestartTool) Validate(_ context.Context, input types.ToolInput) error {
	name, ok := input["name"].(string)
	if !ok || strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: service name is required", types.ErrInvalidInput)
	}
	return nil
}

func (t RestartTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	name := strings.TrimSpace(input["name"].(string))

	if err := exec.CommandContext(ctx, "/etc/init.d/"+name, "restart").Run(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"service": name,
				"success": true,
				"method":  "init.d",
			},
		}, nil
	}

	if err := exec.CommandContext(ctx, "systemctl", "restart", name+".service").Run(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"service": name,
				"success": true,
				"method":  "systemctl",
			},
		}, nil
	}

	return types.ToolResult{}, fmt.Errorf("failed to restart service %s", name)
}
