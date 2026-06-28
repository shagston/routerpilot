package dhcp

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type LeasesTool struct{}

func (LeasesTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "dhcp.leases",
		Version:        "0.1.0",
		Category:       "dhcp",
		Description:    "List active DHCP leases from /tmp/dhcp.leases or ubus.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (LeasesTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields:              map[string]types.FieldSchema{},
	}
}

func (LeasesTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"leases": {Type: types.FieldArray},
			"count":  {Type: types.FieldInteger},
			"source": {Type: types.FieldString},
		},
	}
}

func (t LeasesTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

type dhcpLease struct {
	Expiry    string `json:"expires"`
	MAC       string `json:"mac"`
	IP        string `json:"ip"`
	Hostname  string `json:"hostname"`
}

func (t LeasesTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	if leases, err := readUbusLeases(ctx); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"leases": leases,
				"count":  len(leases),
				"source": "ubus",
			},
		}, nil
	}

	if leases, err := readDHCPLeasesFile(); err == nil {
		return types.ToolResult{
			Success: true,
			Output: types.ToolOutput{
				"leases": leases,
				"count":  len(leases),
				"source": "dhcp.leases",
			},
		}, nil
	}

	return types.ToolResult{}, fmt.Errorf("no DHCP lease source available (tried ubus and /tmp/dhcp.leases)")
}

func readUbusLeases(ctx context.Context) ([]dhcpLease, error) {
	out, err := exec.CommandContext(ctx, "ubus", "call", "dhcp", "leases").Output()
	if err != nil {
		return nil, err
	}
	return parseUbusLeases(string(out))
}

func parseUbusLeases(output string) ([]dhcpLease, error) {
	var leases []dhcpLease
	scanner := bufio.NewScanner(strings.NewReader(output))
	var current dhcpLease
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, `"leases":`) {
			continue
		}
		if strings.Contains(line, `"device"`) {
			if current.MAC != "" {
				leases = append(leases, current)
			}
			current = dhcpLease{}
			continue
		}
		if strings.Contains(line, `"mac"`) {
			current.MAC = extractJSONString(line)
		} else if strings.Contains(line, `"ip"`) {
			current.IP = extractJSONString(line)
		} else if strings.Contains(line, `"hostname"`) {
			current.Hostname = extractJSONString(line)
		} else if strings.Contains(line, `"expires"`) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				current.Expiry = strings.TrimSpace(strings.TrimRight(parts[1], ", "))
			}
		}
	}
	if current.MAC != "" {
		leases = append(leases, current)
	}
	return leases, nil
}

func readDHCPLeasesFile() ([]dhcpLease, error) {
	if err := exec.Command("test", "-f", "/tmp/dhcp.leases").Run(); err != nil {
		return nil, err
	}
	out, err := exec.Command("cat", "/tmp/dhcp.leases").Output()
	if err != nil {
		return nil, err
	}
	return parseDHCPLeasesFile(string(out)), nil
}

func parseDHCPLeasesFile(output string) []dhcpLease {
	var leases []dhcpLease
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 4 {
			leases = append(leases, dhcpLease{
				Expiry:   parts[0],
				MAC:      parts[1],
				IP:       parts[2],
				Hostname: parts[3],
			})
		}
	}
	return leases
}

func extractJSONString(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	val := strings.TrimSpace(parts[1])
	val = strings.Trim(val, `", `)
	return val
}
