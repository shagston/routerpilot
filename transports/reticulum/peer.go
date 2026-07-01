package reticulum

import "time"

type PeerState int

const (
	PeerUnknown    PeerState = 0
	PeerReachable  PeerState = 1
	PeerUnreachable PeerState = 2
)

type Peer struct {
	Identity   Identity          `json:"identity"`
	Address    string            `json:"address"`
	State      PeerState         `json:"state"`
	LastSeen   time.Time         `json:"last_seen"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Uptime     string            `json:"uptime,omitempty"`
	Health     string            `json:"health,omitempty"`
	Transports []string          `json:"transports,omitempty"`
}

type PeerEvent int

const (
	PeerDiscovered PeerEvent = iota
	PeerUpdated
	PeerLost
)

type PeerChange struct {
	Event PeerEvent
	Peer  Peer
}
