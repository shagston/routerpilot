# 11_RUNTIME_LIFECYCLE.md

> Status: Draft
> Version: 0.1

# Runtime Lifecycle

This document specifies the lifecycle of a RouterPilot Runtime instance.

The lifecycle is deterministic and state-driven. Every Runtime instance MUST follow
the same state transitions regardless of deployment platform.

---

# Objectives

The lifecycle model ensures:

- predictable startup
- graceful shutdown
- recoverability
- observable state transitions
- deterministic initialization

---

# State Machine

```text
        ┌────────────┐
        │  Created   │
        └─────┬──────┘
              │
              ▼
      Initializing
              │
              ▼
         Starting
              │
              ▼
           Ready
              │
              ▼
          Running
              │
      ┌───────┴────────┐
      ▼                ▼
Stopping          Failure
      │                │
      ▼                ▼
 Stopped        Recovering
                         │
                         ▼
                      Starting
```

No state may be skipped.

---

# State Definitions

## Created

Runtime object exists but owns no resources.

Allowed actions:

- load configuration
- register bootstrap services

Forbidden:

- execute agents
- load plugins
- publish events

---

## Initializing

The Runtime allocates internal structures.

Tasks:

- logger
- configuration
- dependency graph
- registries
- metrics

Failure returns to Created.

---

## Starting

Subsystem startup occurs in dependency order.

Recommended order:

1. Event Bus
2. Registry
3. Policy Engine
4. Memory
5. Scheduler
6. Transport
7. Plugin Manager
8. Agent Manager

A subsystem MUST NOT start before its dependencies.

---

## Ready

Runtime accepts work but has not yet begun scheduled execution.

Health checks must succeed.

---

## Running

Normal operating state.

Responsibilities:

- execute agents
- evaluate policy
- resolve capabilities
- publish events
- process scheduler
- exchange transport messages

---

## Stopping

Runtime rejects new work.

Must:

- drain queues
- flush events
- persist state
- stop transports
- unload plugins

---

## Stopped

Terminal state.

Resources are released.

No execution permitted.

---

## Failure

Entered when an unrecoverable subsystem error occurs.

Failure must emit:

- runtime.failure
- subsystem.failure

---

## Recovering

Optional state.

Recovery may include:

- plugin reload
- transport reconnect
- registry rebuild
- scheduler recovery

If recovery fails, shutdown should begin.

---

# Lifecycle Events

The Runtime SHOULD emit:

- runtime.created
- runtime.initializing
- runtime.starting
- runtime.ready
- runtime.running
- runtime.stopping
- runtime.stopped
- runtime.failure
- runtime.recovering

These events enable observability and external automation.

---

# Startup Invariants

Before entering Running:

- configuration loaded
- policy initialized
- event bus active
- registries populated
- transports initialized
- plugins validated

---

# Shutdown Guarantees

Graceful shutdown should guarantee:

- no new executions
- event queues drained
- scheduler paused
- transports disconnected
- persistent state flushed

Forced termination may violate these guarantees.

---

# Suggested API

```go
type Lifecycle interface {
    Start(context.Context) error
    Stop(context.Context) error
    State() RuntimeState
    Subscribe(StateListener)
}
```

---

# Related Documents

- 10_RUNTIME.md
- 12_RUNTIME_CONTEXT.md
- 13_RUNTIME_EXECUTION.md

---

# Next

Continue with **12_RUNTIME_CONTEXT.md**.
