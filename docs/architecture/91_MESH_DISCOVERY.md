# 91_MESH_DISCOVERY.md

> Status: Draft
> Version: 0.1

# Mesh Discovery

The Mesh Discovery subsystem is responsible for discovering, tracking and maintaining knowledge about other RouterPilot Runtime instances participating in the distributed mesh.

Discovery is transport-independent. Concrete transports provide discovery mechanisms, while the Runtime consumes normalized peer information.

---

# Goals

The Mesh Discovery subsystem MUST provide:

- decentralized peer discovery
- peer lifecycle management
- topology awareness
- capability advertisement
- health monitoring
- transport independence

Discovery MUST NOT imply trust or authorization.

---

# Architecture

```text
          Runtime
             │
      Discovery Service
             │
    ┌────────┼────────┐
    ▼        ▼        ▼
 Reticulum  MQTT    HTTP
    │        │        │
    └────────┼────────┘
             ▼
         Peer Database
             │
             ▼
        Runtime Services
```

---

# Responsibilities

The Discovery Service MUST:

- discover peers
- maintain peer metadata
- remove stale peers
- publish discovery events
- expose peer information to the Runtime

The Discovery Service MUST NOT:

- execute capabilities
- evaluate authorization
- schedule work
- modify Runtime state

---

# Peer Lifecycle

```text
Discovered
     │
Validated
     │
Known
     │
Healthy
     │
Unavailable
     │
Expired
```

Each transition SHOULD emit a discovery event.

---

# Peer Metadata

Every discovered peer SHOULD expose:

- Runtime ID
- Node Name
- Protocol Version
- Supported Transports
- Exported Capabilities
- Labels
- Health Status
- Last Seen Timestamp

Metadata SHOULD be immutable except for health and timestamps.

---

# Discovery Sources

Discovery MAY originate from:

- Reticulum announcements
- static configuration
- multicast discovery
- bootstrap peers
- future transport-specific mechanisms

The Runtime consumes a unified peer model regardless of source.

---

# Health Monitoring

The Discovery Service SHOULD track:

- heartbeat interval
- last successful contact
- latency
- transport availability

Peers transition to **Unavailable** after configurable timeout thresholds.

---

# Capability Advertisement

Peers MAY advertise exported capabilities.

Advertisements SHOULD include:

- capability name
- version
- namespace
- availability

Advertisements are informational only.

Authorization remains local.

---

# Topology Awareness

The Runtime MAY expose topology information for:

- diagnostics
- routing optimization
- observability

The topology model MUST remain advisory.

Business logic MUST NOT depend on a specific mesh topology.

---

# Suggested Interfaces

```go
type Discovery interface {
    Peers() []Peer
    Lookup(string) (Peer, bool)
    Refresh(context.Context) error
}

type Peer struct {
    RuntimeID string
    Name      string
    Version   string
    Healthy   bool
}
```

---

# Events

The Discovery Service SHOULD emit:

- peer.discovered
- peer.updated
- peer.healthy
- peer.unavailable
- peer.expired

These events integrate with the Runtime Event Bus.

---

# Security

Discovery information is unauthenticated until validated.

Trust decisions are handled by:

- Transport
- Policy Engine
- Runtime identity verification

Discovery alone MUST NOT authorize communication.

---

# Invariants

- Discovery is transport-independent.
- Every peer has a stable Runtime ID.
- Discovery never bypasses authorization.
- Stale peers eventually expire.
- Runtime remains functional without discovered peers.

---

# Related Documents

- 80_TRANSPORT.md
- 81_RETICULUM.md
- 90_DISTRIBUTED_RUNTIME.md
- 120_SECURITY.md

---

# Next

Continue with **100_PLUGIN_SYSTEM.md**.
