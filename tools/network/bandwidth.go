package network

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type BandwidthTool struct{}

func (BandwidthTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.bandwidth",
		Version:        "0.1.0",
		Category:       "diagnostics",
		Description:    "Measure network bandwidth to a target using iperf3, curl, or fallback methods.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        30 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (BandwidthTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"target":    {Type: types.FieldString, Required: true, Description: "Target host or iperf3 server address."},
			"direction": {Type: types.FieldString, Required: false, Description: "Test direction: 'download' (default) or 'upload'."},
			"duration":  {Type: types.FieldInteger, Required: false, Description: "Test duration in seconds (default 10)."},
		},
	}
}

func (BandwidthTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"target":        {Type: types.FieldString},
			"direction":     {Type: types.FieldString},
			"throughput_mbps": {Type: types.FieldNumber},
			"jitter_ms":     {Type: types.FieldNumber},
			"method":        {Type: types.FieldString},
		},
	}
}

func (t BandwidthTool) Validate(_ context.Context, input types.ToolInput) error {
	target, ok := input["target"].(string)
	if !ok || target == "" {
		return fmt.Errorf("'target' is required")
	}
	if d, ok := input["direction"].(string); ok && d != "" {
		if d != "download" && d != "upload" {
			return fmt.Errorf("direction must be 'download' or 'upload'")
		}
	}
	return nil
}

func (t BandwidthTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	target, _ := input["target"].(string)
	direction, _ := input["direction"].(string)
	if direction == "" {
		direction = "download"
	}
	duration := 10
	if d, ok := input["duration"].(int); ok && d > 0 && d <= 30 {
		duration = d
	}

	result := types.ToolOutput{
		"target":    target,
		"direction": direction,
	}

	if out, err := iperf3Test(ctx, target, direction, duration); err == nil {
		result["method"] = "iperf3"
		for k, v := range out {
			result[k] = v
		}
		return types.ToolResult{Success: true, Output: result}, nil
	}

	if out, err := curlSpeedTest(ctx, target, direction, duration); err == nil {
		result["method"] = "curl"
		for k, v := range out {
			result[k] = v
		}
		return types.ToolResult{Success: true, Output: result}, nil
	}

	return types.ToolResult{
		Success: false,
		Error:   "no bandwidth measurement method available (tried iperf3, curl)",
	}, nil
}

func iperf3Test(ctx context.Context, target, direction string, duration int) (types.ToolOutput, error) {
	if _, err := exec.LookPath("iperf3"); err != nil {
		return nil, err
	}

	args := []string{"-c", target, "-t", strconv.Itoa(duration), "-J", "-f", "m"}
	if direction == "upload" {
		args = append(args, "-R")
	}

	cmd := exec.CommandContext(ctx, "iperf3", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iperf3: %w", err)
	}

	return parseIperf3JSON(string(out)), nil
}

func parseIperf3JSON(output string) types.ToolOutput {
	result := types.ToolOutput{}

	re := regexp.MustCompile(`"bits_per_second":\s*([\d.]+)`)
	if matches := re.FindStringSubmatch(output); len(matches) == 2 {
		if bps, err := strconv.ParseFloat(matches[1], 64); err == nil {
			result["throughput_mbps"] = bps / 1_000_000
		}
	}

	reJitter := regexp.MustCompile(`"jitter_ms":\s*([\d.]+)`)
	if matches := reJitter.FindStringSubmatch(output); len(matches) == 2 {
		if jitter, err := strconv.ParseFloat(matches[1], 64); err == nil {
			result["jitter_ms"] = jitter
		}
	}

	reLost := regexp.MustCompile(`"lost_packets":\s*(\d+)`)
	if matches := reLost.FindStringSubmatch(output); len(matches) == 2 {
		if lost, err := strconv.Atoi(matches[1]); err == nil {
			result["lost_packets"] = lost
		}
	}

	return result
}

func curlSpeedTest(ctx context.Context, target, direction string, duration int) (types.ToolOutput, error) {
	if _, err := exec.LookPath("curl"); err != nil {
		return nil, err
	}

	output := types.ToolOutput{}

	if direction == "upload" {
		url := fmt.Sprintf("https://%s/bench", target)
		cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{speed_upload}", "--connect-timeout", "5", "--max-time", strconv.Itoa(duration), "-F", "file=@/dev/null", url)
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("curl upload: %w", err)
		}
		speed := parseCurlSpeed(string(out))
		if speed > 0 {
			output["throughput_mbps"] = speed
			return output, nil
		}
		return nil, fmt.Errorf("curl upload returned zero speed")
	}

	url := fmt.Sprintf("http://%s/bench", target)
	cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{speed_download}", "--connect-timeout", "5", "--max-time", strconv.Itoa(duration), url)
	out, err := cmd.Output()
	if err != nil {
		cmd2 := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{speed_download}", "--connect-timeout", "5", "--max-time", strconv.Itoa(duration), "http://speedtest.tele2.net/100MB.zip")
		out, err = cmd2.Output()
		if err != nil {
			return nil, fmt.Errorf("curl download: %w", err)
		}
	}

	speed := parseCurlSpeed(string(out))
	if speed > 0 {
		output["throughput_mbps"] = speed
		return output, nil
	}

	return nil, fmt.Errorf("curl returned zero speed")
}

func parseCurlSpeed(speedStr string) float64 {
	speedStr = strings.TrimSpace(speedStr)
	if speed, err := strconv.ParseFloat(speedStr, 64); err == nil {
		return speed * 8 / 1_000_000
	}
	return 0
}
