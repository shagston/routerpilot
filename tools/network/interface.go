package network

import (
	"context"
	"fmt"
	"time"

	"github.com/shagston/routerpilot/internal/network"
	"github.com/shagston/routerpilot/sdk/types"
)

type InterfaceStatusTool struct {
	Provider network.Provider
}

func (InterfaceStatusTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.interface_status",
		Version:        "0.1.0",
		Category:       "network",
		Description:    "Get the current status (up/down) and details of a network interface.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (InterfaceStatusTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"interface": {Type: types.FieldString, Required: true, Description: "Name of the interface (e.g., eth0) or 'all'."},
		},
	}
}

func (InterfaceStatusTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"interface": {Type: types.FieldString},
			"status":    {Type: types.FieldString},
			"active":    {Type: types.FieldBoolean},
			"interfaces": {Type: types.FieldArray},
		},
	}
}

func (tool InterfaceStatusTool) Validate(_ context.Context, input types.ToolInput) error {
	if name, ok := input["interface"].(string); !ok || name == "" {
		return fmt.Errorf("%w: interface name is required", types.ErrInvalidInput)
	}
	return nil
}

func (tool InterfaceStatusTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	if err := tool.Validate(ctx, input); err != nil {
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}

	name := input["interface"].(string)
	statuses, err := tool.Provider.GetInterfaceStatus()
	if err != nil {
		return types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get interface status: %v", err),
		}, err
	}

	if name == "all" {
		entries := make([]types.ToolOutput, 0, len(statuses))
		for _, status := range statuses {
			entries = append(entries, types.ToolOutput{
				"interface": status.Name,
				"status":    fmt.Sprintf("%v", status.Up),
				"active":    status.Active,
			})
		}
		return types.ToolResult{
			Success: true,
			Output:  types.ToolOutput{"interfaces": entries},
		}, nil
	}

	for _, status := range statuses {
		if status.Name == name {
			return types.ToolResult{
				Success: true,
				Output: types.ToolOutput{
					"interface": status.Name,
					"status":    fmt.Sprintf("%v", status.Up),
					"active":    status.Active,
				},
			}, nil
		}
	}

	err = fmt.Errorf("interface %s not found", name)
	return types.ToolResult{Success: false, Error: err.Error()}, err
}

type InterfaceSetStateTool struct {
	Provider network.Provider
}

func (InterfaceSetStateTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.interface_set_state",
		Version:        "0.1.0",
		Category:       "network",
		Description:    "Set the state of a network interface (up or down).",
		Permissions:    []types.Permission{types.PermissionWrite},
		Timeout:        5 * time.Second,
		Risk:           types.RiskMedium,
		SupportsDryRun: true,
	}
}

func (InterfaceSetStateTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"interface": {Type: types.FieldString, Required: true, Description: "Name of the interface."},
			"state":     {Type: types.FieldString, Required: true, Description: "Desired state: 'up' or 'down'."},
		},
	}
}

func (InterfaceSetStateTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"interface": {Type: types.FieldString},
			"new_state": {Type: types.FieldString},
			"success":   {Type: types.FieldBoolean},
		},
	}
}

func (tool InterfaceSetStateTool) Validate(_ context.Context, input types.ToolInput) error {
	if name, ok := input["interface"].(string); !ok || name == "" {
		return fmt.Errorf("%w: interface name is required", types.ErrInvalidInput)
	}
	state, ok := input["state"].(string)
	if !ok || (state != "up" && state != "down") {
		return fmt.Errorf("%w: state must be 'up' or 'down'", types.ErrInvalidInput)
	}
	return nil
}

func (tool InterfaceSetStateTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	if err := tool.Validate(ctx, input); err != nil {
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}

	name := input["interface"].(string)
	state := input["state"].(string)
	up := (state == "up")

	err := tool.Provider.SetInterfaceState(name, up)
	if err != nil {
		return types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to set %s to %s: %v", name, state, err),
		}, err
	}

	return types.ToolResult{
		Success: true,
		Output: types.ToolOutput{
			"interface": name,
			"new_state": state,
			"success":   true,
		},
	}, nil
}
