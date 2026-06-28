package service

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type ListTool struct{}

func (ListTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "service.list",
		Version:        "0.1.0",
		Category:       "service",
		Description:    "List services and their status (OpenWrt: /etc/init.d/*, Linux: systemctl).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        10 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (ListTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (ListTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"services": {Type: types.FieldArray},
			"source":   {Type: types.FieldString},
		},
	}
}

type svcInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Enabled bool  `json:"enabled,omitempty"`
}

func (t ListTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t ListTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	if services, err := listSystemctl(ctx); err == nil {
		return types.ToolResult{Success: true, Output: types.ToolOutput{"services": services, "source": "systemctl"}}, nil
	}

	if services, err := listInitScripts(ctx); err == nil {
		return types.ToolResult{Success: true, Output: types.ToolOutput{"services": services, "source": "init.d"}}, nil
	}

	return types.ToolResult{Success: true, Output: types.ToolOutput{"services": []svcInfo{}, "source": "none"}}, nil
}

func listSystemctl(ctx context.Context) ([]svcInfo, error) {
	out, err := exec.CommandContext(ctx, "systemctl", "list-units", "--type=service", "--all", "--no-pager").Output()
	if err != nil {
		return nil, err
	}

	var services []svcInfo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "UNIT") || strings.HasPrefix(line, "LOAD") || strings.HasPrefix(line, "●") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 && strings.HasSuffix(parts[0], ".service") {
			services = append(services, svcInfo{
				Name:   strings.TrimSuffix(parts[0], ".service"),
				Status: parts[2],
			})
		}
	}
	return services, nil
}

func listInitScripts(ctx context.Context) ([]svcInfo, error) {
	out, err := exec.CommandContext(ctx, "ls", "/etc/init.d/").Output()
	if err != nil {
		return nil, err
	}

	scripts := strings.Fields(string(out))
	var services []svcInfo

	for _, name := range scripts {
		if name == "README" {
			continue
		}
		s := svcInfo{Name: name}
		if out, err := exec.CommandContext(ctx, "/etc/init.d/"+name, "enabled").Output(); err == nil && strings.TrimSpace(string(out)) == "enabled" {
			s.Enabled = true
		}
		if out, err := exec.CommandContext(ctx, "/etc/init.d/"+name, "running").Output(); err == nil && strings.TrimSpace(string(out)) == "running" {
			s.Status = "running"
		} else {
			s.Status = "stopped"
		}
		services = append(services, s)
	}
	return services, nil
}
