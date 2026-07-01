# RouterPilot

> **AI-native runtime for deterministic network automation.**

RouterPilot transforms natural language intent into auditable, safe network operations.  
Separates **planning** (what should happen) from **execution** (how it happens safely).

**33 built-in tools** · REST API · WebSocket · Web UI · Telegram Bot · Plugin system · OpenWrt-ready  
**Zero external dependencies** — only Go standard library.

---

## Quick Start

```bash
# List all 33 tools
go run .\cmd\routerpilot tools

# Ping a host
go run .\cmd\routerpilot ping 8.8.8.8

# Execute an intent (auto-planned)
go run .\cmd\routerpilot plan dns.lookup google.com

# Start REST API + Web UI + WebSocket
go run .\cmd\routerpilot serve

# Start Telegram bot
$env:ROUTERPILOT_TELEGRAM_TOKEN="xxx"
go run .\cmd\routerpilot telegram
```

Or use the prebuilt binary:
```bash
routerpilot serve
curl -X POST http://localhost:8080/intent \
  -H "Content-Type: application/json" \
  -d '{"intent":"dns.lookup","args":{"host":"google.com"}}'
```

---

## Architecture

```
CLI / API / WebSocket / Telegram / Web UI
           │
           ▼
     ┌─────────────┐
     │   Planner    │  Simple (rule-based, ~30 intents) or LLM (OpenAI-compatible)
     └──────┬──────┘
            │ Plan
     ┌──────▼──────┐
     │    Safety   │  Risk levels, permissions, schema validation, dry-run
     └──────┬──────┘
            │
     ┌──────▼──────┐
     │   Runtime   │  DAG scheduler, adaptive re-planning, retry, timeout
     └──────┬──────┘
            │ Tasks
     ┌──────▼──────┐
     │ Tool Registry│── 33 tools across 11 categories
     └──────┬──────┘
            │ Execute
     ┌──────▼──────┐
     │  Network /  │
     │   System    │
     └─────────────┘
```

### Components

| Component      | Responsibility                                       |
| -------------- | ---------------------------------------------------- |
| Planner        | Transform intent into executable Plans               |
| Context Engine | Gather minimal system state for adaptive re-planning |
| Safety Guard   | Enforce risk levels, permissions, and policies       |
| Runtime        | Execute validated Plans deterministically            |
| Tools          | 33 individual operations (network, system, DNS, ...) |
| Registry       | Discover components and capabilities                 |
| Event Bus      | Publish immutable runtime events                     |
| Memory         | Store persistent knowledge                           |
| Knowledge Base | Rule-based diagnostics (11 patterns, no LLM needed)  |
| Plugin System  | Extend via external binary subprocesses              |

---

## Entry Points

### CLI

| Command | Description |
|---------|-------------|
| `routerpilot tools` | List all registered tools |
| `routerpilot ping <host> [count] [--events]` | Quick ping (skips planner, goes directly to runtime) |
| `routerpilot plan <intent> [args...]` | Full intent execution via planner |
| `routerpilot serve` | Start REST API + Web UI + WebSocket server |
| `routerpilot telegram` | Start Telegram bot |

### REST API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Web UI |
| GET | `/api` | API info |
| GET | `/api/config` | Get runtime config |
| PUT | `/api/config` | Update runtime config |
| GET | `/health` | Health check |
| POST | `/intent` | Execute intent |
| POST | `/plan` | Preview plan (no execution) |
| GET | `/tools` | List tools |
| GET | `/status` | Server status |
| GET | `/events` | Execution events |
| GET | `/events/stream` | SSE event stream |
| GET | `/ws` | WebSocket |

### WebSocket

Connect to `ws://localhost:8080/ws`:

```json
{"intent": "ping", "args": {"target": "8.8.8.8"}}
```

Receives:
```json
{"type": "status", "data": "executing ping..."}
{"type": "result", "data": {"result": {...}}}
```

### Web UI

Open `http://localhost:8080/` — dark-themed SPA with console, dashboard, and settings tabs.

### Telegram Bot

Commands: `/ping <host>`, `/plan <intent>`, `/tools`, or any intent directly.

---

## 33 Tools

### Network Diagnostics

| Tool | Description | Example |
|------|-------------|---------|
| `network.ping` | ICMP echo to host | `routerpilot plan ping 8.8.8.8` |
| `network.traceroute` | Trace network path | `routerpilot plan network.traceroute target=google.com` |
| `network.neighbors` | ARP/neighbor table (`ip neigh`) | `routerpilot plan network.neighbors` |
| `network.connections` | Active sockets (`ss -tuln`) | `routerpilot plan network.connections` |
| `network.bandwidth` | Measure throughput (iperf3/curl) | `routerpilot plan network.bandwidth target=10.0.0.1` |

### Interface Management

| Tool | Description | Example |
|------|-------------|---------|
| `network.interface.status` | Interface state and statistics | `routerpilot plan interface.status` |
| `network.interface.set` | Bring interface up/down | `routerpilot plan interface.set interface=eth0 state=down` |
| `network.ip_address_get` | IP addresses on interfaces | `routerpilot plan ip.show` |
| `network.ip_address_set` | Assign IP to interface | `routerpilot plan ip.set interface=eth0 address=192.168.1.1/24` |

### Routing

| Tool | Description | Example |
|------|-------------|---------|
| `network.route_get` | Routing table | `routerpilot plan route.show` |
| `network.route_add` | Add a route | `routerpilot plan route.add destination=10.0.0.0/24 gateway=192.168.1.1` |

### System

| Tool | Description | Example |
|------|-------------|---------|
| `system.info` | OS, kernel, hostname, arch | `routerpilot plan system.info` |
| `system.uptime` | System uptime | `routerpilot plan system.uptime` |
| `system.memory` | RAM usage | `routerpilot plan system.memory` |
| `system.disk` | Disk usage (`df -h`) | `routerpilot plan system.disk` |
| `system.processes` | Top processes (`ps aux`) | `routerpilot plan system.processes sort=mem limit=10` |
| `system.logs` | View logs (journalctl/logread/dmesg) | `routerpilot plan system.logs lines=100 filter=error` |
| `system.reboot` | Reboot the system | `routerpilot plan system.reboot` |

### DNS

| Tool | Description | Example |
|------|-------------|---------|
| `dns.lookup` | Resolve hostname to IP | `routerpilot plan dns.lookup google.com` |
| `dns.status` | Show resolver config | `routerpilot plan dns.status` |
| `dns.flush` | Flush DNS cache | `routerpilot plan dns.flush` |

### Wi-Fi (OpenWrt)

| Tool | Description | Example |
|------|-------------|---------|
| `wifi.scan` | Scan for access points | `routerpilot plan wifi.scan` |
| `wifi.status` | Wi-Fi interface status | `routerpilot plan wifi.status` |
| `wifi.connect` | Connect to a Wi-Fi network | `routerpilot plan wifi.connect ssid=MyNet password=secret` |

### DHCP

| Tool | Description | Example |
|------|-------------|---------|
| `dhcp.leases` | List active DHCP leases | `routerpilot plan dhcp.leases` |
| `dhcp.server` | Show DHCP server config | `routerpilot plan dhcp.server` |

### Firewall

| Tool | Description | Example |
|------|-------------|---------|
| `firewall.status` | Show firewall rules | `routerpilot plan firewall.status` |
| `firewall.reload` | Reload firewall | `routerpilot plan firewall.reload` |

### Services

| Tool | Description | Example |
|------|-------------|---------|
| `service.list` | List services | `routerpilot plan service.list` |
| `service.restart` | Restart a service | `routerpilot plan service.restart name=dnsmasq` |

### Packages, VPN, Bridge

| Tool | Description | Example |
|------|-------------|---------|
| `package.list` | List installed packages | `routerpilot plan package.list` |
| `vpn.status` | Show VPN tunnels (WireGuard, OpenVPN) | `routerpilot plan vpn.status` |
| `bridge.status` | Show bridge interfaces | `routerpilot plan bridge.status` |

### Diagnostics

| Tool | Description | Example |
|------|-------------|---------|
| `diagnose` | Full network diagnostic (if+ip+route+ping) | `routerpilot plan diagnose target=8.8.8.8` |
| `suggest` | Offline KB-based diagnosis (11 patterns) | `routerpilot plan suggest problem="no internet"` |

---

## Safety

RouterPilot's safety is **deterministic** — it never depends on LLM output.

| Layer | Mechanism |
|-------|-----------|
| **Risk Levels** | Each tool has a risk rating: Low, Medium, High, Critical |
| **Permissions** | Scopes: Read, Write, Admin. Configurable at startup |
| **Schema Validation** | All tool inputs validated against declared schemas |
| **Dry-Run Mode** | Preview execution without side effects |
| **Read-Only Mode** | Blocks all write operations |
| **Plan Validation** | Cycle detection, dependency resolution, risk coercion |

---

## Adaptive Re-Planning

RouterPilot can gather system context, merge results, and re-feed to the planner for a refined plan:

1. Runtime detects context-gathering tasks (read-only, no cross-dependencies)
2. Executes them first
3. Merges results into enriched context
4. Re-plans with full system state
5. Falls back to original plan after 3 retries

This enables tools like `diagnose` to run a series of checks and `suggest` to apply KB pattern matching on live data — all without an LLM.

---

## Plugin System

Plugins are standalone executables communicating via stdin/stdout JSON:

```go
func main() {
    if len(os.Args) > 1 && os.Args[1] == "plugin-manifest" {
        json.NewEncoder(os.Stdout).Encode(map[string]any{
            "id": "myplugin.tool",
            "version": "0.1.0",
            "description": "My custom tool",
        })
        return
    }
    // Execute tool logic...
}
```

Place the binary in `plugins/` (or `$ROUTERPILOT_PLUGIN_DIR`).  
See `examples/dns-plugin/` for a complete example.

---

## OpenWrt

RouterPilot targets OpenWrt as a primary platform:

- **Network provider** uses JSON mode (`ip -j`) with automatic fallback to text parsing for busybox `iproute2`
- **Tools** leverage OpenWrt-native utilities: `iwinfo`/`iw`, `logread`, `/etc/init.d/`, `ubus`, `opkg`
- **LuCI package** (`luci-app-routerpilot/`): 4 JS views, Lua controller, CBI settings, init.d script
- **Install** via `install.sh` or as an OpenWrt package

---

## Configuration

### File (`routerpilot.json`)

```json
{
  "server": { "port": ":8080" },
  "planner": { "type": "simple", "model": "gpt-4" },
  "security": { "risk": "low", "permissions": ["read", "write"] }
}
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ROUTERPILOT_PORT` | `:8080` | API server port |
| `ROUTERPILOT_PERMISSIONS` | `read,write` | Allowed permission scopes |
| `ROUTERPILOT_RISK` | `low` | Max allowed risk level |
| `ROUTERPILOT_PLANNER` | `simple` | Planner type (`simple` or `llm`) |
| `ROUTERPILOT_API_KEY` | — | API key for LLM planner |
| `ROUTERPILOT_PLUGIN_DIR` | `plugins` | Plugin directory |
| `ROUTERPILOT_TELEGRAM_TOKEN` | — | Telegram bot token |
| `ROUTERPILOT_LOG_FORMAT` | `text` | Log format (`text` or `json`) |
| `ROUTERPILOT_LOG_LEVEL` | `info` | Log level |

Config can also be updated at runtime via `PUT /api/config`.

---

## Development

```bash
go build ./...
go vet ./...
go test ./...
```

### Project Structure

```
cmd/routerpilot/         # CLI entry point (5 commands)
internal/
  api/                   # REST API + WebSocket + SSE
  app/                   # Application bootstrap, dynamic context
  config/                # Config loading (JSON + env overrides)
  context/               # Context engine
  events/                # In-memory event bus
  kb/                    # Diagnostic knowledge base (11 patterns)
  logger/                # Structured logging (slog wrapper)
  memory/                # In-memory key-value store
  network/               # OS network abstraction (Linux + mock)
  planner/               # Simple + LLM planner
  plugin/                # Plugin loader (subprocess)
  registry/              # Thread-safe tool registry
  runtime/               # Execution engine (DAG scheduling)
  safety/                # Safety guard + validator
  telegram/              # Telegram bot (long-polling)
  webui/                 # Embedded Web UI SPA
sdk/                     # Public interfaces (tool, planner, runtime, events, memory, plugin)
tools/                   # 33 tool implementations
  bridge/ dhcp/ dns/ firewall/ network/ package/ service/ system/ vpn/ wifi/
luci-app-routerpilot/    # OpenWrt LuCI integration
examples/                # Plugin + SDK examples
docs/                    # Architecture, safety, SDK, specs
```

---

## License

MIT
