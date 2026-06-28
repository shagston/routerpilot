package network

import (
	"context"
	"errors"
	"testing"

	"github.com/shagston/routerpilot/sdk/types"
)

func TestPingToolValidateRejectsUnsafeHost(t *testing.T) {
	err := PingTool{}.Validate(context.Background(), types.ToolInput{"host": "example.com && reboot"})
	if !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestParseWindowsPingOutput(t *testing.T) {
	output := `Pinging 127.0.0.1 with 32 bytes of data:
Reply from 127.0.0.1: bytes=32 time<1ms TTL=128

Ping statistics for 127.0.0.1:
    Packets: Sent = 1, Received = 1, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 0ms, Maximum = 0ms, Average = 0ms`

	result := parsePingOutput("127.0.0.1", output)
	if result["packets_sent"] != 1 || result["packets_received"] != 1 {
		t.Fatalf("unexpected packet stats: %#v", result)
	}
	if result["packet_loss"] != float64(0) {
		t.Fatalf("unexpected packet loss: %#v", result)
	}
}

func TestParseUnixPingOutput(t *testing.T) {
	output := `1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 0.031/0.031/0.031/0.000 ms`

	result := parsePingOutput("127.0.0.1", output)
	if result["packets_sent"] != 1 || result["packets_received"] != 1 {
		t.Fatalf("unexpected packet stats: %#v", result)
	}
	if result["latency_avg_ms"] != 0.031 {
		t.Fatalf("unexpected latency: %#v", result)
	}
}
