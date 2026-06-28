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

type ConnectionsTool struct{}

func (ConnectionsTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "network.connections",
		Version:        "0.1.0",
		Category:       "network",
		Description:    "Show active network connections (ss -tuln).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (ConnectionsTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"listening": {Type: types.FieldBoolean, Required: false, Description: "Show only listening sockets (default true)."},
			"tcp":       {Type: types.FieldBoolean, Required: false, Description: "Show TCP sockets."},
			"udp":       {Type: types.FieldBoolean, Required: false, Description: "Show UDP sockets."},
		},
	}
}

func (ConnectionsTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"connections": {Type: types.FieldArray},
		},
	}
}

type connection struct {
	Proto   string `json:"proto"`
	State   string `json:"state,omitempty"`
	Local   string `json:"local"`
	Peer    string `json:"peer"`
	Process string `json:"process,omitempty"`
}

func (t ConnectionsTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t ConnectionsTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	listening, _ := input["listening"].(bool)
	tcp, _ := input["tcp"].(bool)
	udp, _ := input["udp"].(bool)

	args := []string{"-tuln"}
	if !tcp && !udp {
		tcp, udp = true, true
	}
	if !tcp {
		args = []string{"-uln"}
	}
	if !udp {
		args = []string{"-tln"}
	}
	if !listening {
		args = []string{"-tun"}
	}

	if out, err := exec.CommandContext(ctx, "ss", args...).Output(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"connections": parseSS(string(out)),
			},
		}, nil
	}

	if out, err := exec.CommandContext(ctx, "netstat", append([]string{"-n"}, args...)...).Output(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"connections": parseNetstat(string(out)),
			},
		}, nil
	}

	return types.ToolResult{}, fmt.Errorf("no connection source (tried ss, netstat)")
}

func parseSS(output string) []connection {
	var conns []connection
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "State") || strings.HasPrefix(line, "Netid") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 5 {
			c := connection{
				Proto: parts[0],
				State: parts[1],
				Local: parts[4],
				Peer:  parts[5],
			}
			conns = append(conns, c)
		}
	}
	return conns
}

func parseNetstat(output string) []connection {
	var conns []connection
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Proto") || strings.HasPrefix(line, "Active") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 4 {
			c := connection{
				Proto: parts[0],
				Local: parts[3],
				Peer:  parts[4],
			}
			if len(parts) >= 6 {
				c.State = parts[5]
			}
			conns = append(conns, c)
		}
	}
	return conns
}
