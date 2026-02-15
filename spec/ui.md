# Web UI

## Architecture

- Separate container from the probe. Unprivileged (no BPF access).
- Probe(s) write JSON to a shared volume (`/tmp/l2radar/`). UI mounts
  read-only. Multiple probes can share the directory.
- nginx serves the React SPA and the JSON data files.

## Directory Structure

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

## Dashboard Features

- **Combined view** (default): table of all neighbours with "Interface"
  column.
- **Interface tabs**: "All" tab + one tab per interface. Per-interface
  tabs hide the redundant "Interface" column and show an info section
  above the search bar with fields in order: interface name, MAC, IPv4,
  IPv6, last update. Last update shows live relative time (e.g.,
  "2 seconds ago") updating continuously, absolute timestamp on hover
  (title attr).
- **Summary statistics**: total neighbours, count per interface,
  neighbours seen in the last 5 minutes.
- **Search/filter**: filter by MAC or IP (partial match). Present on
  all tabs.
- **OUI vendor names**: MAC addresses shown with vendor name in
  parentheses. UI ships its own IEEE OUI database (`oui.json`) and
  resolves client-side.
- **Sortable columns**: MAC, IPv4, IPv6, first seen, last seen. Default
  sort by last seen (most recent first).
- **Auto-refresh**: polls JSON with `If-Modified-Since`. nginx returns
  304 when unchanged.
- **Design**: dark-themed, compact layout, information-dense. Fully
  responsive (card layout on mobile). Tailwind CSS.

## HTTPS / TLS

- Certs at `/etc/nginx/ssl/cert.pem` and `key.pem` if mounted.
- Otherwise, entrypoint generates self-signed cert.
- Port 443 (HTTPS), port 80 redirects to HTTPS.

## Authentication

- nginx `auth_basic` with htpasswd.
- Credentials via YAML (`/etc/l2radar/auth.yaml`):
  ```yaml
  users:
    - username: admin
      password: changeme
  ```
- Entrypoint generates htpasswd (bcrypt-hashed).

## Container Packaging

- Build: `node:22-alpine`, `npm install && npm run build`.
- Runtime: `nginx:alpine` with static assets + nginx config.
- Entrypoint: `entrypoint.sh`.
- Mounts: `/tmp/l2radar/` (ro), `/etc/l2radar/auth.yaml` (ro),
  `/etc/nginx/ssl/` (optional, ro).
- Ports: 443, 80.
