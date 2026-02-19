# l2rctl — Container Management CLI

Self-contained Go binary that wraps Docker CLI operations for starting,
stopping, monitoring, and inspecting l2radar containers.

## Location

`l2rctl/` — own Go module (`github.com/msune/l2radar/l2rctl`), Go 1.24.
Only dependency: `gopkg.in/yaml.v3` for auth.yaml generation. Shells out
to `docker` CLI (no Docker SDK).

## Subcommands

### `l2rctl start [all|probe|ui]` (default: all)

**Probe flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--iface <name>` | `external` | Interface to monitor (repeatable; `external` = external, `any` = all non-loopback) |
| `--export-dir <dir>` | `/tmp/l2radar` | Host directory for JSON exports |
| `--export-interval <dur>` | `5s` | Export interval |
| `--pin-path <path>` | `/sys/fs/bpf/l2radar` | BPF pin path |
| `--probe-image <image>` | `ghcr.io/msune/l2radar:latest` | Probe image |
| `--probe-docker-args <args>` | | Extra `docker run` arguments |

**UI flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--tls-dir <path>` | | TLS cert directory, mounted as `/etc/nginx/ssl:ro` |
| `--user-file <path>` | | Auth file, mounted as `/etc/l2radar/auth.yaml:ro` |
| `--user <user:pass>` | | Inline user (repeatable), generates temp auth.yaml |
| `--enable-http` | false | Enable HTTP port 80 |
| `--ui-image <image>` | `ghcr.io/msune/l2radar-ui:latest` | UI image |
| `--ui-docker-args <args>` | | Extra `docker run` arguments |

`--user-file` and `--user` are mutually exclusive. If neither is
provided and the target includes the UI, random credentials are
generated automatically (username `admin` + 4 hex chars, 16-char
alphanumeric password) and printed after successful start.

**Container names:** `l2radar` (probe), `l2radar-ui` (UI).

**Probe always gets:**
```
--privileged --network=host \
  -v /sys/fs/bpf:/sys/fs/bpf \
  -v <export-dir>:<export-dir>
```

**UI always gets:**
```
-v <export-dir>:<export-dir>:ro -p 443:443
```

**Pre-start check:** `docker inspect --type container` (to avoid
matching images with the same name). If container is running → error;
if stopped → remove then start.

**Image pull:** `docker pull --quiet <image>` runs before every
`docker run`. Progress output is suppressed; only errors are surfaced.

### `l2rctl install [all|probe|ui]` (default: all)

Same flags and behaviour as `start`, but adds `--restart unless-stopped`
to `docker run` so that Docker automatically restarts the containers
after a system reboot.

**Prerequisites:** the Docker daemon must be enabled at boot
(`systemctl enable docker`).

**Stopping:** `l2rctl stop` stops and removes containers. The
`unless-stopped` policy means stopped containers will *not* be
restarted on reboot.

### `l2rctl stop [all|probe|ui]` (default: all)

`docker stop` + `docker rm` for target containers. Ignores "not found"
errors.

### `l2rctl status`

`docker inspect` both containers. Prints table:

```
CONTAINER    STATUS     STARTED
l2radar      running    2025-06-01T12:00:00Z
l2radar-ui   not found  -
```

### `l2rctl dump --iface <name> [-o json]`

- **Table output** (default): `docker exec l2radar l2radar dump --iface <name>`
- **JSON output** (`-o json`): reads `<export-dir>/neigh-<iface>.json`
  directly from host.

## Auth Generation

`--user admin:secret` generates a temp file at `/tmp/l2rctl-auth-*.yaml`:

```yaml
users:
  - username: admin
    password: secret
```

Mounted as `/etc/l2radar/auth.yaml:ro` into the UI container.

When no `--user` or `--user-file` is provided and the target includes
the UI, `GenerateRandomCredentials()` produces a credential string
(`admin<4-hex>:<16-alphanumeric>`) which is fed into the same
`WriteAuthFile` path. The credentials are printed to stdout after
successful start:

```
Generated UI credentials (no --user or --user-file provided):
  Username: admin3f8a
  Password: aBcDeFgH12345678
```

Validation: split on first `:`, error if missing colon or empty parts.

## Docker Wrapper

```go
type Runner interface {
    Run(args ...string) (stdout, stderr string, err error)
    RunAttached(args ...string) error
}
```

Real implementation uses `os/exec`. Tests use a mock that records calls.

## Testing

All packages test via mock `Runner`. TDD: write failing tests first,
then implement.
