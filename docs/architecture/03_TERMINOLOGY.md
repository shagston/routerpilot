# 03_TERMINOLOGY.md

> Status: Draft
> Version: 0.1

# Terminology

This document defines the canonical vocabulary used throughout the RouterPilot architecture.
Every specification, ADR, SDK and implementation MUST use these terms consistently.

---

# Core Concepts

## Runtime

The Runtime is the execution environment responsible for loading, coordinating and supervising all RouterPilot subsystems.

Responsibilities:

- lifecycle management
- capability execution
- event routing
- plugin loading
- scheduler integration
- policy enforcement
- transport integration

The Runtime is the architectural center of RouterPilot.

---

## Agent

An autonomous software component executed by the Runtime.

An Agent:

- observes
- plans (directly or indirectly)
- requests capabilities
- publishes events
- maintains state

Agents never bypass the Runtime.

---

## Planner

A component that transforms goals into executable plans.

Possible implementations:

- Rule Planner
- Workflow Planner
- LLM Planner
- Remote Planner

Planners never execute work directly.

---

## Capability

An abstract operation that represents *what* should happen instead of *how*.

Examples:

- filesystem.read
- network.scan
- service.restart

Capabilities are resolved by providers.

---

## Capability Provider

A concrete implementation of one or more capabilities.

Examples:

- Linux provider
- OpenWrt provider
- Windows provider
- Mock provider

---

## Registry

A discoverable catalog of runtime resources.

Examples:

- Agent Registry
- Capability Registry
- Plugin Registry

Registries own metadata, not execution.

---

## Event

An immutable message representing something that happened.

Examples:

- agent.started
- network.changed
- capability.completed

Events are transported through the Event Bus.

---

## Event Bus

The internal messaging system connecting runtime components.

The Event Bus provides:

- publish
- subscribe
- routing
- retries
- observability

---

## Policy

A rule determining whether an action is permitted.

Policies evaluate:

- identity
- capability
- scope
- context
- conditions

---

## Transport

A communication layer between RouterPilot nodes.

Examples:

- Reticulum
- MQTT
- HTTP
- NATS

The Runtime interacts only with the Transport interface.

---

## Node

A single RouterPilot runtime instance participating in a distributed system.

Every node may:

- execute agents
- expose capabilities
- consume capabilities
- publish events

---

## Mesh

A collection of RouterPilot nodes connected through one or more transports without requiring a central coordinator.

---

## Envelope

A transport-independent container carrying messages between nodes.

Typical fields:

- ID
- Source
- Destination
- Type
- Payload
- Timestamp
- Signature
- TTL

---

## Scheduler

Subsystem responsible for triggering work according to time or events.

---

## Memory

A subsystem providing contextual state for agents.

Memory layers:

- Working
- Session
- Persistent

---

## Plugin

An independently developed extension loaded through the public SDK.

Plugins extend behavior without modifying the Runtime.

---

## SDK

The public development kit used by plugin authors and application developers.

The SDK defines stable interfaces.

---

## ADR

Architecture Decision Record.

Every significant architectural change MUST be documented as an ADR.

---

# Naming Rules

Preferred terminology:

- Runtime
- Agent
- Capability
- Provider
- Registry
- Event
- Policy
- Transport
- Plugin

Avoid ambiguous terms such as:

- module
- service (unless referring to a distributed service)
- helper
- utility

---

# Glossary Relationships

```
Planner
    │
creates Plan
    │
Runtime
    │
executes Agent
    │
requests Capability
    │
resolved by Provider
    │
authorized by Policy
    │
publishes Event
    │
delivered by Event Bus
    │
transported via Transport
```

---

# Related Documents

- 01_VISION.md
- 02_PHILOSOPHY.md
- 04_ARCHITECTURAL_PRINCIPLES.md

---

# Next

Continue with **04_ARCHITECTURAL_PRINCIPLES.md**.
