package system

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shagston/routerpilot/sdk/types"
)

type InfoTool struct{}

func (InfoTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "system.info",
		Version:        "0.1.0",
		Category:       "system",
		Description:    "Retrieve system information including OS, kernel, architecture, and hostname.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        0,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (InfoTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (InfoTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"hostname":     {Type: types.FieldString},
			"os":           {Type: types.FieldString},
			"kernel":       {Type: types.FieldString},
			"architecture": {Type: types.FieldString},
			"go_version":   {Type: types.FieldString},
		},
	}
}

func (t InfoTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t InfoTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	info := map[string]any{
		"go_version":   runtime.Version(),
		"os":           runtime.GOOS,
		"architecture": runtime.GOARCH,
	}

	if hostname, err := exec.CommandContext(ctx, "hostname").Output(); err == nil {
		info["hostname"] = strings.TrimSpace(string(hostname))
	}

	if uname, err := exec.CommandContext(ctx, "uname", "-r").Output(); err == nil {
		info["kernel"] = strings.TrimSpace(string(uname))
	} else {
		info["kernel"] = fmt.Sprintf("unknown (%v)", err)
	}

	return types.ToolResult{
		Success: true,
		Output:  info,
	}, nil
}
