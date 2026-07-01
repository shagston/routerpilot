# 32_CAPABILITY_PROVIDER.md

> Status: Draft
> Version: 0.1

# Capability Providers

A Capability Provider is the concrete implementation of one or more Capability contracts.

Providers translate platform-independent capability requests into platform-specific operations while remaining invisible to Agents.

Agents never know which Provider executes a request.

---

# Purpose

Providers isolate the Runtime from operating-system details.

Examples:

- Linux Provider
- OpenWrt Provider
- Windows Provider
- Mock Provider
- Remote Provider

The Runtime interacts only with Capability contracts.

---

# Architecture

```text
             Agent
               │
Capability Request
               │
               ▼
      Capability Registry
               │
               ▼
        Provider Resolver
               │
               ▼
      Selected Provider
               │
               ▼
 Platform Implementation
```

---

# Responsibilities

A Provider MUST:

- implement one or more capabilities
- validate provider-specific inputs
- execute platform operations
- return typed results
- expose metadata
- support cancellation

A Provider MUST NOT:

- bypass policy
- access Runtime internals
- communicate directly with Agents
- modify capability contracts

---

# Registration

Providers register themselves during Runtime initialization or plugin loading.

Registration includes:

- provider ID
- supported capabilities
- platform
- version
- priority

Registration MUST succeed before the Provider becomes available.

---

# Execution Flow

```text
Capability Request
        │
Policy Approved
        │
Provider Selected
        │
Execute
        │
Return Result
        │
Publish Event
```

Providers should perform only platform-specific work.

---

# Provider Metadata

Suggested structure:

```go
type ProviderInfo struct {
    ID           string
    Platform     string
    Version      string
    Priority     int
    Capabilities []string
}
```

Metadata is immutable after registration.

---

# Suggested Interface

```go
type CapabilityProvider interface {
    Info() ProviderInfo
    Supports(name string) bool
    Execute(context.Context, CapabilityRequest) (CapabilityResult, error)
}
```

Providers should avoid exposing implementation-specific APIs.

---

# Provider Selection

If multiple Providers implement the same capability, the Runtime selects one using deterministic rules.

Typical order:

1. Policy restrictions
2. Explicit configuration
3. Platform compatibility
4. Provider priority
5. Version compatibility

Selection MUST be reproducible.

---

# Error Handling

Providers should return structured errors.

Typical categories:

- ValidationError
- TimeoutError
- PermissionError
- PlatformError
- TemporaryError
- PermanentError

Panics MUST be recovered by the Runtime boundary.

---

# Cancellation

Providers receive a Context.

Providers SHOULD:

- monitor cancellation
- stop long-running work
- release resources promptly

Cancellation is cooperative.

---

# Testing

Providers should be tested independently from the Runtime.

Recommended tests:

- capability execution
- invalid input
- timeout
- cancellation
- concurrent execution
- platform edge cases

---

# Distributed Providers

Future Runtime versions may expose Providers remotely.

A Remote Provider must appear identical to a local Provider from the Agent's perspective.

Transport details remain hidden.

---

# Invariants

- Providers implement contracts, not business logic.
- Providers are replaceable.
- Providers never bypass authorization.
- Providers never communicate directly with Agents.
- Runtime owns execution.
- Capability contracts remain stable regardless of implementation.

---

# Related Documents

- 30_CAPABILITY_MODEL.md
- 31_CAPABILITY_REGISTRY.md
- 33_CAPABILITY_SECURITY.md
- 50_POLICY_ENGINE.md

---

# Next

Continue with **33_CAPABILITY_SECURITY.md**.
