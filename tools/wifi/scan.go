package wifi

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type ScanTool struct{}

func (ScanTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "wifi.scan",
		Version:        "0.1.0",
		Category:       "wifi",
		Description:    "Scan for nearby Wi-Fi access points using iwinfo (OpenWrt) or iw.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        30 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: false,
	}
}

func (ScanTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"interface": {Type: types.FieldString, Required: false, Description: "Wireless interface name (e.g. wlan0). Default: auto-detect."},
		},
	}
}

func (ScanTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"access_points": {Type: types.FieldArray},
			"interface":     {Type: types.FieldString},
			"source":        {Type: types.FieldString},
		},
	}
}

func (t ScanTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t ScanTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	iface, _ := input["interface"].(string)

	if iface == "" {
		if detected, err := detectWirelessIface(ctx); err == nil {
			iface = detected
		}
	}

	result := types.ToolOutput{
		"interface": iface,
	}

	if out, err := exec.CommandContext(ctx, "iwinfo", iface, "scan").Output(); err == nil {
		result["source"] = "iwinfo"
		result["access_points"] = parseIwinfoScan(string(out))
		return types.ToolResult{Success: true, Output: result}, nil
	}

	if out, err := exec.CommandContext(ctx, "iw", "dev", iface, "scan").Output(); err == nil {
		result["source"] = "iw"
		result["access_points"] = parseIWScan(string(out))
		return types.ToolResult{Success: true, Output: result}, nil
	}

	return types.ToolResult{
		Success: false,
		Error:   "no wifi scanning tool found (tried iwinfo and iw)",
	}, nil
}

func detectWirelessIface(ctx context.Context) (string, error) {
	if out, err := exec.CommandContext(ctx, "iw", "dev").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "Interface") {
				parts := strings.SplitN(line, " ", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1]), nil
				}
			}
		}
	}
	if out, err := exec.CommandContext(ctx, "iwinfo").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.Contains(line, "ESSID") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					return parts[0], nil
				}
			}
		}
	}
	return "", fmt.Errorf("no wireless interface detected")
}

type accessPoint struct {
	SSID     string  `json:"ssid"`
	BSSID    string  `json:"bssid"`
	Channel  int     `json:"channel"`
	Signal   int     `json:"signal"`
	Encryption string `json:"encryption,omitempty"`
}

func parseIwinfoScan(output string) []accessPoint {
	var aps []accessPoint
	var current accessPoint
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Cell") || strings.HasPrefix(line, "Scan") {
			if current.SSID != "" {
				aps = append(aps, current)
			}
			current = accessPoint{}
			continue
		}
		if strings.HasPrefix(line, "ESSID:") {
			current.SSID = strings.TrimSpace(strings.TrimPrefix(line, "ESSID:"))
			current.SSID = strings.Trim(current.SSID, `"'`)
		} else if strings.HasPrefix(line, "Address:") {
			current.BSSID = strings.TrimSpace(strings.TrimPrefix(line, "Address:"))
		} else if strings.HasPrefix(line, "Channel:") {
			fmt.Sscanf(line, "Channel:%d", &current.Channel)
		} else if strings.HasPrefix(line, "Signal:") {
			sigStr := strings.TrimPrefix(line, "Signal:")
			sigStr = strings.TrimSpace(sigStr)
			parts := strings.Fields(sigStr)
			if len(parts) > 0 {
				fmt.Sscanf(parts[0], "%d", &current.Signal)
			}
		} else if strings.HasPrefix(line, "Encryption:") {
			current.Encryption = strings.TrimSpace(strings.TrimPrefix(line, "Encryption:"))
		}
	}
	if current.SSID != "" {
		aps = append(aps, current)
	}
	return aps
}

func parseIWScan(output string) []accessPoint {
	var aps []accessPoint
	var current accessPoint
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "BSS ") {
			if current.BSSID != "" {
				aps = append(aps, current)
			}
			current = accessPoint{}
			bssid := strings.TrimPrefix(line, "BSS ")
			bssid = strings.TrimSpace(bssid)
			if idx := strings.Index(bssid, "("); idx > 0 {
				bssid = strings.TrimSpace(bssid[:idx])
			}
			current.BSSID = bssid
			continue
		}
		if strings.HasPrefix(line, "SSID:") {
			current.SSID = strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		} else if strings.Contains(line, "signal:") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "signal:" && i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d.00", &current.Signal)
				}
			}
		} else if strings.HasPrefix(line, "freq:") {
			var freq int
			fmt.Sscanf(line, "freq:%d", &freq)
			current.Channel = freqToChannel(freq)
		} else if strings.HasPrefix(line, "Channel:") {
			fmt.Sscanf(line, "Channel:%d", &current.Channel)
		}
	}
	if current.BSSID != "" {
		aps = append(aps, current)
	}
	return aps
}

func freqToChannel(freq int) int {
	if freq >= 2412 && freq <= 2484 {
		return (freq - 2412)/5 + 1
	}
	if freq >= 5180 && freq <= 5825 {
		return (freq - 5180)/5 + 36
	}
	return 0
}
