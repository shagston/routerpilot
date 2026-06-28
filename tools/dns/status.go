package dns

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type StatusTool struct{}

func (StatusTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "dns.status",
		Version:        "0.1.0",
		Category:       "dns",
		Description:    "Show current DNS resolver configuration (from /etc/resolv.conf or systemd-resolve).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
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
			"servers":  {Type: types.FieldArray},
			"domain":   {Type: types.FieldString},
			"source":   {Type: types.FieldString},
		},
	}
}

func (t StatusTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t StatusTool) Execute(ctx context.Context, _ types.ToolInput) (types.ToolResult, error) {
	output := types.ToolOutput{}

	if out, err := exec.CommandContext(ctx, "systemd-resolve", "--status").Output(); err == nil {
		output["source"] = "systemd-resolve"
		output["servers"] = parseSystemdResolve(string(out))
		return types.ToolResult{Success: true, Output: output}, nil
	}

	output["source"] = "resolv.conf"
	servers, domain := parseResolvConf()
	output["servers"] = servers
	if domain != "" {
		output["domain"] = domain
	}

	return types.ToolResult{Success: true, Output: output}, nil
}

func parseSystemdResolve(output string) []string {
	var servers []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	inDNSServers := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "DNS Servers:") {
			inDNSServers = true
			continue
		}
		if inDNSServers {
			if line == "" || strings.Contains(line, ":") {
				inDNSServers = false
				continue
			}
			servers = append(servers, strings.TrimSpace(line))
		}
	}
	return servers
}

func parseResolvConf() (servers []string, domain string) {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver ") {
			server := strings.TrimPrefix(line, "nameserver ")
			servers = append(servers, strings.TrimSpace(server))
		} else if strings.HasPrefix(line, "domain ") {
			domain = strings.TrimSpace(strings.TrimPrefix(line, "domain "))
		} else if strings.HasPrefix(line, "search ") {
			if domain == "" {
				parts := strings.Fields(strings.TrimPrefix(line, "search "))
				if len(parts) > 0 {
					domain = parts[0]
				}
			}
		}
	}
	return
}
