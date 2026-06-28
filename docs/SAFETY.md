# SAFETY.md

> RouterPilot Safety Model
>
> Version: 0.1
>
> Status: Draft

---

# Overview

The Safety System is responsible for ensuring that every operation performed by RouterPilot is deterministic, verifiable, and policy-compliant.

Safety is enforced by the Runtime.

The Planner may recommend actions.

The Runtime decides whether those actions are permitted.

---

# Design Goals

The Safety System should be:

* deterministic
* explicit
* auditable
* configurable
* platform independent

Safety rules must never depend on LLM output.

---

# Philosophy

RouterPilot follows one fundamental rule.

> **AI proposes. The Runtime disposes.**

Every action must pass through the Safety Layer before execution.

---

# Safety Pipeline

```text id="d17aw1"
Execution Plan
        │
        ▼
Schema Validation
        │
        ▼
Permission Check
        │
        ▼
Capability Check
        │
        ▼
Policy Evaluation
        │
        ▼
Risk Classification
        │
        ▼
Approval
        │
        ▼
Runtime Execution
```

If any stage fails, execution stops.

---

# Validation Layers

## Schema Validation

Verify:

* required arguments
* types
* ranges
* unknown fields

Invalid requests never reach the Runtime.

---

## Permission Validation

Every Tool declares required permissions.

Example

```text id="xy7a0v"
dns.lookup

↓

read
```

```text id="a5v0mb"
system.reboot

↓

admin
```

Permission checks occur before Tool execution.

---

## Capability Validation

Verify that required platform capabilities exist.

Example

```text id="ndy1kz"
wireguard.connect

↓

WireGuard Installed?
```

Missing capabilities produce validation errors.

---

# Risk Classification

Each Tool defines a default risk level.

| Level    | Description              | Example           |
| -------- | ------------------------ | ----------------- |
| Low      | Read-only                | `network.status`  |
| Medium   | Reversible configuration | `dns.reload`      |
| High     | Service interruption     | `network.restart` |
| Critical | Potential data loss      | `factory.reset`   |

Risk is metadata, not behavior.

---

# Confirmation Policies

Certain operations require explicit confirmation.

Examples

Always require confirmation:

* factory reset
* firmware flash
* storage erase
* credential removal

Read-only operations never require confirmation.

---

# Policy Engine

Policies determine whether an action is allowed.

Example

```yaml id="s6kncp"
allow:

network.status

deny:

factory.reset
```

Policies are evaluated after validation and before execution.

---

# Policy Sources

Policies may originate from:

* system defaults
* administrator configuration
* organization policies
* plugins

Policies are merged deterministically.

---

# Least Privilege

Tools should request only the permissions they need.

Bad

```text id="jbv6my"
admin
```

Good

```text id="z1tn8m"
network.read
```

Smaller permission scopes improve safety.

---

# Dry Run

Whenever supported, state-changing Tools should expose a dry-run mode.

Example

```text id="j4plgw"
Firewall Apply

↓

Validate Rules

↓

Return Diff

↓

Do Not Apply
```

Dry-run is recommended before executing high-risk operations.

---

# Rollback

State-changing Tools should define rollback behavior whenever practical.

Example

```text id="q8nrmx"
Backup

↓

Apply

↓

Verification

↓

Rollback
```

Rollback is coordinated by the Runtime.

---

# Secrets

Secrets require special handling.

Examples

* passwords
* private keys
* API tokens
* VPN credentials

Rules

* never log secrets
* never emit secrets in Events
* never include secrets in Planner context unless required
* redact secrets in error messages

---

# Audit Trail

Every state-changing operation should produce an audit record.

Example

```yaml id="y0bv1l"
execution_id

tool

arguments

user

timestamp

result
```

Audit records must be immutable.

---

# Failure Handling

Safety failures stop execution immediately.

Examples

* invalid schema
* missing permissions
* denied policy
* unsupported capability

These failures do not trigger retries.

---

# Safety Events

Examples

```text id="x7mqlp"
safety.validation.started

safety.validation.failed

policy.denied

confirmation.required

execution.approved
```

Consumers include:

* audit logs
* telemetry
* administration UI

---

# Platform Isolation

The Safety Layer is platform-independent.

OpenWrt-specific checks belong to capability providers, not the Safety Engine.

---

# Configuration

Safety behavior should be configurable.

Examples

```yaml id="z9cpm4"
confirmation:

high_risk: true

critical: always

dry_run:

enabled: true
```

Configuration changes should not require code changes.

---

# Testing

Safety rules require comprehensive testing.

Recommended coverage:

* schema validation
* permission checks
* policy evaluation
* confirmation flow
* rollback triggers
* secret redaction

Every rule should have deterministic test cases.

---

# Future Evolution

Potential enhancements include:

* policy scripting
* signed policy bundles
* organization-wide policy distribution
* execution approval workflows
* risk scoring based on historical failures
* multi-user authorization

These additions should preserve the core validation pipeline.

---

# Safety Philosophy

Safety is enforced by architecture, not by prompts.

The Planner may generate intelligent plans, but only the Runtime has authority to execute them.

By separating reasoning from authorization and enforcing deterministic validation at every stage, RouterPilot minimizes the risk of unintended or unsafe operations while remaining predictable, auditable, and suitable for production networking environments.
