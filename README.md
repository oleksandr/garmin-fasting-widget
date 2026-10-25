# Garmin Fasting Widget

An intermittent fasting tracker widget built in Monkey C for the **Garmin Instinct 3 AMOLED (50mm)**, paired with a lightweight, self-hosted Go API sync server.

<p align="center">
  <img src="docs/assets/preview.png" width="320" alt="Garmin Fasting Widget UI Preview">
</p>

## Highlights

- **AMOLED-Optimized (416 × 416)**: Pure black background (`#000000`) with zero idle battery impact.
- **Glance View**: Shows active state, elapsed time, and progress bar in the widget loop without opening the app.
- **Physical Button Controls**: 100% button-driven navigation designed for the Instinct hardware (no touch needed).
- **On-Wrist Time Adjustments**: Quick-shift windows by ±15 minutes using `UP` / `DOWN`, or fine-tune via the menu.
- **Offline Storage**: Keeps a rolling on-device history of the last 30 completed fasts.
- **Homelab Sync**: Automatically syncs fasts over HTTPS to `https://fasting.home.lobunets.me/api` (or custom LAN host).
- **Web Dashboard**: Self-hosted AMOLED dark dashboard (`/`) displaying streaks, weekly overview, and session history.
- **Flexible Ingestion**: Ingests fasting data from Garmin watches or smartphone apps/shortcuts via `POST /api`.

---

## Controls

| Button | Action |
| :--- | :--- |
| **GPS / START** | Open Action Menu (Start/End Fast, Shift Time, Sync Now, History, Reset) |
| **UP** | Shift Fasting Window (+15 min) |
| **DOWN** | Shift Fasting Window (-15 min) |
| **BACK / SET** | Return to Watch Face / Glance Loop |

---

## Requirements

### Watch Widget
- **Garmin Connect IQ SDK**: 9.2.0+
- **Device Profile**: `instinct3amoled50mm`
- **Java**: OpenJDK 17+
- **Developer Key**: 4096-bit RSA signing key (`developer_key.der`)

### Go Companion Server
- **Go**: 1.27+
- **Docker / Podman** (for containerized deployment on homelab or cloud)

---

## Build & Sideload

### 1. Build PRG Binary
```bash
./scripts/build.sh
```
Compiles `bin/FastingWidget.prg` using Connect IQ SDK `monkeyc`.

### 2. Sideload to Watch (Linux MTP)
Connect your Instinct 3 AMOLED via USB and run:
```bash
./scripts/install.sh
```
The script detects the watch via MTP (`091e:51eb`), copies `Fasting.prg` to `GARMIN/Apps/`, and unmounts cleanly.

### 3. Interactive Web Preview
To test UI layouts locally without deploying to the physical device:
```bash
./scripts/preview.sh
```

---

## Go Sync Server & Dashboard

A standalone server implemented in pure Go (standard library `net/http`, `html/template`, and `encoding/json/v2`) with SQLite storage. Zero external web frameworks.

### Endpoints
- `GET /`: AMOLED dark web dashboard with streaks, 7-day visual overview, and session table.
- `POST /api`: Main data ingestion endpoint (accepts Garmin batch sync, raw JSON arrays, or single-record smartphone shortcut payloads).
- `GET /api`: JSON API summary with live stats, uptime, and documentation.
- `GET /api/fasting/history`: Query recent fasts (`?limit=30`).
- `GET /api/fasting/stats`: Summary metrics (streaks, total fasts, total hours, averages).
- `GET /healthz`: Service health check.

### Run Locally
```bash
cd server
go test -v ./...
go run . -port 8080 -db fasting.db -token "optional-secret-token"
```

### Run with Docker / Podman
```bash
# Build image (~9MB Alpine container)
docker build -t fasting-server:latest .

# Run container with persistent volume
docker run -d \
  --name fasting-server \
  -p 18090:8080 \
  -v fasting_data:/data \
  -e FASTING_PORT=8080 \
  -e FASTING_DB_PATH=/data/fasting.db \
  -e FASTING_API_TOKEN="optional-secret-token" \
  fasting-server:latest
```

### Homelab Deployment (ZimaBoard / Caddy)
Exposed via Caddy reverse proxy at `https://fasting.home.lobunets.me`:
```caddy
fasting.home.lobunets.me {
    import tls_letsencrypt
    reverse_proxy 127.0.0.1:18090
}
```


---

## Documentation

- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md): FSM lifecycle, storage schema, and background alerts.
- [`docs/UI_SPECIFICATION.md`](docs/UI_SPECIFICATION.md): Typography, color palette, and 416 × 416 circular layout.
- [`docs/LOCAL_API_SPEC.md`](docs/LOCAL_API_SPEC.md): Home LAN API specification, Go server, and SSL setup.
- [`docs/DEVELOPMENT_GUIDE.md`](docs/DEVELOPMENT_GUIDE.md): Toolchain setup, device profile installation, and debugging.

---

## License

[MIT](LICENSE) © 2026 Oleksandr
