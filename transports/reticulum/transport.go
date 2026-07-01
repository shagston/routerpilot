package reticulum

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	sdk "github.com/shagston/routerpilot/sdk/transport"
)

type ReticulumTransport struct {
	identity   Identity
	identityPath string
	discovery  *DiscoveryService
	hub        *hub

	mu       sync.RWMutex
	routes   map[string]string
	services map[string]sdk.Info
	peers    map[string]*Peer
	seq      uint64

	inbox    chan sdk.Envelope
	closed   bool
	stopCh   chan struct{}
	wg       sync.WaitGroup

	heartbeatInterval time.Duration
	peerTimeout       time.Duration

	log *slog.Logger
}

type Option func(*ReticulumTransport)

func WithIdentity(id Identity) Option {
	return func(t *ReticulumTransport) { t.identity = id }
}

func WithHeartbeatInterval(d time.Duration) Option {
	return func(t *ReticulumTransport) { t.heartbeatInterval = d }
}

func WithPeerTimeout(d time.Duration) Option {
	return func(t *ReticulumTransport) { t.peerTimeout = d }
}

func New(name, version string, opts ...Option) *ReticulumTransport {
	t := &ReticulumTransport{
		identity:          GenerateIdentity(name, version),
		discovery:         NewDiscoveryService(),
		hub:               globalHub,
		routes:            make(map[string]string),
		services:          make(map[string]sdk.Info),
		peers:             make(map[string]*Peer),
		inbox:             make(chan sdk.Envelope, 256),
		stopCh:            make(chan struct{}),
		heartbeatInterval: 15 * time.Second,
		peerTimeout:       60 * time.Second,
		log:               slog.With("transport", "reticulum"),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t *ReticulumTransport) LocalIdentity() Identity {
	return t.identity
}

func (t *ReticulumTransport) Peers() []Peer {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]Peer, 0, len(t.peers))
	for _, p := range t.peers {
		result = append(result, *p)
	}
	return result
}

func (t *ReticulumTransport) DiscoverPeers(ctx context.Context) error {
	all := t.hub.all()
	for _, tp := range all {
		if tp.identity.ID == t.identity.ID {
			continue
		}
		t.discovery.Discover(ctx, tp.identity)
		t.mu.Lock()
		t.peers[tp.identity.ID] = &Peer{
			Identity: tp.identity,
			Address:  tp.identity.ID,
			State:    PeerReachable,
			LastSeen: time.Now(),
		}
		t.mu.Unlock()
	}
	return nil
}

func (t *ReticulumTransport) Name() string {
	return fmt.Sprintf("reticulum:%s", t.identity.Name)
}

func (t *ReticulumTransport) Start(ctx context.Context) error {
	if err := t.hub.register(t); err != nil {
		return fmt.Errorf("register transport: %w", err)
	}

	t.discovery.Announce(ctx, t.identity)
	t.log.Info("reticulum transport started",
		"id", t.identity.ID,
		"name", t.identity.Name,
	)

	t.wg.Add(1)
	go t.heartbeatLoop(ctx)

	return nil
}

func (t *ReticulumTransport) Stop(ctx context.Context) error {
	t.mu.Lock()
	t.closed = true
	t.mu.Unlock()

	close(t.stopCh)
	t.wg.Wait()

	t.hub.unregister(t.identity.ID)
	t.log.Info("reticulum transport stopped")
	return nil
}

func (t *ReticulumTransport) Endpoint() sdk.Endpoint {
	return t
}

func (t *ReticulumTransport) Router() sdk.Router {
	return t
}

func (t *ReticulumTransport) Discovery() sdk.Discovery {
	return t
}

func (t *ReticulumTransport) ID() string {
	return t.identity.ID
}

func (t *ReticulumTransport) Send(ctx context.Context, env sdk.Envelope) error {
	if t.closed {
		return fmt.Errorf("transport closed")
	}

	if env.ID == "" {
		t.mu.Lock()
		t.seq++
		env.ID = fmt.Sprintf("%s-env-%d", t.identity.ID, t.seq)
		t.mu.Unlock()
	}
	if env.Source == "" {
		env.Source = t.identity.ID
	}

	targetID, err := t.Route(ctx, env)
	if err != nil {
		return fmt.Errorf("no route for %s: %w", env.Target, err)
	}

	if targetID == t.identity.ID {
		select {
		case t.inbox <- env:
			return nil
		default:
			return fmt.Errorf("inbox full")
		}
	}

	targetTp, ok := t.hub.resolve(targetID)
	if !ok {
		return fmt.Errorf("target transport %s not found", targetID)
	}

	select {
	case targetTp.inbox <- env:
		return nil
	default:
		return fmt.Errorf("target %s inbox full", targetID)
	}
}

func (t *ReticulumTransport) Receive(ctx context.Context) (<-chan sdk.Envelope, error) {
	return t.inbox, nil
}

func (t *ReticulumTransport) Close(ctx context.Context) error {
	return nil
}

func (t *ReticulumTransport) Route(ctx context.Context, env sdk.Envelope) (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if target, ok := t.routes[env.Target]; ok {
		return target, nil
	}

	t.mu.RUnlock()
	all := t.hub.all()
	t.mu.RLock()

	for _, tp := range all {
		if tp.identity.ID == env.Target || tp.identity.Name == env.Target {
			return tp.identity.ID, nil
		}
	}

	for pattern, target := range t.routes {
		prefix := pattern
		if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
			prefix = pattern[:len(pattern)-1]
		}
		if len(prefix) > 0 && len(env.Target) >= len(prefix) && env.Target[:len(prefix)] == prefix {
			return target, nil
		}
	}

	return "", fmt.Errorf("no route for target %s", env.Target)
}

func (t *ReticulumTransport) AddRoute(pattern string, endpointID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.routes[pattern] = endpointID
	return nil
}

func (t *ReticulumTransport) RemoveRoute(pattern string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.routes, pattern)
	return nil
}

func (t *ReticulumTransport) Register(ctx context.Context, info sdk.Info) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.services[info.ID] = info
	return nil
}

func (t *ReticulumTransport) Unregister(ctx context.Context, id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.services, id)
	return nil
}

func (t *ReticulumTransport) Lookup(ctx context.Context, svc string) ([]sdk.Info, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []sdk.Info
	for _, info := range t.services {
		if info.Name == svc || info.Type == svc || info.ID == svc {
			result = append(result, info)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("service %s not found", svc)
	}
	return result, nil
}

func (t *ReticulumTransport) heartbeatLoop(ctx context.Context) {
	defer t.wg.Done()

	ticker := time.NewTicker(t.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.broadcastHeartbeat(ctx)
			t.checkPeerTimeouts(ctx)
		}
	}
}

func (t *ReticulumTransport) broadcastHeartbeat(ctx context.Context) {
	all := t.hub.all()
	for _, tp := range all {
		if tp.identity.ID == t.identity.ID {
			continue
		}
		env := sdk.Envelope{
			Source: t.identity.ID,
			Target: tp.identity.ID,
			Type:   "heartbeat",
			Metadata: map[string]string{
				"uptime":  time.Now().String(),
				"health":  "ok",
				"version": t.identity.Version,
			},
		}
		select {
		case tp.inbox <- env:
		default:
		}
	}
}

func (t *ReticulumTransport) checkPeerTimeouts(ctx context.Context) {
	now := time.Now()
	t.mu.Lock()
	for id, p := range t.peers {
		if now.Sub(p.LastSeen) > t.peerTimeout && p.State == PeerReachable {
			p.State = PeerUnreachable
			t.discovery.Lost(ctx, id)
			t.log.Warn("peer unreachable", "peer_id", id)
		}
	}
	t.mu.Unlock()
}
