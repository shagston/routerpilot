# RouterPilot Architecture

> **Status:** Draft  
> **Version:** 0.1  
> **Applies to:** RouterPilot Runtime vNext  
> **Last Updated:** 2026-07-02

---

# Purpose

This directory contains the canonical architecture specification for RouterPilot.

Unlike implementation documentation, these documents describe what the system is, why it exists, and how every subsystem must behave.

Every implementation should conform to this specification.

When implementation and documentation diverge, implementation should be updated to match the specification or the specification should be revised through an Architecture Decision Record (ADR).

---

# Vision

RouterPilot is a **Distributed Edge Agent Runtime**.

RouterPilot is **not**:

- an OpenWrt application
- a collection of shell scripts
- an automation framework
- an AI agent
- an orchestration server

RouterPilot **is**:

- a runtime for autonomous agents
- an execution platform
- a capability-oriented operating environment
- an event-driven runtime
- a distributed execution platform
- transport-independent
- secure by default

OpenWrt is only one supported platform.

Linux, BSD, embedded devices, Raspberry Pi, servers and future operating systems are equal deployment targets.

---

# Documentation Principles

## Architecture First

Architecture defines implementation.

## Runtime First

The Runtime is the center of the system.

## Capability First

Agents express intent through capabilities rather than platform-specific commands.

Examples:

- filesystem.read
- filesystem.write
- network.scan
- service.restart
- wifi.connect

## Event First

Subsystems communicate through events.

## Plugin First

Every replaceable subsystem exposes interfaces.

## Documentation Driven Development

Every architectural change requires:

- Architecture update
- ADR
- Examples
- Tests

---

# Reading Order

## Foundations

- 01_VISION.md
- 02_PHILOSOPHY.md
- 03_TERMINOLOGY.md
- 04_ARCHITECTURAL_PRINCIPLES.md

## Runtime

- 10_RUNTIME.md
- 11_RUNTIME_LIFECYCLE.md
- 12_RUNTIME_CONTEXT.md
- 13_RUNTIME_EXECUTION.md

## Agents

- 20_AGENT_MODEL.md
- 21_AGENT_LIFECYCLE.md
- 22_AGENT_API.md
- 23_AGENT_REGISTRY.md

## Capabilities

- 30_CAPABILITY_MODEL.md
- 31_CAPABILITY_REGISTRY.md
- 32_CAPABILITY_PROVIDER.md
- 33_CAPABILITY_SECURITY.md

## Events

- 40_EVENT_BUS.md
- 41_EVENT_PROTOCOL.md
- 42_EVENT_DELIVERY.md

## Policy

- 50_POLICY_ENGINE.md
- 51_PERMISSION_MODEL.md

## Memory

- 60_MEMORY.md
- 61_MEMORY_PROVIDER.md

## Scheduler

- 70_SCHEDULER.md

## Transport

- 80_TRANSPORT.md
- 81_RETICULUM.md
- 82_TRANSPORT_PROTOCOL.md

## Distributed Runtime

- 90_DISTRIBUTED_RUNTIME.md
- 91_MESH_DISCOVERY.md

## Plugins

- 100_PLUGIN_SYSTEM.md
- 101_PLUGIN_SDK.md

## Planner

- 110_PLANNER.md

## Security

- 120_SECURITY.md

## Repository

- 130_REPOSITORY.md
- 140_ROADMAP.md

---

# Architecture Relationships

```text
Vision
  │
Architectural Principles
  │
Runtime
  ├── Agents
  ├── Capabilities
  ├── Events
  ├── Policy
  ├── Memory
  ├── Scheduler
  ├── Transport
  └── Plugins
```

---

# ADR Index

- ADR-0001 Vision
- ADR-0002 Runtime Architecture
- ADR-0003 Capability Model
- ADR-0004 Event Model
- ADR-0005 Agent Lifecycle
- ADR-0006 Policy Engine
- ADR-0007 Transport Layer
- ADR-0008 Plugin Architecture
- ADR-0009 Distributed Runtime
- ADR-0010 Versioning Strategy

---

# Audience

This specification targets:

- Runtime developers
- SDK maintainers
- Plugin authors
- AI coding agents
- Contributors
- Security reviewers

---

# Scope

The architecture specification defines:

- architecture
- contracts
- interfaces
- responsibilities
- invariants
- subsystem interactions

Implementation details belong in subsystem documents.

---

# Next

Continue with **01_VISION.md**.
