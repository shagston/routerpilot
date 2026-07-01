package capability

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/shagston/routerpilot/sdk/capability"
	"github.com/shagston/routerpilot/sdk/tool"
	"github.com/shagston/routerpilot/sdk/types"
)

var _ capability.Registry = (*Registry)(nil)

type Registry struct {
	mu       sync.RWMutex
	byID     map[capability.ID]capability.Provider
	byCap    map[types.Capability][]capability.ID
}

func NewRegistry() *Registry {
	return &Registry{
		byID:  make(map[capability.ID]capability.Provider),
		byCap: make(map[types.Capability][]capability.ID),
	}
}

func (r *Registry) Register(p capability.Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.ID()
	if _, exists := r.byID[id]; exists {
		return fmt.Errorf("capability provider %s already registered", id)
	}
	r.byID[id] = p
	return nil
}

func (r *Registry) Get(id capability.ID) (capability.Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("%w: capability provider %s", types.ErrNotFound, id)
	}
	return p, nil
}

func (r *Registry) List() []capability.Info {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]capability.Info, 0, len(r.byID))
	for _, p := range r.byID {
		result = append(result, p.Info())
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func (r *Registry) Resolve(cap types.Capability) ([]capability.Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, ok := r.byCap[cap]
	if !ok || len(ids) == 0 {
		return nil, fmt.Errorf("%w: no providers for capability %s", types.ErrCapabilityMissing, cap)
	}

	providers := make([]capability.Provider, 0, len(ids))
	for _, id := range ids {
		if p, exists := r.byID[id]; exists {
			providers = append(providers, p)
		}
	}
	return providers, nil
}

func (r *Registry) RegisterFromTool(t tool.Tool) error {
	meta := t.Metadata()

	caps := meta.Capabilities
	if len(caps) == 0 {
		caps = []types.Capability{types.Capability(meta.ID)}
	}

	r.mu.Lock()
	for _, cap := range caps {
		r.byCap[cap] = append(r.byCap[cap], capability.ID(meta.ID))
	}
	r.mu.Unlock()

	adapter := &toolAdapter{tool: t}
	return r.Register(adapter)
}

type toolAdapter struct {
	tool tool.Tool
}

func (a *toolAdapter) ID() capability.ID {
	return capability.ID(a.tool.Metadata().ID)
}

func (a *toolAdapter) Info() capability.Info {
	meta := a.tool.Metadata()
	return capability.Info{
		ID:          capability.ID(meta.ID),
		Name:        string(meta.ID),
		Description: meta.Description,
		Category:    meta.Category,
		Risk:        meta.Risk,
		Timeout:     meta.Timeout.Milliseconds(),
		Permissions: meta.Permissions,
		InputSchema: a.tool.InputSchema(),
	}
}

func (a *toolAdapter) Validate(ctx context.Context, input map[string]any) error {
	return a.tool.Validate(ctx, types.ToolInput(input))
}

func (a *toolAdapter) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	ctx2, cancel := context.WithTimeout(ctx, resolveTimeout(a.tool.Metadata().Timeout))
	defer cancel()

	result, err := a.tool.Execute(ctx2, types.ToolInput(input))
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}
	return map[string]any(result.Output), nil
}

func resolveTimeout(t time.Duration) time.Duration {
	if t > 0 {
		return t
	}
	return 30 * time.Second
}
