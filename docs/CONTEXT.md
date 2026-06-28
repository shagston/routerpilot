# CONTEXT.md

> RouterPilot Context Engine
>
> Version: 0.1
>
> Status: Draft

---

# Overview

The Context Engine is responsible for constructing the minimal, relevant, and deterministic context required for planning.

It acts as the bridge between the live system, persistent memory, and the Planner.

The Context Engine never performs reasoning.

It only selects and structures information.

---

# Design Goals

The Context Engine should be:

* deterministic
* minimal
* reproducible
* modular
* explainable
* cache-friendly

Every planning request should receive only the information necessary to produce a correct execution plan.

---

# Philosophy

The RouterPilot Planner should never receive the complete system state.

Instead:

```text id="u3nq0x"
Entire System

↓

Context Builder

↓

Relevant Context

↓

Planner
```

Smaller context improves:

* latency
* accuracy
* determinism
* token efficiency

---

# Sources

The Context Engine aggregates information from multiple providers.

```text id="n81v4a"
Live State

↓

Memory

↓

Capabilities

↓

Configuration

↓

Execution History

↓

Context
```

Each source is independent.

---

# Context Pipeline

Every planning request follows the same pipeline.

```text id="3k0pwc"
Intent

↓

Determine Required Domains

↓

Collect Data

↓

Normalize

↓

Deduplicate

↓

Context Object
```

---

# Context Domains

Context is organized into domains.

Examples

```text id="ztb6pr"
System

Network

DNS

Firewall

Wi-Fi

VPN

Memory

History

Capabilities
```

Domains are loaded only when required.

---

# Intent Mapping

Every Intent defines the domains it needs.

Example

Intent

```text id="wzq8gh"
dns.lookup
```

Required

```text id="r9lcf6"
DNS

WAN

Capabilities
```

Ignored

```text id="qk1y2v"
Wi-Fi

Firewall

Package Manager
```

---

# Context Providers

Every domain is implemented by a Context Provider.

Example interface

```go id="h8x7mf"
type ContextProvider interface {
    Collect(ContextRequest) (ContextFragment, error)
}
```

Providers remain independent.

---

# Context Fragments

Each provider returns a fragment.

Example

```yaml id="a5p2ns"
domain: dns

resolver: dnsmasq

servers:

- 1.1.1.1

- 9.9.9.9
```

The Context Engine merges fragments into a single Context object.

---

# Normalization

Providers may expose different data formats.

The Context Engine converts them into canonical representations.

Example

```text id="2s0jya"
OpenWrt

↓

Canonical Interface

↓

Planner
```

The Planner never depends on platform-specific formats.

---

# Deduplication

Duplicate information should be removed.

Example

Memory

```text id="x2n7qe"
Gateway: 192.168.1.1
```

Live State

```text id="b6t0fw"
Gateway: 192.168.1.1
```

Result

```text id="y4r1mj"
Gateway: 192.168.1.1
```

---

# Priority

When multiple sources disagree:

```text id="0mvfcb"
Live State

↓

Verified Runtime Output

↓

Persistent Memory

↓

Historical Records
```

The most recent verified information wins.

---

# Context Object

The final Context contains:

```yaml id="g5w9dk"
intent

state

memory

capabilities

configuration

history

constraints
```

The structure remains stable across platforms.

---

# Size Budget

Context size should be controlled.

Recommended defaults

```text id="n2c5uy"
Maximum domains

Maximum records

Maximum history depth

Maximum token budget
```

The Context Engine may truncate low-priority information.

---

# Context Caching

Frequently requested contexts may be cached.

Example

```text id="w7g4jf"
Network Status

↓

Cache

↓

Planner
```

Caches must be invalidated through Runtime Events.

---

# Freshness

Every fragment includes freshness metadata.

Example

```yaml id="z8f6tn"
generated_at:

ttl:
```

Stale fragments should be refreshed before planning.

---

# Explainability

The Context Engine should explain why a fragment was included.

Example

```text id="d3k9rv"
DNS included

Reason:

Intent requires DNS resolution.
```

This improves debugging and planner transparency.

---

# Security

Sensitive information should be filtered before reaching the Planner.

Examples

Remove

* passwords
* private keys
* API tokens

Keep

* capability names
* interface names
* service status

Secrets should remain outside the planning context unless explicitly required.

---

# Testing

The Context Engine should support:

* deterministic outputs
* provider isolation
* cache validation
* freshness handling
* deduplication tests
* size limit enforcement

Identical inputs should produce identical Context objects.

---

# Future Evolution

Potential enhancements include:

* adaptive context selection
* semantic context ranking
* predictive prefetching
* distributed context providers
* cost-aware token budgeting
* context compression strategies

These additions should preserve the public Context Provider interface.

---

# Context Philosophy

The Context Engine exists to answer one question:

> **What is the smallest amount of trustworthy information the Planner needs to make a correct decision?**

By constructing focused, reproducible, and platform-independent Context objects, RouterPilot reduces token usage, improves planning quality, and keeps reasoning isolated from data collection.
