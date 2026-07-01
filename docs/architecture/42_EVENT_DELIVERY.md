# 42_EVENT_DELIVERY.md

> Status: Draft
> Version: 0.1

# Event Delivery

This document specifies how events are routed, delivered, retried and acknowledged within the RouterPilot Runtime.

Delivery semantics are independent of the transport layer. Whether events are local or remote, the Runtime must expose a consistent delivery model.

---

# Objectives

The delivery subsystem MUST provide:

- reliable routing
- deterministic delivery
- subscriber isolation
- retry support
- back-pressure handling
- observable delivery
- transport-independent semantics

---

# Delivery Pipeline

```text
Publisher
    │
    ▼
 Event Bus
    │
Validation
    │
Routing
    │
Subscription Matching
    │
Queue
    │
Delivery
    │
Acknowledgement
```

Each stage is independent and observable.

---

# Delivery Modes

The Runtime SHOULD support:

## Fire-and-Forget

The publisher does not wait for subscribers.

Typical use:

- telemetry
- metrics
- informational events

---

## Reliable Delivery

Delivery succeeds only after subscriber acknowledgement.

Typical use:

- lifecycle events
- policy changes
- distributed coordination

---

## Broadcast

An event is delivered independently to all matching subscribers.

Subscriber failures MUST NOT prevent delivery to other subscribers.

---

# Subscription Matching

The Runtime SHOULD support:

- exact match
- namespace wildcard
- multiple subscriptions

Examples:

```
runtime.started

runtime.*

agent.*

capability.completed
```

Subscription resolution MUST be deterministic.

---

# Queues

Every subscriber SHOULD own an independent delivery queue.

Benefits:

- isolation
- back-pressure control
- failure containment

A slow subscriber MUST NOT block unrelated subscribers.

---

# Retry Policy

Retries are controlled by delivery policy.

Recommended parameters:

- maximum attempts
- exponential backoff
- retry window
- retryable error classes

Non-retryable failures terminate immediately.

---

# Dead Letter Queue

Events that permanently fail delivery SHOULD be moved to a Dead Letter Queue (DLQ).

DLQ entries SHOULD include:

- original event
- subscriber
- failure reason
- retry count
- timestamp

DLQs improve debugging and replay.

---

# Acknowledgements

Acknowledgement confirms successful processing.

Two models are supported:

- automatic acknowledgement
- explicit acknowledgement

Explicit acknowledgement is recommended for distributed transports.

---

# Ordering Guarantees

Ordering is guaranteed:

- within a subscriber queue
- within a single event stream

Ordering is NOT guaranteed across independent subscribers or event types.

Applications MUST NOT rely on global ordering.

---

# Back-pressure

When subscribers cannot keep up:

Possible strategies:

- bounded queues
- dropping low-priority events
- delaying publishers
- dynamic throttling

The chosen strategy SHOULD be configurable.

---

# Failure Isolation

Subscriber failures are isolated.

Publisher failures do not invalidate already queued events.

Transport failures affect only remote delivery.

---

# Suggested Interfaces

```go
type DeliveryPolicy struct {
    Reliable    bool
    MaxRetries  int
    Backoff     time.Duration
}

type Subscriber interface {
    Handle(context.Context, Event) error
}
```

---

# Metrics

Recommended metrics:

- events published
- events delivered
- acknowledgements
- retries
- dropped events
- queue depth
- subscriber latency
- DLQ size

These metrics SHOULD be exported through the Runtime observability API.

---

# Invariants

- Events remain immutable during delivery.
- Subscribers are isolated.
- Retry decisions are deterministic.
- Delivery policies are explicit.
- Ordering guarantees are documented.
- Delivery semantics are transport-independent.

---

# Related Documents

- 40_EVENT_BUS.md
- 41_EVENT_PROTOCOL.md
- 50_POLICY_ENGINE.md
- 80_TRANSPORT.md

---

# Next

Continue with **50_POLICY_ENGINE.md**.
