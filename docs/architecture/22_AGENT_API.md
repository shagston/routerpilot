# 22_AGENT_API.md

> Status: Draft
> Version: 0.1

# Agent API

This document defines the public Agent API exposed by the RouterPilot SDK.

The API is intentionally small. The Runtime owns execution, while Agents implement business behavior through stable interfaces.

---

# Design Goals

The Agent API MUST be:

- stable
- transport-independent
- planner-independent
- capability-oriented
- testable
- deterministic

---

# Core Interface

```go
type Agent interface {
    ID() string
    Metadata() Metadata
    Execute(AgentContext) error
}
```

The Runtime invokes `Execute` for every scheduled execution.

Agents MUST NOT spawn unmanaged execution loops.

---

# Metadata

Metadata identifies an Agent and describes its behavior.

Recommended fields:

```go
type Metadata struct {
    ID          string
    Name        string
    Version     string
    Namespace   string
    Description string
    Labels      map[string]string
}
```

Metadata MUST be immutable after registration.

---

# AgentContext

The Runtime provides an AgentContext for each execution.

Responsibilities:

- capability access
- event publication
- memory access
- logging
- tracing
- cancellation
- deadlines

Example:

```go
type AgentContext interface {
    Context() context.Context
    Capabilities() CapabilityRegistry
    Events() EventBus
    Memory() MemoryProvider
    Logger() Logger
}
```

---

# Optional Interfaces

Agents MAY implement additional interfaces.

## Initializer

```go
type Initializer interface {
    Initialize(AgentContext) error
}
```

Executed once before the Agent enters Ready.

## ShutdownHook

```go
type ShutdownHook interface {
    Shutdown(AgentContext) error
}
```

Executed during graceful shutdown.

## HealthReporter

```go
type HealthReporter interface {
    Health() HealthStatus
}
```

Provides health information without affecting lifecycle.

---

# Execution Rules

The Runtime guarantees:

- one execution context per invocation
- policy enforcement
- capability resolution
- event publication
- cancellation propagation

Agents MUST assume Execute may be called multiple times over their lifetime.

---

# Error Handling

`Execute` returns:

- `nil` on success
- an error on failure

The Runtime classifies failures and decides whether to retry, recover or stop the Agent.

Agents SHOULD return typed errors where possible.

---

# Thread Safety

Unless explicitly documented otherwise:

- Agent implementations SHOULD be safe for concurrent execution.
- Shared mutable state SHOULD be avoided.
- Internal synchronization belongs to the Agent implementation.

---

# Testing

Agent implementations should be testable without a full Runtime.

Recommended approach:

- mock CapabilityRegistry
- mock EventBus
- mock MemoryProvider
- inject fake AgentContext

---

# Versioning

The Agent API follows Semantic Versioning.

Breaking interface changes require:

- ADR
- migration guide
- major version increment

---

# Invariants

- Runtime always owns invocation.
- Execute is the primary entry point.
- Context is never nil.
- Metadata is immutable.
- Agents depend only on SDK packages.

---

# Related Documents

- 20_AGENT_MODEL.md
- 21_AGENT_LIFECYCLE.md
- 23_AGENT_REGISTRY.md
- 13_RUNTIME_EXECUTION.md

---

# Next

Continue with **23_AGENT_REGISTRY.md**.
