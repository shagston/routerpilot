package reticulum

import (
	"context"
	"log/slog"
	"sync"
)

type DiscoveryService struct {
	mu      sync.RWMutex
	entries map[string]Identity
	changes chan PeerChange
}

func NewDiscoveryService() *DiscoveryService {
	return &DiscoveryService{
		entries: make(map[string]Identity),
		changes: make(chan PeerChange, 64),
	}
}

func (d *DiscoveryService) Announce(ctx context.Context, id Identity) {
	d.mu.Lock()
	d.entries[id.ID] = id
	d.mu.Unlock()
}

func (d *DiscoveryService) Discover(ctx context.Context, id Identity) {
	d.mu.Lock()
	existing, ok := d.entries[id.ID]
	d.entries[id.ID] = id
	d.mu.Unlock()

	peer := Peer{
		Identity: id,
		State:    PeerReachable,
	}

	if !ok {
		d.changes <- PeerChange{Event: PeerDiscovered, Peer: peer}
		slog.Debug("peer discovered", "peer_id", id.ID, "name", id.Name)
	} else if existing.Version != id.Version {
		d.changes <- PeerChange{Event: PeerUpdated, Peer: peer}
		slog.Debug("peer updated", "peer_id", id.ID)
	}
}

func (d *DiscoveryService) Lost(ctx context.Context, peerID string) {
	d.mu.Lock()
	entry, ok := d.entries[peerID]
	if ok {
		delete(d.entries, peerID)
	}
	d.mu.Unlock()

	if ok {
		d.changes <- PeerChange{
			Event: PeerLost,
			Peer:  Peer{Identity: entry, State: PeerUnreachable},
		}
		slog.Debug("peer lost", "peer_id", peerID)
	}
}

func (d *DiscoveryService) Changes() <-chan PeerChange {
	return d.changes
}

func (d *DiscoveryService) Lookup(peerID string) (Identity, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	id, ok := d.entries[peerID]
	return id, ok
}

func (d *DiscoveryService) All() []Identity {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make([]Identity, 0, len(d.entries))
	for _, id := range d.entries {
		result = append(result, id)
	}
	return result
}
