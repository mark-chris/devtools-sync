#!/usr/bin/env bash
set -euo pipefail

# Generate self-signed development certificates for devtools-sync
# Usage: ./scripts/generate-dev-certs.sh [output-dir]

CERT_DIR="${1:-certs}"
DAYS=365
KEY_BITS=2048

mkdir -p "$CERT_DIR"

echo "Generating development certificates in $CERT_DIR/"

# Generate CA key and certificate
openssl req -x509 -newkey "rsa:$KEY_BITS" -nodes \
    -keyout "$CERT_DIR/ca-key.pem" \
    -out "$CERT_DIR/ca-cert.pem" \
    -days "$DAYS" \
    -subj "/C=US/ST=Dev/L=Dev/O=devtools-sync/CN=devtools-sync-ca"

# Generate server key
openssl genrsa -out "$CERT_DIR/server-key.pem" "$KEY_BITS"

# Generate server CSR
openssl req -new \
    -key "$CERT_DIR/server-key.pem" \
    -out "$CERT_DIR/server.csr" \
    -subj "/C=US/ST=Dev/L=Dev/O=devtools-sync/CN=localhost"

# Create SAN config
cat > "$CERT_DIR/san.cnf" <<EOF
[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

# Sign server certificate with CA
openssl x509 -req \
    -in "$CERT_DIR/server.csr" \
    -CA "$CERT_DIR/ca-cert.pem" \
    -CAkey "$CERT_DIR/ca-key.pem" \
    -CAcreateserial \
    -out "$CERT_DIR/server-cert.pem" \
    -days "$DAYS" \
    -extensions v3_req \
    -extfile "$CERT_DIR/san.cnf"

# Clean up intermediate files
rm -f "$CERT_DIR/server.csr" "$CERT_DIR/san.cnf" "$CERT_DIR/ca-cert.srl"

echo ""
echo "Certificates generated:"
echo "  CA certificate:     $CERT_DIR/ca-cert.pem"
echo "  Server certificate: $CERT_DIR/server-cert.pem"
echo "  Server key:         $CERT_DIR/server-key.pem"
echo ""
echo "To use with devtools-sync server:"
echo "  export TLS_ENABLED=true"
echo "  export TLS_CERT_FILE=$CERT_DIR/server-cert.pem"
echo "  export TLS_KEY_FILE=$CERT_DIR/server-key.pem"
echo ""
echo "To trust the CA on your system:"
echo "  macOS:  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $CERT_DIR/ca-cert.pem"
echo "  Ubuntu: sudo cp $CERT_DIR/ca-cert.pem /usr/local/share/ca-certificates/devtools-sync-ca.crt && sudo update-ca-certificates"
