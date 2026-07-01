
# 01_VISION.md

> Status: Draft
> Version: 0.1

# Vision

RouterPilot is a **Distributed Edge Agent Runtime** designed to execute autonomous software agents across heterogeneous edge devices without requiring a centralized control plane.

The project began with OpenWrt, but OpenWrt is no longer considered the architectural center of the system. It is one supported platform among many.

The long-term objective is to provide a small, secure, capability-oriented runtime that enables agents to discover one another, communicate, cooperate and execute work across a distributed mesh.

---

# Mission

Provide a production-grade runtime for autonomous edge agents that is:

- deterministic
- distributed
- offline-capable
- transport-independent
- secure by default
- extensible through plugins

---

# Long-Term Goals

## 1. Build once, run anywhere

Supported targets include:

- OpenWrt
- Linux
- Raspberry Pi
- x86 servers
- virtual machines
- containers
- future embedded systems

Platform-specific details are abstracted behind capabilities.

---

## 2. Distributed by design

Every RouterPilot node is both a client and a server.

Nodes may:

- execute local agents
- expose capabilities
- consume remote capabilities
- exchange events
- synchronize state

No central coordinator is required.

---

## 3. Capability-first Architecture

Applications must not depend on operating-system commands.

Instead they request capabilities.

Example:

```
network.scan
filesystem.read
service.restart
```

The runtime decides how those capabilities are fulfilled.

---

## 4. AI is Optional

LLMs improve planning but are not required for execution.

The runtime must continue to operate with:

- rule-based planners
- workflows
- static plans
- remote planners

Execution must remain deterministic regardless of planner implementation.

---

## 5. Event-driven Runtime

All major subsystems communicate through events.

Benefits:

- loose coupling
- scalability
- observability
- replayability

---

## 6. Security-first

Every action passes through policy evaluation.

Core principles:

- least privilege
- explicit permissions
- auditable execution
- isolated plugins

---

# Non Goals

RouterPilot is not intended to become:

- a Kubernetes replacement
- a shell scripting framework
- a configuration management system
- a cloud orchestration platform
- a monolithic AI assistant

---

# Success Criteria

A successful RouterPilot deployment should allow:

1. Autonomous agents to execute locally.
2. Remote capability invocation.
3. Distributed event propagation.
4. Plugin-based extensibility.
5. Offline operation.
6. Multiple transports without runtime changes.
7. Stable SDKs for third-party developers.

---

# Architecture North Star

```
                Planner
                   │
                   ▼
             Agent Runtime
                   │
        ┌──────────┼──────────┐
        ▼          ▼          ▼
  Capabilities   Events    Policies
        │          │          │
        └──────────┼──────────┘
                   ▼
              Transport API
                   │
     ┌─────────────┴─────────────┐
     ▼                           ▼
 Reticulum                  Other Transports
```

---

# Design Invariants

The following statements are architectural invariants.

- Runtime MUST NOT depend on transport implementations.
- Runtime MUST NOT depend on a concrete planner.
- Agents MUST execute through the runtime.
- Capabilities MUST be authorized before execution.
- Plugins MUST interact through public SDK interfaces.
- Distributed execution MUST appear identical to local execution whenever practical.

Any proposal violating these invariants requires a new ADR.

---

# Related Documents

- 02_PHILOSOPHY.md
- 04_ARCHITECTURAL_PRINCIPLES.md
- ADR-0001 Vision
- ADR-0002 Runtime Architecture

---

# Next

Continue with **02_PHILOSOPHY.md**.
