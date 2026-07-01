# 51_PERMISSION_MODEL.md

> Status: Draft
> Version: 0.1

# Permission Model

This document defines the authorization model used by the RouterPilot Runtime.

Permissions describe **what an identity is allowed to do**. Policies describe **under which conditions those permissions apply**.

Permissions are immutable capability identifiers evaluated by the Policy Engine.

---

# Design Goals

The permission model MUST be:

- explicit
- deterministic
- capability-oriented
- least-privilege by default
- transport-independent
- hierarchical where appropriate

No permission is implied.

---

# Permission Hierarchy

Permissions follow a dotted namespace convention.

Examples:

```
runtime.*

filesystem.*

filesystem.read
filesystem.write

network.*

network.scan
network.configure

wifi.scan
wifi.connect

process.exec

service.restart

transport.send
transport.receive
```

Wildcards MAY be supported by the Policy Engine but MUST resolve deterministically.

---

# Subjects

Permissions may be granted to:

- Agents
- Namespaces
- Plugins
- Runtime components
- Remote Nodes
- Future identities

Permissions SHOULD NOT be attached directly to providers.

---

# Resources

Permissions protect resources including:

- capabilities
- memory
- runtime APIs
- scheduler operations
- plugin lifecycle
- transports
- configuration
- storage

Everything reachable through the Runtime is considered a protected resource.

---

# Permission Resolution

Authorization follows this sequence:

```text
Identity
    │
Permission Set
    │
Policy Engine
    │
Matching Rules
    │
Decision
```

Permission resolution MUST be deterministic.

---

# Default Behavior

The default permission model is:

```
Default = DENY
```

A capability executes only when explicitly authorized.

Implicit Allow is forbidden.

---

# Least Privilege

Agents SHOULD request only the permissions required for their declared behavior.

Example:

```yaml
permissions:
  - network.scan
  - wifi.scan
```

Avoid broad permissions such as:

```
runtime.*
filesystem.*
```

unless strictly necessary.

---

# Permission Manifest

Agents SHOULD declare permissions within their manifest.

Example:

```yaml
id: wifi-monitor

permissions:
  - network.scan
  - wifi.scan

subscriptions:
  - network.changed
```

The Runtime validates manifests during registration.

---

# Permission Categories

Recommended categories:

Runtime

- runtime.*

Filesystem

- filesystem.*

Network

- network.*

Wireless

- wifi.*

Processes

- process.*

Services

- service.*

Memory

- memory.*

Scheduler

- scheduler.*

Transport

- transport.*

Plugins

- plugin.*

---

# Revocation

Permissions MAY be revoked dynamically.

Runtime behavior after revocation:

1. Reject new requests.
2. Allow in-flight execution to complete unless policy requires cancellation.
3. Emit audit event.

---

# Suggested Interfaces

```go
type Permission string

type PermissionSet interface {
    Has(Permission) bool
    List() []Permission
}
```

Permission objects are immutable after creation.

---

# Audit Requirements

Permission evaluation SHOULD record:

- subject
- requested permission
- decision
- policy ID
- timestamp
- correlation ID

Audit events support debugging and compliance.

---

# Security Invariants

- Default decision is Deny.
- Permissions are explicit.
- Permission evaluation is deterministic.
- Agents never grant permissions to themselves.
- Providers never bypass permission checks.
- Runtime owns authorization.

---

# Related Documents

- 50_POLICY_ENGINE.md
- 33_CAPABILITY_SECURITY.md
- 120_SECURITY.md

---

# Next

Continue with **60_MEMORY.md**.
