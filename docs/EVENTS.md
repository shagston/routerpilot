# EVENTS.md

> RouterPilot Event System
>
> Version: 0.1
>
> Status: Draft

---

# Overview

The Event System is the communication backbone of RouterPilot.

Every significant action performed by the Runtime is represented as an immutable Event.

Components communicate through Events rather than direct coupling.

This architecture enables observability, extensibility, and deterministic execution.

---

# Design Goals

The Event System should be:

* asynchronous
* deterministic
* observable
* replayable
* loosely coupled
* extensible

Events are facts.

They never describe intentions.

---

# Philosophy

RouterPilot follows an Event-Driven Architecture.

Instead of

```text id="f2x8zv"
Runtime

↓

Logger

↓

Telemetry

↓

UI
```

the Runtime emits Events.

Consumers subscribe independently.

```text id="0q6twy"
Runtime

↓

Event Bus

↓

Logger

↓

Telemetry

↓

CLI

↓

Web UI

↓

Plugins
```

The Runtime remains unaware of subscribers.

---

# Event Lifecycle

Every Event follows the same lifecycle.

```text id="bd7l4h"
Created

↓

Published

↓

Delivered

↓

Consumed

↓

Archived
```

Events are immutable after publication.

---

# Event Structure

Every Event contains the same core fields.

```yaml id="wxq5w7"
id:

timestamp:

execution_id:

type:

source:

payload:

metadata:
```

Payloads are typed.

Metadata is optional.

---

# Event Categories

Events are grouped by subsystem.

Examples

Execution

```text id="u2z2jq"
execution.created

execution.started

execution.completed

execution.failed
```

Planning

```text id="rw9v0s"
plan.created

plan.validated

plan.rejected
```

Runtime

```text id="e1x4bn"
runtime.started

runtime.stopped

runtime.error
```

Tasks

```text id="v31v0w"
task.created

task.started

task.completed

task.failed

task.cancelled
```

Tools

```text id="p1c9hj"
tool.started

tool.progress

tool.completed

tool.failed
```

Verification

```text id="e5tw5u"
verification.started

verification.completed

verification.failed
```

Rollback

```text id="j8mbt8"
rollback.started

rollback.completed

rollback.failed
```

Plugins

```text id="tqv7ok"
plugin.loaded

plugin.unloaded

plugin.failed
```

---

# Event Bus

The Event Bus is responsible for:

* publishing
* routing
* subscription management
* event ordering

It does not interpret event payloads.

---

# Publishers

Any Runtime component may publish Events.

Typical publishers:

* Runtime
* Planner
* Tool Executor
* Validator
* Registry
* Plugin Loader

Publishers never communicate directly with subscribers.

---

# Subscribers

Subscribers consume Events independently.

Examples:

* Logger
* CLI
* Web UI
* Metrics Exporter
* Notification Service
* Plugin
* Recorder

A failure in one subscriber must not affect others.

---

# Event Ordering

Within a single Execution, Events must preserve causal order.

Example

```text id="fgv0np"
execution.started

↓

task.started

↓

tool.started

↓

tool.completed

↓

task.completed

↓

execution.completed
```

Cross-execution ordering is not guaranteed.

---

# Event Delivery

The Runtime guarantees at-least-once delivery inside a single process.

Subscribers must therefore be idempotent.

---

# Event Persistence

Events may optionally be persisted.

Possible backends:

* memory
* JSON log
* SQLite
* PostgreSQL

Persistence enables:

* replay
* auditing
* debugging

---

# Event Replay

A stored event stream may be replayed.

Use cases:

* debugging
* deterministic testing
* telemetry regeneration
* UI reconstruction

Replay never re-executes Tools.

Only Events are replayed.

---

# Correlation

Every Event belongs to an Execution.

Example

```yaml id="y5gc5e"
execution_id:

task_id:

tool_id:

parent_event:
```

Correlation allows complete execution tracing.

---

# Event Payloads

Payloads should contain only structured data.

Avoid formatted strings.

Good

```yaml id="v7mvkh"
latency: 18

packet_loss: 0
```

Avoid

```text id="0s3wrm"
"Ping completed successfully."
```

Formatting belongs to presentation layers.

---

# Event Filtering

Subscribers may filter by:

* execution
* type
* source
* severity
* category

Filtering should occur before payload deserialization where possible.

---

# Severity

Events may define severity levels.

```text id="9kuxhf"
DEBUG

INFO

WARNING

ERROR

CRITICAL
```

Severity is informational.

It does not affect Runtime behavior.

---

# Event Consumers

Typical consumers include:

Logging

Telemetry

CLI Progress

WebSocket Streams

Metrics

Tracing

Plugins

Audit Logs

Consumers should remain independent.

---

# Thread Safety

Publishing Events must be safe from multiple goroutines.

The Event Bus should minimize locking while preserving ordering guarantees.

---

# Performance

Event publication should have minimal overhead.

Recommendations:

* non-blocking queues
* bounded buffers
* batch persistence
* asynchronous consumers

The Runtime must not block on slow subscribers.

---

# Event API

The public SDK exposes a minimal interface.

```go id="g4i8c2"
type EventPublisher interface {
    Publish(Event) error
}

type EventSubscriber interface {
    Handle(Event) error
}
```

The SDK hides implementation details.

---

# Testing

The Event System should support:

* deterministic ordering
* replay tests
* concurrency tests
* subscriber isolation
* persistence tests

Events are ideal for golden-file regression testing.

---

# Future Evolution

Potential enhancements include:

* distributed event buses
* remote streaming
* OpenTelemetry integration
* Kafka/NATS adapters
* event versioning
* event compression
* real-time dashboards

These additions should not change the public Event interfaces.

---

# Event Philosophy

Events are the source of truth for everything that happens inside RouterPilot.

The Runtime produces Events.

The rest of the system observes them.

By making Events immutable, structured, and replayable, RouterPilot gains transparent execution, simplified debugging, powerful observability, and loose coupling between subsystems.
