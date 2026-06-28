# PLANNER.md

> RouterPilot Planner
>
> Version: 0.1
>
> Status: Draft

---

# Overview

The Planner is the reasoning component of RouterPilot.

Its responsibility is to transform user intent into a deterministic execution plan.

The Planner never executes actions.

It only decides **what should happen**.

Execution belongs exclusively to the Runtime.

---

# Responsibilities

The Planner is responsible for:

* understanding intent
* selecting capabilities
* building execution plans
* ordering tasks
* estimating dependencies
* requesting verification
* minimizing execution cost

The Planner is **not** responsible for:

* executing commands
* accessing the operating system
* modifying router configuration
* retry logic
* rollback
* progress reporting

---

# Design Philosophy

RouterPilot separates intelligence from execution.

```text
User Request
      │
      ▼
 Planner
      │
 Execution Plan
      │
      ▼
 Runtime
      │
 Tool Execution
```

The Planner reasons.

The Runtime acts.

---

# Planning Pipeline

Every request passes through the same planning stages.

```text
Intent Detection
        │
        ▼
Context Collection
        │
        ▼
Capability Discovery
        │
        ▼
Plan Generation
        │
        ▼
Plan Validation
        │
        ▼
Execution Plan
```

Each stage has a single responsibility.

---

# Intent Detection

The Planner begins by classifying user intent.

Examples

```text
"My Internet is slow."

↓

diagnose.network
```

```text
"Restart Wi-Fi"

↓

wifi.restart
```

```text
"Show DHCP leases"

↓

dhcp.list
```

Intent detection should produce structured output.

---

# Context Collection

Planning depends on context.

The Context Engine provides only relevant information.

Example

Intent

```text
dns.lookup
```

Relevant context

```text
DNS server

WAN status

Resolver

Network interfaces
```

Ignored

```text
Wi-Fi channels

Firewall rules

Package list
```

The Planner never requests unnecessary data.

---

# Capability Discovery

The Planner does not assume available functionality.

Instead, it queries the Tool Registry.

Example

```text
Available

✓ dns.lookup

✓ ping

✓ traceroute

✗ speedtest
```

Plans must only reference available capabilities.

---

# Tool Selection

Multiple Tools may solve the same problem.

Example

Internet diagnostics

Possible tools

```text
Ping

Traceroute

DNS Lookup

Gateway Check

Interface Status
```

The Planner selects the minimal set required.

---

# Planning Principles

The generated plan should be:

* minimal
* deterministic
* verifiable
* composable
* observable

Avoid unnecessary actions.

---

# Execution Plan

Plans are immutable after validation.

Example

```yaml
intent: diagnose.network

steps:

- network.status

- network.gateway

- dns.lookup

- ping.internet

- report
```

Plans never contain shell commands.

---

# Dependency Resolution

The Planner identifies dependencies.

Example

```text
Restart DNS

↓

Verify DNS

↓

Verify Internet
```

Dependencies are explicit.

The Runtime schedules them.

---

# Verification Planning

Verification is part of every plan.

Example

```text
Reload Firewall

↓

Verify Firewall

↓

Verify Connectivity
```

A plan without verification is incomplete.

---

# Read vs Write Operations

The Planner distinguishes between:

Read

* inspect
* diagnose
* query

Write

* restart
* modify
* delete

Write operations may require confirmation.

---

# Risk Classification

Every plan receives a risk level.

Example

Low

```text
Read logs
```

Medium

```text
Restart DNS
```

High

```text
Reset router
```

Critical

```text
Flash firmware
```

The Runtime may enforce additional confirmation based on risk.

---

# Cost Optimization

The Planner should minimize:

* execution time
* network traffic
* system load
* token usage
* duplicated work

Example

Avoid

```text
Ping

Ping

Ping
```

Prefer

```text
Ping

Reuse Result
```

---

# Context Compression

The Planner should consume the smallest useful context.

Instead of:

Entire configuration

Prefer:

Relevant configuration fragment

Smaller prompts improve both latency and reliability.

---

# Multi-Step Planning

Complex requests become multiple tasks.

Example

```text
Fix Wi-Fi

↓

Inspect

↓

Diagnose

↓

Apply Fix

↓

Verify

↓

Report
```

Each step should be independently executable.

---

# Replanning

Execution may invalidate assumptions.

Example

```text
Gateway Missing

↓

Original Plan Invalid

↓

Generate New Plan
```

Replanning always starts from the latest verified state.

---

# Failure Strategy

Planning failures differ from execution failures.

Planning failure

```text
Unknown capability
```

Execution failure

```text
Tool timeout
```

Only the Runtime handles execution failures.

---

# Hallucination Prevention

The Planner must never invent:

* Tools
* Parameters
* Capabilities
* Interfaces

Every Tool referenced by a plan must exist in the Tool Registry.

Unknown capabilities must result in a planning error.

---

# Constraint System

The Planner operates under deterministic constraints.

Examples

Never

* bypass validation
* invent permissions
* modify execution plans after approval

Always

* respect tool contracts
* respect permissions
* respect platform capabilities

---

# Plan Validation

Before reaching the Runtime, every plan is validated.

Checks include:

* tool existence
* schema compatibility
* dependency cycles
* permission requirements
* capability availability
* risk classification

Invalid plans are rejected.

---

# Planner Interfaces

The Planner depends only on public SDK interfaces.

Required services

* Context Provider
* Tool Registry
* Memory Provider
* Validator
* Capability Resolver

The Planner never depends on concrete implementations.

---

# Memory Usage

The Planner may use memory for:

* user preferences
* previous executions
* preferred DNS
* known topology

Memory should improve planning, never replace current system state.

---

# Explainability

The Planner should be able to explain its decisions.

Example

```text
Selected:

dns.lookup

Because:

Internet connectivity depends on DNS resolution.
```

Reasoning should be concise and deterministic.

---

# Testing

Planner implementations should be tested for:

* deterministic planning
* invalid input
* missing tools
* dependency ordering
* risk classification
* context reduction

Regression tests should use fixed planning scenarios.

---

# Future Evolution

Potential future capabilities include:

* hierarchical planning
* collaborative multi-agent planning
* plan caching
* execution cost prediction
* adaptive planning based on historical outcomes
* distributed planning across multiple routers
* policy-aware optimization

These enhancements should not change the Planner's external contract.

---

# Planner Philosophy

The Planner is the only component responsible for reasoning.

It transforms intent into a structured plan while remaining completely isolated from execution.

By limiting the Planner to deterministic, verifiable planning and delegating all side effects to the Runtime, RouterPilot maintains a clear separation between intelligence and action, reducing complexity while improving safety, auditability, and testability.
