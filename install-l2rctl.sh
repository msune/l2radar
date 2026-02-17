#!/usr/bin/env bash
# install-l2rctl.sh — install the l2rctl CLI tool
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/msune/l2radar/latest/install-l2rctl.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/msune/l2radar/latest/install-l2rctl.sh | bash -s -- v0.1.0
#
set -euo pipefail

VERSION="${1:-latest}"
MIN_KERNEL="6.6"
MODULE="github.com/msune/l2radar/l2rctl/cmd/l2rctl"

# --- helpers ----------------------------------------------------------------

die() { echo "ERROR: $*" >&2; exit 1; }
info() { echo "==> $*"; }

# --- preflight checks -------------------------------------------------------

# Go
command -v go >/dev/null 2>&1 || die "Go is not installed. Install Go 1.24+ from https://go.dev/dl/"

# Kernel version >= MIN_KERNEL
kernel="$(uname -r)"
kernel_major="${kernel%%.*}"
kernel_rest="${kernel#*.}"
kernel_minor="${kernel_rest%%[.-]*}"
min_major="${MIN_KERNEL%%.*}"
min_minor="${MIN_KERNEL#*.}"

if [ "$kernel_major" -lt "$min_major" ] || \
   { [ "$kernel_major" -eq "$min_major" ] && [ "$kernel_minor" -lt "$min_minor" ]; }; then
    die "Kernel $kernel is too old. l2radar requires Linux $MIN_KERNEL+ (for TCX support)."
fi

# --- install -----------------------------------------------------------------

if [ "$VERSION" = "latest" ]; then
    info "Installing l2rctl@latest ..."
    go install "${MODULE}@latest"
else
    info "Installing l2rctl@${VERSION} ..."
    go install "${MODULE}@${VERSION}"
fi

# Locate the installed binary
GOBIN="${GOBIN:-${GOPATH:-$HOME/go}/bin}"
[ -x "${GOBIN}/l2rctl" ] || die "go install succeeded but ${GOBIN}/l2rctl not found."

# --- symlink into ~/.local/bin -----------------------------------------------

LOCAL_BIN="$HOME/.local/bin"
mkdir -p "$LOCAL_BIN"
ln -sf "${GOBIN}/l2rctl" "${LOCAL_BIN}/l2rctl"
info "Symlinked ${LOCAL_BIN}/l2rctl -> ${GOBIN}/l2rctl"

# --- PATH check --------------------------------------------------------------

case ":${PATH}:" in
    *":${LOCAL_BIN}:"*) ;;
    *)
        echo ""
        echo "WARNING: ${LOCAL_BIN} is not in your PATH."
        echo "Add it by appending this line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo "  export PATH=\"\$PATH:\$HOME/.local/bin\""
        echo ""
        ;;
esac

info "Done! Run 'l2rctl --help' to get started."
