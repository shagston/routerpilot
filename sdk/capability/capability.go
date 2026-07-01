package capability

import (
	"context"

	"github.com/shagston/routerpilot/sdk/types"
)

type ID string

type Info struct {
	ID          ID               `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    string           `json:"category"`
	Risk        types.RiskLevel  `json:"risk"`
	Timeout     int64            `json:"timeout_ms"`
	Permissions []types.Permission `json:"permissions"`
	InputSchema types.Schema     `json:"input_schema"`
}

type Provider interface {
	ID() ID
	Info() Info
	Validate(ctx context.Context, input map[string]any) error
	Execute(ctx context.Context, input map[string]any) (map[string]any, error)
}

type Registry interface {
	Register(Provider) error
	Get(id ID) (Provider, error)
	List() []Info
	Resolve(capability types.Capability) ([]Provider, error)
}
