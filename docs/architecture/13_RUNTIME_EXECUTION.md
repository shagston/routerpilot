# 13_RUNTIME_EXECUTION.md

> Status: Draft
> Version: 0.1

# Runtime Execution Model

This document specifies how work flows through the RouterPilot Runtime.

Execution is deterministic, policy-controlled and capability-oriented.

---

# Objectives

The execution engine MUST:

- execute plans deterministically
- isolate agent executions
- authorize every capability
- publish lifecycle events
- propagate context
- support cancellation and retries

---

# Execution Pipeline

```text
Goal / Trigger
      │
      ▼
 Planner
      │
      ▼
 Execution Plan
      │
      ▼
 Runtime
      │
      ▼
 Agent Execution
      │
      ▼
 Capability Request
      │
      ▼
 Policy Evaluation
      │
      ▼
 Capability Resolution
      │
      ▼
 Provider Execution
      │
      ▼
 Result
      │
      ▼
 Event Publication
```

---

# Execution Unit

The smallest schedulable unit is an **Execution**.

Every execution has:

- Execution ID
- Agent ID
- Context
- State
- Deadline
- Policy Snapshot
- Trace Metadata

Execution IDs MUST be globally unique.

---

# Execution States

```text
Created
   │
Queued
   │
Running
   │
Completed

Running
   │
Failed

Running
   │
Cancelled

Failed
   │
Retrying
```

State transitions are monotonic.

---

# Plan Execution

A Planner produces an immutable execution plan.

The Runtime validates the plan before execution.

Validation includes:

- syntax
- capability existence
- policy pre-check
- dependency graph
- timeout constraints

Invalid plans MUST NOT execute.

---

# Capability Invocation

Capability execution follows this sequence:

1. Create capability context.
2. Evaluate policy.
3. Resolve provider.
4. Execute provider.
5. Collect result.
6. Publish completion event.

Providers never communicate directly with planners.

---

# Parallel Execution

Independent plan steps MAY execute concurrently.

Parallel execution requires:

- no dependency edges
- isolated contexts
- independent capabilities

Ordering constraints MUST be respected.

---

# Failure Handling

Failure categories:

- validation
- authorization
- provider
- transport
- timeout
- cancellation
- internal runtime

Runtime distinguishes transient failures from permanent failures.

---

# Retry Strategy

Retries are controlled by policy and scheduler.

Recommended fields:

- max attempts
- backoff strategy
- retry window
- retryable errors

Retries create new execution attempts while preserving the same Correlation ID.

---

# Cancellation

Cancellation propagates through the execution tree.

```text
Execution
    │
    ├── Capability A
    ├── Capability B
    └── Capability C
```

Cancelling the execution cancels all active children.

---

# Events

The Runtime SHOULD publish:

- execution.created
- execution.started
- execution.completed
- execution.failed
- execution.cancelled
- execution.retried

---

# Suggested Interfaces

```go
type Executor interface {
    Execute(context.Context, Plan) (Result, error)
}

type Execution interface {
    ID() string
    State() ExecutionState
    Context() context.Context
}
```

---

# Observability

Every execution SHOULD expose:

- duration
- retries
- capability count
- policy decisions
- emitted events
- failure reason

These metrics feed tracing, logging and monitoring systems.

---

# Invariants

- Every execution has exactly one root context.
- Every capability request is authorized.
- Execution plans are immutable after validation.
- Provider failures never corrupt runtime state.
- Cancellation is cooperative and propagated.
- Runtime always emits terminal execution events.

---

# Related Documents

- 10_RUNTIME.md
- 11_RUNTIME_LIFECYCLE.md
- 12_RUNTIME_CONTEXT.md
- 20_AGENT_MODEL.md
- 30_CAPABILITY_MODEL.md

---

# Next

Continue with **20_AGENT_MODEL.md**.
