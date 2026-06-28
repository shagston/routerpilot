package safety

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/sdk/types"
)

type safeTool struct {
	metadata types.ToolMetadata
	schema   types.Schema
}

func (t safeTool) Metadata() types.ToolMetadata { return t.metadata }
func (t safeTool) InputSchema() types.Schema    { return t.schema }
func (t safeTool) OutputSchema() types.Schema   { return types.Schema{} }
func (t safeTool) Validate(context.Context, types.ToolInput) error {
	return nil
}
func (t safeTool) Execute(context.Context, types.ToolInput) (types.ToolResult, error) {
	return types.ToolResult{Success: true}, nil
}

func TestValidatorAcceptsAllowedPlan(t *testing.T) {
	reg := registry.NewToolRegistry()
	if err := reg.Register(safeTool{
		metadata: metadata("dns.lookup", []types.Permission{types.PermissionRead}, []types.Capability{"dns"}),
		schema: types.Schema{
			RejectUnknownFields: true,
			Fields: map[string]types.FieldSchema{
				"host": {Type: types.FieldString, Required: true},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	validator := NewValidator(reg, Config{
		Permissions:  []types.Permission{types.PermissionRead},
		Capabilities: []types.Capability{"dns"},
	})

	err := validator.Validate(context.Background(), types.Plan{
		ID:     "plan-1",
		Intent: "dns.lookup",
		Steps:  []types.Task{{ID: "lookup", Tool: "dns.lookup", Arguments: types.ToolInput{"host": "example.com"}}},
		Risk:   types.RiskLow,
	})
	if err != nil {
		t.Fatalf("validate plan: %v", err)
	}
}

func TestValidatorRejectsMissingPermission(t *testing.T) {
	reg := registry.NewToolRegistry()
	if err := reg.Register(safeTool{
		metadata: metadata("system.reboot", []types.Permission{types.PermissionAdmin}, nil),
	}); err != nil {
		t.Fatal(err)
	}
	validator := NewValidator(reg, Config{Permissions: []types.Permission{types.PermissionRead}})

	err := validator.Validate(context.Background(), types.Plan{
		ID:     "plan-1",
		Intent: "system.reboot",
		Steps:  []types.Task{{ID: "reboot", Tool: "system.reboot"}},
		Risk:   types.RiskHigh,
	})
	if !errors.Is(err, types.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestValidatorRejectsUnknownFields(t *testing.T) {
	err := ValidateInput(types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"host": {Type: types.FieldString, Required: true},
		},
	}, types.ToolInput{"host": "example.com", "extra": true})
	if !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestValidatorRejectsMissingCapabilities(t *testing.T) {
	reg := registry.NewToolRegistry()
	if err := reg.Register(safeTool{
		metadata: metadata("wifi.scan", []types.Permission{types.PermissionRead}, []types.Capability{"wifi"}),
	}); err != nil {
		t.Fatal(err)
	}
	validator := NewValidator(reg, Config{Permissions: []types.Permission{types.PermissionRead}})

	err := validator.Validate(context.Background(), types.Plan{
		ID:     "plan-1",
		Intent: "wifi.scan",
		Steps:  []types.Task{{ID: "scan", Tool: "wifi.scan"}},
		Risk:   types.RiskLow,
	})
	if !errors.Is(err, types.ErrCapabilityMissing) {
		t.Fatalf("expected ErrCapabilityMissing, got %v", err)
	}
}

func metadata(id types.ToolID, permissions []types.Permission, capabilities []types.Capability) types.ToolMetadata {
	return types.ToolMetadata{
		ID:           id,
		Version:      "0.1.0",
		Category:     "test",
		Description:  "test tool",
		Permissions:  permissions,
		Capabilities: capabilities,
		Timeout:      time.Second,
		Risk:         types.RiskLow,
	}
}
