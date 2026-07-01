# 130_REPOSITORY.md

> Status: Draft
> Version: 0.1

# Repository Architecture

This document defines the canonical repository layout for RouterPilot.

The repository structure is an architectural contract. It exists to separate stable public APIs from internal implementation details, simplify navigation, and enable long-term maintainability.

---

# Design Goals

The repository MUST:

- separate public and internal APIs
- isolate platform-specific code
- support plugins
- support multiple runtimes
- remain understandable by humans and AI agents
- scale without excessive coupling

---

# High-Level Layout

```text
routerpilot/
├── cmd/
├── internal/
├── pkg/
├── sdk/
├── plugins/
├── transports/
├── docs/
├── examples/
├── test/
├── scripts/
└── configs/
```

---

# Directory Responsibilities

## cmd/

Application entry points.

Examples:

- routerpilot
- routerpilotd
- rpctl

No business logic belongs here.

---

## internal/

Private Runtime implementation.

Contains:

- runtime
- scheduler
- policy
- eventbus
- lifecycle
- registries
- execution engine

Packages under `internal/` MUST NOT be imported by plugins.

---

## pkg/

Reusable packages that are stable enough for internal reuse but are not part of the public SDK.

Examples:

- logging
- metrics
- serialization
- utilities

---

## sdk/

Public compatibility layer.

Contains interfaces for:

- agents
- capabilities
- memory
- transport
- plugins
- runtime
- events

Third-party code SHOULD depend only on this directory.

---

## plugins/

Reference plugin implementations.

Examples:

- reticulum
- sqlite-memory
- rule-planner
- filesystem-provider

Plugins demonstrate SDK usage.

---

## transports/

Transport-specific adapters.

Examples:

- reticulum
- mqtt
- http

Concrete implementations remain isolated from the Runtime.

---

## docs/

Architecture, ADRs, specifications and guides.

Suggested structure:

```text
docs/
  architecture/
  adr/
  sdk/
  runtime/
```

Documentation is considered part of the codebase.

---

## examples/

Small, self-contained SDK examples.

Examples SHOULD compile and remain synchronized with released SDK versions.

---

## test/

Integration, compatibility and conformance tests.

Recommended structure:

- integration/
- e2e/
- benchmarks/
- fixtures/

---

## scripts/

Automation utilities for development, CI and release engineering.

Scripts MUST NOT contain business logic.

---

## configs/

Reference configuration files.

Configuration examples SHOULD be production-oriented.

---

# Dependency Rules

Dependency direction is strictly enforced.

```text
Applications
      │
      ▼
SDK
      │
      ▼
Runtime
      ▲
      │
Plugins
```

Internal packages MUST NOT depend on plugins.

Plugins MUST NOT depend on internal packages.

---

# Package Naming

Preferred:

- runtime
- scheduler
- transport
- capability
- policy
- registry
- plugin

Avoid:

- misc
- common
- helpers
- utils (unless genuinely generic)

---

# Versioning

Repository versioning follows Semantic Versioning.

Architecture documents SHOULD be versioned alongside Runtime releases.

Breaking changes require:

- ADR
- migration guide
- release notes

---

# Coding Standards

The project SHOULD follow:

- small packages
- explicit interfaces
- dependency injection
- immutable contracts
- comprehensive tests
- documentation-first development

---

# CI Requirements

Continuous Integration SHOULD validate:

- formatting
- linting
- unit tests
- integration tests
- architecture conformance
- SDK compatibility

---

# Invariants

- Public SDK remains isolated.
- Runtime internals remain private.
- Plugins depend only on SDK packages.
- Documentation evolves with code.
- Repository structure reflects architectural boundaries.

---

# Related Documents

- 100_PLUGIN_SYSTEM.md
- 101_PLUGIN_SDK.md
- 120_SECURITY.md
- 140_ROADMAP.md

---

# Next

Continue with **140_ROADMAP.md**.
