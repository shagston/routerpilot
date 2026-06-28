# SDK.md

> RouterPilot Software Development Kit
>
> Version: 0.1
>
> Status: Draft

---

# Overview

The RouterPilot SDK defines the public interfaces used to extend the runtime.

The SDK is intentionally small.

Its purpose is not to expose internal implementation details but to provide stable contracts between the Runtime and external components.

Anything outside the SDK should be considered an implementation detail.

---

# Goals

The SDK should be:

* stable
* type-safe
* deterministic
* backwards compatible
* platform independent
* easy to test

---

# Core Philosophy

RouterPilot follows an Interface First design.

The Runtime depends only on interfaces.

Concrete implementations are injected during startup.

Example

```text
Runtime

↓

Tool Interface

↓

OpenWrt Implementation
```

This allows multiple implementations of the same capability.

---

# SDK Modules

The SDK consists of several packages.

```text
sdk/

tool/

runtime/

planner/

memory/

validator/

events/

telemetry/

plugin/

types/
```

Each package exposes only public contracts.

---

# Core Types

Several types are shared across the entire system.

Examples

```text
Execution

Plan

Task

Context

ToolCall

ToolResult

Event

Capability

Permission
```

These types must remain stable.

---

# Tool Interface

Every Tool implements the Tool interface.

Responsibilities

* expose metadata
* validate input
* execute
* describe capabilities

Pseudo-interface

```go
type Tool interface {
    Metadata() Metadata
    Validate(Input) error
    Execute(Context, Input) (Output, error)
}
```

The Runtime interacts only with this interface.

---

# Metadata

Every Tool provides immutable metadata.

```go
type Metadata struct {
    ID
    Version
    Category
    Description
    Permissions
    Capabilities
    Timeout
}
```

Metadata is available without executing the Tool.

---

# Validator Interface

Validation is separated from execution.

```go
type Validator interface {
    Validate(Context) error
}
```

Validators may be attached to:

* Tools
* Plans
* Runtime
* Plugins

---

# Executor Interface

Execution is abstracted through Executors.

```go
type Executor interface {
    Execute(Task) Result
}
```

Possible implementations

* Local
* OpenWrt
* Docker
* Remote SSH
* Future Cluster Executor

---

# Planner Interface

The Runtime never depends on a specific Planner.

```go
type Planner interface {
    Plan(Intent, Context) Plan
}
```

Possible planners

* OpenAI
* Local LLM
* Rule-based
* Hybrid

---

# Memory Provider

Memory is abstracted behind an interface.

```go
type MemoryProvider interface {
    Read(...)
    Write(...)
    Search(...)
}
```

Possible implementations

* SQLite
* BoltDB
* Redis
* PostgreSQL
* In-memory

---

# Context Provider

Context generation is modular.

```go
type ContextProvider interface {
    Build(Intent) Context
}
```

Different providers may contribute independent pieces of context.

---

# Event Publisher

The Runtime never writes directly to consumers.

Instead

```go
type EventPublisher interface {
    Publish(Event)
}
```

Consumers subscribe independently.

---

# Telemetry Interface

Telemetry exporters are optional.

```go
type TelemetryExporter interface {
    Export(Event)
}
```

Possible targets

* OpenTelemetry
* Prometheus
* JSON
* Files

---

# Plugin Interface

Plugins expose capabilities.

```go
type Plugin interface {
    Name()
    Version()
    Register(Registry)
}
```

Registration occurs during startup.

---

# Registry

The Runtime maintains registries.

Examples

```text
Tool Registry

Planner Registry

Validator Registry

Memory Registry

Executor Registry

Plugin Registry
```

Registries support discovery rather than hardcoded wiring.

---

# Dependency Injection

RouterPilot avoids global state.

Components receive dependencies explicitly.

```text
Runtime

↓

Executor

↓

Registry

↓

Tool
```

This simplifies testing.

---

# Errors

Public SDK errors are typed.

Examples

```text
ErrInvalidInput

ErrTimeout

ErrPermissionDenied

ErrCapabilityMissing

ErrExecutionFailed

ErrCancelled
```

Applications should not depend on implementation-specific errors.

---

# Context Object

Every Tool receives a Runtime Context.

The Context includes:

* execution ID
* cancellation signal
* logger
* configuration
* metadata
* event publisher

The Context does not expose internal Runtime state.

---

# Configuration

Components receive immutable configuration objects.

Configuration should never be read from global variables.

---

# Thread Safety

SDK implementations should be safe for concurrent execution unless explicitly documented otherwise.

Shared mutable state should be avoided.

---

# Compatibility

The SDK follows semantic versioning.

Minor releases may add interfaces.

Major releases may change contracts.

Existing interfaces should remain stable whenever possible.

---

# Testing Guidelines

SDK components should support:

* unit testing
* mock implementations
* deterministic behavior
* dependency injection
* isolated execution

Every public interface should be mockable.

---

# Extension Model

Third-party developers should be able to add:

* Tools
* Executors
* Validators
* Memory Providers
* Context Providers
* Event Consumers
* Telemetry Exporters

without modifying the Runtime.

---

# Design Rules

Public interfaces should:

* be small
* expose behavior rather than implementation
* avoid cyclic dependencies
* avoid global state
* remain platform neutral

If an interface grows beyond a single responsibility, it should be split.

---

# Package Dependency Rules

The dependency graph should remain acyclic.

```text
Application
      │
      ▼
Runtime
      │
      ▼
SDK
      │
      ▼
Implementations
```

Implementation packages must never be imported by the SDK.

---

# Future Evolution

Planned SDK additions include:

* capability negotiation
* remote execution protocol
* distributed runtime support
* plugin signing
* version compatibility checks
* generated client libraries

---

# SDK Philosophy

The SDK defines the language spoken by every component of RouterPilot.

The Runtime executes it.

Tools implement it.

Plugins extend it.

Applications depend on it.

By keeping the SDK minimal, stable, and interface-driven, RouterPilot remains extensible without coupling extensions to internal implementation details.