package memory

import "context"

type Record struct {
	Key      string         `json:"key"`
	Value    map[string]any `json:"value"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type Provider interface {
	Read(context.Context, string) (Record, error)
	Write(context.Context, Record) error
	Search(context.Context, string, int) ([]Record, error)
}
