package kb

import (
	"context"
	"fmt"
	"time"

	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/sdk/types"
)

type EvidenceCollector struct {
	registry *registry.ToolRegistry
}

func NewEvidenceCollector(reg *registry.ToolRegistry) *EvidenceCollector {
	return &EvidenceCollector{registry: reg}
}

func (ec *EvidenceCollector) Collect(ctx context.Context, checks []string) (map[string]any, error) {
	evidence := make(map[string]any)

	checkFns := map[string]func(context.Context) (any, error){
		"interface_status": ec.collectInterfaceStatus,
		"ip_address":       ec.collectIPAddress,
		"default_route":    ec.collectDefaultRoute,
		"ping_external":    ec.collectPingExternal,
		"dns_resolve":      ec.collectDNSResolve,
		"dhcp_leases":      ec.collectDHCPLeases,
		"connections":      ec.collectConnections,
		"wifi_status":      ec.collectWiFiStatus,
		"system_memory":    ec.collectSystemMemory,
		"system_disk":      ec.collectSystemDisk,
		"system_processes": ec.collectSystemProcesses,
		"vpn_status":       ec.collectVPNStatus,
	}

	for _, check := range checks {
		if fn, ok := checkFns[check]; ok {
			checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			result, err := fn(checkCtx)
			cancel()
			if err != nil {
				evidence[check] = map[string]any{"error": err.Error(), "success": false}
			} else {
				evidence[check] = result
			}
		}
	}

	return evidence, nil
}

func (ec *EvidenceCollector) runTool(ctx context.Context, toolID types.ToolID, input types.ToolInput) (types.ToolResult, error) {
	t, err := ec.registry.Get(toolID)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("tool %s not found: %w", toolID, err)
	}

	if err := t.Validate(ctx, input); err != nil {
		return types.ToolResult{}, fmt.Errorf("validate %s: %w", toolID, err)
	}

	return t.Execute(ctx, input)
}

func (ec *EvidenceCollector) collectInterfaceStatus(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "network.interface_status", types.ToolInput{"interface": "all"})
	if err != nil {
		return nil, err
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}

func (ec *EvidenceCollector) collectIPAddress(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "network.ip_address_get", types.ToolInput{})
	if err != nil {
		return nil, err
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}

func (ec *EvidenceCollector) collectDefaultRoute(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "network.route_get", types.ToolInput{})
	if err != nil {
		return nil, err
	}

	hasDefault := false
	if output := result.Output; output != nil {
		if routes, ok := output["routes"]; ok {
			if routeList, ok := routes.([]any); ok {
				for _, r := range routeList {
					if route, ok := r.(map[string]any); ok {
						if dest, ok := route["destination"]; ok && fmt.Sprintf("%v", dest) == "0.0.0.0/0" {
							hasDefault = true
							break
						}
					}
				}
			}
		}
	}

	return map[string]any{
		"success":     true,
		"has_default": hasDefault,
	}, nil
}

func (ec *EvidenceCollector) collectPingExternal(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "network.ping", types.ToolInput{"host": "8.8.8.8", "count": 2})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}

func (ec *EvidenceCollector) collectDNSResolve(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "dns.lookup", types.ToolInput{"host": "google.com"})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}

func (ec *EvidenceCollector) collectDHCPLeases(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "dhcp.leases", types.ToolInput{})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}

func (ec *EvidenceCollector) collectConnections(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "network.connections", types.ToolInput{"listening": false})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}

func (ec *EvidenceCollector) collectWiFiStatus(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "wifi.status", types.ToolInput{})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}

func (ec *EvidenceCollector) collectSystemMemory(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "system.memory", types.ToolInput{})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}

func (ec *EvidenceCollector) collectSystemDisk(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "system.disk", types.ToolInput{})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}

func (ec *EvidenceCollector) collectSystemProcesses(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "system.processes", types.ToolInput{"sort": "cpu", "limit": 5})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}

func (ec *EvidenceCollector) collectVPNStatus(ctx context.Context) (any, error) {
	result, err := ec.runTool(ctx, "vpn.status", types.ToolInput{})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	output := result.Output
	if output == nil {
		output = types.ToolOutput{}
	}
	output["success"] = result.Success
	return output, nil
}
