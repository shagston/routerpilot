package vpn

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type StatusTool struct{}

func (StatusTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "vpn.status",
		Version:        "0.1.0",
		Category:       "vpn",
		Description:    "Show VPN tunnel status (OpenVPN, WireGuard).",
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
			"openvpn":  {Type: types.FieldArray},
			"wireguard": {Type: types.FieldArray},
		},
	}
}

type vpnTunnel struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	Local   string `json:"local,omitempty"`
	Remote  string `json:"remote,omitempty"`
	Status  string `json:"status,omitempty"`
}

func (t StatusTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t StatusTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	result := types.ToolOutput{}

	if wg, err := getWireGuardStatus(ctx); err == nil {
		result["wireguard"] = wg
	}

	if ovpn, err := getOpenVPNStatus(ctx); err == nil {
		result["openvpn"] = ovpn
	}

	return types.ToolResult{Success: true, Output: result}, nil
}

func getWireGuardStatus(ctx context.Context) ([]vpnTunnel, error) {
	out, err := exec.CommandContext(ctx, "wg", "show").Output()
	if err != nil {
		return nil, err
	}

	var tunnels []vpnTunnel
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var current vpnTunnel
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "interface:") {
			if current.Name != "" {
				tunnels = append(tunnels, current)
			}
			current = vpnTunnel{
				Name: strings.TrimSpace(strings.TrimPrefix(line, "interface:")),
			}
		} else if strings.Contains(line, "latest handshake:") {
			current.State = "connected"
		} else if strings.Contains(line, "endpoint:") {
			current.Remote = strings.TrimSpace(strings.TrimPrefix(line, "endpoint:"))
		}
	}
	if current.Name != "" {
		tunnels = append(tunnels, current)
	}
	if len(tunnels) == 0 {
		tunnels = append(tunnels, vpnTunnel{State: "down"})
	}
	for i, t := range tunnels {
		if t.State == "" {
			tunnels[i].State = "down"
		}
	}
	return tunnels, nil
}

func getOpenVPNStatus(ctx context.Context) ([]vpnTunnel, error) {
	out, err := exec.CommandContext(ctx, "systemctl", "list-units", "--type=service", "--all", "--no-pager").Output()
	if err != nil {
		return nil, err
	}

	var tunnels []vpnTunnel
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "openvpn") && strings.HasSuffix(strings.Fields(line)[0], ".service") {
			parts := strings.Fields(line)
			t := vpnTunnel{
				Name:   strings.TrimSuffix(parts[0], ".service"),
				State:  parts[2],
			}
			if len(parts) > 3 {
				t.Status = strings.Join(parts[3:], " ")
			}
			tunnels = append(tunnels, t)
		}
	}
	return tunnels, nil
}
