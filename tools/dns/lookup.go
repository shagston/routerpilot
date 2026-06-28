package dns

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type LookupTool struct{}

func (LookupTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:             "dns.lookup",
		Version:        "0.1.0",
		Category:       "dns",
		Description:    "Resolve a hostname to IP addresses (A/AAAA records).",
		Permissions:    []types.Permission{types.PermissionRead},
		Timeout:        10 * time.Second,
		Risk:           types.RiskLow,
		SupportsDryRun: true,
	}
}

func (LookupTool) InputSchema() types.Schema {
	return types.Schema{
		RejectUnknownFields: true,
		Fields: map[string]types.FieldSchema{
			"host": {Type: types.FieldString, Required: true, Description: "Hostname to resolve."},
		},
	}
}

func (LookupTool) OutputSchema() types.Schema {
	return types.Schema{
		Fields: map[string]types.FieldSchema{
			"host":      {Type: types.FieldString},
			"addresses": {Type: types.FieldArray},
		},
	}
}

func (t LookupTool) Validate(_ context.Context, input types.ToolInput) error {
	host, ok := input["host"].(string)
	if !ok || strings.TrimSpace(host) == "" {
		return fmt.Errorf("%w: host is required", types.ErrInvalidInput)
	}
	return nil
}

func (t LookupTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	if err := t.Validate(ctx, input); err != nil {
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}

	host := strings.TrimSpace(input["host"].(string))
	var r net.Resolver
	ips, err := r.LookupHost(ctx, host)
	if err != nil {
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}

	return types.ToolResult{
		Success: true,
		Output: types.ToolOutput{
			"host":      host,
			"addresses": ips,
		},
	}, nil
}
