# 70_SCHEDULER.md

> Status: Draft
> Version: 0.1

# Scheduler

The Scheduler is responsible for initiating Agent execution according to time, events, dependencies and runtime conditions.

The Scheduler never performs business logic. It creates execution requests and submits them to the Runtime.

---

# Goals

The Scheduler MUST provide:

- deterministic scheduling
- event-driven execution
- cron scheduling
- interval scheduling
- one-shot execution
- dependency-aware execution
- retry orchestration
- cancellation support

---

# Responsibilities

The Scheduler MUST:

- register schedules
- trigger executions
- respect dependencies
- enforce concurrency limits
- emit scheduling events

The Scheduler MUST NOT:

- execute capabilities
- authorize requests
- invoke providers directly

---

# Scheduling Sources

Execution may be triggered by:

- cron expressions
- fixed intervals
- runtime startup
- external API
- incoming transport events
- Event Bus subscriptions
- dependency completion

---

# Scheduling Pipeline

```text
Trigger
   │
Schedule Evaluation
   │
Execution Request
   │
Runtime Queue
   │
Agent Execution
```

---

# Schedule Types

## Cron

Calendar-based recurring execution.

Example:

```yaml
cron: "0 */5 * * * *"
```

---

## Interval

Relative recurring execution.

Example:

```yaml
interval: 30s
```

---

## One-shot

Executed once.

```yaml
run_at: 2026-07-03T10:00:00Z
```

---

## Event Trigger

Execution begins after receiving matching events.

```yaml
on:
  - network.changed
  - runtime.ready
```

---

# Dependency Graph

Schedules MAY depend on completion of other schedules.

```text
Collect Metrics
      │
      ▼
Analyze Metrics
      │
      ▼
Publish Report
```

Circular dependencies MUST be rejected.

---

# Concurrency

Scheduler SHOULD support:

- max concurrent executions
- per-agent concurrency
- global limits
- queue priorities

---

# Retry Policy

Retry configuration MAY include:

- maximum attempts
- backoff strategy
- retry window
- retryable errors

Retries create new execution attempts while preserving correlation metadata.

---

# Cancellation

The Scheduler cooperates with Runtime cancellation.

Cancelling a scheduled execution MUST:

- cancel queued work
- propagate cancellation to active execution
- emit scheduler.cancelled

---

# Suggested Interfaces

```go
type Scheduler interface {
    Register(Schedule) error
    Remove(string) error
    Trigger(string) error
    List() []Schedule
}

type Schedule struct {
    ID       string
    Type     ScheduleType
    Priority int
}
```

---

# Events

The Scheduler SHOULD emit:

- scheduler.registered
- scheduler.triggered
- scheduler.started
- scheduler.completed
- scheduler.failed
- scheduler.cancelled

---

# Observability

Recommended metrics:

- active schedules
- queued executions
- execution latency
- trigger count
- retries
- skipped executions

---

# Invariants

- Scheduler never executes business logic.
- Runtime owns execution.
- Schedules are deterministic.
- Dependency graphs are acyclic.
- Cancellation propagates correctly.
- Scheduling is observable.

---

# Related Documents

- 13_RUNTIME_EXECUTION.md
- 20_AGENT_MODEL.md
- 40_EVENT_BUS.md

---

# Next

Continue with **80_TRANSPORT.md**.
