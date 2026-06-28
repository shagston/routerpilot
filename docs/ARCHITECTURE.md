# ARCHITECTURE.md

> **RouterPilot Architecture**
>
> Version: **0.1.0**
>
> Status: **Draft**
>
> This document describes the architectural principles, runtime model, and internal components of RouterPilot.

---

# Table of Contents

1. Vision
2. Goals
3. Non-Goals
4. Design Principles
5. System Architecture
6. Runtime Lifecycle
7. AI Pipeline
8. Context Engine
9. Memory
10. Planner
11. Runtime Engine
12. Tool System
13. Safety Model
14. Event System
15. Plugin System
16. Repository Layout
17. Scalability
18. Security
19. Observability
20. Future Evolution

---

# Vision

RouterPilot is an AI-native networking runtime that translates human intent into safe, deterministic networking operations.

Unlike chatbot wrappers around shell commands, RouterPilot separates:

* reasoning
* planning
* execution
* verification

into independent layers.

The language model is treated as a planner—not as the operating system.

---

# Goals

The project is designed to satisfy the following requirements.

## Lightweight

Must run on:

* OpenWrt
* Raspberry Pi
* Docker
* Linux
* small VPS

without requiring GPU acceleration.

---

## Deterministic

Identical inputs should produce identical execution plans whenever possible.

Random behavior belongs only inside LLM reasoning.

---

## Safe

No AI-generated command should execute directly.

Every action must pass through deterministic validation.

---

## Explainable

Every decision should be inspectable.

The runtime must be able to answer:

* Why?
* What happened?
* What changed?
* What failed?

---

## Extensible

Adding new functionality should require adding new tools rather than modifying the core runtime.

---

# Non-Goals

RouterPilot is **not**:

* a shell replacement
* another ChatGPT wrapper
* an autonomous operating system
* a remote access platform
* a configuration management system like Ansible

It complements these tools rather than replacing them.

---

# Design Principles

## Separation of Concerns

Planning and execution are different responsibilities.

LLM:

* understands intent
* produces plans

Runtime:

* executes
* validates
* retries
* logs
* verifies

---

## Explicit Interfaces

Every subsystem communicates using typed interfaces.

No hidden dependencies.

---

## Immutable Plans

Execution plans become immutable after validation.

Runtime may cancel a plan.

Runtime may retry a plan.

Runtime never silently rewrites a validated plan.

---

## Tools First

Business logic lives inside tools.

The runtime orchestrates tools.

The planner selects tools.

---

## Offline First

Cloud inference is optional.

RouterPilot should support:

* local LLMs
* remote APIs
* hybrid inference

---

# System Architecture

```text
                User
                  │
                  ▼
          Intent Detection
                  │
                  ▼
          Context Builder
                  │
                  ▼
            AI Planner
                  │
          Execution Plan
                  │
                  ▼
         Runtime Validation
                  │
                  ▼
          Runtime Scheduler
                  │
      ┌───────────┼───────────┐
      ▼           ▼           ▼
   Tool A      Tool B      Tool C
      │           │           │
      └───────────┼───────────┘
                  ▼
          Verification Layer
                  ▼
             Final Report
```

---

# Runtime Lifecycle

Each request moves through fixed stages.

```text
NEW

↓

CONTEXT_READY

↓

PLANNED

↓

VALIDATED

↓

RUNNING

↓

VERIFYING

↓

COMPLETED
```

Possible failure states:

```text
FAILED

CANCELLED

ROLLED_BACK

TIMEOUT
```

---

# AI Pipeline

The AI subsystem consists of several specialized agents.

## Intent Agent

Responsible for classification.

Example:

```
"My Wi-Fi disappeared."

↓

wifi.diagnose
```

---

## Context Agent

Builds prompt context.

Typical inputs:

* firmware
* packages
* router model
* interfaces
* routes
* firewall
* DNS
* DHCP

---

## Planning Agent

Produces structured execution plans.

Example:

```yaml
plan:

- wifi.status

- wifi.scan

- network.routes

- report
```

---

## Explanation Agent

Converts runtime output into human-readable language.

No execution.

No planning.

Only explanation.

---

# Context Engine

The Context Engine minimizes token usage.

Instead of dumping the full configuration, it provides only relevant information.

Example:

```
Intent:

diagnose DNS

↓

Include:

✓ resolv.conf

✓ dnsmasq

✓ WAN

Exclude:

✗ Wi-Fi

✗ Firewall

✗ VPN
```

This dramatically reduces prompt size.

---

# Memory

RouterPilot separates memory into layers.

## Static Memory

Rarely changes.

Examples:

* router model
* ISP
* preferred DNS
* LAN subnet

---

## Dynamic Memory

Frequently updated.

Examples:

* latest WAN IP
* interface state
* connected clients

---

## Historical Memory

Stores execution history.

Useful for:

* anomaly detection
* regression detection
* user explanations

---

# Planner

The planner outputs typed plans.

Example:

```yaml
intent:

network.restart

steps:

- verify interfaces

- stop services

- reload config

- verify connectivity

rollback:

restore previous config
```

No shell commands are generated.

---

# Runtime Engine

The Runtime is the heart of RouterPilot.

Responsibilities include:

* dependency graph
* retries
* rollback
* scheduling
* timeout
* cancellation
* progress updates
* logging

---

## Dependency Graph

Example

```
Reload Firewall

↓

Restart DNS

↓

Restart DHCP

↓

Verify Connectivity
```

Tasks execute only when dependencies succeed.

---

## Retry Strategy

Every tool defines:

```
retry_count

retry_delay

timeout
```

Example:

```
Ping

↓

Fail

↓

Retry

↓

Retry

↓

Report failure
```

---

# Tool System

Every capability is implemented as a tool.

Example metadata:

```yaml
name:

network.status

version:

1.0

permissions:

read

timeout:

5s

rollback:

none
```

---

## Tool Categories

Examples include:

Networking

* routes
* interfaces
* bridges

Wi-Fi

* scan
* connect
* disconnect

Firewall

* reload
* inspect

DNS

* lookup
* flush
* test

System

* reboot
* logs
* packages

Diagnostics

* traceroute
* speedtest
* ping

---

# Safety Model

Safety never depends on the LLM.

Rules are deterministic.

Example:

```
flash firmware

↓

confirmation required
```

```
factory reset

↓

confirmation required
```

```
read logs

↓

allowed
```

```
network restart

↓

allowed
```

---

## Validation Layers

Before execution:

Schema Validation

↓

Permission Validation

↓

Capability Validation

↓

Conflict Detection

↓

Dry Run

↓

Execution

---

# Event System

Every important action emits an event.

Examples:

```
intent.detected

context.ready

plan.created

plan.validated

tool.started

tool.completed

tool.failed

execution.finished
```

Events are consumed by:

* UI
* CLI
* logs
* telemetry
* plugins

---

# Plugin System

RouterPilot exposes extension points.

Plugins may contribute:

* tools
* validators
* planners
* executors
* telemetry exporters
* memory providers

Plugins communicate only through public interfaces.

---

# Repository Layout

```text
routerpilot/

cmd/

internal/

planner/

runtime/

memory/

context/

events/

telemetry/

executor/

safety/

sdk/

plugins/

tools/

examples/

configs/

docs/

tests/
```

---

# Scalability

The architecture supports future distributed execution.

Possible future topology:

```
Desktop

↓

Coordinator

↓

Router A

↓

Router B

↓

Router C
```

Each runtime remains independent.

---

# Security

Principles:

* least privilege
* typed permissions
* signed plugins
* audit logs
* deterministic validation
* no arbitrary shell execution

Credentials remain outside prompts whenever possible.

---

# Observability

Every execution produces:

* execution ID
* timestamps
* duration
* tool outputs
* logs
* validation report
* final result

The runtime should be replayable for debugging.

---

# Future Evolution

Planned capabilities include:

* distributed execution
* local LLM orchestration
* policy engine
* topology discovery
* voice interface
* workflow automation
* web dashboard
* mobile companion
* multi-router coordination
* cloud synchronization (optional)

---

# Architectural Philosophy

RouterPilot is built around a simple premise:

> AI should decide **what** to do.

The runtime decides **how** to do it safely.

This separation keeps the system predictable, testable, portable, and suitable for production networking environments.
