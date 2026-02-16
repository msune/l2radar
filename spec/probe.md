# eBPF Probe & Go CLI

## Directory Structure

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
│       └── oui.json      # Preparsed IEEE OUI database (prefix→vendor)
├── go.mod
└── go.sum
```

## eBPF Attachment & Map

- Attach via **TCX ingress** (requires kernel 6.6+)
- Can be attached to multiple interfaces simultaneously
- One **BPF_MAP_TYPE_HASH** per interface
- Pin path: `/sys/fs/bpf/l2radar/neigh-<iface>`
- Map pin permissions: `0444` (world-readable)
- Max entries: 4096 (default)
- Return value: always **TC_ACT_UNSPEC** (passive, allows chaining)

## Map Key/Value Schema

- **Key**: `u8 mac[6]` (padded to 8 bytes for alignment)
- **Value**:
  - `__be32 ipv4[4]` — up to 4 IPv4 addresses
  - `struct in6_addr ipv6[4]` — up to 4 IPv6 addresses
  - `u8 ipv4_count`, `u8 ipv6_count`
  - `u64 first_seen` — ktime_get_ns at first observation
  - `u64 last_seen` — ktime_get_ns at most recent observation

## Packet Parsing

- **Multicast filter**: skip if source MAC bit 0 is set (`mac[0] & 0x01`)
  or broadcast (`ff:ff:ff:ff:ff:ff`)
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

## Go Loader (cilium/ebpf + bpf2go)

- `probe/pkg/loader/`: library with `Attach()` / `Detach()` per interface
- `probe/cmd/l2radar/`: CLI with `--iface` (repeatable), `--pin-path`
- Signal handling (SIGINT/SIGTERM) for clean detach + unpin
- Structured logging via slog

## OUI Vendor Lookup

- Package: `probe/pkg/oui/`
- Uses the IEEE MA-L (OUI) database to resolve the first 3 bytes of a
  MAC address to a vendor name.
- `oui.json` is a preparsed prefix→vendor map shared with the UI,
  embedded via `//go:embed`.
- `Lookup(mac net.HardwareAddr) string` — returns vendor name or `""`.
- Used by `dump` subcommand for terminal display.
- The UI ships its own OUI copy and resolves client-side. The JSON
  export does **not** include vendor names — display concern only.

## CLI Interface

- **Default mode** (no subcommand): attach probes, run until signal.
- Usage: `l2radar --iface <name> [--iface <name>...] [--pin-path <path>]
  [--export-dir <dir>] [--export-interval <duration>]`
- Flags:
  - `--iface` (repeatable, required): interface to monitor. `any` =
    external interfaces (excludes loopbacks and virtual interfaces like
    docker*, veth*, br-*, virbr*). `all` = all L2 interfaces except
    loopbacks.
  - `--pin-path`: base path for pinning (default `/sys/fs/bpf/l2radar`).
  - `--export-dir` (optional): periodically export JSON to this dir.
  - `--export-interval`: export frequency (default `5s`).
- Atomic writes (temp file + rename) for JSON export.
- Signal handling (SIGINT/SIGTERM) for clean shutdown.

## `dump` Subcommand

- Reads pinned map at `<pin-path>/neigh-<iface>` (read-only).
- Output: formatted table with columns:
  - MAC address with OUI vendor name (e.g., `dc:4b:a1:69:38:16 (Apple Inc.)`)
  - IPv4 addresses (comma-separated)
  - IPv6 addresses (comma-separated)
  - First seen, Last seen (human-readable timestamps)
- Sorted by last seen (most recent first).

## JSON Export Schema

```json
{
  "interface": "<iface>",
  "timestamp": "<RFC3339>",
  "mac": "aa:bb:cc:dd:ee:ff",
  "ipv4": ["192.168.1.10"],
  "ipv6": ["fe80::1"],
  "stats": {
    "tx_bytes": 123456,
    "rx_bytes": 789012,
    "tx_packets": 1000,
    "rx_packets": 2000,
    "tx_errors": 0,
    "rx_errors": 0,
    "tx_dropped": 0,
    "rx_dropped": 0
  },
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

Top-level `mac`, `ipv4`, `ipv6` are the monitored interface's own
addresses (via `net.InterfaceByName`).

### Interface Stats

The `stats` object contains kernel interface counters read from
`/sys/class/net/<iface>/statistics/` at export time. All values are
`uint64`. The field is `null` when stats are unavailable (e.g.,
interface not found).

## Container Packaging

- Multi-stage build:
  - Build: `golang:1.24-bookworm` + `clang` + `libbpf-dev`.
    `go generate` (bpf2go) then static Go binary.
  - Runtime: `debian:bookworm-slim` with the binary.
- Entrypoint: `["/l2radar"]`.
- Default mode: `--privileged`, `--network=host`,
  `-v /sys/fs/bpf:/sys/fs/bpf`.
- `dump`: `--cap-add=BPF` + bpffs mount (read-only).
- Files: `probe/Dockerfile`, `probe/.dockerignore`.
