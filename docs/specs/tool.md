# Tool Specification

> Status: Draft
>
> Version: 0.1

---

# Purpose

A Tool is the smallest executable unit in RouterPilot.

Every operation executed by the Runtime is represented by exactly one Tool.

Tools are deterministic, stateless, and self-describing.

---

# Design Rules

A Tool:

* performs one responsibility
* has typed input
* has typed output
* validates its own arguments
* is independently testable
* exposes metadata
* never depends on another Tool directly

Complex workflows are composed by the Planner and Runtime.

---

# Lifecycle

```text
Registry

↓

Validation

↓

Execution

↓

Verification

↓

Result
```

---

# Interface

```go
type Tool interface {

    Metadata() Metadata

    Validate(ctx Context, input any) error

    Execute(ctx Context, input any) (Result, error)

}
```

Tools must be safe for concurrent execution.

---

# Metadata

Every Tool exposes immutable metadata.

```go
type Metadata struct {

    ID string

    Version string

    Category string

    Description string

    Permissions []Permission

    Capabilities []Capability

    Timeout time.Duration

    SupportsRollback bool

    SupportsDryRun bool

}
```

Metadata is available without executing the Tool.

---

# Input

Every Tool defines exactly one input schema.

Example

```go
type PingInput struct {

    Host string

    Count int

    Timeout time.Duration

}
```

Unknown fields are rejected.

Validation occurs before execution.

---

# Output

Every Tool defines exactly one output schema.

Example

```go
type PingResult struct {

    Success bool

    Latency time.Duration

    PacketLoss float64

}
```

Outputs must contain structured data only.

Formatting belongs to presentation layers.

---

# Context

Every Tool receives Runtime Context.

The Context contains:

* execution id
* cancellation signal
* logger
* event publisher
* configuration
* memory access
* runtime metadata

Tools never access global state.

---

# Errors

Tools return typed errors.

Example

```go
var (

    ErrInvalidInput

    ErrPermissionDenied

    ErrCapabilityMissing

    ErrTimeout

    ErrExecutionFailed

)
```

Errors should support wrapping.

---

# Cancellation

Execution must observe:

```go
ctx.Done()
```

Long-running operations should terminate promptly after cancellation.

---

# Timeouts

Timeouts are enforced by the Runtime.

Tools should not implement independent timeout logic unless interacting with external systems.

---

# Events

Tools publish Events through the Context.

Typical Events

```text
tool.started

tool.progress

tool.completed

tool.failed
```

The Runtime subscribes to these Events.

---

# Logging

Structured logging only.

Example

```go
logger.Info(

    "ping completed",

    "latency", latency,

    "host", host,

)
```

No formatted console output.

---

# Thread Safety

Multiple instances of the same Tool may execute simultaneously.

Shared mutable state must be synchronized or avoided.

---

# Dependencies

Tools must never invoke other Tools directly.

Instead:

```text
Planner

↓

Runtime

↓

Tool A

↓

Runtime

↓

Tool B
```

This keeps orchestration centralized.

---

# Testing Requirements

Every Tool must include:

* metadata tests
* validation tests
* success path
* failure path
* timeout behavior
* cancellation behavior

Golden tests are recommended for structured outputs.

---

# Platform Abstraction

A Tool describes a capability, not an implementation.

Example

```text
network.status

↓

OpenWrt

↓

Linux

↓

Container
```

The Runtime selects the correct implementation.

---

# Performance

Tools should:

* minimize allocations
* avoid blocking operations
* stream progress when appropriate
* avoid unnecessary memory copies

---

# Versioning

Breaking changes require a new Tool version.

Example

```text
network.status@1

network.status@2
```

Old versions may coexist during migration.

---

# Anti-Patterns

Tools should never:

* call the Planner
* modify Runtime state
* access global variables
* invoke shell commands through AI
* contain orchestration logic
* depend on specific UI implementations

---

# Philosophy

A Tool is the smallest trustworthy execution unit in RouterPilot.

It performs exactly one task, exposes a stable contract, and remains isolated from orchestration and planning.

Everything larger than a Tool belongs to the Runtime.
