package network

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type PingTool struct{}

func (PingTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.ping",
		Version:        "0.1.0",
		Category:       "diagnostics",
		Description:    "Send ICMP echo requests to a host and return packet loss and latency.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        10 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (PingTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"host":  {Type: types.FieldString, Required: true, Description: "Host or IP address to ping."},
			"count": {Type: types.FieldInteger, Description: "Number of echo requests. Defaults to 4."},
		},
	}
}

func (PingTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"host":             {Type: types.FieldString},
			"packets_sent":     {Type: types.FieldInteger},
			"packets_received": {Type: types.FieldInteger},
			"packet_loss":      {Type: types.FieldNumber},
			"latency_min_ms":   {Type: types.FieldNumber},
			"latency_max_ms":   {Type: types.FieldNumber},
			"latency_avg_ms":   {Type: types.FieldNumber},
		},
	}
}

func (tool PingTool) Validate(_ context.Context, input types.ToolInput) error {
	host, ok := input["host"].(string)
	if !ok || strings.TrimSpace(host) == "" {
		return fmt.Errorf("%w: host is required", types.ErrInvalidInput)
	}
	if strings.ContainsAny(host, " \t\r\n") {
		return fmt.Errorf("%w: host must not contain whitespace", types.ErrInvalidInput)
	}

	if value, exists := input["count"]; exists {
		count, ok := asInt(value)
		if !ok || count < 1 || count > 10 {
			return fmt.Errorf("%w: count must be between 1 and 10", types.ErrInvalidInput)
		}
	}
	return nil
}

func (tool PingTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	if err := tool.Validate(ctx, input); err != nil {
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}

	host := strings.TrimSpace(input["host"].(string))
	count := 4
	if value, exists := input["count"]; exists {
		count, _ = asInt(value)
	}

	args := pingArgs(host, count)
	cmd := exec.CommandContext(ctx, "ping", args...)
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)
	parsed := parsePingOutput(host, output)
	if err != nil {
		result := types.ToolResult{Success: false, Output: parsed, Error: err.Error(), Retryable: true}
		if ctx.Err() != nil {
			result.Error = ctx.Err().Error()
		}
		return result, err
	}

	return types.ToolResult{
		Success: true,
		Output:  parsed,
	}, nil
}

func pingArgs(host string, count int) []string {
	if runtime.GOOS == "windows" {
		return []string{"-n", strconv.Itoa(count), host}
	}
	return []string{"-c", strconv.Itoa(count), host}
}

func parsePingOutput(host, output string) types.ToolOutput {
	result := types.ToolOutput{"host": host, "raw": output}

	if matches := regexp.MustCompile(`(?i)Sent = (\d+), Received = (\d+), Lost = \d+ \((\d+)% loss\)`).FindStringSubmatch(output); len(matches) == 4 {
		result["packets_sent"] = mustAtoi(matches[1])
		result["packets_received"] = mustAtoi(matches[2])
		result["packet_loss"] = float64(mustAtoi(matches[3]))
	}
	if matches := regexp.MustCompile(`(?i)Minimum = (\d+)ms, Maximum = (\d+)ms, Average = (\d+)ms`).FindStringSubmatch(output); len(matches) == 4 {
		result["latency_min_ms"] = float64(mustAtoi(matches[1]))
		result["latency_max_ms"] = float64(mustAtoi(matches[2]))
		result["latency_avg_ms"] = float64(mustAtoi(matches[3]))
	}
	if matches := regexp.MustCompile(`(?m)(\d+) packets transmitted, (\d+) (?:packets )?received, ([\d.]+)% packet loss`).FindStringSubmatch(output); len(matches) == 4 {
		result["packets_sent"] = mustAtoi(matches[1])
		result["packets_received"] = mustAtoi(matches[2])
		result["packet_loss"] = mustParseFloat(matches[3])
	}
	if matches := regexp.MustCompile(`(?m)rtt min/avg/max/(?:mdev|stddev) = ([\d.]+)/([\d.]+)/([\d.]+)/`).FindStringSubmatch(output); len(matches) == 4 {
		result["latency_min_ms"] = mustParseFloat(matches[1])
		result["latency_avg_ms"] = mustParseFloat(matches[2])
		result["latency_max_ms"] = mustParseFloat(matches[3])
	}

	return result
}

func asInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int8:
		return int(typed), true
	case int16:
		return int(typed), true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case uint:
		return int(typed), true
	case uint8:
		return int(typed), true
	case uint16:
		return int(typed), true
	case uint32:
		return int(typed), true
	case uint64:
		return int(typed), true
	default:
		return 0, false
	}
}

func mustAtoi(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}

func mustParseFloat(value string) float64 {
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}
