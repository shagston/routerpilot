# 50_POLICY_ENGINE.md

> Status: Draft
> Version: 0.1

# Policy Engine

The Policy Engine is responsible for authorization decisions within the RouterPilot Runtime.

Every operation that may affect the Runtime, operating system, network, storage, transport or distributed mesh MUST be evaluated by the Policy Engine before execution.

The Policy Engine is the single source of truth for authorization.

---

# Goals

The Policy Engine MUST provide:

- deterministic authorization
- explicit allow/deny decisions
- capability-based permissions
- auditability
- transport independence
- plugin isolation
- policy versioning

The Policy Engine MUST NOT execute capabilities.

---

# Responsibilities

The Policy Engine:

- evaluates requests
- resolves policies
- applies conditions
- returns authorization decisions
- generates audit records

The Runtime remains responsible for execution.

---

# Authorization Pipeline

```text
Agent
  │
Capability Request
  │
Context Validation
  │
Policy Engine
  │
Permission Resolution
  │
Decision
  ├── Allow
  ├── Deny
  └── Conditional
  │
Runtime
  │
Capability Provider
```

No capability execution occurs before a decision.

---

# Policy Sources

Policies MAY originate from:

- static YAML files
- Runtime configuration
- signed policy bundles
- future remote policy providers

Regardless of source, policies are normalized into the same internal representation.

---

# Policy Structure

Recommended fields:

```yaml
id: default-network

effect: allow

subjects:
  - namespace: system

capabilities:
  - network.scan
  - wifi.scan

conditions:
  time: any

priority: 100
```

---

# Subjects

Policies apply to subjects.

Supported subject types:

- Agent
- Namespace
- Runtime
- Plugin
- Remote Node

Policies SHOULD avoid platform-specific identifiers.

---

# Resources

Protected resources include:

- capabilities
- memory
- filesystem
- transport
- plugins
- runtime services

Everything exposed by the Runtime should be considered a protected resource.

---

# Conditions

Policies MAY evaluate:

- namespace
- labels
- runtime mode
- transport
- execution origin
- time window
- trust level

Conditions MUST be deterministic.

---

# Decision Model

Every evaluation returns exactly one decision.

```text
Allow

Deny

Conditional
```

Unknown policies default to Deny.

---

# Conflict Resolution

When multiple policies match:

1. Explicit Deny
2. Explicit Allow
3. Conditional
4. Default Deny

Policy ordering MUST be deterministic.

---

# Audit

Every evaluation SHOULD produce:

- timestamp
- subject
- resource
- decision
- matching policy
- correlation ID
- execution ID

Audit records SHOULD be immutable.

---

# Suggested Interfaces

```go
type PolicyEngine interface {
    Evaluate(context.Context, CapabilityRequest) (Decision, error)
}

type Decision interface {
    Allowed() bool
    Reason() string
}
```

---

# Failure Handling

If policy evaluation cannot complete:

- execution MUST stop
- provider MUST NOT execute
- an audit event SHOULD be generated

The Policy Engine fails closed.

---

# Invariants

- Authorization always precedes execution.
- Default decision is Deny.
- Policies are deterministic.
- Runtime never bypasses the Policy Engine.
- Providers never authorize themselves.
- Policy evaluation is observable.

---

# Related Documents

- 51_PERMISSION_MODEL.md
- 33_CAPABILITY_SECURITY.md
- 120_SECURITY.md

---

# Next

Continue with **51_PERMISSION_MODEL.md**.
