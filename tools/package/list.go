package pkg

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
		ID:             "package.list",
		Version:        "0.1.0",
		Category:       "package",
		Description:    "List installed packages (opkg list-installed or dpkg -l).",
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
			"packages": {Type: types.FieldArray},
			"count":   {Type: types.FieldInteger},
		},
	}
}

type pkgInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

func (t ListTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t ListTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	if out, err := exec.CommandContext(ctx, "opkg", "list-installed").Output(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"packages": parseOpkg(string(out)),
				"count":    strings.Count(string(out), "\n"),
				"source":   "opkg",
			},
		}, nil
	}

	if out, err := exec.CommandContext(ctx, "dpkg", "-l").Output(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"packages": parseDpkg(string(out)),
				"source":   "dpkg",
			},
		}, nil
	}

	return types.ToolResult{
		Success: true,
		Output: types.ToolOutput{
			"packages": []pkgInfo{},
			"source":   "none",
		},
	}, nil
}

func parseOpkg(output string) []pkgInfo {
	var pkgs []pkgInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			p := pkgInfo{Name: parts[0]}
			if len(parts) >= 2 {
				p.Version = parts[1]
			}
			pkgs = append(pkgs, p)
		}
	}
	return pkgs
}

func parseDpkg(output string) []pkgInfo {
	var pkgs []pkgInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ii ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				pkgs = append(pkgs, pkgInfo{Name: parts[1], Version: parts[2]})
			}
		}
	}
	return pkgs
}
