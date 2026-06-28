package registry

import (
	"sort"
	"sync"

	"github.com/shagston/routerpilot/sdk/tool"
	"github.com/shagston/routerpilot/sdk/types"
)

type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[types.ToolID]tool.Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[types.ToolID]tool.Tool)}
}

func (r *ToolRegistry) Register(t tool.Tool) error {
	metadata := t.Metadata()
	if metadata.ID == "" {
		return types.ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[metadata.ID]; exists {
		return types.ErrAlreadyExists
	}
	r.tools[metadata.ID] = t
	return nil
}

func (r *ToolRegistry) Get(id types.ToolID) (tool.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.tools[id]
	if !exists {
		return nil, types.ErrNotFound
	}
	return t, nil
}

func (r *ToolRegistry) List() []types.ToolMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata := make([]types.ToolMetadata, 0, len(r.tools))
	for _, t := range r.tools {
		metadata = append(metadata, t.Metadata())
	}
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].ID < metadata[j].ID
	})
	return metadata
}
