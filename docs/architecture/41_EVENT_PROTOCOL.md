# 41_EVENT_PROTOCOL.md

> Status: Draft
> Version: 0.1

# Event Protocol

This document defines the canonical event protocol used internally by RouterPilot and, where applicable, across distributed Runtime instances.

The protocol is transport-independent. Reticulum, MQTT, HTTP or any future transport serialize the same logical event model.

---

# Objectives

The protocol MUST provide:

- immutable events
- versioned payloads
- correlation
- tracing
- transport independence
- forward compatibility

---

# Event Structure

```text
Event
 ├── Header
 ├── Payload
 └── Metadata
```

The payload contains domain data.

The header contains routing information.

Metadata contains optional execution information.

---

# Header

Required fields:

- EventID
- EventType
- Version
- Timestamp
- SourceRuntime
- SourceAgent
- CorrelationID

Optional fields:

- ParentEventID
- TraceID
- TTL
- Priority

---

# Payload

Payloads are domain-specific.

Examples:

runtime.started

```json
{
  "runtime_id":"node-1",
  "version":"1.0.0"
}
```

agent.started

```json
{
  "agent":"wifi-monitor"
}
```

Payload schemas SHOULD be versioned independently from the transport.

---

# Metadata

Metadata may contain:

- labels
- tags
- diagnostics
- execution metrics
- transport hints

Consumers MUST ignore unknown metadata fields.

---

# Event Types

Recommended namespaces:

```
runtime.*
agent.*
capability.*
policy.*
scheduler.*
memory.*
transport.*
plugin.*
system.*
```

Event type names MUST remain stable.

---

# Correlation

Every execution chain SHOULD reuse a single Correlation ID.

Example:

```text
Planner
  │
Plan Created
  │
Execution Started
  │
Capability Requested
  │
Capability Completed
```

All events share one Correlation ID while keeping unique Event IDs.

---

# Trace Propagation

Trace IDs enable distributed tracing.

If a Trace ID exists, child events inherit it.

Correlation and Trace IDs serve different purposes and MUST NOT be conflated.

---

# Serialization

The protocol is serialization-agnostic.

Supported encodings MAY include:

- JSON
- CBOR
- MessagePack
- Protocol Buffers

Encoding selection is a transport concern.

---

# Compatibility

Event schemas follow Semantic Versioning.

Rules:

- additive fields are backward compatible
- field removal requires a major version
- consumers ignore unknown fields

---

# Suggested Interfaces

```go
type Event struct {
    Header   EventHeader
    Payload  any
    Metadata map[string]any
}

type EventHeader struct {
    ID            string
    Type          string
    Version       string
    CorrelationID string
    TraceID       string
}
```

---

# Invariants

- Event IDs are globally unique.
- Payloads are immutable.
- Headers never change after publication.
- Unknown metadata is ignored.
- Event protocol is independent of transport.

---

# Related Documents

- 40_EVENT_BUS.md
- 42_EVENT_DELIVERY.md
- 80_TRANSPORT.md

---

# Next

Continue with **42_EVENT_DELIVERY.md**.
