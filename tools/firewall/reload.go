package firewall

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type ReloadTool struct{}

func (ReloadTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "firewall.reload",
		Version:        "0.1.0",
		Category:       "firewall",
		Description:    "Reload firewall rules (OpenWrt: /etc/init.d/firewall reload).",
		Permissions:    []types.Permission{types.PermissionAdmin},
		Timeout:        30 * time.Second,
		Risk:           types.RiskMedium,
		SupportsDryRun: false,
	}
}

func (ReloadTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (ReloadTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"success": {Type: types.FieldBoolean},
			"output":  {Type: types.FieldString},
		},
	}
}

func (t ReloadTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t ReloadTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	cmds := []struct {
		bin  string
		args []string
	}{
		{"/etc/init.d/firewall", []string{"reload"}},
		{"fw3", []string{"reload"}},
		{"ufw", []string{"reload"}},
	}

	for _, cmd := range cmds {
		out, err := exec.CommandContext(ctx, cmd.bin, cmd.args...).CombinedOutput()
		if err == nil {
			return types.ToolResult{
				Success: true,
				Output: types.ToolOutput{
					"success": true,
					"output":  string(out),
				},
			}, nil
		}
	}

	return types.ToolResult{}, fmt.Errorf("no firewall reload method found (tried /etc/init.d/firewall, fw3, ufw)")
}
