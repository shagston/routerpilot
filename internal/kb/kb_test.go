package kb

import (
	"testing"
)

func TestEvaluateNoInternet_PingFailed(t *testing.T) {
	evidence := map[string]any{
		"ping_external": map[string]any{
			"success": false,
		},
	}
	match, desc := EvaluateNoInternet(evidence)
	if !match {
		t.Fatal("expected match for failed ping")
	}
	if desc == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestEvaluateNoInternet_PingSuccess(t *testing.T) {
	evidence := map[string]any{
		"ping_external": map[string]any{
			"success": true,
		},
	}
	match, _ := EvaluateNoInternet(evidence)
	if match {
		t.Fatal("expected no match for successful ping")
	}
}

func TestEvaluateNoInternet_NoEvidence(t *testing.T) {
	match, _ := EvaluateNoInternet(map[string]any{})
	if match {
		t.Fatal("expected no match with no evidence")
	}
}

func TestEvaluateInterfaceDown_StateDown(t *testing.T) {
	evidence := map[string]any{
		"interface_status": map[string]any{
			"state": "down",
			"name":  "eth0",
		},
	}
	match, desc := EvaluateInterfaceDown(evidence)
	if !match {
		t.Fatal("expected match for interface down")
	}
	if desc == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestEvaluateInterfaceDown_StateUp(t *testing.T) {
	evidence := map[string]any{
		"interface_status": map[string]any{
			"state": "up",
		},
	}
	match, _ := EvaluateInterfaceDown(evidence)
	if match {
		t.Fatal("expected no match for interface up")
	}
}

func TestEvaluateNoIP_EmptyAddresses(t *testing.T) {
	evidence := map[string]any{
		"ip_address": map[string]any{
			"addresses": []any{},
		},
	}
	match, desc := EvaluateNoIP(evidence)
	if !match {
		t.Fatal("expected match for no IP addresses")
	}
	if desc == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestEvaluateNoIP_HasAddresses(t *testing.T) {
	evidence := map[string]any{
		"ip_address": map[string]any{
			"addresses": []any{"192.168.1.1/24"},
		},
	}
	match, _ := EvaluateNoIP(evidence)
	if match {
		t.Fatal("expected no match when IP assigned")
	}
}

func TestEvaluateDNSFailure_Failed(t *testing.T) {
	evidence := map[string]any{
		"dns_resolve": map[string]any{
			"success": false,
		},
		"interface_status": map[string]any{
			"state": "up",
		},
		"ip_address": map[string]any{
			"addresses": []any{"192.168.1.1/24"},
		},
	}
	match, desc := EvaluateDNSFailure(evidence)
	if !match {
		t.Fatal("expected match for DNS failure")
	}
	if desc == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestEvaluateDNSFailure_Success(t *testing.T) {
	evidence := map[string]any{
		"dns_resolve": map[string]any{
			"success": true,
		},
	}
	match, _ := EvaluateDNSFailure(evidence)
	if match {
		t.Fatal("expected no match for successful DNS")
	}
}

func TestEvaluateNoDefaultRoute_Missing(t *testing.T) {
	evidence := map[string]any{
		"default_route": map[string]any{
			"has_default": false,
		},
	}
	match, desc := EvaluateNoDefaultRoute(evidence)
	if !match {
		t.Fatal("expected match for missing default route")
	}
	if desc == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestEvaluateNoDefaultRoute_Present(t *testing.T) {
	evidence := map[string]any{
		"default_route": map[string]any{
			"has_default": true,
		},
	}
	match, _ := EvaluateNoDefaultRoute(evidence)
	if match {
		t.Fatal("expected no match when default route present")
	}
}

func TestEvaluateWiFiDisconnected_StateDown(t *testing.T) {
	evidence := map[string]any{
		"wifi_status": map[string]any{
			"interfaces": []any{
				map[string]any{"name": "wlan0", "state": "down"},
			},
		},
	}
	match, desc := EvaluateWiFiDisconnected(evidence)
	if !match {
		t.Fatal("expected match for wifi down")
	}
	if desc == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestEvaluateWiFiDisconnected_StateConnected(t *testing.T) {
	evidence := map[string]any{
		"wifi_status": map[string]any{
			"interfaces": []any{
				map[string]any{"name": "wlan0", "state": "connected"},
			},
		},
	}
	match, _ := EvaluateWiFiDisconnected(evidence)
	if match {
		t.Fatal("expected no match for connected wifi")
	}
}

func TestEvaluateHighLatency_HighLatency(t *testing.T) {
	evidence := map[string]any{
		"ping_external": map[string]any{
			"latency_avg_ms": float64(600),
		},
	}
	match, desc := EvaluateHighLatency(evidence)
	if !match {
		t.Fatal("expected match for high latency")
	}
	if desc == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestEvaluateHighLatency_PacketLoss(t *testing.T) {
	evidence := map[string]any{
		"ping_external": map[string]any{
			"latency_avg_ms": float64(30),
			"packet_loss":    float64(20),
		},
	}
	match, desc := EvaluateHighLatency(evidence)
	if !match {
		t.Fatal("expected match for packet loss")
	}
	if desc == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestEvaluateHighLatency_Normal(t *testing.T) {
	evidence := map[string]any{
		"ping_external": map[string]any{
			"latency_avg_ms": float64(20),
			"packet_loss":    float64(0),
		},
	}
	match, _ := EvaluateHighLatency(evidence)
	if match {
		t.Fatal("expected no match for normal latency")
	}
}

func TestEvaluateDHCPExhausted_NearFull(t *testing.T) {
	evidence := map[string]any{
		"dhcp_leases": map[string]any{
			"count": 90,
			"total": 100,
		},
	}
	match, desc := EvaluateDHCPExhausted(evidence)
	if !match {
		t.Fatal("expected match for near-full DHCP")
	}
	if desc == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestEvaluateDHCPExhausted_Normal(t *testing.T) {
	evidence := map[string]any{
		"dhcp_leases": map[string]any{
			"count": 30,
			"total": 100,
		},
	}
	match, _ := EvaluateDHCPExhausted(evidence)
	if match {
		t.Fatal("expected no match for normal DHCP usage")
	}
}

func TestEvaluatePatternSimple(t *testing.T) {
	evidence := map[string]any{
		"ping_external": map[string]any{
			"success": true,
		},
	}
	match, desc := EvaluatePatternSimple("no-internet", evidence)
	if match {
		t.Fatal("expected no match when ping succeeds")
	}
	if desc != "" {
		t.Fatalf("expected empty desc, got %s", desc)
	}
}

func TestPatternByID_Found(t *testing.T) {
	p, ok := PatternByID("no-internet")
	if !ok {
		t.Fatal("expected to find no-internet pattern")
	}
	if p.Title == "" {
		t.Fatal("expected non-empty title")
	}
}

func TestPatternByID_NotFound(t *testing.T) {
	_, ok := PatternByID("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestPatterns_NotEmpty(t *testing.T) {
	patterns := Patterns()
	if len(patterns) == 0 {
		t.Fatal("expected non-empty patterns")
	}
}

func TestMatchProblemKeywords_Internet(t *testing.T) {
	matches := MatchProblemKeywords("no internet")
	if len(matches) == 0 {
		t.Fatal("expected at least one match for 'no internet'")
	}
	found := false
	for _, m := range matches {
		if m == "no-internet" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'no-internet' in matches, got %v", matches)
	}
}

func TestMatchProblemKeywords_DNS(t *testing.T) {
	matches := MatchProblemKeywords("dns not working")
	if len(matches) == 0 {
		t.Fatal("expected at least one match for 'dns not working'")
	}
	found := false
	for _, m := range matches {
		if m == "dns-failure" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'dns-failure' in matches, got %v", matches)
	}
}

func TestMatchProblemKeywords_Unknown(t *testing.T) {
	matches := MatchProblemKeywords("something completely different")
	if len(matches) != 0 {
		t.Fatalf("expected no matches for unknown problem, got %v", matches)
	}
}

func TestFormatMarkdown_Matched(t *testing.T) {
	r := DiagnosisResult{
		Pattern: ProblemPattern{
			ID:          "test",
			Title:       "Test Problem",
			Description: "A test problem",
			Solutions: []Solution{
				{Priority: 1, Title: "Do X", Steps: []string{"step 1", "step 2"}},
			},
		},
		Match:     true,
		MatchDesc: "something is wrong",
	}
	md := r.FormatMarkdown()
	if len(md) == 0 {
		t.Fatal("expected non-empty markdown")
	}
}

func TestFormatMarkdown_NotMatched(t *testing.T) {
	r := DiagnosisResult{
		Pattern: ProblemPattern{
			ID:    "test",
			Title: "Test Problem",
		},
		Match: false,
	}
	md := r.FormatMarkdown()
	if len(md) == 0 {
		t.Fatal("expected non-empty markdown")
	}
}
