package network

import (
	"testing"
)

func TestParseIperf3JSON(t *testing.T) {
	output := `{
	"end": {
		"sum_sent": {
			"bits_per_second": 52428800
		},
		"sum_received": {
			"bits_per_second": 51200000,
			"jitter_ms": 2.34,
			"lost_packets": 5,
			"packets": 1000
		}
	}
}`

	result := parseIperf3JSON(output)
	tput, ok := result["throughput_mbps"]
	if !ok {
		t.Fatal("expected throughput_mbps")
	}
	mbps, ok := tput.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", tput)
	}
	if mbps < 50 || mbps > 53 {
		t.Fatalf("expected throughput ~52 Mbps, got %f", mbps)
	}

	jitter, ok := result["jitter_ms"]
	if !ok {
		t.Fatal("expected jitter_ms")
	}
	j, ok := jitter.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", jitter)
	}
	if j != 2.34 {
		t.Fatalf("expected jitter 2.34, got %f", j)
	}

	lost, ok := result["lost_packets"]
	if !ok {
		t.Fatal("expected lost_packets")
	}
	l, ok := lost.(int)
	if !ok {
		t.Fatalf("expected int, got %T", lost)
	}
	if l != 5 {
		t.Fatalf("expected 5 lost packets, got %d", l)
	}
}

func TestParseIperf3JSON_Empty(t *testing.T) {
	result := parseIperf3JSON(`{}`)
	if _, ok := result["throughput_mbps"]; ok {
		t.Fatal("expected no throughput for empty output")
	}
}

func TestParseCurlSpeed(t *testing.T) {
	speed := parseCurlSpeed("1048576")
	if speed < 8 {
		t.Fatalf("expected speed > 8 Mbps for 1048576 B/s, got %f", speed)
	}
}

func TestParseCurlSpeed_Zero(t *testing.T) {
	speed := parseCurlSpeed("0")
	if speed != 0 {
		t.Fatalf("expected 0, got %f", speed)
	}
}

func TestParseCurlSpeed_Empty(t *testing.T) {
	speed := parseCurlSpeed("")
	if speed != 0 {
		t.Fatalf("expected 0, got %f", speed)
	}
}
