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

type StatusTool struct{}

func (StatusTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "wifi.status",
		Version:        "0.1.0",
		Category:       "wifi",
		Description:    "Show Wi-Fi interface status (iwinfo info or iw dev link).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        5 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (StatusTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"interface": {Type: types.FieldString, Required: false, Description: "Wireless interface name. Default: auto-detect."},
		},
	}
}

func (StatusTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"interfaces": {Type: types.FieldArray},
		},
	}
}

func (t StatusTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

type wifiIface struct {
	Name      string `json:"name"`
	SSID      string `json:"ssid,omitempty"`
	Frequency string `json:"frequency,omitempty"`
	Signal    int    `json:"signal,omitempty"`
	Noise     int    `json:"noise,omitempty"`
	Channel   int    `json:"channel,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Bitrate   string `json:"bitrate,omitempty"`
	State     string `json:"state,omitempty"`
	MAC       string `json:"mac,omitempty"`
}

func (t StatusTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	iface, _ := input["interface"].(string)

	var ifaces []wifiIface

	if iface != "" {
		if info, err := getIfaceInfo(ctx, iface); err == nil {
			ifaces = append(ifaces, info)
		}
	} else {
		detected, err := detectWirelessIface(ctx)
		if err == nil {
			if info, err := getIfaceInfo(ctx, detected); err == nil {
				ifaces = append(ifaces, info)
			}
		}
		ifaces = append(ifaces, getAllIfaces(ctx)...)
	}

	if len(ifaces) == 0 {
		return types.ToolResult{
			Success: false,
			Error:   "no wireless interfaces found",
		}, nil
	}

	return types.ToolResult{
		Success: true,
		Output:  types.ToolOutput{"interfaces": ifaces},
	}, nil
}

func getIfaceInfo(ctx context.Context, iface string) (wifiIface, error) {
	info := wifiIface{Name: iface}

	if out, err := exec.CommandContext(ctx, "iwinfo", iface, "info").Output(); err == nil {
		return parseIwinfoInfo(string(out), info), nil
	}

	if out, err := exec.CommandContext(ctx, "iw", "dev", iface, "link").Output(); err == nil {
		return parseIWLink(string(out), info), nil
	}

	return info, fmt.Errorf("no info source for %s", iface)
}

func parseIwinfoInfo(output string, info wifiIface) wifiIface {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if idx := strings.Index(line, "ESSID:"); idx >= 0 {
			info.SSID = strings.Trim(line[idx+6:], ` "'`)
		} else if idx := strings.Index(line, "Mode:"); idx >= 0 {
			info.Mode = strings.TrimSpace(line[idx+5:])
		} else if idx := strings.Index(line, "Frequency:"); idx >= 0 {
			info.Frequency = strings.TrimSpace(line[idx+10:])
		} else if idx := strings.Index(line, "Signal:"); idx >= 0 {
			parts := strings.Fields(line[idx+7:])
			if len(parts) > 0 {
				fmt.Sscanf(parts[0], "%d", &info.Signal)
			}
		} else if idx := strings.Index(line, "Noise:"); idx >= 0 {
			parts := strings.Fields(line[idx+6:])
			if len(parts) > 0 {
				fmt.Sscanf(parts[0], "%d", &info.Noise)
			}
		} else if idx := strings.Index(line, "Bit Rate:"); idx >= 0 {
			info.Bitrate = strings.TrimSpace(line[idx+9:])
		} else if idx := strings.Index(line, "Channel:"); idx >= 0 {
			fmt.Sscanf(line[idx+8:], "%d", &info.Channel)
		} else if idx := strings.Index(line, "MAC:"); idx >= 0 {
			info.MAC = strings.TrimSpace(line[idx+4:])
		}
	}
	info.State = "up"
	return info
}

func parseIWLink(output string, info wifiIface) wifiIface {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "SSID:") {
			info.SSID = strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		} else if strings.Contains(line, "freq:") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "freq:" && i+1 < len(parts) {
					info.Frequency = parts[i+1]
				}
			}
		} else if strings.Contains(line, "signal:") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "signal:" && i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d.00", &info.Signal)
				}
			}
		} else if strings.HasPrefix(line, "tx bitrate:") {
			info.Bitrate = strings.TrimSpace(strings.TrimPrefix(line, "tx bitrate:"))
		} else if strings.HasPrefix(line, "Not connected") {
			info.State = "down"
			info.SSID = ""
		}
	}
	if info.State == "" {
		info.State = "connected"
	}
	return info
}

func getAllIfaces(ctx context.Context) []wifiIface {
	var ifaces []wifiIface

	if out, err := exec.CommandContext(ctx, "iw", "dev").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		var current wifiIface
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "Interface") {
				if current.Name != "" {
					ifaces = append(ifaces, current)
				}
				current = wifiIface{}
				parts := strings.SplitN(line, " ", 2)
				if len(parts) == 2 {
					current.Name = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "addr") {
				parts := strings.SplitN(line, " ", 2)
				if len(parts) == 2 {
					current.MAC = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "type") {
				parts := strings.SplitN(line, " ", 2)
				if len(parts) == 2 {
					current.Mode = strings.TrimSpace(parts[1])
				}
			}
		}
		if current.Name != "" {
			ifaces = append(ifaces, current)
		}
	}

	return ifaces
}
