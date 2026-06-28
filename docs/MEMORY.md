# MEMORY.md

> RouterPilot Memory System
>
> Version: 0.1
>
> Status: Draft

---

# Overview

The Memory System provides RouterPilot with persistent knowledge across executions.

Unlike the Context Engine, which represents the current state of the system, Memory stores information that remains useful over time.

Memory is structured, typed, and deterministic.

It is **not** a conversation history.

---

# Design Goals

The Memory System should be:

* deterministic
* structured
* queryable
* persistent
* versioned
* platform independent

Memory should improve planning without becoming a hidden source of truth.

---

# Philosophy

RouterPilot distinguishes between **State**, **Context**, and **Memory**.

```text id="m9r2vx"
Current Router

↓

State

↓

Context Builder

↓

Planner

↓

Memory
```

These concepts serve different purposes.

---

# State vs Context vs Memory

## State

Represents the current system.

Examples:

* current WAN IP
* active interfaces
* connected Wi-Fi clients
* routing table

State changes continuously.

---

## Context

A temporary snapshot built for one execution.

Examples:

* router model
* current DNS
* selected interfaces

Context exists only during planning and execution.

---

## Memory

Persistent knowledge.

Examples:

* preferred DNS provider
* ISP name
* router location
* common network topology
* previous successful configuration

Memory survives restarts.

---

# Memory Categories

The Memory System consists of several logical stores.

```text id="x0bxij"
Identity

Configuration

Preferences

Topology

History

Capabilities

Knowledge
```

Each category has different update rules.

---

# Identity Memory

Stores immutable or rarely changing information.

Examples

```yaml id="5dbjlwm"
device_model

serial_number

board_name

firmware_family

hardware_revision
```

Identity should change only after hardware replacement or firmware migration.

---

# Configuration Memory

Represents persistent configuration choices.

Examples

```yaml id="49i9mj7"
preferred_dns

wan_type

hostname

lan_subnet

ntp_servers
```

Configuration may change through Runtime actions.

---

# Preference Memory

Stores user preferences.

Examples

```yaml id="7w6z6lw"
preferred_language

confirmation_policy

favorite_tools

diagnostic_level
```

Preferences influence planning but never override runtime validation.

---

# Topology Memory

Stores known network topology.

Examples

```yaml id="m6o9c0v"
gateway

switches

access_points

mesh_nodes

subnets
```

Topology evolves gradually over time.

---

# Capability Memory

Caches discovered capabilities.

Examples

```yaml id="l8wpyw5"
dnsmasq

firewall4

wireguard

tailscale

podkop
```

Capability discovery should be refreshed when software changes.

---

# Historical Memory

Stores previous executions.

Examples

```yaml id="g1y6k8a"
successful_plan

failed_plan

execution_duration

failure_reason
```

History enables trend analysis and smarter planning.

---

# Knowledge Memory

Stores semantic information learned from the environment.

Examples

```yaml id="u0clgkn"
ISP maintenance window

known gateway address

observed MTU

stable DNS server
```

Knowledge must remain verifiable.

---

# Memory Object

Each record has common metadata.

```yaml id="kq3jvzi"
id

type

value

created_at

updated_at

source

confidence

version
```

Metadata supports auditing and conflict resolution.

---

# Sources

Memory may originate from:

* Runtime
* User
* Planner
* Discovery Tools
* Plugins
* External APIs

Every record includes its origin.

---

# Confidence

Some memory entries are inferred.

Example

```yaml id="y1ewl2s"
value: "ISP uses CGNAT"

confidence: 0.84
```

Confidence never replaces verification.

The Runtime should always prefer current State when available.

---

# Mutability

Different memory classes have different update policies.

Immutable

* serial number
* hardware revision

Rarely changing

* ISP
* topology

Frequently changing

* preferred diagnostics
* learned capabilities

---

# Read Path

Planning begins by querying Memory.

```text id="u4h1rpb"
Planner

↓

Memory Provider

↓

Relevant Records

↓

Context Builder
```

Only relevant records should be returned.

---

# Write Path

Memory updates occur after successful verification.

```text id="y76v5zt"
Runtime

↓

Verification

↓

Memory Update

↓

Event
```

Failed executions must not update Memory automatically.

---

# Memory Queries

Supported query types include:

* exact lookup
* prefix search
* category search
* execution history
* capability lookup

Queries should return structured objects rather than formatted text.

---

# Versioning

Memory records are versioned.

Example

```yaml id="r8mpm2q"
version: 4
```

Versioning enables conflict detection and future synchronization.

---

# Expiration

Some records become stale.

Examples

* discovered clients
* temporary routes
* transient failures

Such records may include expiration metadata.

```yaml id="tq1k4ea"
expires_at:
```

Expired records should not participate in planning.

---

# Event Integration

Memory changes generate Events.

Examples

```text id="hzcn4yv"
memory.created

memory.updated

memory.deleted

memory.expired
```

This allows telemetry and plugins to react to changes.

---

# Security

Sensitive records should be marked explicitly.

Examples

```yaml id="sx5bjlwm"
classification:

public

internal

secret
```

Secrets should never be included in LLM prompts unless absolutely required.

---

# Storage Backends

The SDK supports multiple implementations.

Examples

```text id="x7q4l6r"
In-Memory

SQLite

BoltDB

PostgreSQL

Redis
```

Backends must expose identical behavior through the Memory Provider interface.

---

# Caching

Frequently accessed records may be cached.

Caches should be transparent to callers.

Cache invalidation occurs through Events.

---

# Backup and Restore

Memory should support export and import.

Possible formats

* JSON
* YAML

This enables migration between devices.

---

# Testing

Memory implementations should support:

* deterministic reads
* concurrent writes
* version conflicts
* expiration handling
* backup/restore
* event generation

Tests should avoid backend-specific assumptions.

---

# Future Evolution

Potential enhancements include:

* encrypted storage
* distributed synchronization
* semantic indexing
* vector search for documentation
* policy-based retention
* multi-device memory replication

These features should extend, not replace, the existing Memory Provider interface.

---

# Memory Philosophy

Memory exists to preserve durable knowledge, not transient state.

The Planner uses Memory to make better decisions.

The Runtime updates Memory only after verified outcomes.

By keeping Memory structured, typed, and separate from live system state, RouterPilot remains predictable, explainable, and resilient while steadily improving its understanding of the managed environment.
