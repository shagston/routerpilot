# REGISTRY.md

> RouterPilot Registry
>
> Version: 0.1
>
> Status: Draft

---

# Overview

The Registry is the central discovery and dependency resolution mechanism of RouterPilot.

Its primary responsibility is to maintain a catalog of runtime components and provide deterministic lookup without introducing compile-time coupling.

The Registry never executes logic.

It only stores, validates, and resolves component registrations.

---

# Design Goals

The Registry should be:

* deterministic
* type-safe
* immutable after initialization
* thread-safe
* extensible
* platform independent

---

# Philosophy

RouterPilot follows a **registration-based architecture**.

Components announce themselves during startup.

The Runtime discovers capabilities through the Registry rather than through hardcoded dependencies.

```text
Plugin

↓

Registry

↓

Runtime

↓

Planner

↓

Execution
```

---

# Registered Components

The Registry maintains independent collections.

```text
Tools

Executors

Validators

Context Providers

Memory Providers

Plugins

Policies

Capabilities
```

Each category has its own namespace.

---

# Registration Lifecycle

```text
Component

↓

Validate

↓

Register

↓

Freeze

↓

Runtime Starts
```

Once the Runtime starts, registrations become read-only.

---

# Tool Registry

Each Tool registers:

```yaml
id:

version:

category:

metadata:

implementation:
```

IDs must be globally unique.

---

# Executor Registry

Executors provide platform-specific implementations.

Example

```text
network.status

↓

Linux Executor

↓

OpenWrt Executor

↓

Container Executor
```

The Runtime resolves the correct Executor at execution time.

---

# Validator Registry

Validators are attached to:

* Plans
* Tools
* Runtime
* Plugins

Validators are executed in deterministic order.

---

# Context Provider Registry

Context Providers contribute fragments.

Example

```text
DNS Provider

↓

Network Provider

↓

Firewall Provider

↓

Merged Context
```

Providers never communicate directly.

---

# Memory Provider Registry

Multiple storage backends may coexist.

Example

```text
SQLite

BoltDB

Redis

In-Memory
```

The active provider is selected through configuration.

---

# Capability Registry

Capabilities describe platform features.

Examples

```text
dnsmasq

firewall4

wireguard

uci

ubus
```

Capabilities are discovered during startup.

---

# Plugin Registry

Plugins register through a single entry point.

```go
type Plugin interface {
    Register(*Registry) error
}
```

Plugins may contribute:

* Tools
* Validators
* Context Providers
* Event Subscribers
* Policies

---

# Dependency Resolution

Components never resolve dependencies directly.

Instead:

```text
Runtime

↓

Registry

↓

Component
```

This avoids cyclic dependencies.

---

# Namespaces

Every component belongs to a namespace.

Example

```text
tool/network.status

tool/dns.lookup

validator/schema

memory/sqlite

plugin/openwrt
```

Namespaces prevent collisions.

---

# Version Resolution

Multiple versions may coexist.

Example

```text
network.status@1

network.status@2
```

Resolution strategy is deterministic.

---

# Duplicate Registration

Duplicate IDs are rejected during startup.

Example

```text
network.status

↓

Already Registered

↓

Startup Error
```

The Runtime should never silently replace components.

---

# Registry Freeze

After initialization:

```text
Mutable

↓

Freeze

↓

Read Only
```

Read-only registries eliminate synchronization complexity during execution.

---

# Lookup API

Typical operations:

```go
FindTool(id)

FindExecutor(id)

FindValidator(id)

FindCapability(id)

FindPlugin(id)
```

Lookups must be O(1) whenever practical.

---

# Events

Registry changes emit Events during startup.

Examples

```text
registry.initialized

tool.registered

plugin.loaded

capability.discovered
```

After freeze, no further registration events should occur.

---

# Thread Safety

Registration is single-threaded.

Lookup is concurrent.

This allows lock-free reads after initialization.

---

# Startup Sequence

```text
Load Configuration

↓

Discover Plugins

↓

Register Components

↓

Validate Registry

↓

Freeze Registry

↓

Start Runtime
```

The Runtime must never start with an invalid Registry.

---

# Testing

The Registry should support:

* duplicate detection
* namespace validation
* version resolution
* concurrent lookups
* plugin registration
* freeze behavior

---

# Future Evolution

Possible enhancements include:

* lazy component loading
* remote registries
* signed component manifests
* dependency graphs
* hot-reload for development mode

Production mode should remain immutable after startup.

---

# Registry Philosophy

The Registry is the directory of RouterPilot.

It knows **what exists**, but never decides **what should happen**.

By centralizing discovery while keeping execution separate, the Registry enables a modular architecture where new capabilities can be added without modifying the Runtime or Planner.
