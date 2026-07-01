# 90_DISTRIBUTED_RUNTIME.md

> Status: Draft
> Version: 0.1

# Distributed Runtime

The Distributed Runtime extends RouterPilot beyond a single process or device.

Each Runtime instance is an autonomous node capable of executing agents, exposing capabilities, exchanging events and cooperating with peers without requiring a central coordinator.

Distributed execution must preserve the same programming model as local execution whenever practical.

---

# Goals

The Distributed Runtime MUST provide:

- decentralized operation
- peer discovery
- remote capability execution
- event propagation
- workload distribution
- fault isolation
- transport independence

---

# Architecture

```text
             ┌─────────────── Mesh ───────────────┐

        Runtime A  ◄──────────────►  Runtime B
             ▲                            ▲
             │                            │
             ▼                            ▼
        Runtime C  ◄──────────────►  Runtime D

      Every Runtime is both a client and a server.
```

No permanent control plane is required.

---

# Node Model

Every Runtime node contains:

- Runtime Core
- Agent Manager
- Capability Registry
- Policy Engine
- Event Bus
- Scheduler
- Memory
- Transport

Nodes remain fully functional when disconnected from the mesh.

---

# Responsibilities

Each Runtime MAY:

- execute local agents
- publish events
- expose capabilities
- consume remote capabilities
- synchronize metadata
- participate in discovery

Each Runtime MUST remain authoritative for its own local state.

---

# Peer Discovery

Peers are discovered through the active Transport.

Discovery information SHOULD include:

- Runtime ID
- Node Name
- Supported Transports
- Protocol Version
- Health Status
- Exported Capabilities

Discovery does not imply trust.

---

# Remote Capability Execution

Execution flow:

```text
Agent
   │
Capability Request
   │
Policy Evaluation
   │
Transport
   │
Remote Runtime
   │
Policy Evaluation
   │
Capability Provider
   │
Response
```

Authorization occurs independently on the sending and receiving Runtime.

---

# Distributed Events

Events MAY be propagated to peers.

Propagation policy determines:

- which topics are exported
- which peers receive events
- filtering rules
- loop prevention

Event semantics remain unchanged.

---

# Node Identity

Every Runtime has a persistent identity.

Identity SHOULD include:

- Runtime ID
- Public Identity
- Labels
- Version
- Capabilities

Identity persistence enables stable trust relationships.

---

# Fault Isolation

Node failures MUST NOT cascade.

Examples:

- unreachable peer
- transport outage
- provider crash
- scheduler failure

A local Runtime continues operating whenever possible.

---

# Synchronization

Synchronization MAY include:

- exported capabilities
- node metadata
- health information
- protocol versions

Application state synchronization is outside the scope of the Runtime.

---

# Suggested Interfaces

```go
type RemoteRuntime interface {
    ID() string
    Capabilities() []CapabilityDescriptor
    Health() HealthStatus
}

type Mesh interface {
    Peers() []RemoteRuntime
}
```

---

# Observability

The Distributed Runtime SHOULD expose:

- connected peers
- disconnected peers
- remote executions
- synchronization latency
- transport failures
- exported capabilities

---

# Invariants

- Every Runtime remains autonomous.
- Local execution remains possible without peers.
- Remote execution preserves capability contracts.
- Transport implementations remain replaceable.
- Distributed failures do not corrupt local Runtime state.

---

# Related Documents

- 80_TRANSPORT.md
- 81_RETICULUM.md
- 82_TRANSPORT_PROTOCOL.md
- 91_MESH_DISCOVERY.md

---

# Next

Continue with **91_MESH_DISCOVERY.md**.
