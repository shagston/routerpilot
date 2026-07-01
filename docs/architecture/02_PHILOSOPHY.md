# 02_PHILOSOPHY.md

> Status: Draft
> Version: 0.1

# Engineering Philosophy

This document defines the engineering philosophy behind RouterPilot. It explains *why* architectural decisions are made and provides guiding principles for future development.

---

# Core Philosophy

RouterPilot is designed as a runtime rather than an application.

The runtime should provide stable execution semantics while allowing higher-level components (agents, planners, plugins, transports) to evolve independently.

The runtime owns execution.
Everything else is replaceable.

---

# Fundamental Principles

## Runtime over Framework

Applications should depend on the Runtime.

The Runtime should not depend on applications.

---

## Composition over Integration

Subsystems communicate through contracts rather than implementation details.

Preferred:

- interfaces
- events
- capabilities

Avoid:

- package coupling
- shared globals
- hidden dependencies

---

## Capabilities over Commands

Agents describe *intent*.

Example:

```
filesystem.read
```

rather than:

```
cat /etc/config/network
```

The runtime resolves intent into platform-specific implementations.

---

## Events over Direct Calls

Whenever practical, communication should occur through the Event Bus.

Benefits:

- loose coupling
- observability
- replayability
- distributed execution

---

## Policy before Execution

Every capability request must pass authorization.

Execution is impossible without an explicit policy decision.

---

## Deterministic Runtime

Planning may be non-deterministic.

Execution must be deterministic.

Two identical execution plans should produce identical runtime behavior.

---

## Offline First

The runtime must continue operating without Internet connectivity.

Networking enhances the runtime but must never be a prerequisite.

---

## Distributed by Default

Every node should be capable of acting as:

- execution node
- service provider
- service consumer
- event participant

No architectural assumption should require a permanent central server.

---

## Stable SDK

SDKs evolve slowly.

Internal implementations may change frequently.

Public contracts must remain stable whenever possible.

---

## Documentation Driven Development

Major architectural work follows this order:

1. Update architecture documentation.
2. Record an ADR.
3. Implement.
4. Test.
5. Update examples.

---

# Trade-offs

RouterPilot intentionally favors:

- simplicity over cleverness
- explicitness over magic
- interfaces over inheritance
- long-term maintainability over short-term convenience

---

# Decision Checklist

Before introducing a new subsystem, contributors should ask:

- Does this reduce coupling?
- Can it become a plugin?
- Does it expose capabilities instead of implementation?
- Can it work offline?
- Is it transport-independent?
- Does it require an ADR?
- Can it be tested independently?

---

# Anti-Patterns

Avoid introducing:

- global mutable state
- transport-specific runtime logic
- planner-specific runtime logic
- shell commands as public API
- circular dependencies
- hidden side effects

---

# Philosophy Statement

RouterPilot should remain a small, predictable, extensible runtime that enables distributed autonomous agents without sacrificing determinism, security or portability.

---

# Related Documents

- 01_VISION.md
- 03_TERMINOLOGY.md
- 04_ARCHITECTURAL_PRINCIPLES.md

---

# Next

Continue with **03_TERMINOLOGY.md**.
