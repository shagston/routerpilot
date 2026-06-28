package bridge

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type StatusTool struct{}

func (StatusTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "bridge.status",
		Version:        "0.1.0",
		Category:       "bridge",
		Description:    "Show bridge interfaces and their ports (ip link show type bridge or brctl show).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (StatusTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (StatusTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"bridges": {Type: types.FieldArray},
		},
	}
}

type bridgeInfo struct {
	Name      string   `json:"name"`
	Interfaces []string `json:"interfaces"`
	MAC       string   `json:"mac,omitempty"`
	State     string   `json:"state,omitempty"`
}

func (t StatusTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t StatusTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	if bridges, err := getBridgeFromIPLink(ctx); err == nil {
		return types.ToolResult{Success: true, Output: types.ToolOutput{"bridges": bridges}}, nil
	}

	if bridges, err := getBridgeFromBrctl(ctx); err == nil {
		return types.ToolResult{Success: true, Output: types.ToolOutput{"bridges": bridges}}, nil
	}

	return types.ToolResult{}, fmt.Errorf("no bridge source (tried ip link, brctl)")
}

func getBridgeFromIPLink(ctx context.Context) ([]bridgeInfo, error) {
	out, err := exec.CommandContext(ctx, "ip", "link", "show", "type", "bridge").Output()
	if err != nil {
		return nil, err
	}

	var bridges []bridgeInfo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, ":") && (strings.Contains(line, "br-") || strings.Contains(line, "bridge")) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(strings.Split(parts[1], "@")[0])
				state := "up"
				if strings.Contains(line, "DOWN") {
					state = "down"
				}
				mac := ""
				if scanner.Scan() {
					macLine := strings.TrimSpace(scanner.Text())
					if strings.Contains(macLine, "link/ether") {
						macParts := strings.Fields(macLine)
						if len(macParts) >= 2 {
							mac = macParts[1]
						}
					}
				}
				ports, _ := getBridgePorts(ctx, name)
				bridges = append(bridges, bridgeInfo{
					Name:       name,
					State:      state,
					MAC:        mac,
					Interfaces: ports,
				})
			}
		}
	}
	return bridges, nil
}

func getBridgePorts(ctx context.Context, bridge string) ([]string, error) {
	out, err := exec.CommandContext(ctx, "ip", "link", "show", "master", bridge).Output()
	if err != nil {
		return nil, err
	}
	var ports []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, ":") && !strings.Contains(line, "link/") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(strings.Split(parts[1], "@")[0])
				if name != bridge {
					ports = append(ports, name)
				}
			}
		}
	}
	return ports, nil
}

func getBridgeFromBrctl(ctx context.Context) ([]bridgeInfo, error) {
	out, err := exec.CommandContext(ctx, "brctl", "show").Output()
	if err != nil {
		return nil, err
	}
	var bridges []bridgeInfo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var current *bridgeInfo
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "bridge name") || strings.Contains(line, "---") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			if current == nil || current.Name != parts[0] {
				if current != nil {
					bridges = append(bridges, *current)
				}
				current = &bridgeInfo{Name: parts[0]}
			}
			if len(parts) >= 4 {
				current.Interfaces = append(current.Interfaces, parts[3])
			}
		}
	}
	if current != nil {
		bridges = append(bridges, *current)
	}
	return bridges, nil
}
