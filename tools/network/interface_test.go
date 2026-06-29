package network

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/shagston/routerpilot/internal/network"
	"github.com/shagston/routerpilot/sdk/types"
)

func TestInterfaceStatusTool_All(t *testing.T) {
	mock := network.NewMockProvider()
	tool := InterfaceStatusTool{Provider: mock}

	result, err := tool.Execute(context.Background(), types.ToolInput{"interface": "all"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	output := result.Output
	ifaces, ok := output["interfaces"].([]types.ToolOutput)
	if !ok {
		t.Fatalf("expected interfaces array, got %T", output["interfaces"])
	}
	if len(ifaces) != 3 {
		t.Fatalf("expected 3 interfaces, got %d", len(ifaces))
	}
}

func TestInterfaceStatusTool_Specific(t *testing.T) {
	mock := network.NewMockProvider()
	tool := InterfaceStatusTool{Provider: mock}

	result, err := tool.Execute(context.Background(), types.ToolInput{"interface": "eth0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Output["interface"] != "eth0" {
		t.Fatalf("expected eth0, got %v", result.Output["interface"])
	}
}

func TestInterfaceStatusTool_NotFound(t *testing.T) {
	mock := network.NewMockProvider()
	tool := InterfaceStatusTool{Provider: mock}

	result, err := tool.Execute(context.Background(), types.ToolInput{"interface": "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent interface")
	}
	if result.Success {
		t.Fatal("expected failure")
	}
}

func TestInterfaceStatusTool_ValidateRejectsEmpty(t *testing.T) {
	mock := network.NewMockProvider()
	tool := InterfaceStatusTool{Provider: mock}

	err := tool.Validate(context.Background(), types.ToolInput{})
	if !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestInterfaceSetStateTool_Up(t *testing.T) {
	mock := network.NewMockProvider()
	tool := InterfaceSetStateTool{Provider: mock}

	result, err := tool.Execute(context.Background(), types.ToolInput{
		"interface": "eth0",
		"state":     "up",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Output["new_state"] != "up" {
		t.Fatalf("expected new_state=up, got %v", result.Output["new_state"])
	}
}

func TestInterfaceSetStateTool_Down(t *testing.T) {
	mock := network.NewMockProvider()
	tool := InterfaceSetStateTool{Provider: mock}

	result, err := tool.Execute(context.Background(), types.ToolInput{
		"interface": "eth0",
		"state":     "down",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Output["new_state"] != "down" {
		t.Fatalf("expected new_state=down, got %v", result.Output["new_state"])
	}
}

func TestInterfaceSetStateTool_ValidateRejectsInvalidState(t *testing.T) {
	mock := network.NewMockProvider()
	tool := InterfaceSetStateTool{Provider: mock}

	err := tool.Validate(context.Background(), types.ToolInput{
		"interface": "eth0",
		"state":     "invalid",
	})
	if !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestRouteGetTool(t *testing.T) {
	mock := network.NewMockProvider()
	tool := RouteGetTool{Provider: mock}

	result, err := tool.Execute(context.Background(), types.ToolInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	routes, ok := result.Output["routes"].(string)
	if !ok {
		t.Fatalf("expected string output, got %T", result.Output["routes"])
	}
	if !strings.Contains(routes, "default") {
		t.Fatalf("expected default route in output, got: %s", routes)
	}
}

func TestRouteAddTool(t *testing.T) {
	mock := network.NewMockProvider()
	tool := RouteAddTool{Provider: mock}

	result, err := tool.Execute(context.Background(), types.ToolInput{
		"destination": "10.0.0.0/24",
		"gateway":     "192.168.1.1",
		"interface":   "eth1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestRouteAddTool_ValidateRejectsMissing(t *testing.T) {
	mock := network.NewMockProvider()
	tool := RouteAddTool{Provider: mock}

	err := tool.Validate(context.Background(), types.ToolInput{
		"destination": "10.0.0.0/24",
	})
	if !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestIPAddressGetTool(t *testing.T) {
	mock := network.NewMockProvider()
	tool := IPAddressGetTool{Provider: mock}

	result, err := tool.Execute(context.Background(), types.ToolInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	addrs, ok := result.Output["addresses"].(string)
	if !ok {
		t.Fatalf("expected string output, got %T", result.Output["addresses"])
	}
	if !strings.Contains(addrs, "192.168.1.100") {
		t.Fatalf("expected 192.168.1.100 in output, got: %s", addrs)
	}
}

func TestIPAddressSetTool(t *testing.T) {
	mock := network.NewMockProvider()
	tool := IPAddressSetTool{Provider: mock}

	result, err := tool.Execute(context.Background(), types.ToolInput{
		"interface": "eth0",
		"address":   "10.0.0.1/24",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestIPAddressSetTool_ValidateRejectsMissing(t *testing.T) {
	mock := network.NewMockProvider()
	tool := IPAddressSetTool{Provider: mock}

	err := tool.Validate(context.Background(), types.ToolInput{
		"interface": "eth0",
	})
	if !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
