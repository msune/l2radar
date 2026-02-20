# Changelog

All notable changes to this project are documented in this file.

## Next
- [`45ff683`](https://github.com/msune/l2radar/commit/45ff683) docs: add changelog
- [`49ffd70`](https://github.com/msune/l2radar/commit/49ffd70) docs: update README for named volume

## v0.1.0 (2026-02-20)
- [`8db5646`](https://github.com/msune/l2radar/commit/8db5646) l2rctl/start: implement EnsureCleanVolume, wire into CLI (green)
- [`25641d1`](https://github.com/msune/l2radar/commit/25641d1) l2rctl/start: add EnsureCleanVolume stub and tests (red)
- [`26787f8`](https://github.com/msune/l2radar/commit/26787f8) l2rctl/stop: remove named volume on stop all (green)
- [`20fb092`](https://github.com/msune/l2radar/commit/20fb092) l2rctl/stop: add Opts struct, volume removal tests (red)
- [`a14983f`](https://github.com/msune/l2radar/commit/a14983f) ui/nginx: update data alias to /var/lib/l2radar
- [`958f5d5`](https://github.com/msune/l2radar/commit/958f5d5) l2rctl/cli: add --volume-name flag, update --export-dir default
- [`3f44870`](https://github.com/msune/l2radar/commit/3f44870) l2rctl/start: switch probe/UI to named Docker volume (green)
- [`599cce7`](https://github.com/msune/l2radar/commit/599cce7) l2rctl/start: add VolumeName to opts, update tests for named volume (red)
- [`0577c34`](https://github.com/msune/l2radar/commit/0577c34) Makefile: fix use of "all" in dev Makefile
- [`11c5efd`](https://github.com/msune/l2radar/commit/11c5efd) Revert "l2rctl: attempt to fix race on new version check"
- [`f5489fe`](https://github.com/msune/l2radar/commit/f5489fe) README.md: remove Alpha notice
- [`2642a0a`](https://github.com/msune/l2radar/commit/2642a0a) Rename --iface any -> external, all -> any

## v0.0.15 (2026-02-19)
- [`b1ed1a6`](https://github.com/msune/l2radar/commit/b1ed1a6) l2rctl|l2radar: manually fix -o json
- [`b1ebfcc`](https://github.com/msune/l2radar/commit/b1ebfcc) l2rctl/test: test auth tmp files perm. 0600

## v0.0.14 (2026-02-18)
- [`7735542`](https://github.com/msune/l2radar/commit/7735542) fix(l2rctl): rework dump command interface

## v0.0.13 (2026-02-18)
- [`93146c3`](https://github.com/msune/l2radar/commit/93146c3) ci: add multi-arch (amd64 + arm64) Docker image builds

## v0.0.12 (2026-02-17)
- [`6e1a809`](https://github.com/msune/l2radar/commit/6e1a809) feat(bpf): add non-linear skb support and optimize L3 dispatch
- [`d9a6c92`](https://github.com/msune/l2radar/commit/d9a6c92) Makefile: add a dev Makefile

## v0.0.9 (2026-02-18)
- [`d9b71d3`](https://github.com/msune/l2radar/commit/d9b71d3) l2rctl: attempt to fix race on new version check

## v0.0.7 (2026-02-17)
- [`ebf2030`](https://github.com/msune/l2radar/commit/ebf2030) ci: create l2rctl/ submodule tag on semver release push

## v0.0.6 (2026-02-17)
- [`6116a94`](https://github.com/msune/l2radar/commit/6116a94) feat(l2rctl): add version subcommand and background update check
- [`839a110`](https://github.com/msune/l2radar/commit/839a110) feat(l2rctl): add version check core (internal/version)
- [`fdc911b`](https://github.com/msune/l2radar/commit/fdc911b) fix(ui): filter git describe to only match semver tags
- [`9c54aea`](https://github.com/msune/l2radar/commit/9c54aea) README.md: cleanup

## v0.0.5 (2026-02-17)
- [`0827c37`](https://github.com/msune/l2radar/commit/0827c37) ci: update latest annotated tag on semver release push

## v0.0.4 (2026-02-17)
- [`9cb1bee`](https://github.com/msune/l2radar/commit/9cb1bee) feat: add install script with symlink and kernel check

## v0.0.3 (2026-02-17)
- [`c8747b8`](https://github.com/msune/l2radar/commit/c8747b8) docs: fix l2rctl module path in README and spec
- [`723e796`](https://github.com/msune/l2radar/commit/723e796) fix(l2rctl): rename module to github.com/msune/l2radar/l2rctl

## v0.0.2 (2026-02-17)
- [`58ee3ef`](https://github.com/msune/l2radar/commit/58ee3ef) docs: update install instructions to use go install
- [`808035b`](https://github.com/msune/l2radar/commit/808035b) chore: remove install-l2rctl.sh script
- [`439e5cd`](https://github.com/msune/l2radar/commit/439e5cd) ci: use bleeding-edge tag for main, latest for stable releases
- [`16940c3`](https://github.com/msune/l2radar/commit/16940c3) docs: add install subcommand to l2rctl spec
- [`ea0e153`](https://github.com/msune/l2radar/commit/ea0e153) feat(l2rctl): add install subcommand
- [`af9344a`](https://github.com/msune/l2radar/commit/af9344a) feat(start): add RestartPolicy to ProbeOpts and UIOpts
- [`7d21af6`](https://github.com/msune/l2radar/commit/7d21af6) docs: add l2rctl as architectural component, fix stale paths
- [`1d391f2`](https://github.com/msune/l2radar/commit/1d391f2) ui: fix IPv6 overlap with last update
- [`0f8bf46`](https://github.com/msune/l2radar/commit/0f8bf46) ui: fix splash screen version
- [`11c3df1`](https://github.com/msune/l2radar/commit/11c3df1) ci: always generate l2rctl latest pkg
- [`e1d770a`](https://github.com/msune/l2radar/commit/e1d770a) README.md: minor fixes
- [`0984ec0`](https://github.com/msune/l2radar/commit/0984ec0) fix(l2rctl): remove credentials from displayed access URLs
- [`f81e7b6`](https://github.com/msune/l2radar/commit/f81e7b6) fix(l2rctl): fall back to local image when pull fails
- [`0a2f583`](https://github.com/msune/l2radar/commit/0a2f583) refactor(probe): migrate from stdlib flag to cobra
- [`1d908dc`](https://github.com/msune/l2radar/commit/1d908dc) refactor(l2rctl): migrate from stdlib flag to cobra
- [`fae45a4`](https://github.com/msune/l2radar/commit/fae45a4) fix(l2rctl): show component as positional argument in help output
- [`f4ffcfe`](https://github.com/msune/l2radar/commit/f4ffcfe) refactor(l2rctl): move to dedicated l2rctl/ directory
- [`967990c`](https://github.com/msune/l2radar/commit/967990c) feat(l2rctl): show clickable access URL(s) after UI starts
- [`af367d4`](https://github.com/msune/l2radar/commit/af367d4) feat(l2rctl): add --bind, --https-port, --http-port flags for UI container
- [`c567845`](https://github.com/msune/l2radar/commit/c567845) README.md: minor improvements
- [`195ad07`](https://github.com/msune/l2radar/commit/195ad07) README.md: misc
- [`7d0b8d1`](https://github.com/msune/l2radar/commit/7d0b8d1) docs(spec): document image pull behavior
- [`edba6ef`](https://github.com/msune/l2radar/commit/edba6ef) feat(l2rctl): always pull latest image before starting containers
- [`047081c`](https://github.com/msune/l2radar/commit/047081c) docs(spec): update l2rctl spec with random credentials and --type container
- [`6d13245`](https://github.com/msune/l2radar/commit/6d13245) feat(l2rctl): generate random credentials when no auth is configured
- [`40fa47e`](https://github.com/msune/l2radar/commit/40fa47e) feat(auth): add GenerateRandomCredentials for default UI auth
- [`4825b11`](https://github.com/msune/l2radar/commit/4825b11) fix(l2rctl): use docker inspect --type container to avoid image collision
- [`2354ee1`](https://github.com/msune/l2radar/commit/2354ee1) ui: add lucide icons to stat cards and interface tabs
- [`2244267`](https://github.com/msune/l2radar/commit/2244267) license: add BSD-2-Clause, dual BSD/GPL for eBPF code
- [`23edbc8`](https://github.com/msune/l2radar/commit/23edbc8) docs: add README and architecture documentation
- [`709de2e`](https://github.com/msune/l2radar/commit/709de2e) feat(l2rctl): add install script
- [`6ac1b50`](https://github.com/msune/l2radar/commit/6ac1b50) ci(l2rctl): add GitHub Release for static binaries
- [`85965c1`](https://github.com/msune/l2radar/commit/85965c1) ui: show version in splash screen
- [`4af899b`](https://github.com/msune/l2radar/commit/4af899b) ui: show splash screen only once per 2-hour window
- [`8bb54cd`](https://github.com/msune/l2radar/commit/8bb54cd) spec: update probe and l2rctl docs for any/all interface keywords
- [`f7dab7d`](https://github.com/msune/l2radar/commit/f7dab7d) feat(probe): redefine 'any' as external-only, add 'all' for all interfaces
- [`4455e87`](https://github.com/msune/l2radar/commit/4455e87) ci(l2rctl): add test and build jobs
- [`a998023`](https://github.com/msune/l2radar/commit/a998023) feat(l2rctl): implement dump subcommand and wire up CLI
- [`e119c9f`](https://github.com/msune/l2radar/commit/e119c9f) feat(l2rctl): implement status subcommand
- [`a936e35`](https://github.com/msune/l2radar/commit/a936e35) feat(l2rctl): implement stop subcommand
- [`055b348`](https://github.com/msune/l2radar/commit/055b348) feat(l2rctl): implement start subcommand
- [`40c6940`](https://github.com/msune/l2radar/commit/40c6940) feat(l2rctl): add auth.yaml generation
- [`7e87746`](https://github.com/msune/l2radar/commit/7e87746) feat(l2rctl): scaffold Go module, CLI dispatch, and docker wrapper
- [`057e832`](https://github.com/msune/l2radar/commit/057e832) spec(l2rctl): add CLI specification
- [`b0b646e`](https://github.com/msune/l2radar/commit/b0b646e) ui: make splash size responsive
- [`6215729`](https://github.com/msune/l2radar/commit/6215729) ui: tweak logo, splash size/opacity
- [`0ded92d`](https://github.com/msune/l2radar/commit/0ded92d) ci: refix (manual) fetch tags on checkout
- [`57acec3`](https://github.com/msune/l2radar/commit/57acec3) ui: disable HTTP port 80 by default, HTTPS only
- [`1d03ffe`](https://github.com/msune/l2radar/commit/1d03ffe) ci: ensure no tests are ever skipped
- [`90bfe43`](https://github.com/msune/l2radar/commit/90bfe43) ui: add splash screen with logo on app load
- [`b0373a1`](https://github.com/msune/l2radar/commit/b0373a1) ui: drop --dirty from git describe version string
- [`1081a6c`](https://github.com/msune/l2radar/commit/1081a6c) Add .claudeignore to exclude large generated/binary files
- [`87f51e3`](https://github.com/msune/l2radar/commit/87f51e3) ci: fix UI Docker build context to use repo root
- [`89feacf`](https://github.com/msune/l2radar/commit/89feacf) ui: use logo_small.png in header, track assets/img
- [`3ba7dcd`](https://github.com/msune/l2radar/commit/3ba7dcd) ui: add favicons from assets/icons/ to Docker build
- [`81d9574`](https://github.com/msune/l2radar/commit/81d9574) fix: ensure git describe resolves version correctly
- [`16e28b7`](https://github.com/msune/l2radar/commit/16e28b7) ui: add collapsible interface stats row to InterfaceInfo
- [`d08ac5c`](https://github.com/msune/l2radar/commit/d08ac5c) feat: pass interface stats through JS parser
- [`b406019`](https://github.com/msune/l2radar/commit/b406019) feat: add interface stats (TX/RX counters) to JSON export
- [`8c410db`](https://github.com/msune/l2radar/commit/8c410db) spec: add interface stats (TX/RX counters) to JSON schema and UI spec
- [`ceea94a`](https://github.com/msune/l2radar/commit/ceea94a) ui: add powered-by eBPF and tweak footer
- [`f348975`](https://github.com/msune/l2radar/commit/f348975) feat: add header menu with username, version, and logout

## v0.0.1 (2026-02-15)
- [`bf31efb`](https://github.com/msune/l2radar/commit/bf31efb) ui: add sticky footer with copyright, attribution, and GitHub link
- [`06292b1`](https://github.com/msune/l2radar/commit/06292b1) refactor: use preparsed oui.json instead of oui.txt for probe OUI lookup
- [`5879110`](https://github.com/msune/l2radar/commit/5879110) feat: add export_interval to JSON, turn last update red when overdue
- [`bb6bf55`](https://github.com/msune/l2radar/commit/bb6bf55) ui: highlight fresh updates and dim stale neighbours
- [`aa52ae5`](https://github.com/msune/l2radar/commit/aa52ae5) spec: add freshness highlight and stale dimming requirements
- [`9d30937`](https://github.com/msune/l2radar/commit/9d30937) ui: show "less than a minute ago" for sub-minute times, refresh every 5s
- [`fff16f1`](https://github.com/msune/l2radar/commit/fff16f1) ui: show relative time for last update and last seen
- [`956cac1`](https://github.com/msune/l2radar/commit/956cac1) ui: reorder InterfaceInfo fields to name, MAC, IPv4, IPv6, last update
- [`f16f004`](https://github.com/msune/l2radar/commit/f16f004) refactor: split CLAUDE.md into spec/ files to reduce context size
- [`eb071d6`](https://github.com/msune/l2radar/commit/eb071d6) feat: hide redundant Interface column on per-interface tabs
- [`1f31c5c`](https://github.com/msune/l2radar/commit/1f31c5c) docs: update UI requirements for interface tabs and info section
- [`09823ae`](https://github.com/msune/l2radar/commit/09823ae) feat: display OUI vendor names in UI neighbour table
- [`7b8391b`](https://github.com/msune/l2radar/commit/7b8391b) feat: show OUI vendor names in dump table output
- [`076b814`](https://github.com/msune/l2radar/commit/076b814) feat: add OUI vendor lookup package with embedded IEEE database
- [`f46a7c7`](https://github.com/msune/l2radar/commit/f46a7c7) feat: add interface tabs with info section and address metadata
- [`113410e`](https://github.com/msune/l2radar/commit/113410e) docs: clarify TDD red phase and git rebase rules in CLAUDE.md
- [`abd1860`](https://github.com/msune/l2radar/commit/abd1860) fix: set world-readable permissions on exported JSON files
- [`6bb29f6`](https://github.com/msune/l2radar/commit/6bb29f6) ci: add UI test, build and publish jobs
- [`cc74239`](https://github.com/msune/l2radar/commit/cc74239) ui: add Dockerfile and .dockerignore
- [`70cf1b2`](https://github.com/msune/l2radar/commit/70cf1b2) ui: add nginx config and entrypoint script
- [`b948287`](https://github.com/msune/l2radar/commit/b948287) ui: add search/filter and interface filter
- [`3116f4e`](https://github.com/msune/l2radar/commit/3116f4e) ui: add neighbour table with sorting
- [`a2a9671`](https://github.com/msune/l2radar/commit/a2a9671) ui: add summary statistics component
- [`4d22bd8`](https://github.com/msune/l2radar/commit/4d22bd8) ui: add data fetching with If-Modified-Since polling and contract tests
- [`728620e`](https://github.com/msune/l2radar/commit/728620e) ui: scaffold Vite + React + Tailwind project
- [`7e69aa8`](https://github.com/msune/l2radar/commit/7e69aa8) probe: add optional JSON export to default mode
- [`94b60d2`](https://github.com/msune/l2radar/commit/94b60d2) CLAUDE.md: add Web UI and JSON export requirements
- [`ec0ef2f`](https://github.com/msune/l2radar/commit/ec0ef2f) probe: use CLOCK_BOOTTIME for suspend-safe timestamps
- [`0155b21`](https://github.com/msune/l2radar/commit/0155b21) probe: fix ktime-to-wallclock conversion in dump
- [`f1576a7`](https://github.com/msune/l2radar/commit/f1576a7) ci: add GitHub Actions pipeline
- [`17d0f03`](https://github.com/msune/l2radar/commit/17d0f03) CLAUDE.md: add CI pipeline requirements
- [`3f8825e`](https://github.com/msune/l2radar/commit/3f8825e) CLAUDE.md: fix dump container requires CAP_BPF
- [`a3d97f8`](https://github.com/msune/l2radar/commit/a3d97f8) probe: add Dockerfile for containerized deployment
- [`febdc0a`](https://github.com/msune/l2radar/commit/febdc0a) CLAUDE.md: add container packaging requirements
- [`dabfce3`](https://github.com/msune/l2radar/commit/dabfce3) probe: filter out frames with unknown ethertypes
- [`4e6dd44`](https://github.com/msune/l2radar/commit/4e6dd44) probe: add dump subcommand
- [`c121fd7`](https://github.com/msune/l2radar/commit/c121fd7) CLAUDE.md: add map dump tool requirement
- [`e099c4c`](https://github.com/msune/l2radar/commit/e099c4c) probe: fix IPv4 endianness in test helper
- [`5037feb`](https://github.com/msune/l2radar/commit/5037feb) probe: add CLI entrypoint with multi-interface support
- [`26a0361`](https://github.com/msune/l2radar/commit/26a0361) probe: add Go loader library with TCX attach, pin, permissions
- [`6dccc0d`](https://github.com/msune/l2radar/commit/6dccc0d) probe: add 802.1Q VLAN tag support
- [`24a24c9`](https://github.com/msune/l2radar/commit/24a24c9) probe: add NDP RS/RA parser support
- [`b629ab9`](https://github.com/msune/l2radar/commit/b629ab9) probe: add NDP NS/NA parser with unsolicited support
- [`eb3b828`](https://github.com/msune/l2radar/commit/eb3b828) probe: add ARP parser with gratuitous ARP support
- [`580645e`](https://github.com/msune/l2radar/commit/580645e) probe: add multicast filter, MAC tracking, and tests
- [`4f9e00e`](https://github.com/msune/l2radar/commit/4f9e00e) probe: define eBPF map and data structures
- [`bf37dc5`](https://github.com/msune/l2radar/commit/bf37dc5) probe: scaffold project structure with bpf2go
- [`53f3bc6`](https://github.com/msune/l2radar/commit/53f3bc6) CLAUDE.md: add detailed eBPF component requirements
- [`22c027f`](https://github.com/msune/l2radar/commit/22c027f) CLAUDE.md: add initial version
- [`7409ba9`](https://github.com/msune/l2radar/commit/7409ba9) Initial commit
