package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/shagston/routerpilot/sdk/types"
)

type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

type Policy struct {
	ID          string     `json:"id"`
	Description string     `json:"description,omitempty"`
	Effect      Effect     `json:"effect"`
	Match       Match      `json:"match"`
	Condition   *Condition `json:"condition,omitempty"`
}

type Match struct {
	Permissions  []string `json:"permissions,omitempty"`
	Risk         []string `json:"risk,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Agents       []string `json:"agents,omitempty"`
}

type Condition struct {
	AgentPermissions []string `json:"agent_permissions,omitempty"`
}

type PolicySet struct {
	Policies []Policy `json:"policies"`
}

type Request struct {
	Capability  string
	Risk        types.RiskLevel
	Permissions []types.Permission
	AgentID     types.AgentID
	AgentPerms  []types.Permission
}

type Result struct {
	Allowed  bool   `json:"allowed"`
	PolicyID string `json:"policy_id,omitempty"`
	Effect   Effect `json:"effect"`
	Reason   string `json:"reason,omitempty"`
}

type Engine struct {
	policies []Policy
}

func NewEngine() *Engine {
	return &Engine{}
}

func NewEngineWithPolicies(policies []Policy) *Engine {
	return &Engine{policies: policies}
}

func (e *Engine) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read policy file: %w", err)
	}

	var set PolicySet
	if err := json.Unmarshal(data, &set); err != nil {
		return fmt.Errorf("parse policy file: %w", err)
	}

	e.policies = append(e.policies, set.Policies...)
	return nil
}

func (e *Engine) AddPolicy(p Policy) {
	e.policies = append(e.policies, p)
}

func (e *Engine) Policies() []Policy {
	return append([]Policy(nil), e.policies...)
}

func (e *Engine) Evaluate(ctx context.Context, req Request) Result {
	reqCap := strings.ToLower(req.Capability)
	reqRisk := strings.ToLower(string(req.Risk))

	for _, p := range e.policies {
		if !e.matchPolicy(p, req, reqCap, reqRisk) {
			continue
		}
		if p.Condition != nil && !e.evaluateCondition(*p.Condition, req) {
			continue
		}

		switch p.Effect {
		case EffectAllow:
			return Result{Allowed: true, PolicyID: p.ID, Effect: EffectAllow}
		case EffectDeny:
			return Result{
				Allowed:  false,
				PolicyID: p.ID,
				Effect:   EffectDeny,
				Reason:   fmt.Sprintf("denied by policy %s: %s", p.ID, p.Description),
			}
		}
	}

	return Result{
		Allowed:  false,
		Effect:   EffectDeny,
		Reason:   "no matching allow policy",
	}
}

func (e *Engine) matchPolicy(p Policy, req Request, reqCap, reqRisk string) bool {
	if len(p.Match.Capabilities) > 0 {
		if !matchGlobList(reqCap, p.Match.Capabilities) {
			return false
		}
	}

	if len(p.Match.Risk) > 0 {
		if !stringInList(reqRisk, p.Match.Risk) {
			return false
		}
	}

	if len(p.Match.Permissions) > 0 {
		matched := false
		for _, rp := range req.Permissions {
			if stringInList(strings.ToLower(string(rp)), p.Match.Permissions) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(p.Match.Agents) > 0 {
		if !stringInList(string(req.AgentID), p.Match.Agents) {
			return false
		}
	}

	return true
}

func (e *Engine) evaluateCondition(c Condition, req Request) bool {
	if len(c.AgentPermissions) > 0 {
		for _, ap := range req.AgentPerms {
			if stringInList(strings.ToLower(string(ap)), c.AgentPermissions) {
				return true
			}
		}
		return false
	}
	return true
}

func stringInList(s string, list []string) bool {
	for _, item := range list {
		if s == item {
			return true
		}
	}
	return false
}

func matchGlobList(s string, patterns []string) bool {
	for _, p := range patterns {
		if p == "*" || p == s {
			return true
		}
		if strings.HasSuffix(p, "*") && strings.HasPrefix(s, strings.TrimSuffix(p, "*")) {
			return true
		}
	}
	return false
}

func DefaultPolicies() []Policy {
	return []Policy{
		{
			ID:          "allow-read-low",
			Description: "Allow read capabilities with low risk",
			Effect:      EffectAllow,
			Match: Match{
				Permissions: []string{"read"},
				Risk:        []string{"low"},
				Capabilities: []string{"*"},
			},
		},
		{
			ID:          "deny-critical-default",
			Description: "Deny critical risk capabilities by default",
			Effect:      EffectDeny,
			Match: Match{
				Risk: []string{"critical"},
			},
		},
	}
}
