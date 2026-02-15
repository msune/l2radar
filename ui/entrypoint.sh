#!/bin/sh
set -e

SSL_DIR="/etc/nginx/ssl"
CERT="$SSL_DIR/cert.pem"
KEY="$SSL_DIR/key.pem"
AUTH_CONFIG="/etc/l2radar/auth.yaml"
HTPASSWD="/etc/nginx/.htpasswd"

# Generate self-signed certificate if none provided
if [ ! -f "$CERT" ] || [ ! -f "$KEY" ]; then
    echo "No TLS certificates found, generating self-signed..."
    mkdir -p "$SSL_DIR"
    openssl req -x509 -nodes -days 365 \
        -newkey rsa:2048 \
        -keyout "$KEY" \
        -out "$CERT" \
        -subj "/CN=l2radar" \
        2>/dev/null
fi

# Generate htpasswd from YAML config
if [ -f "$AUTH_CONFIG" ]; then
    echo "Generating htpasswd from $AUTH_CONFIG..."
    > "$HTPASSWD"
    # Parse YAML using yq and generate bcrypt hashes
    USERS=$(yq -r '.users[] | .username + ":" + .password' "$AUTH_CONFIG")
    echo "$USERS" | while IFS=: read -r username password; do
        htpasswd -bB "$HTPASSWD" "$username" "$password"
    done
else
    echo "WARNING: No auth config found at $AUTH_CONFIG"
    echo "Creating default credentials (admin/changeme)..."
    htpasswd -cbB "$HTPASSWD" "admin" "changeme"
fi

# Enable plain HTTP on port 80 if --enable-http is passed
for arg in "$@"; do
    if [ "$arg" = "--enable-http" ]; then
        cp /etc/nginx/http-plain.conf /etc/nginx/conf.d/http-plain.conf
        echo "Plain HTTP (port 80) enabled"
        break
    fi
done

# Create data directory if it doesn't exist
mkdir -p /tmp/l2radar

echo "Starting nginx..."
exec nginx -g "daemon off;"
