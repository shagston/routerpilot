package distributed

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/shagston/routerpilot/sdk/capability"
	"github.com/shagston/routerpilot/sdk/events"
	sdkTp "github.com/shagston/routerpilot/sdk/transport"
	"github.com/shagston/routerpilot/sdk/types"
)

type CapabilityHandler func(ctx context.Context, source string, providerID string, input map[string]any) (map[string]any, error)

type Mesh struct {
	mu       sync.RWMutex
	peers    map[string]*Peer
	localID  string
	trans    sdkTp.Transport
	pub      events.Publisher
	executor *RemoteExecutor

	capHandlers map[string]CapabilityHandler

	advertisedCapabilities []CapabilityDescriptor

	heartbeatInterval time.Duration
	peerTimeout       time.Duration

	stopCh chan struct{}
	wg     sync.WaitGroup
	log    *slog.Logger
}

type MeshOption func(*Mesh)

func WithHeartbeatInterval(d time.Duration) MeshOption {
	return func(m *Mesh) { m.heartbeatInterval = d }
}

func WithPeerTimeout(d time.Duration) MeshOption {
	return func(m *Mesh) { m.peerTimeout = d }
}

func WithAdvertisedCapabilities(caps []CapabilityDescriptor) MeshOption {
	return func(m *Mesh) { m.advertisedCapabilities = caps }
}

func NewMesh(localID string, trans sdkTp.Transport, pub events.Publisher, opts ...MeshOption) *Mesh {
	m := &Mesh{
		peers:             make(map[string]*Peer),
		localID:           localID,
		trans:             trans,
		pub:               pub,
		capHandlers:       make(map[string]CapabilityHandler),
		heartbeatInterval: 30 * time.Second,
		peerTimeout:       90 * time.Second,
		stopCh:            make(chan struct{}),
		log:               slog.With("component", "mesh"),
	}
	for _, opt := range opts {
		opt(m)
	}
	m.executor = NewRemoteExecutor(m)
	return m
}

func (m *Mesh) Start(ctx context.Context) error {
	m.wg.Add(1)
	go m.healthLoop(ctx)

	m.wg.Add(1)
	go m.discoveryListener(ctx)

	m.log.Info("mesh started", "local_id", m.localID)
	return nil
}

func (m *Mesh) Stop(ctx context.Context) error {
	close(m.stopCh)
	m.wg.Wait()
	m.log.Info("mesh stopped")
	return nil
}

func (m *Mesh) Peers() []Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Peer, 0, len(m.peers))
	for _, p := range m.peers {
		result = append(result, *p)
	}
	return result
}

func (m *Mesh) Lookup(runtimeID string) (Peer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.peers[runtimeID]
	if !ok {
		return Peer{}, false
	}
	return *p, true
}

func (m *Mesh) Refresh(ctx context.Context) error {
	if disc := m.trans.Discovery(); disc != nil {
		allInfo, err := disc.Lookup(ctx, "")
		if err == nil {
			for _, info := range allInfo {
				m.upsertPeer(ctx, info)
			}
		}
	}
	return nil
}

func (m *Mesh) DiscoverPeer(ctx context.Context, info sdkTp.Info) {
	m.upsertPeer(ctx, info)
}

func (m *Mesh) RemovePeer(ctx context.Context, runtimeID string) {
	m.mu.Lock()
	p, ok := m.peers[runtimeID]
	if ok {
		delete(m.peers, runtimeID)
	}
	m.mu.Unlock()

	if ok {
		m.emitPeerEvent(ctx, EventPeerExpired, *p)
		m.log.Info("peer expired", "peer_id", runtimeID, "name", p.Name)
	}
}

func (m *Mesh) RegisterCapabilityHandler(capID string, handler CapabilityHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.capHandlers[capID] = handler
}

func (m *Mesh) ExecuteRemotely(ctx context.Context, targetID string, provider capability.Provider, input map[string]any) (map[string]any, error) {
	return m.executor.Execute(ctx, targetID, provider, input)
}

func (m *Mesh) upsertPeer(ctx context.Context, info sdkTp.Info) {
	m.mu.Lock()

	existing, exists := m.peers[info.ID]
	now := time.Now()

	peer := &Peer{
		RuntimeID:  info.ID,
		Name:       info.Name,
		Version:    info.Metadata["version"],
		State:      PeerKnown,
		Labels:     info.Metadata,
		Health:     "unknown",
		LastSeen:   now,
	}
	if info.Type != "" {
		peer.Transports = []string{info.Type}
	}

	if !exists {
		peer.FirstSeen = now
		peer.State = PeerDiscovered
		m.peers[info.ID] = peer
		m.mu.Unlock()
		m.emitPeerEvent(ctx, EventPeerDiscovered, *peer)
		m.log.Info("peer discovered", "peer_id", info.ID, "name", info.Name)
		return
	}

	peer.FirstSeen = existing.FirstSeen
	peer.Transports = existing.Transports
	if len(info.Type) > 0 {
		peer.Transports = append(peer.Transports, info.Type)
	}

	if existing.State == PeerUnavailable {
		peer.State = PeerHealthy
		m.peers[info.ID] = peer
		m.mu.Unlock()
		m.emitPeerEvent(ctx, EventPeerHealthy, *peer)
		m.log.Info("peer became healthy", "peer_id", info.ID)
		return
	}

	peer.State = PeerKnown
	m.peers[info.ID] = peer
	m.mu.Unlock()
}

func (m *Mesh) emitPeerEvent(ctx context.Context, event PeerEvent, peer Peer) {
	if m.pub == nil {
		return
	}

	m.pub.Publish(types.Event{
		Timestamp: time.Now(),
		Type:      types.EventType(event.String()),
		Source:    m.localID,
		Severity:  types.SeverityInfo,
		Priority:  types.PriorityNormal,
		Payload: map[string]any{
			"peer_id":   peer.RuntimeID,
			"peer_name": peer.Name,
			"state":     peer.State.String(),
		},
	})
}

func (m *Mesh) healthLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkPeerHealth(ctx)
		}
	}
}

func (m *Mesh) checkPeerHealth(ctx context.Context) {
	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, p := range m.peers {
		switch {
		case now.Sub(p.LastSeen) > m.peerTimeout && (p.State == PeerHealthy || p.State == PeerKnown):
			p.State = PeerUnavailable
			m.emitPeerEvent(ctx, EventPeerUnavailable, *p)
			m.log.Warn("peer unavailable", "peer_id", id, "name", p.Name)

		case now.Sub(p.LastSeen) > m.peerTimeout && (p.State == PeerDiscovered || p.State == PeerValidated):
			p.State = PeerUnavailable
			m.emitPeerEvent(ctx, EventPeerUnavailable, *p)
			m.log.Debug("peer never healthy, marking unavailable", "peer_id", id, "name", p.Name)

		case now.Sub(p.LastSeen) > m.peerTimeout*2 && p.State == PeerUnavailable:
			delete(m.peers, id)
			m.emitPeerEvent(ctx, EventPeerExpired, *p)
			m.log.Warn("peer expired", "peer_id", id, "name", p.Name)
		}
	}
}

func (m *Mesh) discoveryListener(ctx context.Context) {
	defer m.wg.Done()

	inbox, err := m.trans.Endpoint().Receive(ctx)
	if err != nil {
		m.log.Error("discovery listener: cannot receive", "error", err)
		return
	}

	for {
		select {
		case <-m.stopCh:
			return
		case env, ok := <-inbox:
			if !ok {
				return
			}
			m.handleIncoming(ctx, env)
		}
	}
}

func (m *Mesh) handleIncoming(ctx context.Context, env sdkTp.Envelope) {
	switch env.Type {
	case "capability.request":
		m.handleRemoteCapabilityRequest(ctx, env)
	case "capability.response":
		m.executor.handleResponse(ctx, env)
	default:
		if env.Source != "" && env.Source != m.localID {
			m.mu.Lock()
			_, known := m.peers[env.Source]
			if !known {
				m.peers[env.Source] = &Peer{
					RuntimeID: env.Source,
					Name:      env.Source,
					State:     PeerDiscovered,
					LastSeen:  time.Now(),
					FirstSeen: time.Now(),
				}
				peer := *m.peers[env.Source]
				m.mu.Unlock()
				m.emitPeerEvent(ctx, EventPeerDiscovered, peer)
			} else {
				m.mu.Unlock()
			}
		}
	}
}

func (m *Mesh) handleRemoteCapabilityRequest(ctx context.Context, env sdkTp.Envelope) {
	m.log.Info("remote capability request", "source", env.Source)

	req, err := parseCapabilityRequest(env.Payload)
	if err != nil {
		m.log.Error("failed to parse capability request", "error", err)
		m.sendErrorResponse(ctx, env.Source, env.Metadata["request_id"], "malformed request")
		return
	}

	m.mu.RLock()
	handler, ok := m.capHandlers[req.Capability]
	m.mu.RUnlock()

	if !ok {
		m.log.Warn("no handler for capability", "capability", req.Capability)
		m.sendErrorResponse(ctx, env.Source, req.RequestID, fmt.Sprintf("capability %s not available", req.Capability))
		return
	}

	result, err := handler(ctx, env.Source, req.Capability, req.Input)
	if err != nil {
		m.sendErrorResponse(ctx, env.Source, req.RequestID, err.Error())
		return
	}

	m.sendCapabilityResponse(ctx, env.Source, req.RequestID, result)
}

func (m *Mesh) sendErrorResponse(ctx context.Context, target, requestID, errMsg string) {
	respPayload := map[string]any{
		"request_id": requestID,
		"status":     "error",
		"error":      errMsg,
	}
	respBytes, _ := serializePayload(respPayload)
	env := sdkTp.Envelope{
		Source: m.localID,
		Target: target,
		Type:   "capability.response",
		Payload: respBytes,
		Metadata: map[string]string{
			"request_id": requestID,
			"status":     "error",
		},
	}
	m.trans.Endpoint().Send(ctx, env)
}

func (m *Mesh) sendCapabilityResponse(ctx context.Context, target, requestID string, output map[string]any) {
	respPayload := map[string]any{
		"request_id": requestID,
		"status":     "success",
		"output":     output,
	}
	respBytes, _ := serializePayload(respPayload)
	env := sdkTp.Envelope{
		Source: m.localID,
		Target: target,
		Type:   "capability.response",
		Payload: respBytes,
		Metadata: map[string]string{
			"request_id": requestID,
			"status":     "success",
		},
	}
	m.trans.Endpoint().Send(ctx, env)
}

type capabilityRequest struct {
	RequestID  string         `json:"request_id"`
	Capability string         `json:"capability"`
	Name       string         `json:"name,omitempty"`
	Input      map[string]any `json:"input"`
	TimeoutMs  int64          `json:"timeout_ms,omitempty"`
}
