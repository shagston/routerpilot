# 10_RUNTIME.md

> Status: Draft
> Version: 0.1

# Runtime Architecture

The Runtime is the execution kernel of RouterPilot.

Every subsystem ultimately interacts with the Runtime. It is responsible for coordinating execution, enforcing policies, routing events, loading plugins, and providing a stable execution environment for agents.

---

# Responsibilities

The Runtime MUST:

- initialize core subsystems
- load configuration
- initialize transports
- load plugins
- initialize registries
- start schedulers
- execute agents
- resolve capabilities
- enforce policies
- publish lifecycle events
- expose public APIs
- monitor subsystem health

The Runtime MUST NOT:

- implement platform-specific logic
- depend on a concrete planner
- depend on a concrete transport
- execute shell commands directly

---

# Architectural Position

```text
                Applications
                      │
                Public SDK/API
                      │
               ┌──────▼──────┐
               │   Runtime   │
               └──────┬──────┘
     ┌─────────┬──────┼──────┬─────────┐
     ▼         ▼      ▼      ▼         ▼
 Agents  Capabilities Events Policy Scheduler
     │         │      │      │         │
     └─────────┴──────┼──────┴─────────┘
                      ▼
                  Transport
```

---

# Runtime Components

## Runtime Core

Coordinates subsystem lifecycle.

## Registry Manager

Owns discovery and registration of:

- agents
- capabilities
- plugins
- transports

## Execution Engine

Executes validated capability requests.

## Lifecycle Manager

Transitions runtime and subsystem states.

## Health Manager

Tracks subsystem readiness and liveness.

---

# Startup Sequence

```text
Load Configuration
        │
Initialize Runtime
        │
Initialize Registries
        │
Load Plugins
        │
Initialize Policy Engine
        │
Initialize Event Bus
        │
Initialize Scheduler
        │
Initialize Transport
        │
Start Agent Manager
        │
Runtime Ready
```

---

# Shutdown Sequence

1. Stop accepting new work.
2. Drain queues.
3. Flush events.
4. Persist state.
5. Stop transports.
6. Stop plugins.
7. Release resources.

Shutdown should be graceful whenever possible.

---

# Runtime State

```text
Created
   │
Initializing
   │
Starting
   │
Ready
   │
Running
   │
Stopping
   │
Stopped

Failure
   │
Recovering
```

---

# Public Responsibilities

The Runtime exposes services for:

- Agent execution
- Capability execution
- Event publication
- Policy evaluation
- Scheduler integration
- Plugin lifecycle
- Transport abstraction

---

# Internal Rules

- Runtime owns execution.
- Runtime owns lifecycle.
- Runtime never bypasses policy.
- Runtime never invokes providers directly without registry resolution.
- Runtime publishes lifecycle events for every major transition.

---

# Suggested Interfaces

```go
type Runtime interface {
    Start(context.Context) error
    Stop(context.Context) error
    Execute(Plan) (Result, error)
    Events() EventBus
    Capabilities() CapabilityRegistry
    Agents() AgentRegistry
}
```

Concrete implementations may extend the interface but should preserve compatibility.

---

# Invariants

- Exactly one Runtime exists per process.
- Runtime state transitions are monotonic.
- Every capability request is authorized.
- Runtime components communicate through interfaces.
- Public SDK packages must not import internal packages.

---

# Related Documents

- 11_RUNTIME_LIFECYCLE.md
- 12_RUNTIME_CONTEXT.md
- 13_RUNTIME_EXECUTION.md
- 20_AGENT_MODEL.md
- 30_CAPABILITY_MODEL.md

---

# Next

Continue with **11_RUNTIME_LIFECYCLE.md**.
