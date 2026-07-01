package distributed

import "time"

type PeerState int

const (
	PeerDiscovered  PeerState = 0
	PeerValidated   PeerState = 1
	PeerKnown       PeerState = 2
	PeerHealthy     PeerState = 3
	PeerUnavailable PeerState = 4
	PeerExpired     PeerState = 5
)

func (s PeerState) String() string {
	switch s {
	case PeerDiscovered:
		return "discovered"
	case PeerValidated:
		return "validated"
	case PeerKnown:
		return "known"
	case PeerHealthy:
		return "healthy"
	case PeerUnavailable:
		return "unavailable"
	case PeerExpired:
		return "expired"
	default:
		return "unknown"
	}
}

type CapabilityDescriptor struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Namespace string `json:"namespace"`
	Available bool   `json:"available"`
}

type Peer struct {
	RuntimeID  string                `json:"runtime_id"`
	Name       string                `json:"name"`
	Version    string                `json:"version"`
	State      PeerState             `json:"state"`
	Transports []string              `json:"transports"`
	Labels     map[string]string     `json:"labels,omitempty"`
	Capabilities []CapabilityDescriptor `json:"capabilities,omitempty"`
	Health     string                `json:"health"`
	LastSeen   time.Time             `json:"last_seen"`
	FirstSeen  time.Time             `json:"first_seen"`
}

type PeerEvent int

const (
	EventPeerDiscovered PeerEvent = iota
	EventPeerUpdated
	EventPeerHealthy
	EventPeerUnavailable
	EventPeerExpired
)

func (e PeerEvent) String() string {
	switch e {
	case EventPeerDiscovered:
		return "peer.discovered"
	case EventPeerUpdated:
		return "peer.updated"
	case EventPeerHealthy:
		return "peer.healthy"
	case EventPeerUnavailable:
		return "peer.unavailable"
	case EventPeerExpired:
		return "peer.expired"
	default:
		return "peer.unknown"
	}
}
