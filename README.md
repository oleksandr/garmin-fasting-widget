# Garmin Instinct 3 AMOLED Fasting Widget

An intermittent fasting companion widget built specifically for the **Garmin Instinct 3 AMOLED (50mm)** using Monkey C and the official Garmin Connect IQ SDK.

Designed for flexible schedules (e.g. 16/8), this app runs as an interactive **Glance Widget**:
* **Zero Idle Battery Drain**: Stays completely unloaded while on your standard watch face.
* **Instant Glance Access**: One press of the `DOWN` button shows fasting status & progress bar in your widget list.
* **Physical 5-Button Controls**: Optimized for button navigation (no touch needed). Quick-shift fasting windows (+/- 15 mins or +/- 1 hour) anytime.
* **Wrist Alerts**: Schedules background temporal wake-ups when fasting or eating windows finish.
* **Offline-First Rolling Storage**: Retains the last 30 completed fasts directly in persistent watch memory.
* **Local Network Sync**: Pushes history to your private LAN API endpoint via Bluetooth through Garmin Connect Mobile.

---

## Repository Structure

```
garmin-fasting-widget/
├── manifest.xml                  # App manifest & permissions
├── monkey.jungle                 # Build configuration
├── developer_key.der             # 4096-bit RSA build signing key
├── docs/                         # Detailed technical documentation
│   ├── ARCHITECTURE.md           # FSM state machine, lifecycle, storage schema
│   ├── UI_SPECIFICATION.md       # 390x390 AMOLED layout, colors, button mapping
│   ├── LOCAL_API_SPEC.md         # Local API schema, SSL & split-DNS guide
│   └── DEVELOPMENT_GUIDE.md      # Build, MTP sideload, and troubleshooting
├── resources/                    # Drawables, strings, and settings
│   ├── drawables/                # 40x40 launcher icon & drawables.xml
│   ├── settings/                 # Configurable API URL, token, and hours
│   └── strings/                  # UI and menu localization strings
├── scripts/
│   ├── build.sh                  # One-click build script
│   ├── install.sh                # 1-click Linux MTP sideload script
│   └── sample_local_server.py    # Local FastAPI test sync receiver
└── source/                       # Monkey C source code
    ├── FastingApp.mc             # Application lifecycle & background registration
    ├── FastingBackground.mc      # Temporal event wake-up delegate
    ├── FastingGlanceView.mc      # Glance loop view
    ├── FastingMainView.mc        # 390x390 circular AMOLED interface
    ├── FastingMenuDelegate.mc    # Physical button handler & action menus
    ├── FastingStorage.mc         # Persistent FIFO storage (last 30 fasts)
    └── FastingSync.mc            # Communications HTTP sync client
```

---

## Getting Started

### 1. Build the Widget
```bash
bash scripts/build.sh
```
This produces `bin/FastingWidget.prg` (~20 KB).

### 2. Sideload to Watch (1-Click USB Sideload)
Connect your Instinct 3 AMOLED with its USB cable and run:
```bash
bash scripts/install.sh
```
This automatically detects the watch over Linux MTP (`Vendor 0x091e`, `Product 0x51eb`), mounts it, copies `Fasting.prg` to `GARMIN/Apps/`, and safely unmounts it.

### 3. Activate on Watch
1. Disconnect the USB cable.
2. From the watch face, press **`DOWN`** to browse through your **Glances**.
3. Look for **Fasting**. *(If not present, scroll to the bottom -> **Edit** -> **+ Add** -> **Fasting**)*.
4. Press **`GPS / ENTER`** on Fasting to launch the full 390 × 390 AMOLED screen.

---

## Testing Local Network Sync

To test the sync functionality with your computer acting as the local server:
1. Start the sample server:
   ```bash
   python3 scripts/sample_local_server.py
   ```
2. In the Garmin Connect app on your smartphone, open **App Settings** for the Fasting widget.
3. Set **Local API Endpoint** to your server's HTTPS endpoint.
4. From the watch widget, press `GPS / ENTER` -> select **Sync Now**.

For detailed troubleshooting, architecture notes, and SSL certificates guides, see [`docs/DEVELOPMENT_GUIDE.md`](docs/DEVELOPMENT_GUIDE.md) and [`docs/LOCAL_API_SPEC.md`](docs/LOCAL_API_SPEC.md).
