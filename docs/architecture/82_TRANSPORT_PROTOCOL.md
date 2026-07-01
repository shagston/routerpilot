# 82_TRANSPORT_PROTOCOL.md

> Status: Draft
> Version: 0.1

# Transport Protocol

This document defines the transport-neutral protocol used between RouterPilot Runtime instances.

The protocol specifies the logical message format exchanged across any supported transport implementation. Concrete transports serialize and deliver these messages but MUST NOT change their semantics.

---

# Goals

The protocol MUST provide:

- transport independence
- version negotiation
- request/response correlation
- extensibility
- integrity metadata
- forward compatibility

---

# Architecture

```text
Runtime
   │
Transport Protocol
   │
Envelope
   │
Transport Adapter
   │
Physical Network
```

---

# Envelope

Every message is wrapped in a transport-neutral envelope.

Required fields:

- MessageID
- Type
- ProtocolVersion
- SourceRuntime
- DestinationRuntime
- CorrelationID
- Timestamp
- Payload

Optional fields:

- TraceID
- TTL
- Compression
- Signature
- Metadata

---

# Message Types

Core message categories:

- discovery.request
- discovery.response
- heartbeat
- event.publish
- capability.request
- capability.response
- sync.request
- sync.response
- error

New message types SHOULD use namespaced identifiers.

---

# Capability Request

A capability request SHOULD contain:

- capability name
- input payload
- timeout
- execution options
- correlation ID

The receiving Runtime performs policy evaluation before execution.

---

# Capability Response

Responses SHOULD contain:

- request ID
- status
- output payload
- execution duration
- error (if any)

Responses MUST preserve the original Correlation ID.

---

# Version Negotiation

Protocol compatibility follows Semantic Versioning.

Rules:

- same major version => compatible
- newer minor versions => additive features only
- unknown optional fields => ignored

Unsupported major versions SHOULD produce a protocol error.

---

# Integrity

The protocol SHOULD support:

- signatures
- checksums
- message authentication
- replay protection

Integrity mechanisms are transport-agnostic.

---

# Error Model

Standard protocol errors include:

- unsupported.version
- invalid.envelope
- unknown.message
- authorization.denied
- destination.unreachable
- timeout
- malformed.payload

Errors SHOULD be machine-readable.

---

# Suggested Interfaces

```go
type Envelope struct {
    Header  Header
    Payload []byte
}

type Header struct {
    MessageID        string
    Type             string
    ProtocolVersion  string
    CorrelationID    string
}
```

---

# Observability

Implementations SHOULD expose:

- messages sent
- messages received
- protocol errors
- serialization failures
- negotiation failures
- average latency

---

# Invariants

- Protocol semantics are transport-independent.
- Envelopes are immutable after transmission.
- Correlation IDs are preserved end-to-end.
- Unknown optional fields are ignored.
- Authorization always occurs after receipt and before execution.

---

# Related Documents

- 80_TRANSPORT.md
- 81_RETICULUM.md
- 90_DISTRIBUTED_RUNTIME.md

---

# Next

Continue with **90_DISTRIBUTED_RUNTIME.md**.
