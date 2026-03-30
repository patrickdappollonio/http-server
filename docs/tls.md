# TLS / HTTPS support

`http-server` supports serving content over HTTPS using your own TLS certificate and key.

## Quick start

```bash
http-server --tls-cert cert.pem --tls-key key.pem --hostname example.com -d ./site
```

This starts two listeners:

- **HTTPS** on port 443 serving your content
- **HTTP** on port 80 redirecting all requests to HTTPS

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--tls-cert` | *(none)* | Path to the TLS certificate file (PEM format) |
| `--tls-key` | *(none)* | Path to the TLS private key file (PEM format) |
| `--hostname` | *(none, required with TLS)* | Hostname used in HTTP-to-HTTPS redirect URLs |
| `--https-port` | `443` | Port for the HTTPS listener |
| `--http-port` | `80` | Port for the HTTP redirect listener (use `0` to disable) |

Both `--tls-cert` and `--tls-key` must be provided together. When TLS is active, the `--port` flag cannot be used. Use `--http-port` and `--https-port` instead.

## Custom ports

To use non-privileged ports (no root required):

```bash
http-server --tls-cert cert.pem --tls-key key.pem --hostname localhost \
  --https-port 8443 --http-port 8080 -d ./site
```

## Disabling the HTTP redirect

If you only want HTTPS with no HTTP listener at all:

```bash
http-server --tls-cert cert.pem --tls-key key.pem --hostname example.com \
  --http-port 0 -d ./site
```

## Certificate file hiding

If your certificate and key files are inside the directory being served, they are automatically hidden from directory listings and blocked from direct download. A warning is printed at startup recommending you move them outside the served directory.

## Certificate reload

You can reload the TLS certificate without restarting the server:

- **Unix:** send `SIGHUP` to the process: `kill -HUP $(pgrep http-server)`
- **Any platform:** send a POST request to the `/_/tls/reload` endpoint

The reload endpoint returns JSON:

```json
{"reloaded": true}
```

If the new certificate is invalid, the old certificate is preserved and an error is returned.

## Certificate metadata endpoint

When TLS is active, a `GET /_/tls` endpoint returns JSON metadata about the currently loaded certificate:

```json
{
  "tls_mode": "byo",
  "tls_cert_subject": "example.com",
  "tls_cert_sans": ["example.com", "www.example.com"],
  "tls_cert_issuer": "Let's Encrypt",
  "tls_cert_not_after": "2026-06-15T00:00:00Z",
  "tls_cert_not_before": "2026-03-15T00:00:00Z"
}
```

When TLS is not active, this endpoint is not available.

Both `/_/tls` and `/_/tls/reload` are protected by the same authentication (basic auth or JWT) as the rest of the server, if configured.

## Validation

At startup, `http-server` validates:

- Both `--tls-cert` and `--tls-key` are provided together
- The certificate and key files exist and form a valid pair
- The certificate has not expired (hard error)
- The certificate is not future-dated (hard error)
- The certificate expiry is within 30 days (warning, server still starts)
- `--port` is not explicitly set alongside TLS flags
- `--http-port` and `--https-port` differ (unless `--http-port` is `0`)
- `--hostname` is provided

## Docker usage

```yaml
services:
  http-server:
    image: ghcr.io/patrickdappollonio/docker-http-server:v2
    restart: unless-stopped
    ports:
      - "443:443"
      - "80:80"
    volumes:
      - ./site:/html:ro
      - ./certs:/certs:ro
    environment:
      - TLS_CERT=/certs/cert.pem
      - TLS_KEY=/certs/key.pem
      - HOSTNAME=example.com
```

Or with custom ports:

```yaml
    ports:
      - "8443:8443"
      - "8080:8080"
    environment:
      - TLS_CERT=/certs/cert.pem
      - TLS_KEY=/certs/key.pem
      - HOSTNAME=example.com
      - HTTPS_PORT=8443
      - HTTP_PORT=8080
```
