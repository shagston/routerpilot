package network

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type NeighborsTool struct{}

func (NeighborsTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.neighbors",
		Version:        "0.1.0",
		Category:       "network",
		Description:    "Show ARP/neighbor table (ip neigh).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (NeighborsTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (NeighborsTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"neighbors": {Type: types.FieldArray},
		},
	}
}

type neighbor struct {
	IP       string `json:"ip"`
	Dev      string `json:"dev"`
	MAC      string `json:"mac,omitempty"`
	State    string `json:"state"`
}

func (t NeighborsTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t NeighborsTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	if out, err := exec.CommandContext(ctx, "ip", "neigh").Output(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"neighbors": parseIPNeigh(string(out)),
			},
		}, nil
	}

	if out, err := exec.CommandContext(ctx, "arp", "-n").Output(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"neighbors": parseArp(string(out)),
			},
		}, nil
	}

	return types.ToolResult{}, fmt.Errorf("no neighbor table source (tried ip neigh, arp -n)")
}

func parseIPNeigh(output string) []neighbor {
	var neighs []neighbor
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		n := neighbor{IP: parts[0], Dev: ""}
		for i, p := range parts {
			if p == "dev" && i+1 < len(parts) {
				n.Dev = parts[i+1]
			}
			if p == "lladdr" && i+1 < len(parts) {
				n.MAC = parts[i+1]
			}
			if p == "REACHABLE" || p == "STALE" || p == "DELAY" || p == "PROBE" || p == "FAILED" || p == "INCOMPLETE" || p == "PERMANENT" || p == "NOARP" {
				n.State = p
			}
		}
		neighs = append(neighs, n)
	}
	return neighs
}

func parseArp(output string) []neighbor {
	var neighs []neighbor
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Address") || strings.HasPrefix(line, "?") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			n := neighbor{IP: parts[0], MAC: parts[1], State: parts[2]}
			if len(parts) >= 4 {
				n.Dev = parts[3]
			}
			neighs = append(neighs, n)
		}
	}
	return neighs
}
