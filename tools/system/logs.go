package system

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type LogsTool struct{}

func (LogsTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "system.logs",
		Version:        "0.1.0",
		Category:       "system",
		Description:    "View system logs (OpenWrt: logread, Linux: journalctl or dmesg).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        10 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (LogsTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"lines":   {Type: types.FieldInteger, Required: false, Description: "Number of lines to return (default 50, max 500)."},
			"filter":  {Type: types.FieldString, Required: false, Description: "Filter logs by keyword."},
			"service": {Type: types.FieldString, Required: false, Description: "Filter by service name (journalctl only)."},
		},
	}
}

func (LogsTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"entries": {Type: types.FieldArray},
			"source":  {Type: types.FieldString},
			"count":   {Type: types.FieldInteger},
		},
	}
}

func (t LogsTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

type logEntry struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp,omitempty"`
	Service   string `json:"service,omitempty"`
}

func (t LogsTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	lines := 50
	if v, ok := input["lines"].(int); ok && v > 0 && v <= 500 {
		lines = v
	}
	filter, _ := input["filter"].(string)
	service, _ := input["service"].(string)

	if entries, err := readJournalctl(ctx, lines, filter, service); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"entries": entries,
				"source":  "journalctl",
				"count":   len(entries),
			},
		}, nil
	}

	if entries, err := readLogread(ctx, lines, filter); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"entries": entries,
				"source":  "logread",
				"count":   len(entries),
			},
		}, nil
	}

	if entries, err := readDmesg(ctx, lines, filter); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"entries": entries,
				"source":  "dmesg",
				"count":   len(entries),
			},
		}, nil
	}

	return types.ToolResult{}, fmt.Errorf("no log source available (tried journalctl, logread, dmesg)")
}

func readJournalctl(ctx context.Context, lines int, filter, service string) ([]logEntry, error) {
	args := []string{"-n", fmt.Sprintf("%d", lines), "--no-pager", "-o", "short"}
	if service != "" {
		args = append(args, "-u", service)
	}
	out, err := exec.CommandContext(ctx, "journalctl", args...).Output()
	if err != nil {
		return nil, err
	}
	return parseJournalctl(string(out), filter), nil
}

func parseJournalctl(output, filter string) []logEntry {
	var entries []logEntry
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if filter != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(filter)) {
			continue
		}
		entries = append(entries, logEntry{Message: line})
	}
	return entries
}

func readLogread(ctx context.Context, lines int, filter string) ([]logEntry, error) {
	out, err := exec.CommandContext(ctx, "logread", "-l", fmt.Sprintf("%d", lines)).Output()
	if err != nil {
		return nil, err
	}
	return parseLogread(string(out), filter), nil
}

func parseLogread(output, filter string) []logEntry {
	var entries []logEntry
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if filter != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(filter)) {
			continue
		}
		entry := logEntry{Message: line}
		if idx := strings.Index(line, " "); idx > 0 {
			entry.Timestamp = line[:idx]
			entry.Message = strings.TrimSpace(line[idx:])
		}
		entries = append(entries, entry)
	}
	return entries
}

func readDmesg(ctx context.Context, lines int, filter string) ([]logEntry, error) {
	args := []string{"-T", "-l", fmt.Sprintf("%d", lines)}
	out, err := exec.CommandContext(ctx, "dmesg", args...).Output()
	if err != nil {
		return nil, err
	}
	return parseDmesg(string(out), filter), nil
}

func parseDmesg(output, filter string) []logEntry {
	var entries []logEntry
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if filter != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(filter)) {
			continue
		}
		entries = append(entries, logEntry{Message: line})
	}
	return entries
}
