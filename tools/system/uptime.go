package system

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type UptimeTool struct{}

func (UptimeTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "system.uptime",
		Version:        "0.1.0",
		Category:       "system",
		Description:    "Get system uptime in seconds and human-readable format.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (UptimeTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (UptimeTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"uptime_seconds": {Type: types.FieldNumber},
			"uptime":         {Type: types.FieldString},
		},
	}
}

func (t UptimeTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t UptimeTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	uptimeSeconds := getUptime(ctx)

	output := types.ToolOutput{
		"uptime_seconds": uptimeSeconds,
		"uptime":         formatDuration(time.Duration(uptimeSeconds) * time.Second),
	}

	return types.ToolResult{Success: true, Output: output}, nil
}

func getUptime(ctx context.Context) float64 {
	if out, err := exec.CommandContext(ctx, "uptime").Output(); err == nil {
		return parseUptimeOutput(string(out))
	}
	return 0
}

func parseUptimeOutput(output string) float64 {
	dayRe := regexp.MustCompile(`up\s+(\d+)\s+day`)
	if matches := dayRe.FindStringSubmatch(output); len(matches) > 1 {
		days, _ := strconv.ParseFloat(matches[1], 64)
		return days * 86400
	}
	hmRe := regexp.MustCompile(`up\s+(\d+):(\d+)`)
	if matches := hmRe.FindStringSubmatch(output); len(matches) > 2 {
		hours, _ := strconv.ParseFloat(matches[1], 64)
		mins, _ := strconv.ParseFloat(matches[2], 64)
		return hours*3600 + mins*60
	}
	return 0
}

func formatDuration(d time.Duration) string {
	days := int64(d.Hours()) / 24
	hours := int64(d.Hours()) % 24
	mins := int64(math.Mod(d.Minutes(), 60))

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
