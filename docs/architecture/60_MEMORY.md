# 60_MEMORY.md

> Status: Draft
> Version: 0.1

# Memory Architecture

The Memory subsystem provides durable and transient knowledge for Agents.

Memory is a Runtime service. Agents never access storage implementations directly.

The Memory subsystem separates *what* is remembered from *how* it is stored.

---

# Goals

The Memory subsystem MUST provide:

- deterministic APIs
- provider independence
- persistence
- session isolation
- concurrent access
- transport independence

The Runtime owns Memory lifecycle.

---

# Memory Layers

RouterPilot defines four logical layers.

```text
Working Memory
      │
Session Memory
      │
Persistent Memory
      │
External Providers
```

Each layer has different lifetime guarantees.

---

# Working Memory

Working Memory exists only during a single execution.

Contains:

- intermediate values
- execution state
- temporary variables
- planner artifacts

Destroyed when execution completes.

---

# Session Memory

Session Memory spans multiple executions belonging to the same Agent session.

Typical contents:

- recent observations
- cached results
- dialogue state
- temporary workflow data

Session Memory may expire automatically.

---

# Persistent Memory

Persistent Memory survives Runtime restarts.

Examples:

- configuration
- learned mappings
- indexes
- historical state

Persistence is delegated to Memory Providers.

---

# External Memory

External systems MAY implement the Memory Provider interface.

Examples:

- SQLite
- PostgreSQL
- BadgerDB
- Redis
- Object Storage
- Vector Databases

The Runtime remains provider-independent.

---

# Memory Objects

Memory records SHOULD include:

- Key
- Namespace
- Owner
- Value
- Version
- CreatedAt
- UpdatedAt
- Labels

The internal representation is provider-defined.

---

# Namespaces

Memory MUST support namespaces.

Examples:

runtime/

agent/

plugin/

system/

Namespaces prevent accidental collisions.

---

# Access Model

```text
Agent
   │
Memory API
   │
Policy Engine
   │
Memory Provider
```

Memory access is subject to authorization.

---

# Suggested Interfaces

```go
type Memory interface {
    Get(context.Context, string) (Value, error)
    Put(context.Context, string, Value) error
    Delete(context.Context, string) error
    List(context.Context, Prefix) ([]Value, error)
}
```

---

# Consistency

Working Memory is strongly consistent.

Persistent Memory consistency depends on the selected provider.

Applications SHOULD avoid assumptions beyond documented provider guarantees.

---

# Observability

Memory operations SHOULD emit:

- memory.read
- memory.write
- memory.delete
- memory.compacted
- memory.error

These events support tracing and diagnostics.

---

# Security

Memory operations MUST pass through the Policy Engine.

Sensitive values SHOULD support encryption through provider capabilities.

---

# Invariants

- Agents access Memory only through the Runtime.
- Memory Providers are replaceable.
- Keys are namespaced.
- Working Memory never survives execution.
- Persistent Memory survives Runtime restart unless explicitly removed.

---

# Related Documents

- 61_MEMORY_PROVIDER.md
- 50_POLICY_ENGINE.md
- 20_AGENT_MODEL.md

---

# Next

Continue with **61_MEMORY_PROVIDER.md**.
