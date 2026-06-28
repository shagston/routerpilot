package planner

import (
	"encoding/json"

	"github.com/shagston/routerpilot/internal/registry"
)

type toolDescription struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Risk        string `json:"risk"`
	InputSchema any    `json:"input_schema"`
}

func FormatToolSchemas(reg *registry.ToolRegistry) string {
	descriptions := make([]toolDescription, 0, len(reg.List()))
	for _, meta := range reg.List() {
		t, err := reg.Get(meta.ID)
		if err != nil {
			continue
		}
		descriptions = append(descriptions, toolDescription{
			ID:          string(meta.ID),
			Description: meta.Description,
			Category:    meta.Category,
			Risk:        string(meta.Risk),
			InputSchema: t.InputSchema(),
		})
	}

	encoded, err := json.MarshalIndent(descriptions, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(encoded)
}
