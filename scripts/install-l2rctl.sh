#!/bin/sh
set -e

REPO="msune/l2radar"
INSTALL_DIR="$HOME/.local/bin"
BINARY="l2rctl"

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)
    echo "Error: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

URL="https://github.com/${REPO}/releases/latest/download/l2rctl-linux-${ARCH}"

echo "Downloading l2rctl for linux/${ARCH}..."
mkdir -p "$INSTALL_DIR"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "${INSTALL_DIR}/${BINARY}" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "${INSTALL_DIR}/${BINARY}" "$URL"
else
  echo "Error: curl or wget is required" >&2
  exit 1
fi

chmod +x "${INSTALL_DIR}/${BINARY}"
echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"

# Check if INSTALL_DIR is in PATH
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "Note: ${INSTALL_DIR} is not in your PATH."
    echo "Add it with:  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac
