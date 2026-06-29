package dhcp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type ServerTool struct{}

func (ServerTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "dhcp.server",
		Version:        "0.1.0",
		Category:       "dhcp",
		Description:    "Show DHCP server configuration and status (dnsmasq / UCI).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (ServerTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (ServerTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"config":      {Type: types.FieldString},
			"running":     {Type: types.FieldBoolean},
			"source":      {Type: types.FieldString},
		},
	}
}

func (t ServerTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t ServerTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	if cfg, err := readUciDHCP(ctx); err == nil {
		running := isDnsmasqRunning(ctx)
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"source":  "uci",
				"config":  cfg,
				"running": running,
			},
		}, nil
	}

	if cfg, err := readDnsmasqConf(ctx); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"source":  "dnsmasq.conf",
				"config":  cfg,
				"running": true,
			},
		}, nil
	}

	return types.ToolResult{}, fmt.Errorf("no DHCP config source available (tried uci and dnsmasq.conf)")
}

func readUciDHCP(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "uci", "show", "dhcp").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func readDnsmasqConf(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "cat", "/etc/dnsmasq.conf").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func isDnsmasqRunning(ctx context.Context) bool {
	if out, err := exec.CommandContext(ctx, "pgrep", "-x", "dnsmasq").Output(); err == nil && len(out) > 0 {
		return true
	}
	if out, err := exec.CommandContext(ctx, "pidof", "dnsmasq").Output(); err == nil && len(out) > 0 {
		return true
	}
	return false
}
