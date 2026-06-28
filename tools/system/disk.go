package system

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type DiskTool struct{}

func (DiskTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "system.disk",
		Version:        "0.1.0",
		Category:       "system",
		Description:    "Show disk usage (df -h).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (DiskTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (DiskTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"mounts": {Type: types.FieldArray},
		},
	}
}

type mountUsage struct {
	Filesystem string `json:"filesystem"`
	Size       string `json:"size"`
	Used       string `json:"used"`
	Available  string `json:"available"`
	UsePercent string `json:"use_percent"`
	MountedOn  string `json:"mounted_on"`
}

func (t DiskTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t DiskTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	if out, err := exec.CommandContext(ctx, "df", "-h").Output(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"mounts": parseDF(string(out)),
			},
		}, nil
	}

	if out, err := exec.CommandContext(ctx, "df", "-k").Output(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"mounts": parseDF(string(out)),
			},
		}, nil
	}

	return types.ToolResult{}, fmt.Errorf("df not available")
}

func parseDF(output string) []mountUsage {
	var mounts []mountUsage
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Filesystem") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 6 {
			mounts = append(mounts, mountUsage{
				Filesystem: parts[0],
				Size:       parts[1],
				Used:       parts[2],
				Available:  parts[3],
				UsePercent: parts[4],
				MountedOn:  parts[5],
			})
		}
	}
	return mounts
}

func mustParseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
