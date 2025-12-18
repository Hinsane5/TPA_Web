#!/bin/bash

# Generate self-signed certificates for local development
# This script creates SSL certificates for localhost and api.hoshi.localhost

CERTS_DIR="./certs"
mkdir -p "$CERTS_DIR"

echo "Generating self-signed SSL certificate for local development..."

# Generate private key
openssl genrsa -out "$CERTS_DIR/localhost.key" 4096

# Generate certificate signing request (CSR)
openssl req -new -key "$CERTS_DIR/localhost.key" \
  -out "$CERTS_DIR/localhost.csr" \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Create a config file for Subject Alternative Names (SAN)
cat > "$CERTS_DIR/san.cnf" << EOF
[req]
default_bits = 4096
prompt = no
default_md = sha256
distinguished_name = dn
req_extensions = v3_req

[dn]
C=US
ST=State
L=City
O=Organization
CN=localhost

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = api.hoshi.localhost
DNS.3 = hoshi.localhost
DNS.4 = *.hoshi.localhost
IP.1 = 127.0.0.1
EOF

# Generate self-signed certificate with SAN
openssl x509 -req -days 365 \
  -in "$CERTS_DIR/localhost.csr" \
  -signkey "$CERTS_DIR/localhost.key" \
  -out "$CERTS_DIR/localhost.crt" \
  -extensions v3_req \
  -extfile "$CERTS_DIR/san.cnf"

# Clean up CSR and config file
rm "$CERTS_DIR/localhost.csr" "$CERTS_DIR/san.cnf"

echo "✅ SSL certificates generated successfully in $CERTS_DIR"
echo ""
echo "Certificate: $CERTS_DIR/localhost.crt"
echo "Private Key: $CERTS_DIR/localhost.key"
echo ""
echo "⚠️  Note: This is a self-signed certificate for local development only."
echo "Your browser will show a security warning. You can safely proceed."
echo ""
echo "To trust this certificate on macOS, run:"
echo "sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $CERTS_DIR/localhost.crt"
