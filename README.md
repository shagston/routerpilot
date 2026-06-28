# RouterPilot

> **AI-native runtime for deterministic network automation.**

RouterPilot transforms intent into deterministic, auditable, and safe network operations.  
Separates **planning** (what should happen) from **execution** (how it happens safely).

**30 built-in tools** · REST API · WebSocket · Web UI · Telegram Bot · Plugin system · OpenWrt-ready

---

## Quick Start

```bash
# List all 30 tools
go run .\cmd\routerpilot tools

# Ping a host
go run .\cmd\routerpilot ping 8.8.8.8

# Execute an intent (auto-planned)
go run .\cmd\routerpilot plan dns.lookup google.com

# Start REST API + Web UI + WebSocket
go run .\cmd\routerpilot serve

# Start Telegram bot
set ROUTERPILOT_TELEGRAM_TOKEN=xxx
go run .\cmd\routerpilot telegram
```

---

## Architecture

```text
CLI / API / WebSocket / Telegram / Web UI
           │
           ▼
     ┌─────────────┐
     │   Planner    │  (Simple or LLM)
     └──────┬──────┘
            │ Plan
     ┌──────▼──────┐
     │   Runtime   │  (DAG scheduler, retry, timeout, dry-run)
     └──────┬──────┘
            │ Tasks
     ┌──────▼──────┐
     │ Tool Registry│── 30 tools
     └──────┬──────┘
            │ Execute
     ┌──────▼──────┐
     │  Network /  │
     │   System    │
     └─────────────┘
```

### Components

| Component      | Responsibility                                |
| -------------- | --------------------------------------------- |
| Planner        | Transform intent into executable Plans        |
| Runtime        | Execute validated Plans deterministically     |
| Tools          | 30 individual operations                      |
| Registry       | Discover components and capabilities          |
| Context Engine | Build minimal planning context                |
| Memory         | Store persistent knowledge                    |
| Event Bus      | Publish immutable runtime events              |
| Safety Layer   | Enforce permissions and risk policies         |
| Plugin System  | Extend via external binaries                  |

---

## 30 Tools

### Network Diagnostics

| Tool | Description | Example |
|------|-------------|---------|
| `network.ping` | ICMP echo to host | `routerpilot plan ping 8.8.8.8` |
| `network.traceroute` | Trace network path | `routerpilot plan network.traceroute target=google.com` |
| `network.neighbors` | ARP/neighbor table (`ip neigh`) | `routerpilot plan network.neighbors` |
| `network.connections` | Active sockets (`ss -tuln`) | `routerpilot plan network.connections` |

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
| `system.memory` | RAM usage (free/meminfo) | `routerpilot plan system.memory` |
| `system.disk` | Disk usage (df -h) | `routerpilot plan system.disk` |
| `system.processes` | Top processes (ps aux) | `routerpilot plan system.processes sort=mem limit=10` |
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

### DHCP

| Tool | Description | Example |
|------|-------------|---------|
| `dhcp.leases` | List active DHCP leases | `routerpilot plan dhcp.leases` |

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

### Packages

| Tool | Description | Example |
|------|-------------|---------|
| `package.list` | List installed packages | `routerpilot plan package.list` |

### VPN

| Tool | Description | Example |
|------|-------------|---------|
| `vpn.status` | Show VPN tunnels (WireGuard, OpenVPN) | `routerpilot plan vpn.status` |

### Bridge

| Tool | Description | Example |
|------|-------------|---------|
| `bridge.status` | Show bridge interfaces | `routerpilot plan bridge.status` |

### Diagnostics

| Tool | Description | Example |
|------|-------------|---------|
| `diagnose` | Full network diagnostic (if+ip+route+ping) | `routerpilot plan diagnose target=8.8.8.8` |

---

## Entry Points

### CLI

```bash
routerpilot tools              # List all 30 tools
routerpilot ping <host> [n]    # Quick ping (skips planner)
routerpilot plan <intent>      # Execute via planner
routerpilot serve              # REST API + Web UI + WebSocket
routerpilot telegram           # Telegram bot
```

### REST API

Start with `routerpilot serve`. Endpoints:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Web UI |
| GET | `/api` | API info |
| GET | `/health` | Health check |
| POST | `/intent` | Execute intent |
| POST | `/plan` | Preview plan |
| GET | `/tools` | List tools |
| GET | `/status` | Server status |
| GET | `/events` | Execution events |
| GET | `/events/stream` | SSE event stream |
| GET | `/ws` | WebSocket |

Example:
```bash
curl -X POST http://localhost:8080/intent \
  -H "Content-Type: application/json" \
  -d '{"intent":"dns.lookup","args":{"host":"google.com"}}'
```

### WebSocket

Connect to `ws://localhost:8080/ws`, send JSON:

```json
{"intent": "ping", "args": {"target": "8.8.8.8"}}
```

Receives:
```json
{"type": "status", "data": "executing ping..."}
{"type": "result", "data": {"result": {...}}}
```

### Web UI

Open `http://localhost:8080/` in browser.  
Includes sidebar with quick commands, input field, and output panel.

### Telegram Bot

```bash
set ROUTERPILOT_TELEGRAM_TOKEN=your_bot_token
routerpilot telegram
```

Commands:
- `/ping <host>` — Ping a host
- `/plan <intent>` — Execute an intent
- `/tools` — List tools
- `<intent>` — Execute any intent directly

---

## Configuration

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `ROUTERPILOT_PORT` | `:8080` | API server port |
| `ROUTERPILOT_PERMISSIONS` | `read,write` | Allowed permission scopes |
| `ROUTERPILOT_RISK` | `low` | Max allowed risk level |
| `ROUTERPILOT_PLANNER` | `simple` | Planner type (`simple` or `llm`) |
| `ROUTERPILOT_API_KEY` | — | API key for LLM planner |
| `ROUTERPILOT_PLUGIN_DIR` | `plugins` | Plugin directory |
| `ROUTERPILOT_TELEGRAM_TOKEN` | — | Telegram bot token |
| `ROUTERPILOT_LOG_FORMAT` | `text` | Log format (`text` or `json`) |
| `ROUTERPILOT_LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |

---

## Plugin System

RouterPilot loads external tools as subprocess plugins.

### Creating a Plugin

See `examples/dns-plugin/` for a complete example.

```go
// plugins/myplugin/main.go
package main

import (
    "encoding/json"
    "fmt"
    "os"
)

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

---

## OpenWrt

RouterPilot targets OpenWrt as a primary platform. The Linux network provider  
(`internal/network/linux.go`) uses JSON mode (`ip -j`) with automatic fallback  
to text parsing for busybox `iproute2`.

Supported OpenWrt tools:
- `iwinfo` / `iw` — Wi-Fi scan and status
- `logread` — System logs
- `/etc/init.d/` — Service management
- `ubus` — DHCP leases
- `opkg` — Package management

---

## Development

```bash
go build ./...
go vet ./...
go test ./...
```

### Project Structure

```
cmd/routerpilot/         # CLI entry point
internal/
  api/                   # REST API + WebSocket
  app/                   # Application bootstrap
  context/               # Context engine
  events/                # Event bus
  logger/                # Structured logging
  memory/                # Memory provider
  network/               # Linux network provider
  planner/               # Simple + LLM planner
  plugin/                # Plugin loader
  registry/              # Tool registry
  runtime/               # Execution engine
  safety/                # Safety guards
  telegram/              # Telegram bot
  webui/                 # Embedded Web UI
sdk/                     # Public interfaces
tools/
  bridge/                # bridge.status
  dhcp/                  # dhcp.leases
  dns/                   # dns.lookup, dns.status, dns.flush
  firewall/              # firewall.status, firewall.reload
  network/               # 10 network tools
  package/               # package.list
  service/               # service.list, service.restart
  system/                # 7 system tools
  vpn/                   # vpn.status
  wifi/                  # wifi.scan, wifi.status
web/                     # Web UI (moved to internal/webui/)
```

---

## License

MIT
