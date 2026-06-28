package network

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type TracerouteTool struct{}

func (TracerouteTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.traceroute",
		Version:        "0.1.0",
		Category:       "diagnostics",
		Description:    "Trace network path to a host (traceroute/tracert).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        30 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (TracerouteTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"host": {Type: types.FieldString, Required: true, Description: "Host or IP to trace."},
			"max_hops": {Type: types.FieldInteger, Required: false, Description: "Maximum hops (default 30)."},
		},
	}
}

func (TracerouteTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"hops": {Type: types.FieldArray},
			"host": {Type: types.FieldString},
		},
	}
}

func (t TracerouteTool) Validate(_ context.Context, input types.ToolInput) error {
	host, ok := input["host"].(string)
	if !ok || strings.TrimSpace(host) == "" {
		return fmt.Errorf("%w: host is required", types.ErrInvalidInput)
	}
	return nil
}

func (t TracerouteTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	host := strings.TrimSpace(input["host"].(string))
	maxHops := 30
	if v, ok := input["max_hops"].(int); ok && v > 0 && v <= 64 {
		maxHops = v
	}

	var args []string
	bin := "traceroute"
	if runtime.GOOS == "windows" {
		bin = "tracert"
		args = []string{"-h", fmt.Sprintf("%d", maxHops), host}
	} else {
		args = []string{"-m", fmt.Sprintf("%d", maxHops), "-n", host}
	}

	out, err := exec.CommandContext(ctx, bin, args...).Output()
	if err != nil {
		return types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("%s failed: %v", bin, err),
			Retryable: ctx.Err() == nil,
		}, nil
	}

	return types.ToolResult{
		Success: true,
		Output: types.ToolOutput{
			"host": host,
			"hops": strings.Split(strings.TrimSpace(string(out)), "\n"),
		},
	}, nil
}
