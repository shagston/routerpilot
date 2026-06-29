package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shagston/routerpilot/internal/kb"
	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/sdk/types"
)

type SuggestTool struct {
	Registry *registry.ToolRegistry
}

func (SuggestTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "suggest",
		Version:        "0.1.0",
		Category:       "diagnostics",
		Description:    "Analyze system state and suggest solutions for common network problems. No LLM required.",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        30 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (SuggestTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"problem": {Type: types.FieldString, Required: false, Description: "Problem description (e.g. 'no internet', 'dns not working')."},
		},
	}
}

func (SuggestTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"matches":     {Type: types.FieldArray},
			"problems_found": {Type: types.FieldInteger},
			"full_report": {Type: types.FieldString},
		},
	}
}

func (t SuggestTool) Validate(_ context.Context, _ types.ToolInput) error {
	return nil
}

func (t SuggestTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	problem, _ := input["problem"].(string)

	collector := kb.NewEvidenceCollector(t.Registry)
	diagnoser := kb.NewLocalDiagnoser()

	allChecks := []string{"interface_status", "ip_address", "default_route", "ping_external", "dns_resolve", "dhcp_leases", "connections", "wifi_status"}
	evidence, err := collector.Collect(ctx, allChecks)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("collect evidence: %w", err)
	}

	var matches []kb.DiagnosisResult

	if problem != "" && problem != "diagnose" {
		keywordMatchIDs := kb.MatchProblemKeywords(problem)
		for _, id := range keywordMatchIDs {
			pattern, ok := kb.PatternByID(id)
			if ok {
				match, desc := kb.EvaluatePatternSimple(id, evidence)
				matches = append(matches, kb.DiagnosisResult{
					Pattern:   pattern,
					Match:     match,
					MatchDesc: desc,
					Evidence:  evidence,
				})
			}
		}
	} else {
		matches = diagnoser.Suggest(ctx, problem, evidence)
	}

	if len(matches) == 0 {
		for _, id := range []string{"no-internet", "interface-down", "no-ip", "dns-failure", "no-default-route", "wifi-disconnected", "dhcp-exhausted", "high-latency"} {
			pattern, ok := kb.PatternByID(id)
			if !ok {
				continue
			}
			match, desc := kb.EvaluatePatternSimple(id, evidence)
			if match {
				matches = append(matches, kb.DiagnosisResult{
					Pattern:   pattern,
					Match:     true,
					MatchDesc: desc,
					Evidence:  evidence,
				})
			}
		}
	}

	var report strings.Builder
	report.WriteString("🔍 **RouterPilot Diagnosis Report**\n\n")
	report.WriteString("_Running without LLM — rule-based analysis_\n\n")

	if len(matches) == 0 {
		report.WriteString("✅ No problems detected.\n")
		report.WriteString("\nSystem appears to be healthy:\n")
		report.WriteString("- Interfaces are up\n")
		report.WriteString("- IP addresses assigned\n")
		report.WriteString("- Default route present\n")
		report.WriteString("- DNS resolving correctly\n")
		report.WriteString("- External connectivity confirmed\n")
	} else {
		report.WriteString(fmt.Sprintf("⚠️ Found **%d** potential issue(s):\n\n", len(matches)))
		for _, r := range matches {
			report.WriteString(r.FormatMarkdown())
			report.WriteString("---\n\n")
		}
	}

	if problem != "" && problem != "diagnose" {
		report.WriteString(fmt.Sprintf("\n_Problem described: \"%s\"_\n", problem))
		report.WriteString("_Matched by keyword: ")
		keywordMatches := kb.MatchProblemKeywords(problem)
		for i, km := range keywordMatches {
			if i > 0 {
				report.WriteString(", ")
			}
			report.WriteString(km)
		}
		report.WriteString("_\n")
	}

	return types.ToolResult{
		Success: true,
		Output: types.ToolOutput{
			"matches":        matches,
			"problems_found": len(matches),
			"full_report":    report.String(),
		},
	}, nil
}


