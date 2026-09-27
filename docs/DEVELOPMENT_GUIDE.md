# Development & Deployment Guide

This guide explains how to compile, test, sideload, and debug the **Garmin Fasting Widget** for the **Garmin Instinct 3 AMOLED (50mm)**.

---

## 1. Prerequisites & Environment Setup

1. **Java Runtime**: OpenJDK 17+ installed (`java -version`).
2. **Connect IQ SDK**: Connect IQ SDK 9.2.0+ located in `~/.Garmin/ConnectIQ/Sdks/`.
3. **Developer Key**: A 4096-bit RSA key exists at `developer_key.der` (already generated in this repository).
4. **Device Definition (`instinct3amoled50mm`)**:
   - Garmin requires an authenticated SSO login to download device profiles.
   - A patched CLI binary is provided at `~/Work/bin/connect-iq-sdk-manager` (with full 2FA/MFA support enabled).
   - Login and download:
     ```bash
     connect-iq-sdk-manager login
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
├── docs/                         # Architecture & technical documentation
│   ├── ARCHITECTURE.md           # FSM state machine, lifecycle, storage schema
│   ├── UI_SPECIFICATION.md       # 390x390 AMOLED layout, colors, button mapping
│   ├── LOCAL_API_SPEC.md         # Local API schema, SSL & split-DNS guide
│   └── DEVELOPMENT_GUIDE.md      # This guide
├── resources/                    # Layouts, strings, and settings
│   ├── drawables/
│   │   ├── drawables.xml
│   │   └── launcher_icon.png
│   ├── menus/
│   │   └── menu.xml
│   ├── settings/
│   │   ├── properties.xml
│   │   └── settings.xml
│   └── strings/
│       └── strings.xml
├── scripts/                      # Build & deployment utilities
│   ├── build.sh                  # Automated compiler script
│   ├── install.sh                # 1-click Linux MTP sideload script
│   └── sample_local_server.py    # Local test sync API server
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

### Automated Build
Run the build script:
```bash
bash scripts/build.sh
```
This produces `bin/FastingWidget.prg` (~20 KB).

### Manual Build Command
If building from terminal directly using `monkeyc`:
```bash
~/.Garmin/ConnectIQ/Sdks/connectiq-sdk-lin-9.2.0-2026-06-09-92a1605b2/bin/monkeyc \
  -o bin/FastingWidget.prg \
  -f monkey.jungle \
  -y developer_key.der \
  -d instinct3amoled50mm \
  -r
```

---

## 4. Sideloading to the Instinct 3 AMOLED (Linux MTP)

Modern Garmin watches (including the Instinct 3 AMOLED) connect using the **MTP (Media Transfer Protocol)** rather than legacy USB Mass Storage.

* **USB IDs**: Vendor `0x091e` (Garmin), Product `0x51eb` (Instinct 3).
* **Storage Path**: `Internal Storage/GARMIN/Apps/Fasting.prg`.

### Method A: Automated 1-Click Sideload (Recommended)
1. Plug the watch into your computer using the USB cable.
2. Run:
   ```bash
   bash scripts/install.sh
   ```
   *The script automatically detects the MTP device, mounts it via GIO/GVFS, copies the binary, and safely unmounts it.*

### Method B: Manual Sideload via GIO CLI
```bash
# 1. Mount the watch
gio mount mtp://091e_51eb_0000d779f07a/

# 2. Copy binary
gio copy -f bin/FastingWidget.prg "mtp://091e_51eb_0000d779f07a/Internal Storage/GARMIN/Apps/Fasting.prg"

# 3. Safely flush and unmount
gio mount -u mtp://091e_51eb_0000d779f07a/
```

---

## 5. On-Watch Activation & Usage

1. **Unplug the watch** from USB.
2. **Find the Widget in Glances**:
   * From the watch face, press **`DOWN`** to scroll through your Glance list.
   * Look for **Fasting**.
   * *If not present by default*: Scroll to the very bottom of the glance list, select **Edit** -> **+ Add**, and choose **Fasting**.
3. **Open the Widget**:
   * Highlight **Fasting** and press **`GPS / ENTER`**.
4. **Physical Button Controls**:
   * **`GPS / ENTER`**: Open action menu (*Start Fast*, *End Fast & Eat*, *+1 Hour*, *-1 Hour*, *Sync Now*, *Reset*).
   * **`UP`**: Quick-shift +15 min (started 15m earlier) with haptic vibration.
   * **`DOWN`**: Quick-shift -15 min (started 15m later) with haptic vibration.
   * **`SET / BACK`**: Exit back to the watch face.

---

## 6. Troubleshooting & Platform Gotchas

### Issue 1: Garmin SSO Error `427` During Login
* **Cause**: Garmin’s Single Sign-On returns status code `427` when an account has Two-Factor Authentication (2FA/MFA) enabled, but the client passes `mfaRequired=false` in the OAuth query string.
* **Fix**: The tool at `~/Work/bin/connect-iq-sdk-manager` was patched to enforce `mfaRequired=true&performMFACheck=true`, allowing the browser to present the 2FA SMS/Authenticator code screen properly.

### Issue 2: `Invalid permission provided: Toybox.Application.Storage`
* **Cause**: In Connect IQ, `Toybox.Application.Storage` is a built-in core API available to all apps without permission. Declaring it in `<iq:permissions>` causes a manifest compiler error.
* **Fix**: Keep only necessary permissions in `manifest.xml` using bare names:
  ```xml
  <iq:permissions>
      <iq:uses-permission id="Background"/>
      <iq:uses-permission id="Communications"/>
  </iq:permissions>
  ```

### Issue 3: `No enum constant ... Setting.Type.STRING`
* **Cause**: In `resources/settings/settings.xml`, string inputs must specify `type="alphaNumeric"`, not `type="string"`.
* **Fix**: Use `<settingConfig type="alphaNumeric" maxLength="128" />`.

### Issue 4: Monkey C Type Checking on Dictionaries
* **Cause**: Reading values from `Dictionary` returns `Toybox.Lang.Object?`. Operations like `== false` or passing to `Array<Number>` methods trigger compiler errors in strict mode.
* **Fix**: Explicitly cast dictionary lookups:
  ```monkeyc
  var isSynced = item.get("synced") as Boolean?;
  var id = item.get("id") as Number?;
  ```
