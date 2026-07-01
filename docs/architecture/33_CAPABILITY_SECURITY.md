# 33_CAPABILITY_SECURITY.md

> Status: Draft
> Version: 0.1

# Capability Security Model

This document defines the security architecture governing Capability execution.

Every Capability invocation MUST be evaluated before execution.

Security is enforced by the Runtime and the Policy Engine, not by individual Agents or Providers.

---

# Security Objectives

The Capability layer MUST provide:

- least privilege
- explicit authorization
- auditability
- deterministic policy evaluation
- provider isolation
- transport-independent enforcement

Capability Providers MUST assume requests are already authorized, but MUST still validate inputs.

---

# Security Pipeline

```text
Agent
  │
Capability Request
  │
Context Validation
  │
Policy Evaluation
  │
Permission Resolution
  │
Capability Resolution
  │
Provider Execution
  │
Audit Event
```

Execution MUST stop immediately after a denied policy decision.

---

# Trust Boundaries

RouterPilot separates trust into distinct boundaries.

Boundary A

Agent
↓

Runtime

Boundary B

Runtime
↓

Capability Provider

Boundary C

Runtime
↓

Transport

Boundary D

Runtime
↓

Plugins

Each boundary is validated independently.

---

# Permissions

Capabilities are protected through explicit permissions.

Example:

```yaml
permissions:

  - filesystem.read
  - filesystem.write
  - network.scan
  - service.restart
```

Permissions SHOULD be additive.

Implicit permissions MUST NOT exist.

---

# Permission Scopes

Recommended scopes:

- runtime
- system
- network
- filesystem
- services
- transport
- plugins

Scopes simplify policy management while preserving fine-grained capability names.

---

# Policy Decisions

Every request results in one of:

- Allow
- Deny
- Conditional Allow

Conditional decisions may require:

- time limits
- resource limits
- transport restrictions
- namespace restrictions

---

# Capability Classification

Capabilities SHOULD be categorized by risk.

Example:

Low Risk

- filesystem.read
- telemetry.publish

Medium Risk

- network.scan
- package.query

High Risk

- process.exec
- service.restart
- filesystem.write

Critical

- firmware.flash
- runtime.shutdown

Risk classification assists policy generation.

---

# Provider Isolation

Providers MUST NOT:

- share mutable Runtime state
- modify Registry contents
- bypass authorization
- expose internal implementation objects

Providers operate only through public SDK interfaces.

---

# Audit Requirements

Every execution SHOULD generate:

- timestamp
- agent ID
- capability
- provider
- policy result
- duration
- execution status
- correlation ID

Audit records SHOULD be immutable.

---

# Suggested Interfaces

```go
type AuthorizationResult struct {
    Allowed bool
    Reason  string
}

type PolicyDecision interface {
    Evaluate(CapabilityRequest) AuthorizationResult
}
```

---

# Failure Modes

Security failures include:

- missing permission
- invalid signature
- malformed request
- unknown capability
- revoked provider
- policy evaluation failure

The Runtime MUST fail closed whenever authorization cannot be determined.

---

# Distributed Runtime

Remote capability execution follows identical policy rules.

Authorization occurs on the executing Runtime.

Transport MUST NOT weaken policy enforcement.

---

# Security Invariants

- Every capability request is authorized.
- Authorization precedes provider execution.
- Policies are deterministic.
- Providers never elevate privileges.
- Agents never bypass Runtime security.
- Audit events are generated for security-relevant actions.
- Unknown capabilities are rejected.

---

# Related Documents

- 30_CAPABILITY_MODEL.md
- 31_CAPABILITY_REGISTRY.md
- 32_CAPABILITY_PROVIDER.md
- 50_POLICY_ENGINE.md
- 51_PERMISSION_MODEL.md
- 120_SECURITY.md

---

# Next

Continue with **40_EVENT_BUS.md**.
