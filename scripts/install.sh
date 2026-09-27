#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PRG="$PROJECT_DIR/bin/FastingWidget.prg"

if [ ! -f "$PRG" ]; then
    echo "Binary not found. Building first..."
    bash "$SCRIPT_DIR/build.sh"
fi

echo "=== Sideloading Fasting Widget to Garmin Watch ==="

# 1. Detect Garmin MTP URI from gio mount list
DEVICE_ID=$(gio mount -l -i | grep -o 'mtp://[a-zA-Z0-9_]*' | head -n 1)

if [ -z "$DEVICE_ID" ]; then
    echo "Error: Garmin Instinct 3 AMOLED not detected over USB."
    echo "Please connect the watch with its USB cable and ensure it's unlocked."
    exit 1
fi

DEVICE_URI="${DEVICE_ID}/"
echo "Detected Garmin device: $DEVICE_URI"

# 2. Mount MTP volume
gio mount "$DEVICE_URI" 2>/dev/null || true

# 3. Locate Apps directory
TARGET_PATH="${DEVICE_URI}Internal Storage/GARMIN/Apps/Fasting.prg"

echo "Copying Fasting.prg to watch..."
gio copy "$PRG" "$TARGET_PATH"

echo "Flushing and safely unmounting..."
gio mount -u "$DEVICE_URI" 2>/dev/null || true

echo ""
echo "SUCCESS! Fasting widget has been installed onto your watch."
echo "You can now unplug the cable."
