# RUNTIME.md

> RouterPilot Runtime
>
> Version: 0.1
>
> Status: Draft

---

# Overview

The Runtime is the execution engine of RouterPilot.

It is responsible for turning validated execution plans into deterministic operations while enforcing safety, observability, and reproducibility.

Unlike the Planner, the Runtime never reasons about user intent. Its responsibility begins only after a plan has been approved.

---

# Responsibilities

The Runtime owns:

* execution
* scheduling
* dependency resolution
* retries
* rollback
* cancellation
* progress reporting
* verification
* telemetry
* event emission

The Runtime never:

* interprets natural language
* modifies execution plans
* generates shell commands using AI

---

# Runtime Pipeline

Every request follows the same lifecycle.

```text
User Request
        │
        ▼
 Intent Parser
        │
        ▼
 Context Builder
        │
        ▼
 Planner
        │
        ▼
 Plan Validation
        │
        ▼
 Runtime
        │
        ▼
 Tool Execution
        │
        ▼
 Verification
        │
        ▼
 Final Result
```

---

# Runtime State Machine

Each execution has a finite state.

```text
NEW

↓

PLANNING

↓

VALIDATING

↓

READY

↓

RUNNING

↓

VERIFYING

↓

COMPLETED
```

Failure states

```text
FAILED

CANCELLED

TIMEOUT

ROLLED_BACK
```

Transitions are deterministic.

No state may be skipped.

---

# Execution Object

Every request creates an Execution.

```yaml
Execution

id

created_at

started_at

finished_at

state

plan

context

events

tasks

result
```

Execution is immutable after completion.

---

# Plan

The Runtime consumes a validated Plan.

A Plan contains:

```yaml
plan_id

intent

steps

rollback

metadata

requirements
```

The Runtime never edits a Plan.

---

# Task

Each Plan is decomposed into Tasks.

Example

```text
Restart DNS

↓

Reload Firewall

↓

Verify WAN

↓

Check Internet
```

Every Task has:

```yaml
id

tool

arguments

status

dependencies

timeout

retry_policy

rollback
```

---

# Scheduler

The Scheduler determines execution order.

Possible execution models:

Sequential

```text
A

↓

B

↓

C
```

Parallel

```text
      A
     / \
    B   C
     \ /
      D
```

The scheduler respects dependency constraints.

---

# Dependency Graph

Tasks form a DAG (Directed Acyclic Graph).

Cycles are rejected during validation.

Example

```text
Reload Network

↓

Restart DHCP

↓

Restart DNS

↓

Verify Connectivity
```

---

# Executor

The Executor invokes Tools.

Responsibilities:

* argument validation
* timeout handling
* retries
* collecting outputs
* event emission

The Executor has no networking logic.

It delegates work to Tools.

---

# Tool Invocation

Execution flow

```text
Task

↓

Permission Check

↓

Schema Validation

↓

Capability Check

↓

Execute Tool

↓

Capture Result

↓

Emit Events
```

---

# Retry Policy

Every Tool defines its retry behavior.

Example

```yaml
retry:

attempts: 3

delay: 2s

strategy: exponential
```

Transient failures may be retried.

Permanent failures stop execution.

---

# Timeouts

Every Tool has a timeout.

Example

```yaml
timeout: 10s
```

Timeout expiration generates

```text
tool.timeout
```

---

# Rollback

Some Tasks define rollback operations.

Example

```text
Apply Configuration

↓

Verification Failed

↓

Restore Previous Configuration
```

Rollback itself is executed as Tasks.

---

# Cancellation

Execution may be cancelled by:

* user
* API
* timeout
* runtime policy

Cancellation never interrupts a Tool in an undefined state.

The Runtime waits until the Tool reaches a safe boundary.

---

# Verification

Execution is not considered complete until verification succeeds.

Example

```text
Restart DNS

↓

DNS Query

↓

Success
```

If verification fails

↓

Execution becomes FAILED

or

ROLLED_BACK

---

# Context Snapshot

Every Execution receives an immutable Context Snapshot.

Example

```yaml
router

firmware

interfaces

dns

routes

firewall

packages
```

This guarantees reproducibility.

---

# Event Bus

Every significant action emits an event.

Examples

```text
execution.created

execution.started

task.started

task.completed

task.failed

rollback.started

rollback.finished

execution.completed
```

Consumers include:

* CLI
* Web UI
* Logs
* Plugins
* Telemetry

---

# Logging

Every Execution produces structured logs.

Example

```yaml
timestamp

execution_id

task

tool

duration

status

message
```

Logs must be machine-readable.

---

# Error Handling

Errors are classified into categories.

Recoverable

* timeout
* temporary network failure
* DNS lookup failure

Non-recoverable

* invalid arguments
* permission denied
* unsupported capability
* corrupted configuration

Only recoverable errors may trigger retries.

---

# Concurrency

Independent Tasks may execute in parallel.

Example

```text
Scan Wi-Fi

Read DHCP

Read Firewall
```

↓

Merge Results

↓

Generate Report

The scheduler determines safe parallelism.

---

# Resource Limits

The Runtime should support constrained devices.

Recommended defaults

```text
Max concurrent tasks: 4

Default timeout: 30 seconds

Default retries: 2

Memory budget: configurable
```

---

# Observability

Every Execution exposes:

* current state
* completed tasks
* running tasks
* failed tasks
* elapsed time
* progress percentage

This allows real-time UI updates.

---

# Extension Points

The Runtime exposes interfaces for:

* Tool Executors
* Validators
* Event Listeners
* Memory Providers
* Telemetry Exporters
* Progress Reporters

The Runtime itself remains independent of concrete implementations.

---

# Runtime Philosophy

The Runtime is intentionally deterministic.

Planning is probabilistic.

Execution is not.

By isolating AI from execution, RouterPilot ensures predictable behavior, easier testing, complete auditability, and safe operation on networking infrastructure.
