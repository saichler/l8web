# TLS Reverse Proxy

A Go-based reverse proxy that routes HTTPS traffic to different backend services based on domain names.

## Features

- Routes incoming HTTPS traffic on port 443 to different backend services
- Supports multiple domain names per backend service
- SNI-based TLS certificate selection
- Automatic certificate loading per domain

## Configuration

The proxy is configured to route traffic as follows:

| Domain | Backend Port | Certificate Path |
|--------|--------------|------------------|
| www.layer8vibe.dev, layer8vibe.dev | 1443 | layer8vibe.dev/ |
| www.probler.dev, probler.dev | 2443 | probler.dev/ |

## Prerequisites

1. **Backend Services**: Ensure your backend services are running on the configured ports (1443 and 2443)
2. **TLS Certificates**: Place your TLS certificates in the current working directory where the binary will be executed:
   - `layer8vibe.dev/domain.cert.pem` - Certificate chain for layer8vibe.dev
   - `layer8vibe.dev/private.key.pem` - Private key for layer8vibe.dev
   - `probler.dev/domain.cert.pem` - Certificate chain for probler.dev
   - `probler.dev/private.key.pem` - Private key for probler.dev

## Installation

1. Navigate to the project directory:
```bash
cd /path/to/l8web
```

2. Build the proxy:
```bash
go build -o reverse-proxy go/web/proxy/main/main.go
```

## Usage

### Running the Proxy

1. Create the certificate directories in your working directory:
```bash
mkdir -p layer8vibe.dev probler.dev
# Copy your certificates to these directories
```

2. Start the proxy (requires root/sudo for binding to port 443):
```bash
sudo ./reverse-proxy
```

3. The proxy will start listening on port 443 and route traffic based on the incoming domain name.

### Running as a Service (systemd)

Create a systemd service file `/etc/systemd/system/reverse-proxy.service`:

```ini
[Unit]
Description=TLS Reverse Proxy
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/path/where/certificates/are
ExecStart=/path/to/reverse-proxy
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Note: The `WorkingDirectory` should be the directory containing the certificate subdirectories (layer8vibe.dev/, probler.dev/)

Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable reverse-proxy
sudo systemctl start reverse-proxy
```

Check service status:
```bash
sudo systemctl status reverse-proxy
```

### Testing

Test the proxy with curl:
```bash
# Test layer8vibe.dev
curl -k https://layer8vibe.dev
curl -k https://www.layer8vibe.dev

# Test probler.dev
curl -k https://probler.dev
curl -k https://www.probler.dev
```

## Logs

The proxy logs all incoming requests and routing decisions to stdout. When running as a systemd service, logs can be viewed with:
```bash
sudo journalctl -u reverse-proxy -f
```

## Security Considerations

- The proxy uses `InsecureSkipVerify` for backend connections since they're on localhost. In production with remote backends, consider proper certificate validation.
- Ensure proper file permissions on certificate files (readable only by the proxy user)
- Consider implementing rate limiting and DDoS protection
- Add health checks for backend services

## Troubleshooting

### Port 443 Already in Use
If port 443 is already in use, stop the conflicting service:
```bash
sudo lsof -i :443
sudo systemctl stop nginx  # or apache2, or other service
```

### Certificate Errors
Ensure certificates are in PEM format and have correct permissions:
```bash
sudo chmod 600 */private.key.pem
sudo chmod 644 */domain.cert.pem
```

### Backend Connection Failed
Verify backend services are running:
```bash
netstat -tlnp | grep -E "1443|2443"
```

## Adding New Routes

To add new domains, modify the `NewReverseProxy()` function in `reverse_proxy.go`:

```go
Routes: []RouteConfig{
    // ... existing routes ...
    {
        Domains:    []string{"www.example.com", "example.com"},
        TargetPort: "3443",
        CertFile:   "example.com/domain.cert.pem",
        KeyFile:    "example.com/private.key.pem",
    },
}
```

Then rebuild and restart the proxy. Make sure to create the certificate directory in your working directory:
```bash
mkdir example.com
# Copy domain.cert.pem and private.key.pem to example.com/
```