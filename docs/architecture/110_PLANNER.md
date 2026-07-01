# 110_PLANNER.md

> Status: Draft
> Version: 0.1

# Planner

The Planner transforms goals into executable plans.

Planning is intentionally separated from execution. The Runtime executes plans, while Planners generate them. This separation allows RouterPilot to support deterministic rule engines, workflows, LLM-based planners and future planning algorithms without modifying the Runtime.

---

# Goals

The Planner subsystem MUST provide:

- planner independence
- deterministic execution contracts
- immutable execution plans
- pluggable implementations
- policy-aware planning
- transport independence

The Planner MUST NOT execute capabilities directly.

---

# Architecture

```text
           Goal / Trigger
                 │
                 ▼
             Planner API
                 │
      ┌──────────┼──────────┐
      ▼          ▼          ▼
 Rule Planner Workflow   LLM Planner
                 │
                 ▼
        Immutable Execution Plan
                 │
                 ▼
              Runtime
                 │
                 ▼
          Agent Execution
```

---

# Responsibilities

A Planner MUST:

- receive goals or triggers
- evaluate available context
- construct an execution plan
- declare dependencies
- estimate constraints
- return an immutable plan

A Planner MUST NOT:

- invoke capability providers
- bypass the Runtime
- modify Runtime state
- execute Agents

---

# Planner Types

## Rule Planner

Executes deterministic rule sets.

Best suited for:

- infrastructure
- monitoring
- automation
- embedded systems

---

## Workflow Planner

Produces execution DAGs.

Suitable for:

- orchestration
- long-running workflows
- dependency graphs

---

## LLM Planner

Uses language models to generate plans.

The Runtime treats generated plans identically to any other plan after validation.

LLMs are optional.

---

## Remote Planner

Planning performed by another Runtime.

Returned plans remain subject to local validation.

---

# Execution Plan

An execution plan SHOULD include:

- Plan ID
- Version
- Steps
- Dependencies
- Timeouts
- Retry Policy
- Metadata

Plans become immutable after validation.

---

# Planning Pipeline

```text
Goal
  │
Planner
  │
Draft Plan
  │
Validation
  │
Immutable Plan
  │
Runtime Queue
```

Validation failures reject the plan before execution.

---

# Suggested Interfaces

```go
type Planner interface {
    Name() string
    Version() string
    Plan(context.Context, Goal) (Plan, error)
}

type Plan struct {
    ID       string
    Version  string
    Steps    []Step
}
```

The Runtime depends only on the Planner interface.

---

# Validation

The Runtime validates:

- capability existence
- policy compatibility
- dependency graph
- timeout constraints
- schema correctness

Invalid plans MUST NOT execute.

---

# Observability

Planners SHOULD emit:

- planner.started
- planner.completed
- planner.failed
- planner.validation.failed

Metrics SHOULD include:

- planning latency
- generated plans
- rejected plans
- planner failures

---

# Security

Planning does not grant authority.

Every capability in a plan is authorized independently during execution.

Plans originating from remote or LLM planners MUST be treated as untrusted until validated.

---

# Invariants

- Planning is separate from execution.
- Plans are immutable after validation.
- Runtime owns execution.
- Planner implementations are replaceable.
- Every plan is validated before execution.
- Authorization occurs during execution, not planning.

---

# Related Documents

- 13_RUNTIME_EXECUTION.md
- 50_POLICY_ENGINE.md
- 100_PLUGIN_SYSTEM.md

---

# Next

Continue with **120_SECURITY.md**.
