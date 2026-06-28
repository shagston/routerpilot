package registry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type fakeTool struct {
	id types.ToolID
}

func (t fakeTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:          t.id,
		Version:     "0.1.0",
		Category:    "diagnostics",
		Description: "fake test tool",
		Permissions: []types.Permission{types.PermissionRead},
		Timeout:     time.Second,
		Risk:        types.RiskLow,
	}
}

func (t fakeTool) InputSchema() types.Schema  { return types.Schema{} }
func (t fakeTool) OutputSchema() types.Schema { return types.Schema{} }
func (t fakeTool) Validate(context.Context, types.ToolInput) error {
	return nil
}
func (t fakeTool) Execute(context.Context, types.ToolInput) (types.ToolResult, error) {
	return types.ToolResult{Success: true}, nil
}

func TestToolRegistryRegistersAndListsTools(t *testing.T) {
	reg := NewToolRegistry()

	if err := reg.Register(fakeTool{id: "network.ping"}); err != nil {
		t.Fatalf("register tool: %v", err)
	}

	got, err := reg.Get("network.ping")
	if err != nil {
		t.Fatalf("get tool: %v", err)
	}
	if got.Metadata().ID != "network.ping" {
		t.Fatalf("unexpected tool id: %s", got.Metadata().ID)
	}

	list := reg.List()
	if len(list) != 1 {
		t.Fatalf("expected one tool, got %d", len(list))
	}
}

func TestToolRegistryRejectsDuplicates(t *testing.T) {
	reg := NewToolRegistry()

	if err := reg.Register(fakeTool{id: "network.ping"}); err != nil {
		t.Fatalf("register tool: %v", err)
	}
	if err := reg.Register(fakeTool{id: "network.ping"}); !errors.Is(err, types.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestToolRegistryRejectsEmptyID(t *testing.T) {
	reg := NewToolRegistry()

	if err := reg.Register(fakeTool{}); !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
