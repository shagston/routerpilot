# RouterPilot

> **AI-native runtime for deterministic network automation.**

RouterPilot is an open-source execution platform that transforms natural-language intent into deterministic, auditable, and safe network operations.

Rather than allowing an LLM to interact directly with the operating system, RouterPilot separates **reasoning** from **execution**.

The Planner decides **what** should happen.

The Runtime decides **how** it happens safely.

---

# Why RouterPilot?

Traditional AI agents often execute commands directly, making behavior difficult to audit, reproduce, and secure.

RouterPilot introduces a layered architecture:

```text
User Request
      │
      ▼
Planner
      │
Execution Plan
      │
      ▼
Runtime
      │
      ▼
Tools
      │
      ▼
Operating System
```

Every operation is:

* deterministic
* validated
* observable
* replayable
* policy-controlled

---

# Design Principles

RouterPilot is built around several architectural principles:

* Runtime owns execution.
* Planning has no side effects.
* Everything executable is a Tool.
* Tools never orchestrate other Tools.
* Events are immutable.
* Memory never replaces live state.
* Platform logic lives outside the Runtime.
* The Core remains intentionally small.

See `docs/ARCHITECTURE.md` and `docs/IMPLEMENTATION_PLAN.md` for the current specification.

---

# Architecture

The project consists of several independent subsystems.

| Component      | Responsibility                                |
| -------------- | --------------------------------------------- |
| Planner        | Transform user intent into executable Plans   |
| Runtime        | Execute validated Plans deterministically     |
| Tools          | Perform individual operations                 |
| Registry       | Discover components and capabilities          |
| Context Engine | Build minimal planning context                |
| Memory         | Store persistent knowledge                    |
| Event Bus      | Publish immutable runtime events              |
| Safety Layer   | Enforce permissions and policies              |
| Plugin System  | Extend RouterPilot without modifying the Core |

---

# Documentation

The project documentation is organized as an architecture handbook.

```text
```text
docs/
  ARCHITECTURE.md
  RUNTIME.md
  PLANNER.md
  TOOLS.md
  SDK.md
  CONTEXT.md
  MEMORY.md
  EVENTS.md
  SAFETY.md
  IMPLEMENTATION_PLAN.md
  specs/
    PLUGINS.md
    REGISTRY.md
    tool.md
```

---

# Current Status

Current stage:

**Core SDK + Runtime vertical slice**

RouterPilot now has a small runnable Go implementation:

* public SDK interfaces and shared types under `sdk/`
* in-memory tool registry under `internal/registry`
* runtime engine with plan validation, dependency ordering, retries, timeouts and event publishing
* deterministic safety validation for schemas, permissions and capabilities
* first real Tool: `network.ping`
* CLI entrypoint: `cmd/routerpilot`

The current executable path is:

```text
CLI
  |
  v
Runtime
  |
  v
Safety Validator
  |
  v
Tool Registry
  |
  v
network.ping
  |
  v
Structured Result + Events
```

Try it locally:

```powershell
go run .\cmd\routerpilot tools
go run .\cmd\routerpilot ping 127.0.0.1 1 --events
```

Run tests:

```powershell
go test ./...
```

Note: on restricted Windows environments, set `GOCACHE` to a writable directory before running Go commands.

---

# Roadmap

High-level milestones:

1. Core SDK
2. Runtime
3. Registry
4. Planner
5. CLI
6. First Tool
7. OpenWrt integration
8. REST API
9. Plugin ecosystem
10. v1.0

See `docs/IMPLEMENTATION_PLAN.md` for details.

---

# Project Goals

RouterPilot aims to provide:

* deterministic execution
* AI-assisted planning
* platform-independent architecture
* OpenWrt-first support
* plugin-based extensibility
* replayable execution
* production-grade observability

---

# Contributing

The architecture is intentionally interface-driven.

Before contributing, please read:

1. `docs/ARCHITECTURE.md`
2. `docs/SDK.md`
3. `docs/RUNTIME.md`
4. `docs/TOOLS.md`

Architectural changes should be accompanied by an ADR.

---

# License

License information will be added before the first public release.
