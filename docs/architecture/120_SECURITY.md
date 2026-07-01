# 120_SECURITY.md

> Status: Draft
> Version: 0.1

# Security Architecture

Security is a foundational property of the RouterPilot Runtime.

Every subsystem is designed assuming that nodes, transports, plugins and execution plans may be untrusted until verified. Authorization, isolation and auditability are enforced by architecture rather than by convention.

---

# Security Goals

The Runtime MUST provide:

- secure-by-default behavior
- least privilege
- explicit authorization
- deterministic policy evaluation
- plugin isolation
- transport-independent security
- end-to-end auditability
- defense in depth

The Runtime MUST fail closed whenever authorization cannot be established.

---

# Security Layers

```text
           Agent
             │
     Identity & Context
             │
       Policy Engine
             │
   Capability Authorization
             │
      Runtime Execution
             │
      Provider Isolation
             │
      Transport Security
             │
        Persistent Storage
```

Every layer validates its own trust assumptions.

---

# Trust Model

Trust is established incrementally.

Trust domains include:

- Local Runtime
- Local Agents
- Plugins
- Remote Runtimes
- Transport Peers
- Memory Providers

Discovery never implies trust.

Authentication never implies authorization.

Authorization never implies unrestricted access.

---

# Identity

Every Runtime SHOULD have a persistent identity.

Every Agent MUST have a unique identifier.

Remote identities SHOULD be cryptographically verifiable where supported by the transport.

Identity is immutable for the lifetime of a Runtime instance.

---

# Authorization

All privileged actions require authorization.

Protected operations include:

- capability execution
- memory access
- transport communication
- plugin lifecycle
- scheduler control
- runtime administration

The Policy Engine is the only component that grants authorization decisions.

---

# Least Privilege

Agents and plugins SHOULD request only the permissions they require.

Broad wildcard permissions SHOULD be avoided.

Permissions are additive and explicitly declared.

---

# Plugin Security

Plugins execute through the public SDK.

Plugins MUST NOT:

- import internal Runtime packages
- modify Runtime state directly
- bypass policy evaluation
- impersonate other plugins

Plugin failures remain isolated from the Runtime.

---

# Transport Security

Transport implementations SHOULD provide:

- peer authentication
- encryption
- integrity verification
- replay protection

Transport security complements but never replaces Runtime authorization.

---

# Secrets

Secrets SHOULD be:

- encrypted at rest
- redacted from logs
- unavailable to unauthorized plugins
- accessed through dedicated providers

Secrets MUST NOT be embedded in source code or manifests.

---

# Audit Logging

Security-relevant operations SHOULD generate immutable audit events.

Recommended fields:

- timestamp
- runtime ID
- agent ID
- capability
- decision
- policy ID
- correlation ID
- outcome

Audit records support incident response and compliance.

---

# Threat Model

Primary threats include:

- unauthorized capability execution
- malicious plugins
- compromised peers
- replay attacks
- privilege escalation
- data tampering
- denial of service

Mitigations SHOULD be documented through ADRs and subsystem specifications.

---

# Incident Handling

The Runtime SHOULD:

1. detect anomalies
2. isolate affected components
3. emit security events
4. preserve audit evidence
5. continue safe operation when possible

---

# Suggested Interfaces

```go
type SecurityService interface {
    Authorize(context.Context, CapabilityRequest) (Decision, error)
    Audit(SecurityEvent) error
}
```

---

# Security Invariants

- Default authorization is Deny.
- Every capability is authorized.
- Plugins cannot bypass the SDK.
- Transport does not bypass Policy.
- Audit events are immutable.
- Secrets are never exposed through public APIs.
- Runtime remains the root of trust for execution.

---

# Related Documents

- 33_CAPABILITY_SECURITY.md
- 50_POLICY_ENGINE.md
- 51_PERMISSION_MODEL.md
- 80_TRANSPORT.md
- 100_PLUGIN_SYSTEM.md

---

# Next

Continue with **130_REPOSITORY.md**.
