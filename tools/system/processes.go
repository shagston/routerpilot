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

type ProcessesTool struct{}

func (ProcessesTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "system.processes",
		Version:        "0.1.0",
		Category:       "system",
		Description:    "List top processes by CPU or memory usage (ps aux or top).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (ProcessesTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"sort":  {Type: types.FieldString, Required: false, Description: "Sort by: cpu, mem, pid (default cpu)."},
			"limit": {Type: types.FieldInteger, Required: false, Description: "Max processes (default 20, max 100)."},
		},
	}
}

func (ProcessesTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"processes": {Type: types.FieldArray},
			"count":     {Type: types.FieldInteger},
		},
	}
}

type process struct {
	User  string `json:"user"`
	PID   int    `json:"pid"`
	CPU   string `json:"cpu"`
	Mem   string `json:"mem"`
	VSZ   string `json:"vsz,omitempty"`
	RSS   string `json:"rss,omitempty"`
	TTY   string `json:"tty,omitempty"`
	Stat  string `json:"stat,omitempty"`
	Start string `json:"start,omitempty"`
	Time  string `json:"time,omitempty"`
	Command string `json:"command"`
}

func (t ProcessesTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t ProcessesTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	sort, _ := input["sort"].(string)
	if sort == "" {
		sort = "cpu"
	}
	limit := 20
	if v, ok := input["limit"].(int); ok && v > 0 && v <= 100 {
		limit = v
	}

	args := []string{"aux"}
	switch sort {
	case "mem":
		args = []string{"aux", "--sort=-%mem"}
	case "cpu":
		args = []string{"aux", "--sort=-%cpu"}
	}

	if out, err := exec.CommandContext(ctx, "ps", args...).Output(); err == nil {
		procs := parsePS(string(out), limit)
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"processes": procs,
				"count":     len(procs),
			},
		}, nil
	}

	if out, err := exec.CommandContext(ctx, "top", "-b", "-n", "1").Output(); err == nil {
		procs := parseTop(string(out), limit)
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"processes": procs,
				"count":     len(procs),
			},
		}, nil
	}

	return types.ToolResult{}, fmt.Errorf("no process listing source (tried ps, top)")
}

func parsePS(output string, limit int) []process {
	var procs []process
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "USER") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 10 {
			p := process{
				User:    parts[0],
				PID:     mustParseInt(parts[1]),
				CPU:     parts[2],
				Mem:     parts[3],
				VSZ:     parts[4],
				RSS:     parts[5],
				TTY:     parts[6],
				Stat:    parts[7],
				Start:   parts[8],
				Time:    parts[9],
				Command: strings.Join(parts[10:], " "),
			}
			procs = append(procs, p)
			if len(procs) >= limit {
				break
			}
		}
	}
	return procs
}

func parseTop(output string, limit int) []process {
	var procs []process
	scanner := bufio.NewScanner(strings.NewReader(output))
	inProcList := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "PID USER") || strings.Contains(line, "PID %CPU") {
			inProcList = true
			continue
		}
		if !inProcList || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 12 {
			p := process{
				PID:     mustParseInt(parts[0]),
				User:    parts[1],
				CPU:     parts[8],
				Mem:     parts[9],
				Command: strings.Join(parts[11:], " "),
			}
			procs = append(procs, p)
			if len(procs) >= limit {
				break
			}
		}
	}
	return procs
}

func mustParseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
