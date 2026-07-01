# 31_CAPABILITY_REGISTRY.md

> Status: Draft
> Version: 0.1

# Capability Registry

The Capability Registry is the authoritative catalog of all capabilities available to a Runtime.

It maps abstract capability contracts to one or more providers while remaining independent of operating systems, transports and planners.

The Registry owns **discovery**, **registration** and **resolution metadata**. It never executes capabilities.

---

# Objectives

The Capability Registry MUST provide:

- capability discovery
- provider registration
- version management
- capability lookup
- provider selection metadata
- runtime introspection

The Registry MUST NOT:

- authorize requests
- execute providers
- perform scheduling
- contain platform-specific logic

---

# Architecture

```text
               Runtime
                  │
                  ▼
        Capability Registry
          │             │
          ▼             ▼
 Capability Metadata  Provider Metadata
          │             │
          └──────┬──────┘
                 ▼
          Resolution Engine
                 │
                 ▼
          Selected Provider
```

---

# Registration Model

Every provider registers one or more capability descriptors.

Registration occurs during Runtime initialization or plugin loading.

Registration sequence:

```text
Load Plugin
      │
Validate Descriptor
      │
Register Capability
      │
Register Provider
      │
Publish capability.registered
```

Duplicate registrations SHOULD be rejected unless explicitly versioned.

---

# Capability Descriptor

A descriptor defines the public contract.

Suggested structure:

```go
type CapabilityDescriptor struct {
    Name        string
    Version     string
    Category    string
    Description string
    InputType   reflect.Type
    OutputType  reflect.Type
}
```

Descriptors are immutable after registration.

---

# Provider Descriptor

Providers describe implementation-specific information.

Suggested fields:

```go
type ProviderDescriptor struct {
    ID         string
    Capability string
    Platform   string
    Priority   int
    Version    string
}
```

Multiple providers may implement the same capability.

---

# Resolution

Resolution determines which provider should satisfy a request.

Selection criteria may include:

1. capability name
2. capability version
3. platform
4. provider priority
5. policy restrictions
6. runtime configuration

Resolution MUST be deterministic.

---

# Lookup Operations

Recommended operations:

- Get(name)
- Exists(name)
- List()
- ListByCategory()
- Providers(name)
- Resolve(request)

Lookups are read-only.

---

# Suggested Interface

```go
type CapabilityRegistry interface {
    Register(CapabilityDescriptor) error
    RegisterProvider(ProviderDescriptor) error
    Resolve(CapabilityRequest) (ProviderDescriptor, error)
    Get(string) (CapabilityDescriptor, bool)
    List() []CapabilityDescriptor
}
```

---

# Events

The Registry SHOULD emit:

- capability.registered
- capability.updated
- capability.removed
- provider.registered
- provider.removed

These events enable dynamic plugin loading.

---

# Thread Safety

The Registry MUST support concurrent reads.

Registration and removal operations MUST be synchronized.

Lookups SHOULD be lock-efficient because they occur on every execution path.

---

# Distributed Considerations

Each Runtime owns its local registry.

Distributed runtimes MAY expose registry summaries to peers.

Remote registry synchronization MUST exchange metadata only.

Providers remain local execution units unless explicitly exported.

---

# Invariants

- Capability contracts are immutable.
- Provider metadata never replaces capability metadata.
- Runtime performs execution.
- Registry performs lookup only.
- Resolution is deterministic.
- Authorization occurs after resolution metadata is available and before provider execution.

---

# Related Documents

- 30_CAPABILITY_MODEL.md
- 32_CAPABILITY_PROVIDER.md
- 33_CAPABILITY_SECURITY.md
- 50_POLICY_ENGINE.md

---

# Next

Continue with **32_CAPABILITY_PROVIDER.md**.
