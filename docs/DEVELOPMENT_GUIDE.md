# Development & Deployment Guide

This guide explains how to compile, test, and sideload the **Garmin Fasting Widget** to your **Garmin Instinct 3 AMOLED (50mm)**.

---

## 1. Prerequisites

1. **Java Runtime**: OpenJDK 17+ installed (`java -version`).
2. **Connect IQ SDK**: Connect IQ SDK 9.x+ is installed in `~/.Garmin/ConnectIQ/Sdks/`.
3. **Developer Key**: A 4096-bit RSA key exists at `developer_key.der` (already generated in this repo).
4. **Device Definition (`instinct3amoled50mm`)**:
   - Because downloading device definitions requires your Garmin Developer Account SSO login, run the following command once in your terminal:
     ```bash
     connect-iq-sdk-manager login
     ```
   - Then download your device profile:
     ```bash
     connect-iq-sdk-manager device download -d instinct3amoled50mm
     ```

---

## 2. Project Directory Structure

```
garmin-fasting-widget/
├── bin/                          # Output .prg binaries
├── developer_key.der             # RSA build signing key (DO NOT COMMIT)
├── developer_key.pem             # Private PEM key
├── manifest.xml                  # Connect IQ app manifest & permissions
├── monkey.jungle                 # Project build config
├── docs/                         # Architecture & specifications
│   ├── ARCHITECTURE.md
│   ├── UI_SPECIFICATION.md
│   ├── LOCAL_API_SPEC.md
│   └── DEVELOPMENT_GUIDE.md
├── resources/                    # Layouts, strings, and settings
│   ├── drawables/
│   │   └── drawables.xml
│   ├── menus/
│   │   └── menu.xml
│   ├── settings/
│   │   ├── properties.xml
│   │   └── settings.xml
│   └── strings/
│       └── strings.xml
├── scripts/                      # Build & test utilities
│   ├── build.sh
│   └── sample_local_server.py
└── source/                       # Monkey C source code
    ├── FastingApp.mc             # Main Application entrypoint
    ├── FastingBackground.mc      # Temporal background wake service
    ├── FastingGlanceView.mc      # 390x390 Glance loop view
    ├── FastingMainView.mc        # Main AMOLED circular UI
    ├── FastingMenuDelegate.mc    # 5-button interaction delegate
    ├── FastingStorage.mc         # Persistent FIFO storage (last 30 fasts)
    └── FastingSync.mc            # Communications HTTP sync client
```

---

## 3. Building the Application

Run the automated build script:
```bash
bash scripts/build.sh
```

Or invoke the compiler manually:
```bash
~/.Garmin/ConnectIQ/Sdks/connectiq-sdk-lin-9.2.0-2026-06-09-92a1605b2/bin/monkeyc \
  -o bin/FastingWidget.prg \
  -f monkey.jungle \
  -y developer_key.der \
  -d instinct3amoled50mm \
  -r
```

---

## 4. Sideloading to your Instinct 3 AMOLED (USB)

No need to publish to the Garmin Connect IQ Store:
1. Connect your Instinct 3 AMOLED to your computer using the Garmin USB charging cable.
2. The watch will mount as a mass storage device (or via MTP on Linux).
3. Copy `bin/FastingWidget.prg` into the **`GARMIN/APPS/`** directory on your watch.
4. Safely eject/unmount the watch and unplug the cable.
5. The **Fasting** widget will now appear in your widget/glance list!
