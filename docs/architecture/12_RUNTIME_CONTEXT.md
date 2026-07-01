# 12_RUNTIME_CONTEXT.md

> Status: Draft
> Version: 0.1

# Runtime Context

This document defines how execution context is created, propagated and terminated within the RouterPilot Runtime.

The Runtime Context is the primary mechanism for carrying execution metadata across subsystem boundaries.

---

# Goals

The context model MUST provide:

- cancellation
- deadlines
- identity propagation
- trace propagation
- authorization context
- correlation IDs
- request-scoped metadata

The context MUST NOT be used as a generic data store.

---

# Context Hierarchy

```text
Runtime Context
      │
      ├── Scheduler Context
      │
      ├── Transport Context
      │
      ├── Agent Context
      │        │
      │        └── Capability Context
      │
      └── Plugin Context
```

Every child context inherits cancellation from its parent.

---

# Runtime Context

Created once during Runtime startup.

Contains:

- runtime identifier
- startup timestamp
- logger
- metrics provider
- configuration handle
- tracing provider

The Runtime Context lives until shutdown.

---

# Agent Context

Created for each Agent execution.

Contains:

- agent ID
- execution ID
- permissions
- planner metadata
- execution deadline

An Agent Context MUST NOT be shared between concurrent executions.

---

# Capability Context

Created for every capability request.

Contains:

- capability name
- provider identifier
- authorization result
- request metadata
- execution timeout

Every capability execution receives a fresh context.

---

# Transport Context

Carries transport-specific metadata without exposing transport implementations to the Runtime.

Typical fields:

- peer ID
- transport ID
- message ID
- receive timestamp

---

# Plugin Context

Provides plugins with runtime services while restricting direct access to internal implementation.

Plugins receive only public SDK abstractions.

---

# Cancellation

Cancellation propagates downward.

```text
Runtime Cancel
      │
      ▼
Agent Cancel
      │
      ▼
Capability Cancel
```

Cancellation MUST be cooperative.

Providers should stop work as soon as practical.

---

# Deadlines

Deadlines may originate from:

- scheduler
- API request
- transport message
- policy

Child contexts inherit the earliest deadline.

---

# Metadata

Recommended metadata:

- Correlation ID
- Trace ID
- Request ID
- Agent ID
- Capability Name
- Planner ID
- Transport ID

Metadata should remain immutable after creation.

---

# Error Propagation

Errors flow upward.

```text
Provider
    │
Capability
    │
Agent
    │
Runtime
```

Context cancellation should preserve the original cause whenever possible.

---

# Suggested Interface

```go
type ExecutionContext interface {
    Context() context.Context
    CorrelationID() string
    TraceID() string
    AgentID() string
    Deadline() (time.Time, bool)
}
```

---

# Invariants

- Every execution has exactly one root context.
- Child contexts inherit cancellation.
- Context values are immutable after creation.
- Context objects must not be reused across independent executions.
- Internal runtime objects are never exposed through context.

---

# Related Documents

- 10_RUNTIME.md
- 11_RUNTIME_LIFECYCLE.md
- 13_RUNTIME_EXECUTION.md

---

# Next

Continue with **13_RUNTIME_EXECUTION.md**.
