# 40_EVENT_BUS.md

> Status: Draft
> Version: 0.1

# Event Bus

The Event Bus is the internal communication backbone of the RouterPilot Runtime.

All major Runtime subsystems communicate through immutable events rather than direct dependencies. This minimizes coupling, improves observability, and enables transparent distributed execution.

---

# Goals

The Event Bus MUST provide:

- publish / subscribe messaging
- asynchronous delivery
- deterministic routing
- transport independence
- context propagation
- replay-friendly events
- observability

The Event Bus MUST NOT contain business logic.

---

# Architecture

```text
              Runtime
                 │
 ┌───────────────┼────────────────┐
 ▼               ▼                ▼
Agents      Capability Layer   Scheduler
     \          |              //
      \         |             //
       └──────── Event Bus ─────┘
                 │
         Subscription Engine
                 │
        Local / Remote Consumers
```

---

# Core Concepts

## Event

An immutable record describing something that happened.

Examples:

- runtime.started
- agent.ready
- capability.completed
- scheduler.tick
- transport.connected

## Publisher

A component emitting events.

Examples:

- Runtime
- Agent
- Scheduler
- Policy Engine
- Transport

## Subscriber

A component interested in one or more event types.

Subscribers never modify published events.

---

# Event Lifecycle

```text
Create
   │
Validate
   │
Publish
   │
Route
   │
Deliver
   │
Acknowledge (optional)
```

Events are immutable after publication.

---

# Event Envelope

Every event SHOULD contain:

- Event ID
- Type
- Timestamp
- Source
- Correlation ID
- Payload
- Metadata
- Version

Payload schemas belong to the event definition.

---

# Topics

Events are grouped by namespaces.

Examples:

```
runtime.*
agent.*
capability.*
scheduler.*
transport.*
policy.*
memory.*
plugin.*
```

Topic hierarchies should remain stable.

---

# Subscription Model

Supported patterns:

- exact topic
- namespace wildcard
- multiple subscriptions

Examples:

```
agent.started
agent.*
runtime.*
```

---

# Ordering

Ordering is guaranteed only within a single event stream.

Cross-stream ordering is not guaranteed.

Subscribers must not rely on global ordering.

---

# Failure Handling

Publisher failures MUST NOT corrupt the Event Bus.

Subscriber failures SHOULD be isolated.

Runtime may retry delivery according to delivery policy.

---

# Suggested Interfaces

```go
type EventBus interface {
    Publish(context.Context, Event) error
    Subscribe(Subscription) error
    Unsubscribe(string) error
}
```

---

# Observability

The Event Bus SHOULD expose metrics:

- published events
- delivered events
- dropped events
- subscriber latency
- queue depth
- retries

---

# Invariants

- Events are immutable.
- Publishers do not know subscribers.
- Subscribers do not modify events.
- Routing is deterministic.
- Event Bus remains transport-independent.
- Event processing is observable.

---

# Related Documents

- 41_EVENT_PROTOCOL.md
- 42_EVENT_DELIVERY.md
- 10_RUNTIME.md
- 20_AGENT_MODEL.md

---

# Next

Continue with **41_EVENT_PROTOCOL.md**.
