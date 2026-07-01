# 30_CAPABILITY_MODEL.md

> Status: Draft
> Version: 0.1

# Capability Model

The Capability Model is the primary abstraction layer between Agents and platform implementations.

Agents describe **what** they need to accomplish. Capability Providers decide **how** that request is fulfilled on a specific platform.

This separation is one of the core architectural principles of RouterPilot.

---

# Motivation

Traditional automation systems expose implementation details:

- shell commands
- REST endpoints
- platform APIs
- RPC methods

RouterPilot instead exposes stable capability contracts.

Example:

```
network.scan
```

rather than

```
iw dev wlan0 scan
```

This allows the same Agent to execute on different operating systems without modification.

---

# Capability Definition

A Capability represents an abstract operation with a well-defined contract.

A Capability:

- has a globally unique name
- exposes a typed input
- produces a typed result
- may fail
- requires authorization
- is transport-independent

Capabilities are immutable contracts.

---

# Capability Lifecycle

```text
Declared
    │
Registered
    │
Resolved
    │
Authorized
    │
Executed
    │
Completed
```

Authorization MUST occur before execution.

---

# Naming Convention

Capabilities use dotted namespaces.

Examples:

```
filesystem.read
filesystem.write

network.scan
network.configure

wifi.scan
wifi.connect

service.start
service.stop
service.restart

process.exec
```

Names MUST remain stable across releases.

---

# Capability Contract

Every Capability defines:

- unique identifier
- version
- input schema
- output schema
- timeout behavior
- retry semantics
- required permissions

Implementations MUST conform to the contract.

---

# Capability Resolution

Resolution is performed by the Runtime.

```text
Agent
   │
Capability Request
   │
Capability Registry
   │
Policy Engine
   │
Capability Provider
```

Agents never resolve providers directly.

---

# Capability Categories

Recommended categories:

- filesystem
- network
- wifi
- service
- process
- package
- storage
- transport
- telemetry
- crypto
- runtime

New categories should be introduced through ADRs.

---

# Suggested Interfaces

```go
type Capability interface {
    Name() string
    Version() string
}

type CapabilityRequest struct {
    Name    string
    Input   any
    Context context.Context
}

type CapabilityResult struct {
    Output any
    Error  error
}
```

---

# Error Model

Capability execution may fail because of:

- authorization
- validation
- timeout
- provider failure
- transport failure
- cancellation

Failures MUST be reported using structured errors.

---

# Versioning

Capability contracts evolve independently from providers.

Breaking changes require:

- new major version
- migration notes
- updated SDK
- updated provider implementations

---

# Invariants

- Agents never invoke providers directly.
- Capability names are stable contracts.
- Providers are replaceable.
- Every execution is authorized.
- Capability execution is observable.
- Runtime owns resolution.

---

# Related Documents

- 31_CAPABILITY_REGISTRY.md
- 32_CAPABILITY_PROVIDER.md
- 33_CAPABILITY_SECURITY.md
- 50_POLICY_ENGINE.md

---

# Next

Continue with **31_CAPABILITY_REGISTRY.md**.
