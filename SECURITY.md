# Security Policy

## Supported Versions

| Version | Supported          |
|---------|-------------------|
| 0.x     | No (pre-release)  |
| 1.0.x   | Yes               |

## Reporting a Vulnerability

Report vulnerabilities to the maintainers by opening a security advisory on GitHub. Do not disclose security issues publicly until they have been addressed.

## Security Principles

RouterPilot follows these security principles:

1. **Default Deny** - All actions require explicit authorization
2. **Least Privilege** - Agents and plugins get minimum required permissions
3. **Defense in Depth** - Multiple independent security layers
4. **Fail Closed** - When authorization cannot be determined, deny
5. **Separation of Concerns** - Policy Engine is the sole authorization authority

## Security Architecture

- **Policy Engine**: Centralizes authorization decisions for all capabilities
- **Capability Authorization**: Each capability execution is independently authorized
- **Plugin Isolation**: Plugins interact only through the public SDK
- **Transport Security**: Transport layer provides encryption and authentication
- **Audit Logging**: All security-relevant operations emit events

## Security Boundaries

- Runtime boundary: Local Runtime <-> Plugin
- Transport boundary: Runtime <-> Remote Peer  
- Memory boundary: Runtime <-> Memory Provider
- Execution boundary: Agent <-> Capability Provider

## Disclosure Policy

Vulnerabilities are disclosed after a fix is released and all supported versions are updated. Security advisories are published on the GitHub repository.

## Related Documents

- `docs/architecture/120_SECURITY.md` - Detailed security architecture
- `docs/architecture/50_POLICY_ENGINE.md` - Policy engine specification
- `docs/architecture/51_PERMISSION_MODEL.md` - Permission model
