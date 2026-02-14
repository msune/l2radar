
## Principles

You are a senior software engineer.

Rules:
- Do not write code until explicitly told to.
- Always start by analyzing requirements and asking clarifying questions.
- Propose a step-by-step implementation plan.
- Wait for confirmation before proceeding.
- Implement only the requested scope—nothing extra.
- After writing code, perform a self-review and list potential issues.
- Prefer clarity and correctness over cleverness.
- If uncertain, explicitly say so.
- When developing new code, use the Test Driven Development (TDD) pattern,
  and follow this strict order:
  1. Red phase: create unit tests. Tests must be exhaustive and must fail.
     Tests must compile and run — failures must be assertion failures, not
     undefined variables, missing imports, or missing glue code.
  2. Green phase: implement the feature, making sure that all tests pass.
  3. Refactor: refactor code if necessary.
- When implementing use git. Create a good incremental commit structure
  for every feature, module, etc. being developed
- Git commits: scoped, clear message (what and why)
- Do not amend or rebase commits unless explicitly told to do so.

Assume production-quality standards.

## Project contract

### Goal

Create a tool that passively monitors L2 neighbours using eBPF (TC/TCX).

The architecture has the following components:

1. eBPF probe: passively inspects packets and writes unicast MAC addresses
   into an eBPF map (one per interface). Multicast addresses MUST NOT be tracked. When
   ARP / Neighbour Discovery Protocol (NDP) packets are observed, the eBPF
   program should add the observed IP addresses associated with the MAC address.
   The program MUST NOT assume regular unicast IPv4/IPv6 can be used to deduce
   IP addresses, as intermediate routers can exist. The eBPF map must be
   read-only by any user.
2. A Go CLI tool (`l2radar`) that attaches the eBPF programs to interfaces,
   optionally exports JSON, and provides a `dump` subcommand for inspection.
3. A web-based dashboard to display the neighbours. Runs in a separate
   container from the probe(s). Multiple probes can feed a single UI
   instance via a shared volume of JSON files.

### Tech stack

- eBPF C programs for passive monitoring
- Web UI: React + Tailwind CSS, built with Vite, served by nginx
- Golang for any command-line tool (if needed)

### eBPF Component — Detailed Requirements

#### Directory Structure

```
probe/
├── bpf/
│   ├── headers/          # vmlinux.h or minimal kernel headers
│   └── l2radar.c         # eBPF C program
├── cmd/
│   └── l2radar/
│       └── main.go       # CLI entrypoint
├── pkg/
│   ├── loader/
│   │   ├── loader.go     # Load, attach, pin logic
│   │   ├── loader_test.go
│   │   └── generate.go   # //go:generate bpf2go directive
│   └── oui/
│       ├── oui.go        # OUI lookup from IEEE MA-L database
│       ├── oui_test.go
│       └── oui.txt       # Cached IEEE OUI database (MA-L)
├── go.mod
└── go.sum
```

#### Attachment & Map

- Attach via **TCX ingress** (requires kernel 6.6+)
- Can be attached to multiple interfaces simultaneously
- One **BPF_MAP_TYPE_HASH** per interface
- Pin path: `/sys/fs/bpf/l2radar/neigh-<iface>`
- Map pin permissions: `0444` (world-readable)
- Max entries: 4096 (default)
- Return value: always **TC_ACT_UNSPEC** (passive, allows chaining)

#### Map Key/Value Schema

- **Key**: `u8 mac[6]` (padded to 8 bytes for alignment)
- **Value**:
  - `__be32 ipv4[4]` — up to 4 IPv4 addresses
  - `struct in6_addr ipv6[4]` — up to 4 IPv6 addresses
  - `u8 ipv4_count`, `u8 ipv6_count`
  - `u64 first_seen` — ktime_get_ns at first observation
  - `u64 last_seen` — ktime_get_ns at most recent observation

#### Packet Parsing

- **Multicast filter**: drop (skip tracking) if source MAC bit 0 is set
  (`mac[0] & 0x01`) or broadcast (`ff:ff:ff:ff:ff:ff`)
- **VLAN support**: handle 802.1Q-tagged frames (ethertype `0x8100`),
  skip 4-byte tag to read inner ethertype
- **MAC tracking**: every valid unicast frame upserts the source MAC
  (set `first_seen` on creation, update `last_seen` always)
- **ARP** (ethertype `0x0806`):
  - Validate: htype=1, ptype=0x0800, hlen=6, plen=4
  - Extract sender MAC + sender IP (request and reply)
  - Extract target MAC + target IP (reply, opcode 2)
  - Handle gratuitous ARP (sender IP == target IP)
  - Dedup IPs; respect cap of 4 IPv4 per MAC
- **NDP** (ethertype `0x86DD`, next_header=58 ICMPv6):
  - Parse all NDP types: NS (135), NA (136), RS (133), RA (134)
  - Parse NDP TLV options for Source Link-Layer Address (type 1)
    and Target Link-Layer Address (type 2)
  - Associate IPv6 source address with link-layer address from options
  - Unsolicited NA: extract target address from NA body
  - Unsolicited NS: extract source address
  - Dedup IPs; respect cap of 4 IPv6 per MAC

#### Go Loader (cilium/ebpf + bpf2go)

- `probe/pkg/loader/`: library with `Attach()` / `Detach()` per interface
- `probe/cmd/l2radar/`: CLI with `--iface` (repeatable), `--pin-path`
- Signal handling (SIGINT/SIGTERM) for clean detach + unpin
- Structured logging via slog

#### OUI Vendor Lookup

- Package: `probe/pkg/oui/`
- Uses the IEEE MA-L (OUI) database to resolve the first 3 bytes of a
  MAC address to a vendor name.
- The OUI database file (`oui.txt`) is cached in the repository at
  `probe/pkg/oui/oui.txt` and embedded into the Go binary via
  `//go:embed`. Source: `https://standards-oui.ieee.org/oui/oui.txt`.
- Provides a `Lookup(mac net.HardwareAddr) string` function that returns
  the vendor name or an empty string if not found.
- Used by the `dump` subcommand (terminal output) to display vendor
  names next to MAC addresses.
- The UI ships its own copy of the OUI database and resolves vendor
  names client-side. The JSON export does **not** include vendor names
  — OUI lookup is a display concern only.

#### CLI Interface

- **Default mode** (no subcommand): attach eBPF probes to the specified
  interfaces and keep running until SIGINT/SIGTERM.
- Usage: `l2radar --iface <name> [--iface <name>...] [--pin-path <path>]
  [--export-dir <dir>] [--export-interval <duration>]`
- Flags:
  - `--iface` (repeatable, required): network interface to monitor.
    The special value `any` means **all L2 interfaces except loopbacks**
    (enumerate via `net.Interfaces()`, skip those with `net.FlagLoopback`).
  - `--pin-path`: base path for pinning eBPF maps (default
    `/sys/fs/bpf/l2radar`).
  - `--export-dir` (optional): if set, periodically export JSON files to
    this directory. When not set, the probe only attaches and populates
    the BPF map — no JSON files are written.
  - `--export-interval`: export frequency (default `5s`). Only meaningful
    when `--export-dir` is set.
- When `--export-dir` is set, the probe writes one JSON file per interface
  to the specified directory, named `neigh-<iface>.json`. Uses atomic
  writes (temp file + rename) so nginx never serves partial files.
- Signal handling (SIGINT/SIGTERM) for clean detach, unpin, and shutdown.

#### `dump` Subcommand

- The only subcommand. Reads the pinned eBPF map for a given interface
  and prints the neighbour table to the terminal.
- Usage: `l2radar dump --iface <name> [--pin-path <path>]`
- Opens the pinned map at `<pin-path>/neigh-<iface>` (read-only, no
  privileges required beyond map pin permissions).
- Output: a formatted table with columns:
  - MAC address with OUI vendor name in parentheses
    (e.g., `dc:4b:a1:69:38:16 (Apple Inc.)`)
  - IPv4 addresses (comma-separated)
  - IPv6 addresses (comma-separated)
  - First seen (human-readable timestamp)
  - Last seen (human-readable timestamp)
- Sorted by last seen (most recent first).

#### Container Packaging

- The probe must be packaged in a Docker container so it can run on any
  host with kernel 6.6+ without a local build.
- Multi-stage build:
  - Build stage: `golang:1.24-bookworm` with `clang` and `libbpf-dev`.
    Runs `go generate` (bpf2go) and builds a static Go binary.
  - Runtime stage: `debian:bookworm-slim`. Contains only the static binary.
- Entrypoint: `["/l2radar"]` — the default mode attaches probes; `dump`
  is the only subcommand.
- Runtime requirements:
  - Default mode: `--privileged` and `--network=host` (BPF syscalls +
    host interface access). `-v /sys/fs/bpf:/sys/fs/bpf` to pin maps.
    Optionally `-v /tmp/l2radar:/tmp/l2radar` when using `--export-dir`.
  - `dump`: `--cap-add=BPF` (the `bpf()` syscall is needed to open
    pinned maps) and the bpffs mount (read-only).
- Files: `probe/Dockerfile`, `probe/.dockerignore`.

#### JSON Export

- Export is **not** a separate subcommand. It is enabled by passing
  `--export-dir` to the default mode.
- Output directory: one JSON file per interface, named
  `neigh-<iface>.json` (mirrors the bpffs pin naming).
- Export interval: configurable via `--export-interval` (default `5s`).
- JSON schema per file:
  ```json
  {
    "interface": "<iface>",
    "timestamp": "<RFC3339>",
    "mac": "aa:bb:cc:dd:ee:ff",
    "ipv4": ["192.168.1.10"],
    "ipv6": ["fe80::1"],
    "neighbours": [
      {
        "mac": "aa:bb:cc:dd:ee:ff",
        "ipv4": ["192.168.1.1"],
        "ipv6": ["fe80::1"],
        "first_seen": "<RFC3339>",
        "last_seen": "<RFC3339>"
      }
    ]
  }
  ```
  The top-level `mac`, `ipv4`, and `ipv6` fields are the monitored
  interface's own addresses (looked up via `net.InterfaceByName`).
- Signal handling (SIGINT/SIGTERM) for clean shutdown.
- Requires `CAP_BPF` and bpffs mount (same as `dump`).

### Web UI Component — Detailed Requirements

#### Architecture

- **Separate container** from the probe. Unprivileged (no BPF access).
- Probe(s) write JSON files to a shared volume (`/tmp/l2radar/`).
  The UI container mounts this volume read-only.
- Multiple probes (on different hosts or interfaces) can write to the
  same shared directory; the UI aggregates all JSON files.
- nginx serves both the static React SPA and the JSON data files.

#### Directory Structure

```
ui/
├── src/
│   ├── components/       # React components
│   ├── App.jsx
│   ├── index.jsx
│   └── index.css         # Tailwind directives
├── public/
├── index.html
├── package.json
├── vite.config.js
├── tailwind.config.js
├── nginx/
│   ├── nginx.conf        # Main nginx config
│   └── default.conf      # Site config (TLS, auth, proxy)
├── entrypoint.sh         # TLS cert generation, htpasswd setup
├── Dockerfile
└── .dockerignore
```

#### Dashboard Features

- **Combined view** (default): single table showing all neighbours across
  all interfaces, with an "Interface" column.
- **Interface tabs**: one "All" tab showing all interfaces, plus one tab
  per interface. Per-interface tabs hide the redundant "Interface" column
  from the table and display an information section above the search bar
  with fields in this order: interface name, MAC address, IPv4 addresses,
  IPv6 addresses, last update. The last update field shows a live
  relative time (e.g., "2 seconds ago") that updates continuously in the
  browser, with the absolute timestamp shown on hover (via title attr).
- **Summary statistics**: total neighbours, count per interface, neighbours
  seen in the last 5 minutes.
- **Search/filter**: filter by MAC address or IP address (partial match).
  Present on all tabs (All and per-interface).
- **OUI vendor names**: MAC addresses are displayed with the OUI vendor
  name in parentheses (e.g., `dc:4b:a1:69:38:16 (Apple Inc.)`). The UI
  ships a copy of the IEEE OUI database and resolves vendor names
  client-side from the MAC address prefix.
- **Sortable columns**: MAC, IPv4, IPv6, first seen, last seen. Default
  sort by last seen (most recent first).
- **Auto-refresh**: client polls JSON files using `If-Modified-Since`.
  Poll interval matches the export interval. nginx returns `304 Not Modified`
  when the file hasn't changed.
- **Modern design**:
  - State-of-the-art design trends: clean, compact layout with minimal
    excessive margins and paddings. Prioritize information density.
  - Custom colour scheme (dark-themed, network/radar aesthetic).
  - Fully responsive: must work on phones, tablets, and desktops.
    Table adapts to small screens (e.g., card layout on mobile).
  - Built with Tailwind CSS.

#### HTTPS / TLS

- If the user mounts certificate and key files at the standard nginx
  SSL path (`/etc/nginx/ssl/cert.pem`, `/etc/nginx/ssl/key.pem`), nginx
  uses them.
- If no certificates are mounted, the entrypoint script generates a
  self-signed certificate at container startup.
- nginx listens on port 443 (HTTPS). HTTP on port 80 redirects to HTTPS.

#### Authentication

- nginx `auth_basic` with htpasswd.
- Credentials are provided via a YAML config file mounted into the
  container (e.g., `/etc/l2radar/auth.yaml`):
  ```yaml
  users:
    - username: admin
      password: changeme
  ```
- The entrypoint script reads the YAML file and generates an htpasswd
  file for nginx. Passwords are stored bcrypt-hashed.

#### Container Packaging

- Multi-stage build:
  - Build stage: `node:22-alpine`. Runs `npm install` and `npm run build`
    to produce static assets.
  - Runtime stage: `nginx:alpine`. Contains static assets and nginx config.
- Entrypoint: `entrypoint.sh` (TLS setup, htpasswd generation, nginx start).
- Runtime mounts:
  - `/tmp/l2radar/` (read-only) — shared volume with JSON files from probe(s).
  - `/etc/l2radar/auth.yaml` (read-only) — authentication credentials.
  - `/etc/nginx/ssl/` (optional, read-only) — TLS certificate and key.
- Ports: 443 (HTTPS), 80 (HTTP redirect).

#### CI Pipeline (GitHub Actions)

- File: `.github/workflows/ci.yml`
- Trigger: push to any branch + pull requests.
- Runner: `ubuntu-24.04` (kernel 6.8+, supports TCX and BPF).
- Jobs:
  1. **test**: Install Go 1.24, clang, llvm, libbpf-dev. Run `go generate`
     then `sudo go test` — BPF tests MUST NOT be skipped.
  2. **build**: Build Docker image via `docker build`.
  3. **publish**: Push to `ghcr.io/${{ github.repository_owner }}/l2radar`.
     Only on push (not PR). Authenticates via `GITHUB_TOKEN`.
     - Tags: `latest` when on main, `<branch>` for other branches.
     - On version tags (`v*`): tag with the version (e.g., `v1.0.0`).
- UI image (`ghcr.io/${{ github.repository_owner }}/l2radar-ui`):
  - **test**: `npm test` (unit tests for React components).
  - **build**: Build Docker image via `docker build`.
  - **publish**: Same tagging strategy as probe image.

### Constraints

- All components MUST have unit tests and they must pass.
- eBPF programs MUST NOT interfere with traffic (passive monitoring only).
- eBPF programs MUST return TC_ACT_UNSPEC to allow program chaining.

### Success criteria

- Unit tests passing
- End-to-end tests passing
