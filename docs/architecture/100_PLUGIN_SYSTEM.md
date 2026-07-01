# 100_PLUGIN_SYSTEM.md

> Status: Draft
> Version: 0.1

# Plugin System

The Plugin System enables RouterPilot to be extended without modifying the Runtime.

All optional functionality should be implemented as plugins whenever practical. The Runtime provides stable extension points through the public SDK while remaining independent of plugin implementations.

---

# Goals

The Plugin System MUST provide:

- runtime extensibility
- hot loading (where supported)
- plugin isolation
- version compatibility
- capability discovery
- lifecycle management
- deterministic initialization

The Runtime MUST remain functional without optional plugins.

---

# Architecture

```text
                Runtime
                   │
             Plugin Manager
                   │
    ┌──────────────┼──────────────┐
    ▼              ▼              ▼
Capability     Transport      Memory
 Plugin          Plugin        Plugin
    ▼              ▼              ▼
 Public SDK   Public SDK    Public SDK
```

Plugins interact only through the SDK.

---

# Plugin Categories

Recommended plugin types:

- Capability Provider
- Transport
- Memory Provider
- Policy Provider
- Planner
- Scheduler Extension
- Observability
- Authentication
- Storage

New categories SHOULD be introduced through ADRs.

---

# Lifecycle

```text
Discovered
     │
Validated
     │
Loaded
     │
Initialized
     │
Running
     │
Stopping
     │
Unloaded
```

Each lifecycle transition SHOULD emit an event.

---

# Responsibilities

A plugin MAY:

- register capabilities
- register providers
- publish events
- consume events
- expose metadata

A plugin MUST NOT:

- modify Runtime internals
- bypass authorization
- access private packages
- alter lifecycle state directly

---

# Manifest

Each plugin SHOULD provide a manifest.

Example:

```yaml
id: reticulum-transport
version: 1.0.0
type: transport

sdk: "1.x"

capabilities:
  - transport.reticulum

permissions:
  - transport.send
  - transport.receive
```

The manifest is validated before loading.

---

# Loading

The Plugin Manager performs:

1. Manifest validation
2. Compatibility checks
3. Dependency resolution
4. Registration
5. Initialization
6. Health verification

A failed plugin MUST NOT prevent unrelated plugins from loading.

---

# Isolation

Plugins are isolated from:

- Runtime internals
- other plugins' private state
- transport implementations
- provider implementations

Communication occurs through SDK interfaces and Event Bus APIs.

---

# Version Compatibility

Plugins declare the SDK version they support.

Compatibility SHOULD follow Semantic Versioning.

Major SDK incompatibility prevents loading.

---

# Suggested Interfaces

```go
type Plugin interface {
    Metadata() PluginMetadata
    Initialize(RuntimeContext) error
    Shutdown(context.Context) error
}

type PluginMetadata struct {
    ID      string
    Version string
    Type    string
}
```

---

# Observability

The Plugin Manager SHOULD emit:

- plugin.discovered
- plugin.loaded
- plugin.initialized
- plugin.failed
- plugin.unloaded

Metrics SHOULD include:

- loaded plugins
- failed plugins
- initialization time
- plugin health

---

# Security

Plugins execute with explicit permissions.

Plugin actions are subject to:

- Policy Engine
- Capability authorization
- Runtime lifecycle rules

Plugins MUST NOT elevate privileges.

---

# Invariants

- Runtime depends only on SDK interfaces.
- Plugins are replaceable.
- Loading is deterministic.
- Plugin failures are isolated.
- Public SDK remains the only supported extension mechanism.

---

# Related Documents

- 101_PLUGIN_SDK.md
- 50_POLICY_ENGINE.md
- 120_SECURITY.md

---

# Next

Continue with **101_PLUGIN_SDK.md**.
