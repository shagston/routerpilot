package network

import (
	"context"
	"fmt"
	"time"

	"github.com/shagston/routerpilot/internal/network"
	"github.com/shagston/routerpilot/sdk/types"
)

type IPAddressGetTool struct {
	Provider network.Provider
}

func (IPAddressGetTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.ip_address_get",
		Version:        "0.1.0",
		Category:       "network",
		Description:    "List all configured IP addresses on all interfaces.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (IPAddressGetTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (IPAddressGetTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"addresses": {Type: types.FieldString, Description: "Formatted list of IP addresses"},
		},
	}
}

func (tool IPAddressGetTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (tool IPAddressGetTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	addrs, err := tool.Provider.GetAddresses()
	if err != nil {
		return types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get addresses: %v", err),
		}, err
	}

	if len(addrs) == 0 {
		return types.ToolResult{
			Success: true,
			Output:  types.ToolOutput{"addresses": "No IP addresses configured"},
		}, nil
	}

	var result string
	for _, a := range addrs {
		result += fmt.Sprintf("Interface: %s, Address: %s\n", a.Interface, a.Address)
	}

	return types.ToolResult{
		Success: true,
		Output:  types.ToolOutput{"addresses": result},
	}, nil
}

type IPAddressSetTool struct {
	Provider network.Provider
}

func (IPAddressSetTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.ip_address_set",
		Version:        "0.1.0",
		Category:       "network",
		Description:    "Assign an IP address to a specific interface.",
		Permissions:    []types.Permission{types.PermissionWrite},
		Timeout:        5 * time.Second,
		Risk:           types.RiskMedium,
		SupportsDryRun: false,
	}
}

func (IPAddressSetTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"interface": {Type: types.FieldString, Required: true, Description: "Name of the interface (e.g., eth0)."},
			"address":   {Type: types.FieldString, Required: true, Description: "IP address with CIDR (e.g., '192.168.1.1/24')."},
		},
	}
}

func (IPAddressSetTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"success": {Type: types.FieldBoolean},
			"message": {Type: types.FieldString},
		},
	}
}

func (tool IPAddressSetTool) Validate(_ context.Context, input types.ToolInput) error {
	if iface, ok := input["interface"].(string); !ok || iface == "" {
		return fmt.Errorf("%w: interface name is required", types.ErrInvalidInput)
	}
	if addr, ok := input["address"].(string); !ok || addr == "" {
		return fmt.Errorf("%w: address is required", types.ErrInvalidInput)
	}
	return nil
}

func (tool IPAddressSetTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	if err := tool.Validate(ctx, input); err != nil {
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}

	iface := input["interface"].(string)
	addr := input["address"].(string)

	err := tool.Provider.AddAddress(iface, addr)
	if err != nil {
		return types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to set IP address: %v", err),
		}, err
	}

	return types.ToolResult{
		Success: true,
		Output: types.ToolOutput{
			"success": true,
			"message": fmt.Sprintf("Successfully assigned %s to %s", addr, iface),
		},
	}, nil
}
