# 61_MEMORY_PROVIDER.md

> Status: Draft
> Version: 0.1

# Memory Providers

The Memory Provider abstraction separates the Runtime memory model from concrete storage technologies.

Providers implement persistence. The Runtime implements semantics.

---

# Purpose

Memory Providers allow RouterPilot to support multiple storage backends without changing Agent or Runtime behavior.

Supported implementations may include:

- In-Memory
- SQLite
- BadgerDB
- PostgreSQL
- Redis
- S3-compatible object storage
- Vector databases

Providers are interchangeable through the public SDK.

---

# Architecture

```text
             Agent
               │
          Memory API
               │
         Policy Engine
               │
        Memory Provider
               │
      Storage Backend
```

The Runtime never accesses backend-specific APIs directly.

---

# Responsibilities

A Memory Provider MUST:

- store values
- retrieve values
- delete values
- enumerate namespaces
- support versioning metadata
- report errors deterministically

A Provider MUST NOT:

- evaluate permissions
- expose storage internals
- bypass Runtime APIs

---

# Data Model

Each stored object SHOULD contain:

- Namespace
- Key
- Value
- Version
- CreatedAt
- UpdatedAt
- Labels
- Metadata

Providers may add implementation-specific metadata but must preserve the logical model.

---

# Provider Capabilities

Optional capabilities include:

- transactions
- encryption at rest
- compression
- TTL expiration
- optimistic locking
- full-text search
- vector similarity search

Capability discovery SHOULD be explicit.

---

# Suggested Interface

```go
type MemoryProvider interface {
    Get(context.Context, string) (Value, error)
    Put(context.Context, string, Value) error
    Delete(context.Context, string) error
    List(context.Context, Prefix) ([]Value, error)
    Capabilities() ProviderCapabilities
}
```

---

# Transactions

Providers MAY support transactions.

If supported, transactions MUST guarantee atomicity for operations executed within a single transaction scope.

The Runtime MUST remain functional on providers without transactional support.

---

# Concurrency

Providers MUST be safe for concurrent access.

Recommended guarantees:

- concurrent reads
- serialized writes per key
- deterministic conflict handling

---

# Error Categories

Providers SHOULD distinguish:

- NotFound
- AlreadyExists
- ValidationError
- Conflict
- Timeout
- StorageFailure
- PermissionDenied

Errors SHOULD be typed where practical.

---

# Migration

Providers SHOULD expose migration utilities.

Migration SHOULD preserve:

- namespaces
- versions
- metadata
- timestamps

Migration MUST NOT require Agent changes.

---

# Observability

Providers SHOULD emit:

- memory.provider.started
- memory.provider.stopped
- memory.provider.error
- memory.provider.compaction
- memory.provider.migration

Metrics SHOULD include:

- read latency
- write latency
- storage utilization
- object count
- cache hit ratio (if applicable)

---

# Security

Providers SHOULD support:

- encryption at rest
- secure deletion
- integrity verification
- access auditing

Secrets MUST NOT be stored unencrypted unless explicitly configured.

---

# Invariants

- Providers are replaceable.
- Runtime owns authorization.
- Agents never access providers directly.
- Storage format is an implementation detail.
- Provider selection does not change Memory semantics.

---

# Related Documents

- 60_MEMORY.md
- 50_POLICY_ENGINE.md
- 120_SECURITY.md

---

# Next

Continue with **70_SCHEDULER.md**.
