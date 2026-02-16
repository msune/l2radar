<p align="left">
  <a href="https://claude.ai/claude-code">
    <img src="https://img.shields.io/badge/Built%20with-Claude%20Code-blueviolet?logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0id2hpdGUiPjxwYXRoIGQ9Ik0xMiAyQzYuNDggMiAyIDYuNDggMiAxMnM0LjQ4IDEwIDEwIDEwIDEwLTQuNDggMTAtMTBTMTcuNTIgMiAxMiAyem0wIDE4Yy00LjQxIDAtOC0zLjU5LTgtOHMzLjU5LTggOC04IDggMy41OSA4IDgtMy41OSA4LTggOHoiLz48L3N2Zz4=" alt="Built with Claude Code" height="28">
  </a>
</p>

<p align="center">
  <img src="assets/img/logo.png" alt="L2 Radar" width="400">
</p>

<p align="center">
  <em>🤖 Generated (mostly) by <a href="https://claude.ai/claude-code">Claude Code</a> · Directed & reviewed by a human 🧑‍💻</em>
</p>

<p align="center">
  <a href="https://github.com/msune/l2radar/actions"><img src="https://github.com/msune/l2radar/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
</p>

---

# 📡 L2 Radar

**Passive L2 neighbour monitor powered by eBPF.** See every device on your
network — MACs, IPs, vendors — without sending a single packet.

L2 Radar attaches eBPF probes to your network interfaces via
[TCX ingress](https://docs.kernel.org/bpf/), silently observes ARP and NDP
traffic, and presents everything in a slick dark-themed dashboard.

## ✨ Features

- 🐝 **eBPF-powered** — zero packet injection, zero interference, zero overhead
- 🔍 **ARP + NDP parsing** — discovers IPv4 and IPv6 neighbours automatically
- 🏭 **OUI vendor lookup** — resolves MAC addresses to manufacturer names
- 🌐 **Web dashboard** — real-time, searchable, sortable, mobile-friendly
- 🔒 **HTTPS + auth** — TLS and basic auth out of the box
- 📦 **Docker-based** — two containers, one command to run

## 🚀 Quick Start

**1. Install `l2rctl`:**

```bash
curl -fsSL https://raw.githubusercontent.com/msune/l2radar/main/scripts/install-l2rctl.sh | sh
```

**2. Start everything:**

```bash
l2rctl start --user admin:changeme
```

**3. Open the dashboard:**

👉 **https://localhost** (accept the self-signed cert)

That's it! L2 Radar is now watching all your external interfaces. 🎉

## 📖 Usage

```bash
# Start only the probe (headless)
l2rctl start probe --iface eth0 --iface wlan0

# Start with custom TLS certs
l2rctl start --tls-dir /etc/mycerts --user admin:secret

# Check what's running
l2rctl status

# Dump the neighbour table from the terminal
l2rctl dump --iface eth0

# Stop everything
l2rctl stop
```

### Interface Keywords

| Keyword | Meaning |
|---------|---------|
| `any` (default) | All external interfaces (skips docker, veth, bridges) |
| `all` | Every non-loopback interface |

## 🏗️ Architecture

L2 Radar has three components:

| Component | Container | What it does |
|-----------|-----------|-------------|
| **eBPF Probe** | `l2radar` | Attaches to NICs, writes neighbour data to BPF maps, exports JSON |
| **Web UI** | `l2radar-ui` | nginx + React dashboard, serves JSON data read-only |
| **l2rctl** | _(host binary)_ | Orchestrates the containers via Docker CLI |

The probe and UI communicate through **JSON files on a shared volume** — no
network calls between them.

```
 ┌──────────────────────┐        /tmp/l2radar/         ┌──────────────────────┐
 │     eBPF Probe       │   neigh-eth0.json            │       Web UI         │
 │                      │──────────────────────────────▶│                      │
 │  TCX ingress hooks   │   neigh-wlan0.json           │  nginx + React SPA   │
 │  ARP/NDP parsing     │──────────────────────────────▶│  auto-refresh polls  │
 │  JSON export loop    │         (read-only)           │  OUI vendor lookup   │
 └──────────────────────┘                               └──────────────────────┘
        privileged                                          ports 443 (80)
        --network=host                                      unprivileged
```

📚 **[Full architecture docs →](docs/architecture.md)**

## 📋 Requirements

- Linux with kernel **6.6+** (for TCX)
- Docker
- `curl` or `wget` (for install script)

## 🛠️ Development

```bash
# Probe tests (requires BPF-capable kernel)
cd probe && go test -v ./...

# UI tests
cd ui && npm test

# l2rctl tests
cd cmd/l2rctl && go test -v ./...
```

## 📄 License

See [LICENSE](LICENSE).

---

<p align="center">
  <sub>Made with ❤️ from Barcelona · Powered by 🐝 eBPF</sub>
</p>
