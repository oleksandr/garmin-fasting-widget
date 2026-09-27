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

# 1. Check if Garmin MTP device is connected
DEVICE_ID=$(gio mount -l -i | grep -B 2 -i "091e 51eb" | grep "activation_root" | head -n 1 | awk -F'=' '{print $2}')

if [ -z "$DEVICE_ID" ]; then
    # Try looking for any Garmin volume
    DEVICE_ID=$(gio mount -l -i | grep -B 2 -i "091e" | grep "activation_root" | head -n 1 | awk -F'=' '{print $2}')
fi

if [ -z "$DEVICE_ID" ]; then
    echo "Error: Garmin Instinct 3 AMOLED not detected over USB."
    echo "Please connect the watch with its USB cable and ensure it's unlocked."
    exit 1
fi

echo "Detected Garmin device: $DEVICE_ID"

# 2. Mount MTP volume if needed
gio mount "$DEVICE_ID" 2>/dev/null || true

TARGET_PATH="${DEVICE_ID}Internal Storage/GARMIN/Apps/Fasting.prg"

echo "Copying Fasting.prg to watch..."
gio copy -f "$PRG" "$TARGET_PATH"

echo "Flushing and safely unmounting..."
gio mount -u "$DEVICE_ID" 2>/dev/null || true

echo ""
echo "SUCCESS! Fasting widget has been installed onto your watch."
echo "You can now unplug the cable."
