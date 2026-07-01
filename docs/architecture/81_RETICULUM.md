# 81_RETICULUM.md

> Status: Draft
> Version: 0.1

# Reticulum Transport

Reticulum is the reference distributed transport for RouterPilot.

It provides decentralized, encrypted communication without requiring permanent servers,
public IP addresses or centralized brokers.

Within RouterPilot, Reticulum is treated as **one implementation** of the Transport API,
not as a special case inside the Runtime.

---

# Objectives

The Reticulum transport MUST provide:

- decentralized discovery
- encrypted communication
- node identities
- reliable message exchange
- capability forwarding
- event forwarding
- offline-first operation

The Runtime MUST remain unaware of Reticulum-specific implementation details.

---

# Architectural Position

```text
             Runtime
                │
         Transport API
                │
        Reticulum Adapter
                │
         Reticulum Network
                │
          Remote Runtime
```

---

# Responsibilities

The Reticulum adapter MUST:

- start and stop Reticulum services
- maintain local identity
- discover peers
- exchange Runtime envelopes
- expose peer state
- translate Runtime events to Reticulum packets

The adapter MUST NOT:

- authorize requests
- execute capabilities
- modify Runtime semantics

---

# Node Identity

Every Runtime SHOULD have a persistent Reticulum identity.

Identity SHOULD survive restarts.

Identity metadata MAY include:

- Runtime ID
- Node name
- Version
- Labels
- Supported capabilities

---

# Discovery

Discovery SHOULD be automatic.

Discovered peers become Runtime peers after validation.

Discovery events:

- peer.discovered
- peer.updated
- peer.lost

---

# Routing

The adapter routes envelopes using Reticulum destinations.

Routing decisions remain deterministic.

Loop prevention SHOULD be implemented.

---

# Message Types

Recommended envelope categories:

- event
- capability.request
- capability.response
- heartbeat
- discovery
- sync
- error

Runtime message schemas remain transport-independent.

---

# Heartbeats

The adapter SHOULD periodically publish heartbeat messages.

Heartbeat metadata MAY include:

- uptime
- Runtime version
- health
- supported transports

Missing heartbeats eventually transition peers to an unavailable state.

---

# Failure Handling

Transport failures SHOULD be isolated.

Typical failures:

- unreachable peer
- timeout
- route unavailable
- identity validation failure

The Runtime decides whether retries are appropriate.

---

# Security

Reticulum provides encrypted communication.

RouterPilot additionally enforces:

- Policy Engine authorization
- capability permissions
- audit logging
- runtime identity validation

Transport encryption does not replace Runtime authorization.

---

# Suggested Interface

```go
type ReticulumTransport interface {
    Transport
    LocalIdentity() Identity
    DiscoverPeers(context.Context) error
}
```

---

# Invariants

- Reticulum is an implementation of Transport.
- Runtime never imports Reticulum APIs directly.
- Peer identities remain stable.
- Runtime envelopes are preserved end-to-end.
- Authorization occurs after message receipt and before execution.

---

# Related Documents

- 80_TRANSPORT.md
- 82_TRANSPORT_PROTOCOL.md
- 90_DISTRIBUTED_RUNTIME.md

---

# Next

Continue with **82_TRANSPORT_PROTOCOL.md**.
