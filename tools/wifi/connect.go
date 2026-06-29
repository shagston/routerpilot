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

type ConnectTool struct{}

func (ConnectTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "wifi.connect",
		Version:        "0.1.0",
		Category:       "wifi",
		Description:    "Connect to a Wi-Fi network. Supports open and WPA-PSK networks via iw, OpenWrt /etc/config/wireless, or wpa_supplicant.",
		Permissions:    []types.Permission{types.PermissionWrite},
		Timeout:        30 * time.Second,
		Risk:           types.RiskMedium,
		SupportsDryRun: false,
	}
}

func (ConnectTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"ssid":      {Type: types.FieldString, Required: true, Description: "Wi-Fi SSID to connect to."},
			"password":  {Type: types.FieldString, Required: false, Description: "WPA-PSK passphrase. Omit for open networks."},
			"interface": {Type: types.FieldString, Required: false, Description: "Wireless interface name. Default: auto-detect."},
		},
	}
}

func (ConnectTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"ssid":      {Type: types.FieldString},
			"interface": {Type: types.FieldString},
			"method":    {Type: types.FieldString},
			"bssid":     {Type: types.FieldString},
		},
	}
}

func (t ConnectTool) Validate(_ context.Context, input types.ToolInput) error {
	ssid, _ := input["ssid"].(string)
	if ssid == "" {
		return fmt.Errorf("'ssid' is required")
	}
	return nil
}

func (t ConnectTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	ssid, _ := input["ssid"].(string)
	password, _ := input["password"].(string)
	iface, _ := input["interface"].(string)

	if iface == "" {
		detected, err := detectWirelessIface(ctx)
		if err != nil {
			return types.ToolResult{}, fmt.Errorf("no wireless interface specified and auto-detect failed: %w", err)
		}
		iface = detected
	}

	method, err := tryConnect(ctx, iface, ssid, password)
	if err != nil {
		return types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to connect to %q on %s: %v", ssid, iface, err),
		}, nil
	}

	bssid := getConnectedBSSID(ctx, iface)

	return types.ToolResult{
		Success: true,
		Output: types.ToolOutput{
			"ssid":      ssid,
			"interface": iface,
			"method":    method,
			"bssid":     bssid,
		},
	}, nil
}

func tryConnect(ctx context.Context, iface, ssid, password string) (string, error) {
	if password == "" {
		if err := iwConnectOpen(ctx, iface, ssid); err == nil {
			return "iw", nil
		}
	} else {
		if err := iwConnectPSK(ctx, iface, ssid, password); err == nil {
			return "iw", nil
		}
	}

	if err := wpaSupplicantConnect(ctx, iface, ssid, password); err == nil {
		return "wpa_supplicant", nil
	}

	if err := openwrtConfigConnect(ctx, iface, ssid, password); err == nil {
		return "openwrt_uci", nil
	}

	return "", fmt.Errorf("all connection methods failed")
}

func iwConnectOpen(ctx context.Context, iface, ssid string) error {
	cmd := exec.CommandContext(ctx, "iw", "dev", iface, "connect", ssid)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("iw connect: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func iwConnectPSK(ctx context.Context, iface, ssid, password string) error {
	cmd := exec.CommandContext(ctx, "iw", "dev", iface, "connect", ssid, "key", "0:"+password)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("iw connect with key: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func wpaSupplicantConnect(ctx context.Context, iface, ssid, password string) error {
	if _, err := exec.LookPath("wpa_cli"); err != nil {
		return err
	}

	add := exec.CommandContext(ctx, "wpa_cli", "-i", iface, "add_network")
	addOut, err := add.Output()
	if err != nil {
		return fmt.Errorf("wpa_cli add_network: %w", err)
	}
	netID := strings.TrimSpace(string(addOut))

	exec.CommandContext(ctx, "wpa_cli", "-i", iface, "set_network", netID, "ssid", fmt.Sprintf(`"%s"`, ssid)).Run()
	if password == "" {
		exec.CommandContext(ctx, "wpa_cli", "-i", iface, "set_network", netID, "key_mgmt", "NONE").Run()
	} else {
		exec.CommandContext(ctx, "wpa_cli", "-i", iface, "set_network", netID, "psk", fmt.Sprintf(`"%s"`, password)).Run()
	}
	exec.CommandContext(ctx, "wpa_cli", "-i", iface, "select_network", netID).Run()
	exec.CommandContext(ctx, "wpa_cli", "-i", iface, "enable_network", netID).Run()

	return nil
}

func openwrtConfigConnect(ctx context.Context, iface, ssid, password string) error {
	if _, err := exec.LookPath("uci"); err != nil {
		return err
	}
	if _, err := exec.LookPath("wifi"); err != nil {
		return err
	}

	if err := exec.CommandContext(ctx, "uci", "set", "wireless.@wifi-iface[0].ssid="+ssid).Run(); err != nil {
		return fmt.Errorf("uci set ssid: %w", err)
	}

	if password != "" {
		exec.CommandContext(ctx, "uci", "set", "wireless.@wifi-iface[0].encryption=psk2").Run()
		exec.CommandContext(ctx, "uci", "set", "wireless.@wifi-iface[0].key="+password).Run()
	} else {
		exec.CommandContext(ctx, "uci", "set", "wireless.@wifi-iface[0].encryption=none").Run()
	}

	exec.CommandContext(ctx, "uci", "commit", "wireless").Run()

	if err := exec.CommandContext(ctx, "wifi").Run(); err != nil {
		return fmt.Errorf("wifi reload: %w", err)
	}

	return nil
}

func getConnectedBSSID(ctx context.Context, iface string) string {
	out, err := exec.CommandContext(ctx, "iw", "dev", iface, "link").Output()
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Connected to") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2]
			}
		}
	}
	return ""
}
