# TOOLS.md

> RouterPilot Tool System
>
> Version: 0.1
>
> Status: Draft

---

# Overview

The Tool System is the capability layer of RouterPilot.

Every operation performed by the Runtime is implemented as a Tool.

The Runtime itself contains no networking logic.

Instead, it orchestrates Tools.

This separation keeps the Runtime generic while allowing capabilities to evolve independently.

---

# Philosophy

RouterPilot follows one simple rule:

> **Everything that performs work is a Tool.**

Examples:

* read interface state
* restart DNS
* reload firewall
* scan Wi-Fi
* run traceroute
* ping gateway
* install package

Even composite operations are decomposed into multiple Tools.

---

# Goals

The Tool System is designed to be:

* deterministic
* discoverable
* type-safe
* self-describing
* testable
* composable
* language-agnostic

---

# Tool Lifecycle

Every Tool follows the same lifecycle.

```text
Requested
    │
    ▼
Validate Arguments
    │
    ▼
Permission Check
    │
    ▼
Capability Check
    │
    ▼
Execute
    │
    ▼
Return Result
```

Failures may emit rollback instructions.

---

# Tool Definition

Every Tool exposes metadata.

Example

```yaml
id: network.status

version: 1.0.0

description: Returns interface status

category: network

permissions:

- read

timeout: 5s

supports_dry_run: true

supports_rollback: false
```

The Runtime never needs to inspect implementation details.

---

# Identity

Every Tool has a globally unique identifier.

Recommended format:

```text
domain.action
```

Examples

```text
network.status

network.routes

network.restart

wifi.scan

wifi.connect

wifi.disconnect

dns.lookup

dns.flush

dhcp.clients

system.logs

system.reboot
```

IDs never change after release.

---

# Versioning

Tools are versioned independently.

Example

```text
network.status@1

network.status@2
```

The Planner may request a specific version.

---

# Categories

Tools belong to logical categories.

Examples

Networking

* interfaces
* routes
* bridges

Wireless

* scan
* connect
* disconnect

DNS

* resolve
* cache
* lookup

DHCP

* leases
* renew

Firewall

* inspect
* reload

VPN

* connect
* disconnect

Diagnostics

* ping
* traceroute
* speedtest

System

* reboot
* packages
* logs

---

# Tool Contract

Every Tool implements the same contract.

Input

↓

Validation

↓

Execution

↓

Structured Output

The Runtime treats every Tool identically.

---

# Input Schema

Inputs are strongly typed.

Example

```yaml
host:

type: string

required: true

timeout:

type: integer

default: 5
```

Unknown fields are rejected.

---

# Output Schema

Every Tool returns structured data.

Example

```yaml
success: true

latency: 12

packet_loss: 0
```

Human-readable formatting is handled later.

---

# Error Model

Errors are typed.

Example

```yaml
INVALID_ARGUMENT

PERMISSION_DENIED

TIMEOUT

NOT_SUPPORTED

NETWORK_FAILURE

EXECUTION_FAILED
```

Errors never contain ambiguous strings.

---

# Permissions

Every Tool declares required permissions.

Examples

Read

```text
network.status

dns.lookup

logs.read
```

Write

```text
network.restart

wifi.connect

firewall.reload
```

Administrative

```text
factory.reset

firmware.flash
```

The Runtime enforces permissions before execution.

---

# Capabilities

Some Tools require platform capabilities.

Example

```yaml
requires:

- dnsmasq

- iproute2

- firewall4
```

Unavailable capabilities prevent execution.

---

# Dry Run

Whenever possible, Tools support dry-run mode.

Example

```text
Apply Firewall

↓

Validate Rules

↓

Return Changes

↓

Do Not Apply
```

Dry-run improves safety and planning accuracy.

---

# Rollback

Tools may expose rollback operations.

Example

```text
Update Configuration

↓

Backup

↓

Apply

↓

Failure

↓

Restore
```

Rollback is optional but recommended for state-changing operations.

---

# Idempotency

Tools should be idempotent whenever practical.

Example

```text
Enable Interface

↓

Already Enabled

↓

Success
```

Repeated execution should not introduce inconsistent state.

---

# Side Effects

Every Tool documents side effects.

Example

```yaml
side_effects:

- restart dns

- reload firewall
```

The Planner can use this information to optimize execution order.

---

# Dependencies

Tools may depend on other Tools.

Example

```text
wifi.connect

↓

network.interface.up

↓

dhcp.renew
```

The Runtime resolves dependencies before execution.

---

# Timeouts

Every Tool defines a timeout.

Example

```yaml
timeout: 10s
```

Long-running Tools must periodically report progress.

---

# Progress Reporting

Tools may emit progress events.

Example

```text
10%

25%

60%

100%
```

This enables responsive user interfaces.

---

# Tool Events

Typical events

```text
tool.requested

tool.validated

tool.started

tool.progress

tool.completed

tool.failed
```

Events are consumed by the Runtime and plugins.

---

# Discovery

Tools are registered dynamically.

The Runtime discovers available Tools during startup.

Each Tool publishes:

* metadata
* schema
* permissions
* capabilities
* version

The Planner does not require hardcoded knowledge.

---

# Composite Tools

Some operations consist of multiple Tools.

Example

```text
Diagnose Internet

↓

Check Interface

↓

Check Gateway

↓

Check DNS

↓

Ping Internet

↓

Generate Report
```

Composite Tools are orchestrated by the Planner rather than embedding execution logic.

---

# Platform Independence

The same Tool interface may have multiple implementations.

Example

```text
network.status

↓

OpenWrt

↓

Linux

↓

Docker
```

The Planner uses the logical Tool, while the Runtime selects the platform-specific implementation.

---

# Testing

Every Tool should support:

* unit tests
* integration tests
* schema validation
* dry-run verification
* failure simulation

Behavior should be deterministic.

---

# SDK Integration

A Tool implementation consists of:

* metadata
* input schema
* output schema
* executor
* validator
* optional rollback handler

The SDK provides helper abstractions to reduce boilerplate.

---

# Design Principles

A Tool should:

* do one thing well
* have a clear input
* produce structured output
* expose deterministic behavior
* document side effects
* avoid hidden state

Large workflows should be composed from many small Tools rather than implemented as a single monolithic Tool.

---

# Future Evolution

Planned enhancements include:

* remote tool execution
* distributed runtimes
* capability negotiation
* signed third-party tools
* semantic tool discovery
* execution cost estimation
* caching of pure read-only tools
* AI-assisted tool selection

---

# Tool Philosophy

Tools are the fundamental building blocks of RouterPilot.

The Planner reasons about Tools.

The Runtime executes Tools.

The SDK defines Tools.

Plugins contribute Tools.

Everything else in the system exists to support this capability model.
