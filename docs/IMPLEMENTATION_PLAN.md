# IMPLEMENTATION_PLAN.md

> RouterPilot Implementation Plan
>
> Version: 0.1
>
> Status: Living Document

---

# Purpose

This document describes the implementation roadmap of RouterPilot.

Unlike the architectural documents, this file focuses on **engineering milestones**, implementation order, and project deliverables.

Every milestone should result in a runnable system.

---

# Development Principles

The project should evolve through small vertical slices.

Avoid implementing large subsystems in isolation.

Instead of building:

```text
Planner

↓

Runtime

↓

Tools

↓

CLI
```

prefer:

```text
CLI

↓

Planner

↓

Runtime

↓

One Tool

↓

Working Result
```

Every milestone should produce something executable.

---

# Overall Roadmap

```text
M0 Documentation

↓

M1 Core SDK

↓

M2 Runtime

↓

M3 Tool Registry

↓

M4 Planner

↓

M5 CLI

↓

M6 First Tool

↓

M7 OpenWrt Integration

↓

M8 Web API

↓

M9 Plugins

↓

M10 v1.0
```

---

# Current Implementation Status

Last updated: 2026-06-28

The repository currently provides a full intent-driven automation platform:

```text
CLI / API
   |
   v
Intent → Planner (Simple/LLM) → Context Engine
   |
   v
Safety Guard (risk-based) + Safety Validator (permission-based)
   |
   v
Runtime (DAG scheduler, retry, timeout, dry-run, event publishing)
   |
   v
Tool Registry → 7 network tools (ping, interface, ip, route)
   |
   v
Linux Network Provider (JSON + text fallback for OpenWrt)
```

Implemented packages:

```text
cmd/routerpilot
internal/api
internal/app
internal/context
internal/events
internal/memory
internal/network
internal/planner
internal/registry
internal/runtime
internal/safety
sdk/events
sdk/memory
sdk/planner
sdk/runtime
sdk/tool
sdk/types
tools/network
```

Working commands:

```powershell
go run .\cmd\routerpilot tools
go run .\cmd\routerpilot ping 127.0.0.1 1 --events
go run .\cmd\routerpilot plan <intent> [args...]
go run .\cmd\routerpilot serve
go test ./...
```

ROUTERPILOT_PORT, ROUTERPILOT_PERMISSIONS, ROUTERPILOT_RISK, ROUTERPILOT_PLANNER env vars supported.

Milestone status:

| Milestone | Status | Notes |
| --------- | ------ | ----- |
| M0 Documentation | In progress | Architecture handbook exists, but some README links still point to planned docs. |
| M1 Core SDK | Implemented | Public contracts for tool, runtime, planner, events, memory and shared types. |
| M2 Runtime | Implemented | DAG execution, retries, timeouts, dry-run, event publishing. Rollback and parallel scheduling remain future work. |
| M3 Tool Registry | Implemented | In-memory registry with registration, lookup and metadata listing. |
| M4 Planner | Implemented | SimplePlanner handles 8 intents (ping, interface.status, interface.set, ip.show, ip.set, route.show, route.add, diagnose). LLM planner also available. Segment-based dependency-safe splitting. |
| M5 CLI | Implemented | `ping`, `plan`, `tools`, `serve` commands. Interactive safety confirmation for high-risk plans. |
| M6 First Production Tool | Implemented | `network.ping` and 6 additional tools (interface, ip, route — get + set). |
| M7 OpenWrt Integration | Implemented | `ip` text output fallback parser when `-j` flag unavailable. |
| M8 REST API | Implemented | `GET /health`, `POST /intent`, `POST /plan`, `GET /tools`, `GET /status`, `GET /events`, `GET /events/stream`. CORS, graceful shutdown, structured JSON errors. |
| M9 Plugin System | Not started | Plugin interface, loader, registry integration remain. |
| M10 v1.0 | Not started | Requires additional tools, telemetry, docs. |

---

# Milestone 0

Documentation

## Goal

Freeze architecture.

Deliverables

* Architecture
* Runtime
* SDK
* Planner
* Tools

Exit Criteria

Architecture no longer changes every week.

Current status

In progress. Core architecture documents exist, and implementation status is now tracked here.

---

# Milestone 1

Core SDK

Goal

Create stable public interfaces.

Deliverables

```text
sdk/

tool/

runtime/

planner/

events/

memory/

types/
```

Exit Criteria

A Tool can compile without Runtime implementation.

Current status

Initial slice implemented. See `sdk/` and `examples/mocktool`.

---

# Milestone 2

Runtime

Goal

Implement deterministic execution.

Deliverables

* Execution object
* Task scheduler
* State machine
* Retry engine
* Cancellation
* Rollback
* Event publishing

Exit Criteria

Runtime executes mocked Tasks.

Current status

Initial slice implemented. The runtime executes mocked tasks and the real `network.ping` tool through the same execution path.

Remaining work

* rollback execution
* parallel scheduling
* richer cancellation boundaries
* persisted execution records

---

# Milestone 3

Tool Registry

Goal

Dynamic capability discovery.

Deliverables

* Registry
* Metadata loader
* Tool lookup
* Capability resolver

Exit Criteria

Runtime discovers Tools automatically.

Current status

Initial in-memory registry implemented. Automatic metadata loading and external discovery are not implemented yet.

---

# Milestone 4

Planner

Goal

Generate executable plans.

Deliverables

* Intent parser
* Context builder
* Planner interface
* Plan validator

Exit Criteria

Simple user requests become valid Plans.

Current status

Implemented. `SimplePlanner` handles 8 intents (ping, interface.status, interface.set, ip.show, ip.set, route.show, route.add, diagnose).
Segment-based planner splits plans into dependency-safe execution segments.
`HasDependencyCycle()` utility integrated into both planner validation and runtime.
Context system (`internal/context/system.go`) gathers system state with validation, event publishing, and per-tool timeout.

---

# Milestone 5

CLI

Goal

Interactive interface.

Examples

```text
routerpilot diagnose

routerpilot dns

routerpilot wifi

routerpilot ping
```

Exit Criteria

CLI can execute a Plan through Runtime.

Current status

Implemented. Commands: `ping <host> [count] [--events]`, `plan <intent> [args...]`, `tools`, `serve`.
CLI supports interactive safety confirmation (prompts for y/N on medium+ risk plans).
Adaptive re-planning retries up to 3 times with context from previous failures.

---

# Milestone 6

First Production Tool

Goal

Complete vertical slice with multiple tools.

Recommended Tool

```text
network.ping
```

Execution Flow

```text
CLI

↓

Planner

↓

Runtime

↓

Registry

↓

Ping Tool

↓

Result
```

Exit Criteria

End-to-end execution works.

Current status

7 production tools implemented:
* `network.ping` — cross-platform ping (Windows + Linux)
* `network.interface.status` — interface state query
* `network.interface.set` — interface up/down
* `network.ip.show` — IP address listing
* `network.ip.set` — IP address assignment
* `network.route.show` — routing table query
* `network.route.add` — route installation

---

# Milestone 7

OpenWrt Support

Goal

Real router execution.

Deliverables

* ubus executor
* uci integration
* package detection
* capability detection

Exit Criteria

RouterPilot runs on OpenWrt.

Current status

Linux network provider (`internal/network/linux.go`) implements `ip` text output fallback parsing.
`GetInterfaceStatus`, `GetAddresses`, and `GetRoutes` try JSON mode (`ip -j`) first; if the `-j` flag is unavailable (busybox iproute2), they fall back to regex-based text parsing.
Configuration via env vars: `ROUTERPILOT_PORT`, `ROUTERPILOT_PERMISSIONS`, `ROUTERPILOT_RISK`, `ROUTERPILOT_PLANNER`.

---

# Milestone 8

REST API

Goal

Machine interface.

Endpoints

```text
GET  /              API info
GET  /health        Health check
POST /intent        Execute an intent
POST /plan          Preview a plan without executing
GET  /tools         List available tools
GET  /status        Server status
GET  /events        List execution events
GET  /events/stream Stream execution events (SSE)
```

Exit Criteria

CLI and API share Runtime.

Current status

Implemented. Server supports CORS, graceful shutdown (SIGINT/SIGTERM), structured JSON error responses.
Configurable via `ROUTERPILOT_PORT` env var (default `:8080`).

---

# Milestone 9

Plugin System

Goal

Third-party extensions.

Deliverables

* Plugin loader
* Registry integration
* Version checking
* Capability negotiation

Exit Criteria

External Tool loads without modifying Runtime.

---

# Milestone 10

Version 1.0

Minimum feature set

Planner

Runtime

Registry

CLI

Plugin System

OpenWrt

20–30 production Tools

Telemetry

Documentation

---

# Suggested Repository Structure

```text
cmd/
internal/
sdk/
tools/
plugins/
configs/
docs/
examples/
tests/
```

This structure should remain stable.

---

# Initial Tools

Recommended implementation order.

Diagnostics

```text
network.ping

network.status

network.routes

network.gateway
```

DNS

```text
dns.lookup

dns.status

dns.flush
```

Wi-Fi

```text
wifi.status

wifi.scan

wifi.connect
```

Firewall

```text
firewall.status

firewall.reload
```

System

```text
system.logs

system.info

system.reboot
```

---

# Testing Strategy

Every milestone must include:

* unit tests
* integration tests
* regression tests

No milestone is complete without automated tests.

---

# Continuous Integration

Recommended pipeline

```text
Lint

↓

Unit Tests

↓

Integration Tests

↓

Coverage

↓

Artifacts
```

Every pull request should pass the pipeline.

---

# Documentation Rule

Documentation is part of the implementation.

No public interface should exist without documentation.

Every new subsystem must include:

* architecture notes
* examples
* API documentation
* tests

---

# Release Strategy

Recommended release cadence

```text
0.1

↓

0.2

↓

0.3

↓

0.5

↓

0.8

↓

1.0
```

Major versions should only introduce intentional breaking changes.

---

# Definition of Done

A milestone is complete only if:

* code compiles
* tests pass
* documentation is updated
* interfaces are stable
* examples work
* CI succeeds

---

# Long-Term Vision

After v1.0, RouterPilot can evolve toward:

* multi-router orchestration
* distributed runtimes
* local LLM execution
* policy-based automation
* autonomous diagnostics
* cloud synchronization
* graphical dashboard
* mobile companion
* community plugin ecosystem

These features should build upon the existing architecture without requiring fundamental redesign.

---

# Guiding Principle

Implement the smallest useful vertical slice first.

Every completed milestone should leave the project in a usable, testable, and extensible state.

Prefer incremental progress over speculative infrastructure.

A working system with one Tool is more valuable than a sophisticated architecture with no executable path.
