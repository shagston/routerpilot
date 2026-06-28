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

type MemoryTool struct{}

func (MemoryTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "system.memory",
		Version:        "0.1.0",
		Category:       "system",
		Description:    "Show system memory usage (free or /proc/meminfo).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (MemoryTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (MemoryTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"total":     {Type: types.FieldInteger},
			"used":      {Type: types.FieldInteger},
			"free":      {Type: types.FieldInteger},
			"available": {Type: types.FieldInteger},
			"swap_total": {Type: types.FieldInteger},
			"swap_used":  {Type: types.FieldInteger},
		},
	}
}

func (t MemoryTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t MemoryTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	if out, err := exec.CommandContext(ctx, "free", "-b").Output(); err == nil {
		return parseFree(string(out))
	}

	if out, err := exec.CommandContext(ctx, "cat", "/proc/meminfo").Output(); err == nil {
		return parseMeminfo(string(out))
	}

	return types.ToolResult{}, fmt.Errorf("no memory source (tried free, /proc/meminfo)")
}

func parseFree(output string) (types.ToolResult, error) {
	result := types.ToolOutput{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Mem:") {
			parts := strings.Fields(line)
			if len(parts) >= 7 {
				result["total"], _ = strconv.ParseInt(parts[1], 10, 64)
				result["used"], _ = strconv.ParseInt(parts[2], 10, 64)
				result["free"], _ = strconv.ParseInt(parts[3], 10, 64)
				result["available"], _ = strconv.ParseInt(parts[6], 10, 64)
			}
		} else if strings.HasPrefix(line, "Swap:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				result["swap_total"], _ = strconv.ParseInt(parts[1], 10, 64)
				result["swap_used"], _ = strconv.ParseInt(parts[2], 10, 64)
			}
		}
	}
	return types.ToolResult{Success: true, Output: result}, nil
}

func parseMeminfo(output string) (types.ToolResult, error) {
	mem := map[string]int64{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			valStr := strings.TrimSpace(parts[1])
			valStr = strings.TrimSuffix(valStr, " kB")
			valStr = strings.TrimSpace(valStr)
			if val, err := strconv.ParseInt(valStr, 10, 64); err == nil {
				mem[key] = val * 1024
			}
		}
	}

	result := types.ToolOutput{}
	if v, ok := mem["MemTotal"]; ok {
		result["total"] = v
		result["free"] = mem["MemFree"]
		result["available"] = mem["MemAvailable"]
		result["used"] = v - mem["MemAvailable"]
	}
	if v, ok := mem["SwapTotal"]; ok {
		result["swap_total"] = v
		result["swap_used"] = v - mem["SwapFree"]
	}

	return types.ToolResult{Success: true, Output: result}, nil
}
