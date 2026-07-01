# 101_PLUGIN_SDK.md

> Status: Draft
> Version: 0.1

# Plugin SDK

The Plugin SDK defines the stable public programming interface for extending the RouterPilot Runtime.

Plugin authors MUST depend exclusively on the SDK. They MUST NOT import Runtime internal packages.

The SDK is the long-term compatibility contract between the Runtime and third-party extensions.

---

# Design Goals

The SDK MUST provide:

- stable APIs
- semantic versioning
- transport independence
- runtime abstraction
- deterministic behavior
- backward compatibility whenever practical

The SDK MUST remain significantly more stable than the Runtime implementation.

---

# Architecture

```text
          Plugin
             │
      RouterPilot SDK
             │
      Runtime Interfaces
             │
      Runtime Internals
```

The dependency direction is strictly downward.

Plugins never depend on Runtime internals.

---

# SDK Modules

Recommended packages:

- sdk/agent
- sdk/capability
- sdk/events
- sdk/memory
- sdk/policy
- sdk/runtime
- sdk/scheduler
- sdk/transport
- sdk/plugin
- sdk/logging
- sdk/metrics

Each module SHOULD have a single responsibility.

---

# Public Interfaces

The SDK SHOULD expose interfaces rather than concrete implementations.

Examples:

```go
type Runtime interface {}

type EventBus interface {}

type CapabilityRegistry interface {}

type Memory interface {}

type Transport interface {}
```

Implementations remain internal to the Runtime.

---

# Plugin Context

Every plugin receives a Runtime Context during initialization.

The context SHOULD expose:

- logger
- event bus
- capability registry
- memory API
- metrics
- configuration
- runtime metadata

The context MUST NOT expose mutable Runtime internals.

---

# Registration

Plugins register themselves through SDK APIs.

Typical registration sequence:

```text
Plugin Start
      │
SDK Registration
      │
Runtime Validation
      │
Plugin Available
```

Registration failures prevent activation.

---

# Compatibility

The SDK follows Semantic Versioning.

Compatibility rules:

- Patch versions are fully compatible.
- Minor versions add functionality without breaking existing plugins.
- Major versions may remove or change APIs.

Plugins declare the SDK version they support.

---

# Error Handling

The SDK SHOULD define typed errors.

Recommended categories:

- ValidationError
- CompatibilityError
- InitializationError
- RegistrationError
- PermissionError

Errors should be deterministic and machine-readable.

---

# Testing

The SDK SHOULD provide test utilities including:

- mock runtime
- mock event bus
- mock memory
- mock capability registry
- fake contexts

Plugins should be testable without a running Runtime.

---

# Documentation

Every exported SDK symbol SHOULD include:

- purpose
- lifecycle expectations
- concurrency guarantees
- compatibility notes
- usage examples

The SDK documentation is part of the public contract.

---

# Suggested Interfaces

```go
type SDK interface {
    Runtime() Runtime
    Events() EventBus
    Memory() Memory
    Capabilities() CapabilityRegistry
}
```

Concrete implementations remain hidden.

---

# Security

SDK consumers are subject to:

- Policy Engine decisions
- permission checks
- capability authorization

The SDK MUST NOT expose APIs capable of bypassing Runtime security.

---

# Invariants

- Plugins depend only on the SDK.
- SDK interfaces are stable.
- Runtime internals remain private.
- Compatibility follows Semantic Versioning.
- SDK abstractions are transport-independent.

---

# Related Documents

- 100_PLUGIN_SYSTEM.md
- 120_SECURITY.md
- 130_REPOSITORY.md

---

# Next

Continue with **110_PLANNER.md**.
