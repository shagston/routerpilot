# 23_AGENT_REGISTRY.md

> Status: Draft
> Version: 0.1

# Agent Registry

The Agent Registry is the authoritative catalog of all agents known to a Runtime.

The Registry owns metadata, discovery and lifecycle registration. It does **not**
execute agents. Execution remains the responsibility of the Runtime.

---

# Objectives

The Agent Registry provides:

- agent discovery
- metadata lookup
- lifecycle registration
- namespace isolation
- dependency lookup
- health visibility

The Registry MUST remain transport-independent.

---

# Responsibilities

The Registry MUST:

- register agents
- unregister agents
- expose immutable metadata
- resolve agents by identifier
- enumerate agents
- expose runtime-visible state

The Registry MUST NOT:

- execute agents
- evaluate policy
- invoke capabilities
- schedule work

---

# Registration Flow

```text
Create Agent
      │
Validate Manifest
      │
Validate Metadata
      │
Register
      │
Publish agent.registered
      │
Available for Runtime
```

Registration MUST fail if:

- Agent ID already exists
- Metadata is invalid
- Manifest validation fails

---

# Agent Descriptor

Every registered agent is represented by a descriptor.

Suggested fields:

```go
type AgentDescriptor struct {
    ID          string
    Name        string
    Version     string
    Namespace   string
    Labels      map[string]string
    State       AgentState
    Health      HealthStatus
}
```

Descriptors are immutable except for runtime state and health.

---

# Lookup Operations

The Registry SHOULD support:

- By ID
- By Namespace
- By Label
- By State
- By Capability (optional)
- List All

Lookups MUST be read-only.

---

# Registration Rules

- Agent IDs are unique within a Runtime.
- Registration is idempotent.
- Duplicate registration returns an error.
- Unregistration is safe to repeat.

---

# Lifecycle Integration

The Runtime updates lifecycle state.

The Registry reflects:

Created
Initialized
Registered
Ready
Running
Sleeping
Stopping
Stopped
Failure

The Registry never drives lifecycle transitions.

---

# Health

Health information is advisory.

Suggested values:

- Healthy
- Degraded
- Unhealthy
- Unknown

Health MUST NOT replace lifecycle state.

---

# Suggested Interface

```go
type AgentRegistry interface {
    Register(Agent) error
    Unregister(string) error
    Get(string) (AgentDescriptor, bool)
    List() []AgentDescriptor
}
```

---

# Events

The Registry SHOULD publish:

- agent.registered
- agent.unregistered
- agent.updated
- agent.health.changed

---

# Distributed Considerations

Local Registry remains authoritative for local agents.

Distributed registries MAY synchronize metadata through the Transport layer.

Synchronization MUST NOT expose runtime internals.

---

# Invariants

- Registry stores metadata only.
- Runtime owns execution.
- Runtime owns lifecycle.
- Registry operations are thread-safe.
- Metadata is immutable after registration.

---

# Related Documents

- 20_AGENT_MODEL.md
- 21_AGENT_LIFECYCLE.md
- 22_AGENT_API.md
- 30_CAPABILITY_MODEL.md

---

# Next

Continue with **30_CAPABILITY_MODEL.md**.
