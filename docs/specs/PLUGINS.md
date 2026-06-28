# PLUGINS.md

> RouterPilot Plugin System
>
> Version: 0.1
>
> Status: Draft

---

# Overview

The Plugin System allows RouterPilot to extend its functionality without modifying the Runtime.

Plugins are first-class citizens of the architecture.

They may contribute capabilities, but they never modify the behavior of the Runtime itself.

The Runtime provides the platform.

Plugins provide functionality.

---

# Design Goals

The Plugin System should be:

* modular
* deterministic
* versioned
* discoverable
* sandbox-friendly
* independently testable

Plugins should be optional.

The Runtime must continue operating if a non-essential plugin is unavailable.

---

# Philosophy

RouterPilot follows an extension-over-modification model.

Instead of changing the Runtime:

```text id="0n0r6n"
Runtime

↓

Plugin

↓

Capability
```

New functionality should almost always be implemented as a Plugin.

---

# Plugin Responsibilities

A Plugin may contribute:

* Tools
* Context Providers
* Memory Providers
* Validators
* Policies
* Event Subscribers
* Telemetry Exporters

Plugins should avoid implementing orchestration logic.

---

# What Plugins Must Not Do

Plugins must never:

* modify Runtime internals
* bypass Safety validation
* replace the Planner
* intercept Tool execution
* alter Event ordering
* mutate Registry entries after startup

The Runtime remains authoritative.

---

# Plugin Lifecycle

```text id="c0u3zm"
Discover

↓

Load

↓

Validate

↓

Register

↓

Initialize

↓

Ready
```

Shutdown follows the reverse order.

---

# Plugin Manifest

Every Plugin provides a manifest.

Example

```yaml id="0l4u6t"
id: openwrt

version: 1.0.0

name: OpenWrt Support

api_version: 1

author: RouterPilot

description: Native OpenWrt integration
```

The manifest is read before initialization.

---

# Plugin Interface

Every Plugin implements the SDK contract.

```go id="z4q2y9"
type Plugin interface {
    Metadata() Metadata

    Register(*Registry) error

    Initialize(Context) error

    Shutdown(context.Context) error
}
```

The Runtime depends only on this interface.

---

# Registration

Plugins register themselves through the Registry.

Typical flow

```text id="ikbkhh"
Plugin

↓

Registry

↓

Tools

↓

Validators

↓

Providers
```

The Runtime does not know plugin implementation details.

---

# Capabilities

Capabilities are declared explicitly.

Example

```yaml id="bzx98e"
capabilities:

- ubus

- uci

- firewall4
```

Capabilities are immutable after registration.

---

# Dependency Declaration

Plugins may depend on other plugins.

Example

```yaml id="4a5s8j"
requires:

- core

- openwrt
```

Dependency resolution occurs before initialization.

Circular dependencies are rejected.

---

# Version Compatibility

Plugins declare supported API versions.

Example

```yaml id="q8t2cf"
api:

minimum: 1

maximum: 2
```

The Runtime validates compatibility during startup.

---

# Configuration

Plugins receive configuration through immutable objects.

Example

```yaml id="4u0h3m"
plugin:

cache: true

timeout: 30s
```

Plugins should not read global configuration directly.

---

# Event Integration

Plugins subscribe to Events through the Event Bus.

Example

```text id="l3d4mz"
tool.completed

↓

Plugin

↓

Telemetry Exporter
```

Plugins should not communicate directly with Runtime components.

---

# Context Integration

Plugins may contribute Context Providers.

Example

```text id="e8d8jg"
WireGuard Plugin

↓

WireGuard Context

↓

Planner
```

The Context Engine merges plugin fragments automatically.

---

# Tool Contribution

Plugins commonly contribute Tools.

Example

```text id="fj9jdu"
OpenWrt Plugin

↓

wifi.scan

↓

uci.commit

↓

ubus.call
```

The Runtime treats plugin Tools identically to built-in Tools.

---

# Memory Integration

Plugins may persist their own structured data through the Memory Provider interface.

They should avoid maintaining independent databases unless absolutely necessary.

---

# Error Handling

Plugin failures are isolated.

Example

```text id="fgm6u1"
Plugin Failure

↓

Disable Plugin

↓

Continue Runtime
```

The failure of one optional plugin should not stop the entire Runtime.

---

# Security

Plugins execute under the same Safety Model as built-in components.

Plugins cannot bypass:

* validation
* permission checks
* policy evaluation
* audit logging

All Tool executions remain subject to Runtime enforcement.

---

# Startup Sequence

```text id="x2f4dr"
Load Plugin

↓

Validate Manifest

↓

Resolve Dependencies

↓

Register Components

↓

Initialize

↓

Ready
```

Initialization order must be deterministic.

---

# Shutdown Sequence

```text id="z5v2pa"
Stop New Work

↓

Flush Events

↓

Release Resources

↓

Shutdown Plugin
```

Shutdown should be graceful and idempotent.

---

# Testing

Every Plugin should provide:

* registration tests
* initialization tests
* dependency tests
* integration tests
* failure isolation tests

Plugins should be testable without requiring a full Runtime.

---

# Distribution

Future distribution models may include:

* local plugin directories
* signed plugin bundles
* remote repositories
* package managers

The loading mechanism should remain independent of the plugin contract.

---

# Future Evolution

Possible enhancements include:

* plugin sandboxing
* hot-reload in development mode
* capability negotiation
* digital signatures
* remote plugins
* marketplace integration

These additions should preserve the existing Plugin API.

---

# Plugin Philosophy

Plugins extend RouterPilot without changing its core.

The Runtime defines the execution model.

The SDK defines the contracts.

The Registry discovers capabilities.

Plugins provide implementations.

By keeping extensions isolated behind stable interfaces, RouterPilot remains modular, maintainable, and adaptable to new networking platforms without increasing the complexity of the core runtime.
