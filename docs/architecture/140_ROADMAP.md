# 140_ROADMAP.md

> Status: Draft
> Version: 0.1

# RouterPilot Roadmap

This document describes the long-term architectural evolution of RouterPilot.

The roadmap focuses on architectural capabilities rather than release dates. Milestones represent maturity levels of the Runtime and its ecosystem.

---

# Vision

RouterPilot evolves from an OpenWrt-focused management tool into a distributed, capability-oriented Edge Agent Runtime.

The Runtime becomes a portable execution environment capable of running autonomous agents across heterogeneous devices without requiring centralized infrastructure.

---

# Guiding Principles

Every milestone should preserve:

- Runtime-first architecture
- Capability-oriented execution
- Event-driven communication
- Transport independence
- Plugin extensibility
- Security by default
- Offline-first operation

---

# Phase 1 — Runtime Foundation

Objective:

Establish the Runtime as the architectural core.

Deliverables:

- Runtime lifecycle
- Agent model
- Capability system
- Event Bus
- Scheduler
- Policy Engine
- Plugin system
- Public SDK

Success criteria:

- Stable Runtime APIs
- Plugin loading
- Deterministic execution
- Comprehensive architecture documentation

---

# Phase 2 — Distributed Runtime

Objective:

Connect Runtime instances into a decentralized mesh.

Deliverables:

- Transport abstraction
- Reticulum transport
- Mesh discovery
- Remote capability execution
- Distributed events
- Node identities

Success criteria:

- Peer discovery
- Secure remote execution
- Offline operation
- Fault isolation

---

# Phase 3 — Intelligent Runtime

Objective:

Support multiple planning strategies.

Deliverables:

- Rule Planner
- Workflow Planner
- LLM Planner
- Remote Planner

Success criteria:

- Planner interchangeability
- Immutable execution plans
- Runtime validation pipeline

---

# Phase 4 — Production Platform

Objective:

Provide production-grade operational capabilities.

Deliverables:

- Metrics
- Tracing
- Audit logging
- Health monitoring
- Backup and restore
- HA-friendly deployment

Success criteria:

- Operational observability
- Stable upgrades
- Disaster recovery procedures

---

# Phase 5 — Ecosystem

Objective:

Enable third-party development.

Deliverables:

- Plugin marketplace
- SDK documentation
- Reference plugins
- Certification tests
- Compatibility suite

Success criteria:

- Independent plugin ecosystem
- Stable SDK
- Long-term API compatibility

---

# Phase 6 — Edge Runtime OS

Objective:

Position RouterPilot as a universal runtime for distributed edge agents.

Potential targets:

- OpenWrt
- Linux
- BSD
- Containers
- Raspberry Pi
- Embedded devices
- Virtual machines

Long-term possibilities:

- Edge clusters
- Smart home
- Industrial automation
- Robotics
- Autonomous infrastructure

---

# Architectural Milestones

The project is considered architecturally mature when:

- Runtime is platform-independent.
- Every subsystem is replaceable.
- Plugins rely only on the SDK.
- Distributed execution is transparent.
- Capability contracts remain stable.
- Security is enforced by architecture.
- Documentation is the canonical source of truth.

---

# Non-Goals

RouterPilot does not aim to become:

- a Kubernetes replacement
- a cloud orchestration platform
- a configuration management language
- a shell scripting framework
- a monolithic AI assistant

---

# Success Metrics

Examples of measurable outcomes:

- SDK compatibility across major releases
- Zero Runtime dependencies on transport implementations
- Full offline functionality
- Deterministic execution under identical plans
- Independent third-party plugin ecosystem

---

# Future Directions

Potential future research areas:

- Federated policy management
- Distributed memory synchronization
- Capability marketplaces
- Agent migration
- Multi-transport routing
- Sandboxed WebAssembly plugins
- Edge AI inference providers
- Formal verification of execution plans

Each future direction should be evaluated through an ADR before implementation.

---

# Related Documents

- 00_INDEX.md
- 01_VISION.md
- 100_PLUGIN_SYSTEM.md
- 101_PLUGIN_SDK.md
- 120_SECURITY.md
- 130_REPOSITORY.md

---

# Completion

This document concludes Version 0.1 of the RouterPilot Architecture Specification.

Future revisions should evolve through Architecture Decision Records (ADRs) while preserving backward compatibility whenever practical.
