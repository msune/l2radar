
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
  2. Green phase: implement the feature, making sure that all tests pass.
  3. Refactor: refactor code if necessary.
- When implementing use git. Create a good incremental commit structure
  for every feature, module, etc. being developed
- Git commits: scoped, clear message (what and why)

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
2. A Go CLI tool to load, attach, and manage the eBPF programs.
3. A graphical user interface to display the neighbours. The UI needs to be
   in a form of Dashboard, with a modern design. It also has to comply with
   the following characteristics:
  - The Web UI must use HTTPs and a basic Authentication. Username and
    password should be passed as a configuration file to the UI.
  - The Web UI should be built as a client-side application only, rendering
    the contents of the JSON file with the contents of the eBPF Map, served by
    the same webserver.
  - The Web UI needs to be packaged in a docker container, where the eBPF
    map will be mounted (read-only).

### Tech stack

- eBPF C programs for passive monitoring
- For the webui:
  - ngninx as webserver.
  - Use react and tailwind for the client-side UI parts
  - Use best tool to read the eBPF map, convert it into JSON and have ngninx
    serve it.
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
│   └── loader/
│       ├── loader.go     # Load, attach, pin logic
│       ├── loader_test.go
│       └── generate.go   # //go:generate bpf2go directive
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

#### Map Dump Tool

- The CLI must support a `dump` subcommand that reads the pinned eBPF map
  for a given interface and prints the neighbour table to the terminal.
- Usage: `l2radar dump --iface <name> [--pin-path <path>]`
- Opens the pinned map at `<pin-path>/neigh-<iface>` (read-only, no
  privileges required beyond map pin permissions).
- Output: a formatted table with columns:
  - MAC address
  - IPv4 addresses (comma-separated)
  - IPv6 addresses (comma-separated)
  - First seen (human-readable timestamp)
  - Last seen (human-readable timestamp)
- Sorted by last seen (most recent first).

### Constraints

- All components MUST have unit tests and they must pass.
- eBPF programs MUST NOT interfere with traffic (passive monitoring only).
- eBPF programs MUST return TC_ACT_UNSPEC to allow program chaining.

### Success criteria

- Unit tests passing
- End-to-end tests passing
