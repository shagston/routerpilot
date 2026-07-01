# RouterPilot Architecture RFC

**Status:** Draft  
**Target Version:** v1.0  
**Audience:** Core maintainers, contributors, AI coding agents (Codex, Claude Code, Gemini CLI)

---

# Table of Contents

1. Vision
2. Project Goals
3. Non-Goals
4. Architectural Principles
5. System Architecture
6. Runtime
7. Agent Model
8. Capability Model
9. Event System
10. Policy Engine
11. Memory
12. Scheduler
13. Transport Layer
14. Distributed Runtime
15. Plugin System
16. Planner
17. SDK
18. API Stability
19. Security Model
20. Repository Layout
21. ADR Index
22. Roadmap
23. Release Strategy
24. Definition of Done
25. AI Development Guidelines

---

# 1. Vision

RouterPilot is a **distributed Edge Agent Runtime**.

It is not an OpenWrt application, nor merely an automation tool.

The runtime should allow autonomous agents to execute locally or across a distributed mesh while exposing abstract capabilities rather than operating-system-specific commands.

Primary transport is **Reticulum**, but the runtime must remain transport-independent.

---

# 2. Goals

- Distributed-first
- Offline-first
- Capability-first
- Event-driven
- Plugin-first
- Secure by default
- AI optional
- Deterministic execution

---

# 3. Non-Goals

- Tight coupling to OpenWrt
- Dependence on cloud infrastructure
- Mandatory LLM execution
- Hardcoded transports
- Direct shell execution as public API

---

# 4. Architectural Principles

1. Runtime owns execution.
2. Agents own behavior.
3. Planner creates plans.
4. Runtime executes plans.
5. Capabilities expose functionality.
6. Policies authorize actions.
7. Events connect components.
8. Transport is replaceable.
9. Plugins extend everything.

---

# 5. High-Level Architecture

```
Runtime
 ├── Agent Runtime
 ├── Planner
 ├── Capability Registry
 ├── Policy Engine
 ├── Event Bus
 ├── Scheduler
 ├── Memory
 ├── Transport Layer
 ├── Plugin Manager
 └── Public API
```

---

# 6. Runtime

Responsibilities:

- lifecycle
- dependency injection
- capability execution
- event routing
- scheduler integration
- transport integration
- plugin loading
- health monitoring

The runtime must never depend directly on any concrete planner.

---

# 7. Agent Model

Every agent has:

- identity
- state
- permissions
- memory
- subscriptions
- capabilities

Lifecycle:

```
Init
 ↓
Start
 ↓
Ready
 ↓
Sleep
 ↓
Wake
 ↓
Stop

Crash
 ↓
Recover
```

---

# 8. Capability Model

Capabilities represent *intent*, not implementation.

Examples:

```
filesystem.read
filesystem.write
network.scan
wifi.connect
service.restart
process.exec
```

Providers implement capabilities.

The runtime invokes providers through the registry.

---

# 9. Event Bus

Requirements:

- publish
- subscribe
- async delivery
- retry
- dead-letter queue
- correlation IDs
- TTL
- priority
- context propagation

No direct agent-to-agent calls.

---

# 10. Policy Engine

All capability execution must pass through policy evaluation.

Support:

- allow
- deny
- conditions
- scopes
- audit log
- YAML policies

---

# 11. Memory

Layers:

- Working Memory
- Session Memory
- Persistent Memory
- Vector Memory Interface

Providers:

- filesystem
- SQLite
- external implementations

---

# 12. Scheduler

Support:

- cron
- interval
- one-shot
- event-triggered
- retries
- pause
- resume
- dependency graph

---

# 13. Transport Layer

Interfaces only:

```
Transport
Endpoint
Envelope
Router
Discovery
```

Runtime must not know transport details.

---

# 14. Distributed Runtime

Supports:

- remote agents
- remote capabilities
- distributed registry
- distributed events
- transparent execution

---

# 15. Plugin System

Plugin types:

- runtime
- capability
- transport
- planner
- policy
- memory

All plugins negotiate versions through SDK metadata.

---

# 16. Planner

Planner interface only.

Possible implementations:

- Rule Planner
- Workflow Planner
- LLM Planner
- Remote Planner

Runtime remains planner-agnostic.

---

# 17. SDK

Stable SDK packages:

```
sdk/runtime
sdk/agent
sdk/plugin
sdk/capability
sdk/events
sdk/memory
sdk/transport
sdk/planner
```

---

# 18. API Stability

Freeze before v1:

- Runtime API
- SDK
- Plugin API
- Capability API
- Transport API

Semantic Versioning required.

---

# 19. Security Model

Security principles:

- least privilege
- capability permissions
- sandbox execution
- audit logging
- panic recovery
- secret isolation
- transport encryption

Required documents:

- SECURITY.md
- THREAT_MODEL.md

---

# 20. Repository Layout

```
cmd/
runtime/
planner/
registry/
sdk/
plugins/
transport/
examples/
docs/
internal/
```

---

# 21. ADR Index

- ADR-0001 Vision
- ADR-0002 Capability-first
- ADR-0003 Transport Independence
- ADR-0004 Agent Lifecycle
- ADR-0005 Plugin Architecture
- ADR-0006 Security
- ADR-0007 Event Bus
- ADR-0008 Memory
- ADR-0009 Distributed Runtime
- ADR-0010 Versioning

---

# 22. Engineering Roadmap

M0 Repository Audit

M1 Runtime Cleanup

M2 Agent Runtime

M3 Capability System

M4 Event Bus

M5 Policy Engine

M6 Transport Interfaces

M7 Reticulum

M8 Distributed Runtime

M9 Memory

M10 Scheduler

M11 Plugin SDK

M12 Planner Interface

M13 Mesh Services

M14 Security Hardening

M15 API Freeze

---

# 23. Release Strategy

```
v0.1 Runtime
v0.2 Agents
v0.3 Capabilities
v0.4 Events
v0.5 Policies
v0.6 Transport
v0.7 Reticulum
v0.8 Distributed Runtime
v0.9 SDK Freeze
v1.0 Stable
```

---

# 24. Definition of Done

Each milestone must:

- compile successfully
- pass all tests
- update documentation
- update examples
- include migration notes
- update ADRs
- avoid unnecessary breaking changes

---

# 25. AI Development Guidelines

For every task:

1. Read existing architecture.
2. Preserve compatibility.
3. Make focused commits.
4. Keep refactoring separate from features.
5. Update documentation.
6. Update tests.
7. Stop after milestone completion.
8. Produce a summary and next-step recommendations.
