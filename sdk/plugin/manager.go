package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type LifecycleState int

const (
	StateDiscovered  LifecycleState = 0
	StateValidated   LifecycleState = 1
	StateLoaded      LifecycleState = 2
	StateInitialized LifecycleState = 3
	StateRunning     LifecycleState = 4
	StateStopping    LifecycleState = 5
	StateUnloaded    LifecycleState = 6
	StateFailed      LifecycleState = 7
)

func (s LifecycleState) String() string {
	switch s {
	case StateDiscovered:
		return "discovered"
	case StateValidated:
		return "validated"
	case StateLoaded:
		return "loaded"
	case StateInitialized:
		return "initialized"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateUnloaded:
		return "unloaded"
	case StateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

type managedPlugin struct {
	plugin Plugin
	state  LifecycleState
}

type Manager struct {
	mu      sync.RWMutex
	plugins map[string]*managedPlugin
	hosts   []Host
	rtCtx   RuntimeContext
	log     *slog.Logger
}

func NewManager(rtCtx RuntimeContext) *Manager {
	return &Manager{
		plugins: make(map[string]*managedPlugin),
		log:     slog.With("component", "plugin-manager"),
		rtCtx:   rtCtx,
	}
}

func (m *Manager) Register(ctx context.Context, p Plugin) error {
	manifest := p.Manifest()
	if manifest.ID == "" {
		return fmt.Errorf("plugin ID is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[manifest.ID]; exists {
		return fmt.Errorf("plugin %s already registered", manifest.ID)
	}

	mp := &managedPlugin{plugin: p, state: StateDiscovered}
	m.plugins[manifest.ID] = mp

	mp.state = StateValidated
	m.log.Info("plugin discovered", "id", manifest.ID, "version", manifest.Version)

	return nil
}

func (m *Manager) InitializeAll(ctx context.Context) error {
	m.mu.RLock()
	ids := make([]string, 0, len(m.plugins))
	for id := range m.plugins {
		ids = append(ids, id)
	}
	m.mu.RUnlock()

	for _, id := range ids {
		if err := m.initialize(ctx, id); err != nil {
			m.log.Error("plugin initialization failed", "id", id, "error", err)
		}
	}

	return nil
}

func (m *Manager) initialize(ctx context.Context, id string) error {
	m.mu.Lock()
	mp, ok := m.plugins[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("plugin %s not found", id)
	}
	mp.state = StateLoaded
	m.mu.Unlock()

	if err := mp.plugin.Init(ctx); err != nil {
		m.mu.Lock()
		mp.state = StateFailed
		m.mu.Unlock()
		return fmt.Errorf("init %s: %w", id, err)
	}

	if err := mp.plugin.Initialize(ctx, m.rtCtx); err != nil {
		m.mu.Lock()
		mp.state = StateFailed
		m.mu.Unlock()
		return fmt.Errorf("initialize %s: %w", id, err)
	}

	m.mu.Lock()
	mp.state = StateInitialized
	m.mu.Unlock()

	m.log.Info("plugin initialized", "id", id)
	return nil
}

func (m *Manager) ShutdownAll(ctx context.Context) error {
	m.mu.RLock()
	ids := make([]string, 0, len(m.plugins))
	for id := range m.plugins {
		ids = append(ids, id)
	}
	m.mu.RUnlock()

	for _, id := range ids {
		m.shutdown(ctx, id)
	}

	return nil
}

func (m *Manager) shutdown(ctx context.Context, id string) {
	m.mu.Lock()
	mp, ok := m.plugins[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	mp.state = StateStopping
	m.mu.Unlock()

	if err := mp.plugin.Shutdown(ctx); err != nil {
		m.log.Warn("plugin shutdown error", "id", id, "error", err)
	}
	if err := mp.plugin.Close(ctx); err != nil {
		m.log.Warn("plugin close error", "id", id, "error", err)
	}

	m.mu.Lock()
	mp.state = StateUnloaded
	m.mu.Unlock()

	m.log.Info("plugin unloaded", "id", id)
}

func (m *Manager) List() []Manifest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Manifest, 0, len(m.plugins))
	for _, mp := range m.plugins {
		result = append(result, mp.plugin.Manifest())
	}
	return result
}

func (m *Manager) Get(id string) (Plugin, LifecycleState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mp, ok := m.plugins[id]
	if !ok {
		return nil, StateUnloaded, false
	}
	return mp.plugin, mp.state, true
}

func (m *Manager) AddHost(host Host) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hosts = append(m.hosts, host)
}

func (m *Manager) LoadFromHosts(ctx context.Context) error {
	m.mu.RLock()
	hosts := append([]Host(nil), m.hosts...)
	m.mu.RUnlock()

	for _, host := range hosts {
		for _, p := range host.List() {
			if err := m.Register(ctx, p); err != nil {
				m.log.Warn("failed to register plugin from host", "error", err)
			}
		}
	}

	return nil
}
