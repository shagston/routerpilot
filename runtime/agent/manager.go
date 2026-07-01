package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type Manager struct {
	mu     sync.RWMutex
	agents map[types.AgentID]*Agent
	seq    uint64
}

func NewManager() *Manager {
	return &Manager{
		agents: make(map[types.AgentID]*Agent),
	}
}

func (m *Manager) nextID() types.AgentID {
	m.seq++
	return types.AgentID(fmt.Sprintf("agent-%d", m.seq))
}

func (m *Manager) Create(ctx context.Context, spec Spec) (*Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.nextID()
	a := New(id, spec)
	m.agents[id] = a
	return a, nil
}

func (m *Manager) Get(id types.AgentID) (*Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	a, ok := m.agents[id]
	if !ok {
		return nil, fmt.Errorf("%w: agent %s", types.ErrNotFound, id)
	}
	return a, nil
}

func (m *Manager) List() []Info {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Info, 0, len(m.agents))
	for _, a := range m.agents {
		result = append(result, a.Info())
	}
	return result
}

func (m *Manager) Update(ctx context.Context, id types.AgentID, update Update) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	a, ok := m.agents[id]
	if !ok {
		return fmt.Errorf("%w: agent %s", types.ErrNotFound, id)
	}

	if update.Name != nil {
		a.info.Name = *update.Name
	}
	if update.Permissions != nil {
		a.info.Permissions = update.Permissions
	}
	if update.Capabilities != nil {
		a.info.Capabilities = update.Capabilities
	}
	if update.Metadata != nil {
		for k, v := range update.Metadata {
			a.info.Metadata[k] = v
		}
	}
	a.info.UpdatedAt = time.Now()
	return nil
}

func (m *Manager) Delete(ctx context.Context, id types.AgentID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.agents[id]; !ok {
		return fmt.Errorf("%w: agent %s", types.ErrNotFound, id)
	}
	delete(m.agents, id)
	return nil
}

func (m *Manager) Default(ctx context.Context) (*Agent, error) {
	m.mu.RLock()
	for _, a := range m.agents {
		if a.info.Name == "default" {
			m.mu.RUnlock()
			return a, nil
		}
	}
	m.mu.RUnlock()

	return m.Create(ctx, DefaultSpec())
}
