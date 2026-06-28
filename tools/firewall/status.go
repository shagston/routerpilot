package firewall

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type StatusTool struct{}

func (StatusTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "firewall.status",
		Version:        "0.1.0",
		Category:       "firewall",
		Description:    "Show firewall rules and status via iptables -L or uci show firewall.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        10 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (StatusTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (StatusTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"rules":       {Type: types.FieldArray},
			"uci_config":  {Type: types.FieldString},
			"source":      {Type: types.FieldString},
		},
	}
}

func (t StatusTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t StatusTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	result := types.ToolOutput{}

	if out, err := exec.CommandContext(ctx, "iptables", "-L", "-n").Output(); err == nil {
		result["source"] = "iptables"
		result["rules"] = parseIPTables(string(out))
		return types.ToolResult{Success: true, Output: result}, nil
	}

	if out, err := exec.CommandContext(ctx, "uci", "show", "firewall").Output(); err == nil {
		result["source"] = "uci"
		result["uci_config"] = string(out)
		return types.ToolResult{Success: true, Output: result}, nil
	}

	result["source"] = "unsupported"
	result["rules"] = []string{}
	return types.ToolResult{Success: true, Output: result}, nil
}

func parseIPTables(output string) []string {
	var rules []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	inChain := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Chain ") {
			inChain = true
			rules = append(rules, line)
			continue
		}
		if inChain && line != "" && !strings.HasPrefix(line, "target") && !strings.HasPrefix(line, "pkts") && !strings.HasPrefix(line, "num") {
			rules = append(rules, line)
		}
	}
	return rules
}
