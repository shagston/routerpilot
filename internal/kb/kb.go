package kb

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/shagston/routerpilot/sdk/types"
)

//go:embed patterns.json
var patternsJSON []byte

type KnowledgeBase struct {
	Patterns []ProblemPattern `json:"patterns"`
}

type ProblemPattern struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Checks      []string   `json:"checks"`
	Solutions   []Solution `json:"solutions"`
}

type Solution struct {
	Priority int      `json:"priority"`
	Title    string   `json:"title"`
	Steps    []string `json:"steps"`
}

type DiagnosisResult struct {
	Pattern   ProblemPattern `json:"pattern"`
	Match     bool           `json:"match"`
	MatchDesc string         `json:"match_desc,omitempty"`
	Evidence  map[string]any `json:"evidence,omitempty"`
}

var defaultKB *KnowledgeBase

func init() {
	defaultKB = &KnowledgeBase{}
	if err := json.Unmarshal(patternsJSON, defaultKB); err != nil {
		panic(fmt.Sprintf("kb: failed to load patterns: %v", err))
	}
}

func Patterns() []ProblemPattern {
	return defaultKB.Patterns
}

func PatternByID(id string) (ProblemPattern, bool) {
	for _, p := range defaultKB.Patterns {
		if p.ID == id {
			return p, true
		}
	}
	return ProblemPattern{}, false
}

type Checker interface {
	Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error)
}

type ContextProvider interface {
	Collect(ctx context.Context, checks []string) (map[string]any, error)
}

func (kb *KnowledgeBase) Diagnose(ctx context.Context, provider ContextProvider) ([]DiagnosisResult, error) {
	var results []DiagnosisResult

	allChecks := uniqueChecks(kb.Patterns)
	evidence, err := provider.Collect(ctx, allChecks)
	if err != nil {
		return nil, fmt.Errorf("collect evidence: %w", err)
	}

	for _, pattern := range kb.Patterns {
		r := DiagnosisResult{
			Pattern:  pattern,
			Match:    false,
			Evidence: filterEvidence(evidence, pattern.Checks),
		}
		r.Match, r.MatchDesc = EvaluatePattern(pattern, evidence)
		results = append(results, r)
	}

	return results, nil
}

func uniqueChecks(patterns []ProblemPattern) []string {
	seen := map[string]bool{}
	var checks []string
	for _, p := range patterns {
		for _, c := range p.Checks {
			if !seen[c] {
				seen[c] = true
				checks = append(checks, c)
			}
		}
	}
	return checks
}

func filterEvidence(evidence map[string]any, checks []string) map[string]any {
	filtered := make(map[string]any)
	for _, c := range checks {
		if v, ok := evidence[c]; ok {
			filtered[c] = v
		}
	}
	return filtered
}

func EvaluatePattern(pattern ProblemPattern, evidence map[string]any) (bool, string) {
	switch pattern.ID {
	case "no-internet":
		return EvaluateNoInternet(evidence)
	case "interface-down":
		return EvaluateInterfaceDown(evidence)
	case "no-ip":
		return EvaluateNoIP(evidence)
	case "dns-failure":
		return EvaluateDNSFailure(evidence)
	case "no-default-route":
		return EvaluateNoDefaultRoute(evidence)
	case "wifi-disconnected":
		return EvaluateWiFiDisconnected(evidence)
	case "dhcp-exhausted":
		return EvaluateDHCPExhausted(evidence)
	case "high-latency":
		return EvaluateHighLatency(evidence)
	case "high-cpu":
		return EvaluateHighCPU(evidence)
	case "low-disk-space":
		return EvaluateLowDiskSpace(evidence)
	case "vpn-disconnected":
		return EvaluateVPNDisconnected(evidence)
	default:
		return false, ""
	}
}

func EvaluateNoInternet(evidence map[string]any) (bool, string) {
	if pingResult, ok := evidence["ping_external"]; ok {
		if m, ok := pingResult.(map[string]any); ok {
			if success, exists := m["success"]; exists {
				if !toBool(success) {
					if routeResult, ok := evidence["default_route"]; ok {
						if r, ok := routeResult.(map[string]any); ok {
							if hasRoute, exists := r["has_default"]; exists && !toBool(hasRoute) {
								return true, "No default route and ping to external host failed"
							}
						}
					}
					if ifResult, ok := evidence["interface_status"]; ok {
						if i, ok := ifResult.(map[string]any); ok {
							if state, exists := i["state"]; exists && fmt.Sprintf("%v", state) == "down" {
								return true, "WAN interface is down, no external connectivity"
							}
						}
					}
					return true, "Ping to external host failed"
				}
			}
		}
	}
	return false, ""
}

func EvaluateInterfaceDown(evidence map[string]any) (bool, string) {
	if ifResult, ok := evidence["interface_status"]; ok {
		if m, ok := ifResult.(map[string]any); ok {
			if state, exists := m["state"]; exists {
				if fmt.Sprintf("%v", state) == "down" {
					return true, "Interface state is 'down'"
				}
			}
		}
	}
	return false, ""
}

func EvaluateNoIP(evidence map[string]any) (bool, string) {
	if ipResult, ok := evidence["ip_address"]; ok {
		if m, ok := ipResult.(map[string]any); ok {
			if addresses, exists := m["addresses"]; exists {
				if arr, ok := addresses.([]any); ok && len(arr) == 0 {
					return true, "No IP addresses assigned"
				}
			}
		}
	}
	if ifResult, ok := evidence["interface_status"]; ok {
		if m, ok := ifResult.(map[string]any); ok {
			if ip, exists := m["ip"]; exists {
				if ipStr := fmt.Sprintf("%v", ip); ipStr == "" || ipStr == "<nil>" {
					return true, "Interface has no IP"
				}
			}
		}
	}
	return false, ""
}

func EvaluateDNSFailure(evidence map[string]any) (bool, string) {
	if dnsResult, ok := evidence["dns_resolve"]; ok {
		if m, ok := dnsResult.(map[string]any); ok {
			if success, exists := m["success"]; exists && !toBool(success) {
				if ifResult, ok := evidence["interface_status"]; ok {
					if i, ok := ifResult.(map[string]any); ok {
						if state, exists := i["state"]; exists && fmt.Sprintf("%v", state) == "up" {
							if ipResult, ok := evidence["ip_address"]; ok {
								if i2, ok := ipResult.(map[string]any); ok {
									if addrs, exists := i2["addresses"]; exists {
										if arr, ok := addrs.([]any); ok && len(arr) > 0 {
											return true, "DNS resolution failed despite interface being up with IP"
										}
									}
								}
							}
						}
					}
				}
				return true, "DNS resolution failed"
			}
		}
	}
	return false, ""
}

func EvaluateNoDefaultRoute(evidence map[string]any) (bool, string) {
	if routeResult, ok := evidence["default_route"]; ok {
		if m, ok := routeResult.(map[string]any); ok {
			if hasRoute, exists := m["has_default"]; exists && !toBool(hasRoute) {
				return true, "No default route in routing table"
			}
		}
	}
	return false, ""
}

func EvaluateWiFiDisconnected(evidence map[string]any) (bool, string) {
	if wifiResult, ok := evidence["wifi_status"]; ok {
		if m, ok := wifiResult.(map[string]any); ok {
			if interfaces, exists := m["interfaces"]; exists {
				if ifaces, ok := interfaces.([]any); ok {
					for _, iface := range ifaces {
						if i, ok := iface.(map[string]any); ok {
							if state, exists := i["state"]; exists {
								if fmt.Sprintf("%v", state) == "down" || fmt.Sprintf("%v", state) == "disconnected" {
									return true, fmt.Sprintf("Wi-Fi interface %v is %v", i["name"], state)
								}
							}
						}
					}
				}
			}
		}
	}
	return false, ""
}

func EvaluateDHCPExhausted(evidence map[string]any) (bool, string) {
	if dhcpResult, ok := evidence["dhcp_leases"]; ok {
		if m, ok := dhcpResult.(map[string]any); ok {
			if count, exists := m["count"]; exists {
				if total, exists := m["total"]; exists {
					used := toInt(count)
					max := toInt(total)
					if max > 0 && float64(used)/float64(max) >= 0.9 {
						return true, fmt.Sprintf("DHCP pool is %.0f%% full (%d/%d)", float64(used)/float64(max)*100, used, max)
					}
				}
			}
		}
	}
	return false, ""
}

func EvaluateHighCPU(evidence map[string]any) (bool, string) {
	if memResult, ok := evidence["system_memory"]; ok {
		if m, ok := memResult.(map[string]any); ok {
			if usage, exists := m["usage_percent"]; exists {
				usageF := toFloat(usage)
				if usageF > 90 {
					return true, fmt.Sprintf("Memory usage is %.0f%%", usageF)
				}
			}
			if load, exists := m["load_avg"]; exists {
				if loadStr, ok := load.(string); ok {
					parts := strings.Fields(loadStr)
					if len(parts) > 0 {
						oneMin := toFloat(parts[0])
						if oneMin > 4.0 {
							return true, fmt.Sprintf("CPU load average is %.2f", oneMin)
						}
					}
				}
			}
		}
	}

	if procResult, ok := evidence["system_processes"]; ok {
		if m, ok := procResult.(map[string]any); ok {
			if count, exists := m["count"]; exists {
				if toInt(count) > 100 {
					return true, fmt.Sprintf("High process count: %d", toInt(count))
				}
			}
		}
	}

	return false, ""
}

func EvaluateLowDiskSpace(evidence map[string]any) (bool, string) {
	if diskResult, ok := evidence["system_disk"]; ok {
		if m, ok := diskResult.(map[string]any); ok {
			if usage, exists := m["usage_percent"]; exists {
				usageF := toFloat(usage)
				if usageF > 90 {
					return true, fmt.Sprintf("Disk usage is %.0f%%", usageF)
				}
			}
		}
	}
	return false, ""
}

func EvaluateVPNDisconnected(evidence map[string]any) (bool, string) {
	if vpnResult, ok := evidence["vpn_status"]; ok {
		if m, ok := vpnResult.(map[string]any); ok {
			if interfaces, exists := m["interfaces"]; exists {
				if ifaces, ok := interfaces.([]any); ok {
					for _, iface := range ifaces {
						if i, ok := iface.(map[string]any); ok {
							if state, exists := i["state"]; exists {
								if state == "down" || state == "disconnected" || state == "error" {
									return true, fmt.Sprintf("VPN interface %v is %v", i["name"], state)
								}
							}
						}
					}
				}
			}
		}
	}
	return false, ""
}

func EvaluateHighLatency(evidence map[string]any) (bool, string) {
	if pingResult, ok := evidence["ping_external"]; ok {
		if m, ok := pingResult.(map[string]any); ok {
			if latency, exists := m["latency_avg_ms"]; exists {
				lat := toFloat(latency)
				if lat > 500 {
					return true, fmt.Sprintf("Average latency is %.0f ms (exceeds 500ms threshold)", lat)
				}
				if lat > 200 {
					// Check connections for possible cause
					return true, fmt.Sprintf("Elevated latency: %.0f ms", lat)
				}
			}
			if loss, exists := m["packet_loss"]; exists {
				if toFloat(loss) > 10 {
					return true, fmt.Sprintf("Packet loss is %.0f%%", toFloat(loss))
				}
			}
		}
	}
	return false, ""
}

func (r DiagnosisResult) FormatMarkdown() string {
	icon := "✅"
	if r.Match {
		icon = "⚠️"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s **%s**", icon, r.Pattern.Title))
	if r.MatchDesc != "" {
		b.WriteString(fmt.Sprintf(" — %s", r.MatchDesc))
	}
	b.WriteString("\n")

	if r.Match {
		b.WriteString(fmt.Sprintf("\n_%s_\n\n", r.Pattern.Description))
		b.WriteString("**Recommended solutions:**\n\n")
		for i, s := range r.Pattern.Solutions {
			b.WriteString(fmt.Sprintf("  **%d. %s**\n", i+1, s.Title))
			for _, step := range s.Steps {
				b.WriteString(fmt.Sprintf("     • %s\n", step))
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

func toBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true" || val == "yes" || val == "ok"
	case int:
		return val != 0
	case float64:
		return val != 0
	}
	return false
}

func toInt(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case int64:
		return int(val)
	case string:
		var i int
		fmt.Sscanf(val, "%d", &i)
		return i
	}
	return 0
}

func toFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	}
	return 0
}

type LocalDiagnoser struct {
	KB *KnowledgeBase
}

func NewLocalDiagnoser() *LocalDiagnoser {
	return &LocalDiagnoser{KB: defaultKB}
}

func (d *LocalDiagnoser) Suggest(ctx context.Context, problem string, evidence map[string]any) []DiagnosisResult {
	var matches []DiagnosisResult

	for _, pattern := range d.KB.Patterns {
		match, desc := EvaluatePattern(pattern, evidence)
		if match {
			matches = append(matches, DiagnosisResult{
				Pattern:   pattern,
				Match:     true,
				MatchDesc: desc,
				Evidence:  filterEvidence(evidence, pattern.Checks),
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Pattern.ID < matches[j].Pattern.ID
	})

	return matches
}

func MatchProblemKeywords(input string) []string {
	input = strings.ToLower(input)
	var matches []string
	for _, pattern := range defaultKB.Patterns {
		keywords := keywordsFor(pattern.ID)
		for _, kw := range keywords {
			if strings.Contains(input, kw) {
				matches = append(matches, pattern.ID)
				break
			}
		}
	}
	return matches
}

func EvaluatePatternSimple(id string, evidence map[string]any) (bool, string) {
	switch id {
	case "no-internet":
		return EvaluateNoInternet(evidence)
	case "interface-down":
		return EvaluateInterfaceDown(evidence)
	case "no-ip":
		return EvaluateNoIP(evidence)
	case "dns-failure":
		return EvaluateDNSFailure(evidence)
	case "no-default-route":
		return EvaluateNoDefaultRoute(evidence)
	case "wifi-disconnected":
		return EvaluateWiFiDisconnected(evidence)
	case "dhcp-exhausted":
		return EvaluateDHCPExhausted(evidence)
	case "high-latency":
		return EvaluateHighLatency(evidence)
	}
	return false, ""
}

func keywordsFor(patternID string) []string {
	switch patternID {
	case "no-internet":
		return []string{"internet", "no internet", "offline", "not working", "no connection", "can't reach"}
	case "interface-down":
		return []string{"interface", "down", "link", "port"}
	case "no-ip":
		return []string{"no ip", "ip address", "dhcp", "no address", "unassigned"}
	case "dns-failure":
		return []string{"dns", "resolve", "resolution", "not resolved", "unknown host"}
	case "no-default-route":
		return []string{"route", "gateway", "default route", "no route"}
	case "wifi-disconnected":
		return []string{"wifi", "wireless", "ssid", "wlan"}
	case "dhcp-exhausted":
		return []string{"dhcp pool", "no leases", "exhausted", "dhcp full"}
	case "high-latency":
		return []string{"latency", "slow", "lag", "packet loss", "timeout"}
	}
	return nil
}
