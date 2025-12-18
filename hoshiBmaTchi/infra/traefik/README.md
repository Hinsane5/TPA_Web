# Traefik HTTPS Setup Guide

This guide explains how to set up HTTPS for local development using Traefik as a reverse proxy with self-signed SSL certificates.

## 🎯 Overview

- **Traefik** acts as the SSL termination point and reverse proxy
- **Backend services** run on HTTP internally
- **External access** is via HTTPS on port 443
- **Self-signed certificates** for local development

## 🚀 Quick Start

### 1. Generate SSL Certificates

First, navigate to the Traefik directory and generate self-signed certificates:

```bash
cd infra/traefik
chmod +x generate-certs.sh
./generate-certs.sh
```

This creates:

- `certs/localhost.crt` - SSL certificate
- `certs/localhost.key` - Private key

The certificates are valid for:

- `localhost`
- `api.hoshi.localhost`
- `*.hoshi.localhost`
- `127.0.0.1`

### 2. Trust the Certificate (Optional but Recommended)

To avoid browser security warnings:

**On macOS:**

```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ./certs/localhost.crt
```

**On Linux:**

```bash
sudo cp ./certs/localhost.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

**On Windows:**

1. Double-click `localhost.crt`
2. Click "Install Certificate"
3. Choose "Local Machine"
4. Select "Place all certificates in the following store"
5. Browse to "Trusted Root Certification Authorities"

### 3. Configure Your Hosts File

Add these entries to `/etc/hosts` (Linux/macOS) or `C:\Windows\System32\drivers\etc\hosts` (Windows):

```
127.0.0.1 api.hoshi.localhost
127.0.0.1 hoshi.localhost
```

### 4. Start the Services

```bash
cd ../..  # Back to project root
docker-compose up --build
```

## 🔧 Configuration Details

### Traefik Setup

The docker-compose.yaml configures Traefik with:

```yaml
services:
  traefik:
    command:
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--entrypoints.web.http.redirections.entrypoint.to=websecure"
      - "--providers.file.directory=/etc/traefik/dynamic"
    volumes:
      - ./infra/traefik/certs:/etc/traefik/certs:ro
      - ./infra/traefik/config:/etc/traefik/dynamic:ro
```

Key points:

- HTTP (port 80) automatically redirects to HTTPS (port 443)
- Certificates are mounted as read-only volumes
- Dynamic configuration from `config/tls.yml`

### Service Labels

Services are configured with Traefik labels for automatic discovery:

**API Gateway Example:**

```yaml
api-gateway:
  labels:
    - "traefik.enable=true"
    - "traefik.http.routers.api-secure.rule=Host(`api.hoshi.localhost`)"
    - "traefik.http.routers.api-secure.entrypoints=websecure"
    - "traefik.http.routers.api-secure.tls=true"
    - "traefik.http.services.api-service.loadbalancer.server.port=8080"
```

**Chat Service (WebSocket) Example:**

```yaml
chat-service:
  labels:
    - "traefik.enable=true"
    - "traefik.http.routers.chat-ws.rule=Host(`api.hoshi.localhost`) && PathPrefix(`/ws`)"
    - "traefik.http.routers.chat-ws.entrypoints=websecure"
    - "traefik.http.routers.chat-ws.tls=true"
```

## 🌐 Accessing Your Services

Once running, access your services via HTTPS:

- **Frontend**: https://hoshi.localhost
- **API Gateway**: https://api.hoshi.localhost
- **WebSocket**: wss://api.hoshi.localhost/ws
- **Traefik Dashboard**: http://localhost:8080

## 🔒 Security Notes

### For Development

- Self-signed certificates will show browser warnings on first access
- Click "Advanced" → "Proceed to localhost" in your browser
- Or trust the certificate using the commands above

### For Production

Replace self-signed certificates with proper SSL certificates from:

- **Let's Encrypt** (free, automated with Traefik)
- **Commercial CA** (e.g., DigiCert, Sectigo)

Traefik can automatically manage Let's Encrypt certificates with this configuration:

```yaml
command:
  - "--certificatesresolvers.letsencrypt.acme.email=your-email@example.com"
  - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
  - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
```

## 🛠️ Troubleshooting

### Certificate Errors

If you see certificate errors:

1. Verify certificates exist in `infra/traefik/certs/`
2. Check file permissions (readable by Docker)
3. Regenerate certificates: `./generate-certs.sh`

### Connection Refused

1. Check Traefik is running: `docker ps | grep traefik`
2. View Traefik logs: `docker logs hoshi-traefik`
3. Verify hosts file entry exists

### Service Not Found

1. Check service labels in docker-compose.yaml
2. Visit Traefik dashboard at http://localhost:8080
3. Verify service is listed in "HTTP Routers"

### WebSocket Connection Issues

For WebSocket connections, ensure:

1. Use `wss://` protocol (not `ws://`)
2. Router has correct PathPrefix in labels
3. No trailing slashes in paths

## 📚 Additional Resources

- [Traefik Documentation](https://doc.traefik.io/traefik/)
- [Docker Labels Reference](https://doc.traefik.io/traefik/routing/providers/docker/)
- [TLS Configuration](https://doc.traefik.io/traefik/https/tls/)

## 🔄 Regenerating Certificates

Certificates expire after 365 days. To regenerate:

```bash
cd infra/traefik
./generate-certs.sh
docker-compose restart traefik
```

## ✅ Verification Checklist

- [ ] Certificates generated in `infra/traefik/certs/`
- [ ] `tls.yml` exists in `infra/traefik/config/`
- [ ] Hosts file configured with `api.hoshi.localhost`
- [ ] Docker services running
- [ ] Can access https://api.hoshi.localhost
- [ ] HTTP redirects to HTTPS
- [ ] Traefik dashboard accessible at http://localhost:8080
