# 21_AGENT_LIFECYCLE.md

> Status: Draft
> Version: 0.1

# Agent Lifecycle

This document specifies the lifecycle of an Agent managed by the RouterPilot Runtime.

The Runtime exclusively owns lifecycle transitions. Agents cannot transition themselves between lifecycle states.

---

# Goals

The lifecycle model provides:

- deterministic initialization
- graceful startup and shutdown
- fault isolation
- recovery support
- observable state transitions

---

# State Machine

```text
Created
   │
Initialized
   │
Registered
   │
Ready
   │
Running
   │
Sleeping
   │
Running
   │
Stopping
   │
Stopped

Running
   │
Failure
   │
Recovering
   │
Ready
```

Transitions outside this graph are invalid.

---

# State Definitions

## Created

The Agent object exists but has not been initialized.

The Runtime may validate metadata and manifests.

---

## Initialized

Static configuration has been loaded.

No capability execution is permitted.

---

## Registered

The Agent has been added to the Agent Registry.

Subscriptions and metadata become discoverable.

---

## Ready

The Agent is eligible for execution.

Prerequisites:

- manifest validated
- permissions resolved
- dependencies available

---

## Running

The Runtime invokes the Agent.

While running an Agent may:

- request capabilities
- publish events
- access memory
- update internal state

---

## Sleeping

The Agent is idle.

It remains loaded but consumes no execution slot until:

- an event arrives
- the scheduler triggers execution
- another runtime trigger occurs

---

## Stopping

The Runtime requests graceful termination.

The Agent should:

- stop new work
- release resources
- persist state
- finish cooperative cleanup

---

## Stopped

Terminal state.

The Agent may later be recreated but not restarted directly.

---

## Failure

Entered when execution cannot continue safely.

Typical causes:

- panic
- timeout
- repeated provider failures
- policy violations

Failure does not imply Runtime failure.

---

## Recovering

Optional recovery stage.

Recovery actions may include:

- rebuilding internal state
- reconnecting external resources
- reloading memory
- clearing temporary caches

If recovery succeeds the Agent returns to Ready.

---

# Lifecycle Events

The Runtime SHOULD publish:

- agent.created
- agent.initialized
- agent.registered
- agent.ready
- agent.started
- agent.sleeping
- agent.stopping
- agent.stopped
- agent.failed
- agent.recovering

---

# Lifecycle Ownership

The Runtime controls:

- creation
- registration
- scheduling
- execution
- cancellation
- shutdown
- recovery

The Agent controls only its domain logic.

---

# Cancellation

Cancellation propagates from Runtime to Agent.

The Agent MUST stop work cooperatively and return control promptly.

---

# Health Checks

Agents SHOULD expose health information.

Recommended states:

- Healthy
- Degraded
- Unhealthy

Health is advisory and independent of lifecycle state.

---

# Suggested Interfaces

```go
type ManagedAgent interface {
    Agent
    Start(AgentContext) error
    Stop(AgentContext) error
    Health() HealthStatus
}
```

Optional interfaces may provide lifecycle hooks such as Initialize or Recover.

---

# Invariants

- Only the Runtime changes lifecycle state.
- Every Agent execution begins in Ready.
- Running Agents always have a valid Agent Context.
- Failures are isolated to the affected Agent whenever possible.
- Lifecycle transitions generate observable events.

---

# Related Documents

- 20_AGENT_MODEL.md
- 22_AGENT_API.md
- 23_AGENT_REGISTRY.md
- 10_RUNTIME.md

---

# Next

Continue with **22_AGENT_API.md**.
