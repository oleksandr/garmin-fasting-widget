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
- **Private LAN Sync**: Automatically pushes completed fasts to a self-hosted Go HTTP API on your home network.

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

### Go Sync Server
- **Go**: 1.27+
- **Docker** (optional, for containerized deployment)

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

## Go Sync Server & SQLite Storage

A standalone HTTP server implemented in pure Go (standard library `net/http` and `encoding/json/v2`) with SQLite storage.

### Endpoints
- `POST /api/fasting/sync` (or `/api/v1/sync`): Receive batch fasting records from the watch.
- `GET /api/fasting/history` (or `/api/v1/history`): Query recent fasts (`?limit=30`).
- `GET /api/fasting/stats` (or `/api/v1/stats`): Summary metrics (total fasts, hours, averages).
- `GET /healthz`: Service health check.

### Run Locally
```bash
cd server
go test -v ./...
go run . -port 8080 -db fasting.db -token "your-secret-token"
```

### Run with Docker
```bash
# Build image (minimal ~9MB Alpine container)
docker build -t fasting-server:latest .

# Run container with persistent volume
docker run -d \
  --name fasting-server \
  -p 8080:8080 \
  -v fasting_data:/data \
  -e FASTING_PORT=8080 \
  -e FASTING_DB_PATH=/data/fasting.db \
  -e FASTING_API_TOKEN="your-secret-token" \
  fasting-server:latest
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
