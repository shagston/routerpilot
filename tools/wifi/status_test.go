package wifi

import (
	"testing"
)

func TestParseIwinfoInfo(t *testing.T) {
	output := `wlan0     ESSID: "MyNetwork"
          Mode: Master
          Frequency: 2.437 GHz
          Signal: -65 dBm
          Noise: -92 dBm
          Bit Rate: 54 MBit/s
          Channel: 6
          MAC: 00:11:22:33:44:55`

	info := parseIwinfoInfo(output, wifiIface{Name: "wlan0"})
	if info.SSID != "MyNetwork" {
		t.Fatalf("expected MyNetwork, got %s", info.SSID)
	}
	if info.State != "up" {
		t.Fatalf("expected state up, got %s", info.State)
	}
	if info.Signal != -65 {
		t.Fatalf("expected signal -65, got %d", info.Signal)
	}
	if info.Noise != -92 {
		t.Fatalf("expected noise -92, got %d", info.Noise)
	}
	if info.Channel != 6 {
		t.Fatalf("expected channel 6, got %d", info.Channel)
	}
	if info.Bitrate != "54 MBit/s" {
		t.Fatalf("expected '54 MBit/s', got %s", info.Bitrate)
	}
	if info.MAC != "00:11:22:33:44:55" {
		t.Fatalf("expected 00:11:22:33:44:55, got %s", info.MAC)
	}
}

func TestParseIWLink(t *testing.T) {
	output := `Connected to 00:11:22:33:44:55 (on wlan0)
	SSID: MyNetwork
	freq: 2437
	signal: -65.00 dBm
	tx bitrate: 65.0 MBit/s`

	info := parseIWLink(output, wifiIface{Name: "wlan0"})
	if info.SSID != "MyNetwork" {
		t.Fatalf("expected MyNetwork, got %s", info.SSID)
	}
	if info.State != "connected" {
		t.Fatalf("expected state connected, got %s", info.State)
	}
	if info.Signal != -65 {
		t.Fatalf("expected signal -65, got %d", info.Signal)
	}
	if info.Bitrate != "65.0 MBit/s" {
		t.Fatalf("expected '65.0 MBit/s', got %s", info.Bitrate)
	}
}

func TestParseIWLink_NotConnected(t *testing.T) {
	output := `Not connected.`

	info := parseIWLink(output, wifiIface{Name: "wlan0"})
	if info.State != "down" {
		t.Fatalf("expected state down, got %s", info.State)
	}
}
