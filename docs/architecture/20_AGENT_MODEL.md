# 20_AGENT_MODEL.md

> Status: Draft
> Version: 0.1

# Agent Model

This document defines the canonical Agent model used by RouterPilot.

An Agent is the primary execution entity of the Runtime.

Agents encapsulate behavior, react to events, request capabilities, maintain local state and collaborate with other agents through the Runtime.

Agents MUST NOT communicate directly with one another.

All interaction occurs through Runtime-managed mechanisms.

---

# Goals

The Agent model provides:

- isolation
- deterministic execution
- composability
- portability
- capability-oriented behavior
- transport transparency
- lifecycle management

---

# Definition

An Agent is an autonomous execution unit consisting of:

- identity
- metadata
- lifecycle
- state
- subscriptions
- permissions
- memory
- planner (optional)
- capability requests

An Agent never owns the Runtime.

The Runtime owns every Agent.

---

# Architectural Position

```text
             Runtime
                │
        ┌───────┼────────┐
        ▼       ▼        ▼
   Agent A  Agent B  Agent C
        │       │        │
        └───────┼────────┘
                ▼
         Capability Layer
                │
                ▼
          Capability Providers
```

---

# Responsibilities

An Agent MAY:

- subscribe to events
- publish events
- request capabilities
- maintain internal state
- expose services
- schedule work

An Agent MUST NOT:

- bypass Runtime
- bypass Policy Engine
- invoke Providers directly
- manipulate Runtime internals

---

# Identity

Every Agent possesses:

- Agent ID
- Name
- Version
- Namespace
- Labels
- Metadata

Example:

```yaml
id: network-monitor
version: 1.0.0
namespace: system
labels:
  role: monitoring
```

Agent IDs are globally unique within a Runtime instance.

---

# Agent State

Recommended states:

```text
Created
 │
Initialized
 │
Ready
 │
Running
 │
Sleeping
 │
Stopping
 │
Stopped

Failure
 │
Recovering
```

State transitions are controlled exclusively by the Runtime.

---

# Agent Context

Each execution receives a dedicated Agent Context.

The context contains:

- execution ID
- runtime reference
- correlation ID
- permissions
- deadlines
- memory access
- logger

Contexts MUST NOT be shared between concurrent executions.

---

# Agent Responsibilities

An Agent should focus exclusively on domain behavior.

Infrastructure concerns belong elsewhere.

Example separation:

Agent:
- monitor Wi-Fi
- detect changes
- publish events

Runtime:
- scheduling
- execution
- retries
- cancellation

Policy:
- authorization

Transport:
- networking

---

# Agent Communication

Preferred communication mechanisms:

1. Events
2. Capability Requests
3. Runtime Services

Avoid:

- direct references
- global registries
- shared mutable objects

---

# Agent Isolation

Each Agent executes independently.

Isolation includes:

- execution context
- state
- permissions
- cancellation
- failures

A failing Agent must not terminate unrelated Agents.

---

# Agent Manifest

Agents SHOULD expose metadata through a manifest.

Example:

```yaml
id: wifi-monitor

version: 1.0.0

permissions:
  - network.scan

subscriptions:
  - network.changed

publishes:
  - wifi.changed

memory:
  persistent

scheduler:
  interval: 30s
```

The manifest describes intent rather than implementation.

---

# Suggested Interface

```go
type Agent interface {
    ID() string
    Metadata() Metadata
    Execute(AgentContext) error
}
```

Optional capabilities may be expressed through additional interfaces.

---

# Design Invariants

- Runtime owns Agent lifecycle.
- Agents never own Runtime state.
- Agents communicate through Runtime abstractions.
- Agents request capabilities rather than implementations.
- Every execution is isolated.
- Every Agent is replaceable.

---

# Related Documents

- 21_AGENT_LIFECYCLE.md
- 22_AGENT_API.md
- 23_AGENT_REGISTRY.md
- 30_CAPABILITY_MODEL.md

---

# Next

Continue with **21_AGENT_LIFECYCLE.md**.
