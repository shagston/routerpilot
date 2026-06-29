package wifi

import (
	"testing"
)

func TestParseIwinfoScan(t *testing.T) {
	output := `Cell 01 - Address: 00:11:22:33:44:55
ESSID: "MyNetwork"
Mode: Master
Channel: 6
Signal: -65 dBm
Encryption: WPA2-PSK
Bit Rate: 54 MBit/s

Cell 02 - Address: 66:77:88:99:AA:BB
ESSID: "GuestNet"
Mode: Master
Channel: 1
Signal: -78 dBm
Encryption: WPA3-SAE
Bit Rate: 130 MBit/s`

	aps := parseIwinfoScan(output)
	if len(aps) != 2 {
		t.Fatalf("expected 2 access points, got %d", len(aps))
	}
	if aps[0].SSID != "MyNetwork" {
		t.Fatalf("expected MyNetwork, got %s", aps[0].SSID)
	}
	if aps[0].BSSID != "00:11:22:33:44:55" {
		t.Fatalf("expected 00:11:22:33:44:55, got %s", aps[0].BSSID)
	}
	if aps[0].Channel != 6 {
		t.Fatalf("expected channel 6, got %d", aps[0].Channel)
	}
	if aps[0].Signal != -65 {
		t.Fatalf("expected signal -65, got %d", aps[0].Signal)
	}
	if aps[0].Encryption != "WPA2-PSK" {
		t.Fatalf("expected WPA2-PSK, got %s", aps[0].Encryption)
	}
	if aps[1].SSID != "GuestNet" {
		t.Fatalf("expected GuestNet, got %s", aps[1].SSID)
	}
}

func TestParseIwinfoScan_Empty(t *testing.T) {
	aps := parseIwinfoScan("")
	if len(aps) != 0 {
		t.Fatalf("expected 0 access points, got %d", len(aps))
	}
}

func TestParseIWScan(t *testing.T) {
	output := `BSS 00:11:22:33:44:55(on wlan0)
SSID: MyNetwork
freq: 2437
signal: -65.00 dBm

BSS 66:77:88:99:AA:BB(on wlan0)
SSID: GuestNet
freq: 2412
signal: -78.00 dBm`

	aps := parseIWScan(output)
	if len(aps) != 2 {
		t.Fatalf("expected 2 access points, got %d", len(aps))
	}
	if aps[0].SSID != "MyNetwork" {
		t.Fatalf("expected MyNetwork, got %s", aps[0].SSID)
	}
	if aps[0].BSSID != "00:11:22:33:44:55" {
		t.Fatalf("expected 00:11:22:33:44:55, got %s", aps[0].BSSID)
	}
	if aps[0].Signal != -65 {
		t.Fatalf("expected signal -65, got %d", aps[0].Signal)
	}
}

func TestFreqToChannel(t *testing.T) {
	tests := []struct {
		freq    int
		channel int
	}{
		{2412, 1},
		{2437, 6},
		{2462, 11},
		{5180, 36},
		{5200, 40},
		{5745, 149},
		{1000, 0},
	}
	for _, tt := range tests {
		ch := freqToChannel(tt.freq)
		if ch != tt.channel {
			t.Errorf("freqToChannel(%d) = %d, want %d", tt.freq, ch, tt.channel)
		}
	}
}
