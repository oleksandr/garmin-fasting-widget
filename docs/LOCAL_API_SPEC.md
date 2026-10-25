# Local Network API Specification & Integration Guide

## 1. Network Topology & Garmin Transport

Because the Garmin watch communicates via Bluetooth Low Energy (BLE) through the **Garmin Connect Mobile** smartphone app:

1. The request originates on the watch.
2. The phone executes the HTTP POST over Wi-Fi when connected to your home LAN.
3. The phone sends the payload to your local server.

```
[ Instinct 3 Watch ] ──(BLE)──► [ Smartphone ] ──(Home Wi-Fi)──► [ Local API Server ]
```

---

## 2. API Endpoint Specification

### Main Ingestion Endpoint: `POST /api` (also aliased as `/api/fasting/sync`)
* **Method**: `POST`
* **Path**: `/api` (or `/api/fasting/sync`)
* **Headers**:
  * `Content-Type: application/json`
  * `Authorization: Bearer <API_SECRET_TOKEN>` (or `X-Api-Key: <TOKEN>`, or query parameter `?token=<TOKEN>`)

#### Format A: Garmin Watch Batch Sync
```json
{
  "device": "Instinct 3 AMOLED 50mm",
  "synced_at": 1727434800,
  "fasts": [
    {
      "id": 1727434800,
      "start": 1727434800,
      "end": 1727492400,
      "target_hours": 16,
      "duration_hours": 16.0,
      "state": "COMPLETED"
    }
  ]
}
```

#### Format B: Smartphone App / iOS Shortcut / Automation Single Record
```json
{
  "start": 1727434800,
  "end": 1727492400,
  "duration_hours": 16.0,
  "target_hours": 16,
  "state": "COMPLETED",
  "device": "iPhone"
}
```

#### Format C: Raw JSON Array
```json
[
  {
    "id": 1727434800,
    "start": 1727434800,
    "end": 1727492400,
    "duration_hours": 16.0,
    "target_hours": 16,
    "state": "COMPLETED"
  }
]
```

### Dashboard Endpoint: `GET /`
Returns the AMOLED-dark, responsive HTML dashboard showing current streak, total fasts, 7-day visual overview chart, active session banner, and rolling fast history table. Zero external JS/CSS dependencies.

### API Summary: `GET /api`
Returns JSON metrics, active session, recent records, and endpoint documentation.


---

## 3. Resolving the Garmin / iOS / Android SSL Requirement

**Important**: Garmin Connect Mobile strictly rejects unencrypted `http://` or self-signed certificates.

### Recommended Setup: Split-Horizon DNS + Let's Encrypt
1. **Domain**: Pick any domain or subdomain you control (e.g. `fasting.lan.yourdomain.com`).
2. **SSL Cert**: Issue a valid Let's Encrypt wildcard or subdomain certificate using ACME DNS-01 validation (zero open inbound router ports required).
3. **Local DNS**: Configure your local DNS (Pi-hole, AdGuard Home, or router) to resolve `fasting.lan.yourdomain.com` directly to your local server IP (e.g. `192.168.1.100`).
4. **Local Proxy**: Terminate SSL on Nginx, Caddy, or Traefik with the valid certificate.

---

## 4. Self-Hosted Go Sync Server & SQLite Storage

A lightweight, self-hosted sync server is implemented in Go under the `server/` directory:
- **Zero Third-Party Web Frameworks**: Pure standard library `net/http` with method-aware routing.
- **Modern JSON**: Uses standard library `encoding/json/v2` for encoding and decoding.
- **Embedded Database**: Pure-Go SQLite (`modernc.org/sqlite`) requiring no CGO.

### Endpoints
* `POST /api/fasting/sync` / `POST /api/v1/sync`: Upserts fasting sessions into SQLite.
* `GET /api/fasting/history` / `GET /api/v1/history`: Returns up to `limit` records (default: 30).
* `GET /api/fasting/stats` / `GET /api/v1/stats`: Returns aggregated duration and count metrics.
* `GET /healthz`: Service health check.

### Running Locally
```bash
cd server
go test -v ./...
go run . -port 8080 -db /path/to/fasting.db -token "your-secret-token"
```

### Running with Docker
```bash
# Build image (~9MB)
docker build -t fasting-server:latest .

# Run with persistent SQLite volume
docker run -d \
  --name fasting-server \
  -p 8080:8080 \
  -v fasting_data:/data \
  -e FASTING_PORT=8080 \
  -e FASTING_DB_PATH=/data/fasting.db \
  -e FASTING_API_TOKEN="your-secret-token" \
  fasting-server:latest
```
