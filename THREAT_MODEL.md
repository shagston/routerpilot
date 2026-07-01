# Threat Model

## Overview

This document describes the primary threats to the RouterPilot Runtime and the mitigations implemented to address them.

## Trust Boundaries

```
[Agent] → [Runtime] → [Policy Engine] → [Capability Provider]
                                      ↘ [Transport] → [Remote Runtime]
                                      ↘ [Plugin SDK] → [Plugin]
                                      ↘ [Memory Provider]
```

Trust boundaries exist at:
1. Runtime ↔ Plugin interface
2. Runtime ↔ Remote Runtime (transport)
3. Runtime ↔ Memory Provider
4. Agent ↔ Capability boundary

## Threats and Mitigations

### T1: Unauthorized Capability Execution

**Description**: An agent or peer executes a capability without proper authorization.

**Mitigations**:
- Policy Engine evaluates every capability request before execution
- Default authorization is Deny
- Capabilities declare required permissions in their metadata
- Agent identity is included in every authorization request

### T2: Malicious Plugin

**Description**: A plugin performs unauthorized operations, accesses private data, or disrupts Runtime operation.

**Mitigations**:
- Plugins interact only through the public SDK
- Plugins cannot import internal Runtime packages
- Plugin initialization goes through a validated lifecycle
- Plugin isolation prevents access to other plugins' state

### T3: Compromised Peer

**Description**: A remote Runtime peer sends malicious capability requests or events.

**Mitigations**:
- Discovery does not imply trust
- Every remote capability request is independently authorized
- Transport provides peer authentication
- Loop prevention in event propagation
- Fault isolation prevents cascade failures

### T4: Replay Attack

**Description**: An attacker intercepts and replays a valid capability request or event.

**Mitigations**:
- Timestamps in protocol envelopes
- Correlation IDs for request/response matching
- Transport-level replay protection (when supported)
- TTL-based event expiration

### T5: Privilege Escalation

**Description**: An agent or plugin gains permissions beyond what was granted.

**Mitigations**:
- Policy Engine is the sole authorization authority
- Permissions are additive and explicitly declared
- Agents have immutable identity
- Policy evaluation includes agent permissions as constraints

### T6: Data Tampering

**Description**: An attacker modifies in-transit or at-rest data.

**Mitigations**:
- Transport encryption (when supported by transport)
- Checksums and signatures in protocol (when supported)
- Memory providers may implement encryption at rest

### T7: Denial of Service

**Description**: An attacker overwhelms the Runtime with requests.

**Mitigations**:
- Concurrency limits in the Scheduler
- Active execution limits in the Runtime
- Queue-based event delivery with backpressure
- Inbox capacity limits in transports

### T8: Information Disclosure

**Description**: Sensitive information is exposed through logs, events, or API responses.

**Mitigations**:
- Secrets are redacted from logs
- Events do not include secret values
- Memory access passes through policy evaluation
- External memory providers may encrypt sensitive values

## Attack Surface

| Component | Attack Vector | Risk Level |
|-----------|--------------|------------|
| Transport | Malformed envelopes, replay, peer spoofing | **High** |
| Plugin SDK | Malicious plugin code | **High** |
| Policy Engine | Bypass attempts | **Critical** |
| Event Bus | Event flooding, unauthorized publish | **Medium** |
| Memory Provider | Data corruption, unauthorized read | **Medium** |
| Scheduler | Resource exhaustion | **Low** |
| API Endpoint | Request flooding, injection | **Medium** |

## Security Controls

| Control | Location | Purpose |
|---------|----------|---------|
| Policy Engine | Runtime | Authorize all capability executions |
| Capability Permissions | Capability metadata | Declare required access |
| Agent Identity | Agent Manager | Identify execution origin |
| Transport Auth | Transport layer | Verify peer identity |
| Plugin SDK boundary | Plugin interface | Isolate plugin code |
| Audit Events | Event Bus | Record security decisions |
| Default Deny | Policy Engine | Fail closed on uncertainty |

## Assumptions

- The local operating system provides basic process isolation
- Cryptographic primitives are implemented by the transport layer
- System administrators configure appropriate policies
- Physical security of Runtime hosts is maintained by operators

## Review Cycle

This threat model should be reviewed when:
- Major architectural changes are introduced
- New transport implementations are added
- Plugin SDK interfaces change
- New capability categories are defined
