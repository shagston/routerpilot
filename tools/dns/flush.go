package dns

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type FlushTool struct{}

func (FlushTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "dns.flush",
		Version:        "0.1.0",
		Category:       "dns",
		Description:    "Flush DNS resolver cache (systemd-resolve, dnsmasq, or nscd).",
		Permissions:    []types.Permission{types.PermissionAdmin},
		Timeout:        10 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: false,
	}
}

func (FlushTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (FlushTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"success": {Type: types.FieldBoolean},
			"method":  {Type: types.FieldString},
		},
	}
}

func (t FlushTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t FlushTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	cmds := []struct {
		name string
		bin  string
		args []string
	}{
		{"systemd-resolve", "systemd-resolve", []string{"--flush-caches"}},
		{"dnsmasq", "/etc/init.d/dnsmasq", []string{"restart"}},
		{"nscd", "nscd", []string{"-i", "hosts"}},
		{"resolvectl", "resolvectl", []string{"flush-caches"}},
	}

	for _, cmd := range cmds {
		err := exec.CommandContext(ctx, cmd.bin, cmd.args...).Run()
		if err == nil {
			return types.ToolResult{
				Success: true,
				Output: types.ToolOutput{
					"success": true,
					"method":  cmd.name,
				},
			}, nil
		}
	}

	return types.ToolResult{}, fmt.Errorf("no DNS flush method found (tried systemd-resolve, dnsmasq, nscd, resolvectl)")
}
