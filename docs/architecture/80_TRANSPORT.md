# 80_TRANSPORT.md

> Status: Draft
> Version: 0.1

# Transport Architecture

The Transport subsystem provides communication between Runtime instances.

The Runtime never communicates directly through sockets, HTTP, Reticulum, MQTT or any other protocol.

Instead, it communicates through a transport abstraction.

Transport implementations become interchangeable plugins.

---

# Design Goals

The Transport layer MUST provide:

- transport independence
- peer communication
- message routing
- authentication hooks
- encryption support
- connection lifecycle
- capability forwarding
- event forwarding

The Runtime MUST NOT depend on transport-specific code.

---

# Architectural Position

```text
                Runtime
                   │
             Transport API
                   │
      ┌────────────┼────────────┐
      ▼            ▼            ▼
  Reticulum      MQTT        HTTP
      │            │            │
      └────────────┼────────────┘
                   ▼
            Remote Runtime
```

Transport is infrastructure.

Business logic never depends on transport implementation.

---

# Responsibilities

The Transport subsystem MUST:

- discover peers
- establish sessions
- exchange envelopes
- route messages
- report connectivity
- expose transport metadata

The Transport subsystem MUST NOT:

- execute agents
- authorize capabilities
- modify payloads
- contain planner logic

---

# Transport Model

Every transport implements the same logical interface.

```text
Runtime
   │
Transport API
   │
Concrete Transport
   │
Physical Network
```

Replacing a transport must not change Runtime behavior.

---

# Core Concepts

## Transport

A concrete implementation capable of exchanging Runtime messages.

Examples:

- Reticulum
- MQTT
- HTTP
- NATS
- Unix Socket

---

## Peer

Another Runtime reachable through one or more transports.

Every Peer has:

- identity
- transport address
- capabilities
- health
- metadata

---

## Session

A logical communication channel between Runtime instances.

Sessions may survive multiple messages.

---

## Envelope

The transport-neutral container used to exchange Runtime messages.

Envelope contents are defined in the Transport Protocol document.

---

# Message Categories

Typical message types:

- event
- capability request
- capability response
- discovery
- heartbeat
- synchronization
- error

Additional categories may be introduced through ADRs.

---

# Discovery

Transport implementations SHOULD support peer discovery.

Discovery methods depend on transport implementation.

The Runtime only consumes normalized peer information.

---

# Routing

Transport routes messages using Runtime identities.

Routing decisions MUST be deterministic.

Transport implementations SHOULD support:

- local routing
- remote routing
- forwarding
- loop prevention

---

# Reliability

The Transport API SHOULD expose:

- delivery status
- connection state
- retry hints
- latency metrics

Reliability policies remain under Runtime control.

---

# Suggested Interface

```go
type Transport interface {
    Name() string
    Start(context.Context) error
    Stop(context.Context) error
    Send(context.Context, Envelope) error
    Peers() []Peer
}
```

The Runtime depends only on this interface.

---

# Events

The Transport SHOULD emit:

- transport.started
- transport.connected
- transport.disconnected
- transport.peer.discovered
- transport.error

---

# Security

Transport implementations SHOULD support:

- authenticated peers
- encrypted channels
- integrity verification

Security policy remains Runtime-driven.

---

# Invariants

- Runtime depends only on Transport interfaces.
- Transport implementations are replaceable.
- Payloads remain transport-independent.
- Discovery is normalized before reaching the Runtime.
- Transport failures never corrupt Runtime state.

---

# Related Documents

- 81_RETICULUM.md
- 82_TRANSPORT_PROTOCOL.md
- 90_DISTRIBUTED_RUNTIME.md

---

# Next

Continue with **81_RETICULUM.md**.
