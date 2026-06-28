package network

import (
	"context"
	"fmt"
	"time"

	"github.com/shagston/routerpilot/internal/network"
	"github.com/shagston/routerpilot/sdk/types"
)

type RouteGetTool struct {
	Provider network.Provider
}

func (RouteGetTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.route_get",
		Version:        "0.1.0",
		Category:       "network",
		Description:    "List all active routes in the routing table.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (RouteGetTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (RouteGetTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"routes": {Type: types.FieldString, Description: "Formatted list of routes"},
		},
	}
}

func (tool RouteGetTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (tool RouteGetTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	routes, err := tool.Provider.GetRoutes()
	if err != nil {
		return types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get routes: %v", err),
		}, err
	}

	if len(routes) == 0 {
		return types.ToolResult{
			Success: true,
			Output:  types.ToolOutput{"routes": "No routes configured"},
		}, nil
	}

	var result string
	for _, r := range routes {
		result += fmt.Sprintf("Dest: %s, Gateway: %s, Interface: %s\n", r.Destination, r.Gateway, r.Interface)
	}

	return types.ToolResult{
		Success: true,
		Output:  types.ToolOutput{"routes": result},
	}, nil
}

type RouteAddTool struct {
	Provider network.Provider
}

func (RouteAddTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.route_add",
		Version:        "0.1.0",
		Category:       "network",
		Description:    "Add a new static route to the routing table.",
		Permissions:    []types.Permission{types.PermissionWrite},
		Timeout:        5 * time.Second,
		Risk:           types.RiskMedium,
		SupportsDryRun: false,
	}
}

func (RouteAddTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"destination": {Type: types.FieldString, Required: true, Description: "Destination network (e.g., '10.0.0.0/24')."},
			"gateway":     {Type: types.FieldString, Required: true, Description: "Next hop gateway IP."},
			"interface":   {Type: types.FieldString, Required: true, Description: "Interface to use for this route."},
		},
	}
}

func (RouteAddTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"success": {Type: types.FieldBoolean},
			"message": {Type: types.FieldString},
		},
	}
}

func (tool RouteAddTool) Validate(_ context.Context, input types.ToolInput) error {
	if dest, ok := input["destination"].(string); !ok || dest == "" {
		return fmt.Errorf("%w: destination is required", types.ErrInvalidInput)
	}
	if gw, ok := input["gateway"].(string); !ok || gw == "" {
		return fmt.Errorf("%w: gateway is required", types.ErrInvalidInput)
	}
	if iface, ok := input["interface"].(string); !ok || iface == "" {
		return fmt.Errorf("%w: interface is required", types.ErrInvalidInput)
	}
	return nil
}

func (tool RouteAddTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	if err := tool.Validate(ctx, input); err != nil {
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}

	dest := input["destination"].(string)
	gw := input["gateway"].(string)
	iface := input["interface"].(string)

	err := tool.Provider.AddRoute(dest, gw, iface)
	if err != nil {
		return types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to add route: %v", err),
		}, err
	}

	return types.ToolResult{
		Success: true,
		Output: types.ToolOutput{
			"success": true,
			"message": fmt.Sprintf("Successfully added route to %s via %s on %s", dest, gw, iface),
		},
	}, nil
}
