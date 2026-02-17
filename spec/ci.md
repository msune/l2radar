# CI Pipeline (GitHub Actions)

- File: `.github/workflows/ci.yml`
- Trigger: push to any branch + pull requests.
- Runner: `ubuntu-24.04` (kernel 6.8+, supports TCX and BPF).

## Probe image (`ghcr.io/<owner>/l2radar`)

1. **test**: Go 1.24, clang, llvm, libbpf-dev. `go generate` then
   `sudo go test` — BPF tests MUST NOT be skipped.
2. **build**: `docker build`.
3. **publish**: push on push (not PR). `GITHUB_TOKEN` auth.
   Tags: `bleeding-edge` on main, `<branch>` on other branches,
   `<version>` + `latest` on `v*` tag push.

## UI image (`ghcr.io/<owner>/l2radar-ui`)

1. **test**: `npm test`.
2. **build**: `docker build`.
3. **publish**: same tagging strategy as probe.

## l2rctl

1. **test**: `go test` — validates the code.

l2rctl is distributed via `go install` (Go module proxy), not GitHub Releases.
