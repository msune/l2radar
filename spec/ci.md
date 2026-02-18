# CI Pipeline (GitHub Actions)

- File: `.github/workflows/ci.yml`
- Trigger: push to any branch + pull requests.
- Runner: `ubuntu-24.04` (kernel 6.8+, supports TCX and BPF).
- Platforms: `linux/amd64`, `linux/arm64` (multi-arch via QEMU + buildx).

## Probe image (`ghcr.io/<owner>/l2radar`)

1. **test**: Go 1.24, clang, llvm, libbpf-dev. `go generate` then
   `sudo go test` — BPF tests MUST NOT be skipped.
2. **build**: multi-arch `docker buildx build` (amd64 + arm64).
3. **publish**: push on push (not PR). `GITHUB_TOKEN` auth.
   Tags: `bleeding-edge` on main, `<branch>` on other branches,
   `<version>` + `latest` on `v*` tag push.

## UI image (`ghcr.io/<owner>/l2radar-ui`)

1. **test**: `npm test`.
2. **build**: multi-arch `docker buildx build` (amd64 + arm64).
3. **publish**: same tagging strategy as probe.

## l2rctl

1. **test**: `go test` — validates the code.

l2rctl is distributed via `go install` (Go module proxy), not GitHub Releases.
Because `l2rctl` is a Go submodule (`github.com/msune/l2radar/l2rctl`), the
proxy requires tags prefixed with `l2rctl/` (e.g., `l2rctl/v0.0.7`). The
`update-release-tags` job creates this tag automatically when a `v*` tag is
pushed.
