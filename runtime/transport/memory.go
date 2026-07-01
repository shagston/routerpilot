package transport

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/shagston/routerpilot/sdk/transport"
)

type MemoryTransport struct {
	id        string
	mu        sync.RWMutex
	routes    map[string]string
	services  map[string]transport.Info
	inbox     chan transport.Envelope
	discovery map[string]transport.Info
	seq       uint64
}

func NewMemory(id string) *MemoryTransport {
	return &MemoryTransport{
		id:        id,
		routes:    make(map[string]string),
		services:  make(map[string]transport.Info),
		inbox:     make(chan transport.Envelope, 256),
		discovery: make(map[string]transport.Info),
	}
}

func (m *MemoryTransport) nextID() string {
	m.seq++
	return fmt.Sprintf("%s-env-%d", m.id, m.seq)
}

func (m *MemoryTransport) Endpoint() transport.Endpoint {
	return m
}

func (m *MemoryTransport) Router() transport.Router {
	return m
}

func (m *MemoryTransport) Discovery() transport.Discovery {
	return m
}

func (m *MemoryTransport) Start(ctx context.Context) error {
	return nil
}

func (m *MemoryTransport) Stop(ctx context.Context) error {
	close(m.inbox)
	return nil
}

func (m *MemoryTransport) ID() string {
	return m.id
}

func (m *MemoryTransport) Send(ctx context.Context, env transport.Envelope) error {
	if env.ID == "" {
		env.ID = m.nextID()
	}

	targetID, err := m.Route(ctx, env)
	if err != nil {
		return fmt.Errorf("no route for %s: %w", env.Target, err)
	}
	if targetID == m.id {
		select {
		case m.inbox <- env:
		default:
			return fmt.Errorf("inbox full")
		}
		return nil
	}

	return fmt.Errorf("no local endpoint for target %s (routed to %s)", env.Target, targetID)
}

func (m *MemoryTransport) Receive(ctx context.Context) (<-chan transport.Envelope, error) {
	return m.inbox, nil
}

func (m *MemoryTransport) Close(ctx context.Context) error {
	return nil
}

func (m *MemoryTransport) Route(ctx context.Context, env transport.Envelope) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if target, ok := m.routes[env.Target]; ok {
		return target, nil
	}

	for pattern, target := range m.routes {
		if strings.HasSuffix(pattern, "*") && strings.HasPrefix(env.Target, strings.TrimSuffix(pattern, "*")) {
			return target, nil
		}
	}

	return "", fmt.Errorf("no route for target %s", env.Target)
}

func (m *MemoryTransport) AddRoute(pattern string, endpointID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes[pattern] = endpointID
	return nil
}

func (m *MemoryTransport) RemoveRoute(pattern string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.routes, pattern)
	return nil
}

func (m *MemoryTransport) Register(ctx context.Context, info transport.Info) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.discovery[info.ID] = info
	return nil
}

func (m *MemoryTransport) Unregister(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.discovery, id)
	return nil
}

func (m *MemoryTransport) Lookup(ctx context.Context, svc string) ([]transport.Info, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []transport.Info
	for _, info := range m.discovery {
		if info.Name == svc || info.Type == svc {
			result = append(result, info)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("service %s not found", svc)
	}
	return result, nil
}

func (m *MemoryTransport) SendTo(ctx context.Context, env transport.Envelope) error {
	if env.ID == "" {
		env.ID = m.nextID()
	}

	select {
	case m.inbox <- env:
	default:
		return fmt.Errorf("inbox full")
	}
	return nil
}
